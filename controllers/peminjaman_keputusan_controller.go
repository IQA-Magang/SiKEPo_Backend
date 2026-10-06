package controllers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/models"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KeputusanPeminjamanRequest struct {
	// Keputusan berisi "setuju" atau "tolak".
	Keputusan string `json:"keputusan"`

	// Catatan adalah arahan saat setuju, dan alasan yang wajib diisi saat tolak.
	Catatan string `json:"catatan"`

	// AlatAlternatifID opsional, hanya dipakai saat Pengelola menolak.
	AlatAlternatifID *uint64 `json:"alat_alternatif_id"`
}

// tahapPersetujuan menggambarkan tahap yang sedang menunggu keputusan.
type tahapPersetujuan struct {
	// prefix adalah awalan nama kolom, misalnya "pengelola" untuk
	// pengelola_status, pengelola_catatan, dan pengelola_at.
	prefix       string
	label        string
	kataSetuju   string
	approverID   uint64
	statusLanjut string
	aksiSetuju   string
	aksiTolak    string
}

// tahapSaatIni menentukan tahap dari status peminjaman. Karena tahap diambil
// dari status, tidak ada cara melompati tahap.
func tahapSaatIni(p *models.Peminjaman) (*tahapPersetujuan, error) {
	switch p.Status {

	case models.StatusPeminjamanMenungguManajerPeminjam:
		return &tahapPersetujuan{
			prefix:       "manajer_peminjam",
			label:        "Manajer Peminjam",
			kataSetuju:   "Divalidasi",
			approverID:   p.ManajerPeminjamID,
			statusLanjut: models.StatusPeminjamanMenungguPengelola,
			aksiSetuju:   models.LogbookAksiDivalidasiManajerPinjam,
			aksiTolak:    models.LogbookAksiDitolakManajerPinjam,
		}, nil

	case models.StatusPeminjamanMenungguPengelola:
		return &tahapPersetujuan{
			prefix:       "pengelola",
			label:        "Pengelola Peralatan",
			kataSetuju:   "Diperiksa dan diteruskan ke Manager Lab",
			approverID:   p.PengelolaID,
			statusLanjut: models.StatusPeminjamanMenungguManagerLab,
			aksiSetuju:   models.LogbookAksiDiteruskanPengelola,
			aksiTolak:    models.LogbookAksiDitolakPengelola,
		}, nil

	case models.StatusPeminjamanMenungguManagerLab:
		return &tahapPersetujuan{
			prefix:       "manager_lab",
			label:        "Manager Lab",
			kataSetuju:   "Disetujui",
			approverID:   p.ManagerLabID,
			statusLanjut: models.StatusPeminjamanDisetujui,
			aksiSetuju:   models.LogbookAksiDisetujuiManagerLab,
			aksiTolak:    models.LogbookAksiDitolakManagerLab,
		}, nil
	}

	return nil, errPeminjaman(fiber.StatusConflict,
		"Peminjaman tidak sedang menunggu keputusan, status saat ini: %s", p.Status)
}

// =====================================================
// KEPUTUSAN
// PUT /api/peminjaman/:id/keputusan
// =====================================================

// Keputusan dipakai ketiga penyetuju secara bergantian. Tahap yang berjalan
// ditentukan dari status peminjaman, dan hanya penyetuju yang ditetapkan untuk
// tahap itu yang boleh memutuskan.
func (c *PeminjamanController) Keputusan(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	var req KeputusanPeminjamanRequest
	if err := ctx.BodyParser(&req); err != nil {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Format JSON tidak valid"), "")
	}

	keputusan := strings.ToLower(strings.TrimSpace(req.Keputusan))
	if keputusan != "setuju" && keputusan != "tolak" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "keputusan harus setuju atau tolak"), "")
	}

	setuju := keputusan == "setuju"
	catatan := strings.TrimSpace(req.Catatan)

	if !setuju && catatan == "" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Alasan penolakan wajib diisi"), "")
	}

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		var p models.Peminjaman
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).
			First(&p).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPeminjaman(fiber.StatusNotFound, "Peminjaman tidak ditemukan")
			}
			return err
		}

		var alat models.Peralatan
		if err := tx.
			Where("id = ? AND deleted_at IS NULL", p.PeralatanID).
			First(&alat).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPeminjaman(fiber.StatusNotFound, "Peralatan tidak ditemukan")
			}
			return err
		}

		tahap, err := tahapSaatIni(&p)
		if err != nil {
			return err
		}

		if tahap.approverID != user.UserID {
			return errPeminjaman(fiber.StatusForbidden,
				"Anda bukan %s yang ditetapkan untuk pengajuan ini", tahap.label)
		}

		// Kondisi alat bisa berubah sejak pengajuan dibuat, misalnya masuk
		// karantina atau melewati jatuh tempo, jadi dicek ulang sebelum setuju.
		if setuju {
			if err := cekKelayakanPeminjaman(tx, &alat, p.RencanaTanggalKembali); err != nil {
				return err
			}
		}

		// Usulan alat lain, hanya saat Pengelola menolak.
		var alternatif *models.Peralatan
		if !setuju && tahap.prefix == "pengelola" && req.AlatAlternatifID != nil && *req.AlatAlternatifID != 0 {
			if *req.AlatAlternatifID == p.PeralatanID {
				return errPeminjaman(fiber.StatusBadRequest,
					"Alat alternatif tidak boleh sama dengan alat yang diajukan")
			}

			var usulan models.Peralatan
			if err := tx.
				Where("id = ? AND deleted_at IS NULL", *req.AlatAlternatifID).
				First(&usulan).
				Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errPeminjaman(fiber.StatusBadRequest, "Alat alternatif tidak ditemukan")
				}
				return err
			}

			alternatif = &usulan
		}

		statusBaru := models.StatusPeminjamanDitolak
		statusTahap := models.TahapRejected

		if setuju {
			statusBaru = tahap.statusLanjut
			statusTahap = models.TahapApproved
		}

		updates := map[string]interface{}{
			"status":                  statusBaru,
			tahap.prefix + "_status":  statusTahap,
			tahap.prefix + "_catatan": catatan,
			tahap.prefix + "_at":      time.Now(),
		}

		if alternatif != nil {
			updates["alat_alternatif_id"] = alternatif.ID
		}

		if err := tx.Model(&models.Peminjaman{}).
			Where("id = ?", p.ID).
			Updates(updates).
			Error; err != nil {
			return err
		}

		p.Status = statusBaru

		aksi := tahap.aksiSetuju
		keterangan := fmt.Sprintf("%s oleh %s (%s).", tahap.kataSetuju, tahap.label, user.Name)

		if !setuju {
			aksi = tahap.aksiTolak
			keterangan = fmt.Sprintf("Ditolak oleh %s (%s). Alasan: %s.", tahap.label, user.Name, catatan)

			if alternatif != nil {
				keterangan += fmt.Sprintf(" Usulan alat lain: %s (%s).", alternatif.NamaPeralatan, alternatif.NomorAset)
			}
		} else if catatan != "" {
			keterangan += fmt.Sprintf(" Catatan: %s.", catatan)
		}

		if err := c.catatLogbook(tx, &p, aksi, user.UserID, keterangan); err != nil {
			return err
		}

		return c.kirimNotifikasiKeputusan(tx, &p, &alat, tahap, setuju, catatan, alternatif)
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal menyimpan keputusan")
	}

	result, err := c.Repository.FindByID(id)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Keputusan tersimpan tetapi gagal mengambil data")
	}

	message := "Keputusan tersimpan"
	switch result.Status {
	case models.StatusPeminjamanDitolak:
		message = "Pengajuan ditolak"
	case models.StatusPeminjamanDisetujui:
		message = "Peminjaman disetujui"
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    c.buildResponse(result),
	})
}

// kirimNotifikasiKeputusan memberi tahu pihak berikutnya. Saat ditolak,
// peminjam yang diberi tahu. Saat setuju, notifikasi menuju tahap berikutnya,
// atau ke peminjam dan Pengelola bila sudah disetujui penuh.
func (c *PeminjamanController) kirimNotifikasiKeputusan(
	tx *gorm.DB,
	p *models.Peminjaman,
	alat *models.Peralatan,
	tahap *tahapPersetujuan,
	setuju bool,
	catatan string,
	alternatif *models.Peralatan,
) error {

	nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

	if !setuju {
		pesan := fmt.Sprintf("Pengajuan %s untuk %s ditolak oleh %s. Alasan: %s.",
			p.Kode, nama, tahap.label, catatan)

		if alternatif != nil {
			pesan += fmt.Sprintf(" Usulan alat lain: %s (%s).", alternatif.NamaPeralatan, alternatif.NomorAset)
		}

		return utils.SendNotification(tx, p.PeminjamID,
			"peminjaman_ditolak", "Pengajuan peminjaman ditolak", pesan)
	}

	switch tahap.prefix {

	case "manajer_peminjam":
		return utils.SendNotification(tx, p.PengelolaID,
			"peminjaman_divalidasi", "Pengajuan peminjaman menunggu pemeriksaan",
			fmt.Sprintf("Pengajuan %s untuk %s sudah divalidasi Manajer Peminjam dan menunggu pemeriksaan Anda.", p.Kode, nama))

	case "pengelola":
		return utils.SendNotification(tx, p.ManagerLabID,
			"peminjaman_diteruskan", "Pengajuan peminjaman menunggu persetujuan",
			fmt.Sprintf("Pengajuan %s untuk %s sudah diperiksa Pengelola dan menunggu persetujuan Anda.", p.Kode, nama))
	}

	pesan := fmt.Sprintf("Pengajuan %s untuk %s disetujui Manager Lab.", p.Kode, nama)

	if err := utils.SendNotification(tx, p.PeminjamID,
		"peminjaman_disetujui", "Peminjaman disetujui", pesan); err != nil {
		return err
	}

	return utils.SendNotification(tx, p.PengelolaID,
		"peminjaman_disetujui", "Peminjaman disetujui, siapkan peralatan",
		pesan+" Silakan siapkan peralatan untuk serah terima.")
}
