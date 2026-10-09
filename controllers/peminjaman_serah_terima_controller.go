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

// Serah terima keluar berjalan dua langkah:
//
//  1. Pengelola mengisi checklist Lampiran A dan menandatangani.
//     Satu butir TS membatalkan peminjaman dan menandai alat Do Not Use.
//     Bila tidak ada TS, status menjadi SIAP_DISERAHKAN dan peminjam diberi tahu.
//  2. Peminjam mengonfirmasi terima di lembar yang sama.
//     Status menjadi SEDANG_DIPINJAM dan status alat menjadi Dipinjam.

// maksUkuranTTD membatasi gambar tanda tangan yang dikirim, sekitar 5 MB.
const maksUkuranTTD = 5 * 1024 * 1024

type SerahTerimaItemRequest struct {
	Status     string `json:"status"`
	Keterangan string `json:"keterangan"`
}

type SerahTerimaRequest struct {
	// LampiranA berisi sepuluh butir, berurutan dari butir 1 sampai 10.
	LampiranA []SerahTerimaItemRequest `json:"lampiran_a"`
	Catatan   string                   `json:"catatan"`

	// TTD opsional, berupa teks gambar seperti yang dikirim kanvas frontend.
	TTD string `json:"ttd"`
}

type KonfirmasiTerimaRequest struct {
	TTD string `json:"ttd"`
}

// validasiLampiranA memeriksa sepuluh butir. Butir ke-i pada array adalah
// butir nomor i+1. Mengembalikan butir siap simpan dan daftar butir TS.
func validasiLampiranA(items []SerahTerimaItemRequest) ([]models.SerahTerimaItem, []string, error) {
	daftar := models.DaftarButirLampiranA

	if len(items) != len(daftar) {
		return nil, nil, errPeminjaman(fiber.StatusBadRequest,
			"lampiran_a harus berisi %d butir", len(daftar))
	}

	hasil := make([]models.SerahTerimaItem, 0, len(items))
	butirTS := []string{}

	for i, butir := range daftar {
		nilai := strings.ToUpper(strings.TrimSpace(items[i].Status))
		keterangan := strings.TrimSpace(items[i].Keterangan)

		switch nilai {
		case "":
			return nil, nil, errPeminjaman(fiber.StatusBadRequest,
				"Butir %d belum diisi", butir.Nomor)

		case models.HasilSesuai:

		case models.HasilTidakBerlaku:
			if !butir.BolehTB {
				return nil, nil, errPeminjaman(fiber.StatusBadRequest,
					"Butir %d tidak boleh bernilai TB", butir.Nomor)
			}

		case models.HasilTidakSesuai:
			if keterangan == "" {
				return nil, nil, errPeminjaman(fiber.StatusBadRequest,
					"Butir %d bernilai TS wajib diberi keterangan", butir.Nomor)
			}
			butirTS = append(butirTS, fmt.Sprintf("butir %d, %s", butir.Nomor, keterangan))

		default:
			return nil, nil, errPeminjaman(fiber.StatusBadRequest,
				"Hasil butir %d harus S, TS, atau TB", butir.Nomor)
		}

		hasil = append(hasil, models.SerahTerimaItem{
			NomorButir: butir.Nomor,
			Butir:      butir.Butir,
			Hasil:      nilai,
			Keterangan: keterangan,
		})
	}

	return hasil, butirTS, nil
}

// ambilPeminjamanDanAlat mengunci pengajuan dan memuat alatnya di dalam transaksi.
func ambilPeminjamanDanAlat(tx *gorm.DB, id uint64) (*models.Peminjaman, *models.Peralatan, error) {
	var p models.Peminjaman
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&p).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errPeminjaman(fiber.StatusNotFound, "Peminjaman tidak ditemukan")
		}
		return nil, nil, err
	}

	var alat models.Peralatan
	if err := tx.
		Where("id = ? AND deleted_at IS NULL", p.PeralatanID).
		First(&alat).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errPeminjaman(fiber.StatusNotFound, "Peralatan tidak ditemukan")
		}
		return nil, nil, err
	}

	return &p, &alat, nil
}

// =====================================================
// LANGKAH 1: PENGELOLA MENGISI CHECKLIST
// POST /api/peminjaman/:id/serah-terima
// =====================================================

func (c *PeminjamanController) SerahTerima(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	var req SerahTerimaRequest
	if err := ctx.BodyParser(&req); err != nil {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Format JSON tidak valid"), "")
	}

	req.Catatan = strings.TrimSpace(req.Catatan)
	req.TTD = strings.TrimSpace(req.TTD)

	if len(req.TTD) > maksUkuranTTD {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Ukuran tanda tangan terlalu besar"), "")
	}

	items, butirTS, err := validasiLampiranA(req.LampiranA)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	adaTS := len(butirTS) > 0

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		p, alat, err := ambilPeminjamanDanAlat(tx, id)
		if err != nil {
			return err
		}

		if p.PengelolaID != user.UserID {
			return errPeminjaman(fiber.StatusForbidden,
				"Hanya Pengelola Peralatan yang ditetapkan yang dapat mengisi checklist serah terima")
		}

		if p.Status != models.StatusPeminjamanDisetujui {
			return errPeminjaman(fiber.StatusConflict,
				"Serah terima hanya bisa dimulai pada pengajuan yang sudah disetujui, status saat ini: %s", p.Status)
		}

		if alat.StatusAlat != "Aktif" {
			return errPeminjaman(fiber.StatusConflict,
				"Peralatan berstatus %s sehingga belum dapat diserahkan", alat.StatusAlat)
		}

		serah := models.SerahTerima{
			PeminjamanID:  p.ID,
			DiperiksaOleh: user.UserID,
			DiperiksaAt:   time.Now(),
			PengelolaTTD:  req.TTD,
			AdaTS:         adaTS,
			Catatan:       req.Catatan,
			Items:         items,
		}

		// GORM ikut menyimpan butir karena relasinya terisi.
		if err := tx.Create(&serah).Error; err != nil {
			return err
		}

		nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

		if adaTS {
			return c.batalkanKarenaTS(tx, p, alat, user, butirTS, nama)
		}

		if err := tx.Model(&models.Peminjaman{}).
			Where("id = ?", p.ID).
			Update("status", models.StatusPeminjamanSiapDiserahkan).Error; err != nil {
			return err
		}

		p.Status = models.StatusPeminjamanSiapDiserahkan

		if err := c.catatLogbook(tx, p, models.LogbookAksiChecklistKeluar, user.UserID,
			fmt.Sprintf("Pemeriksaan serah terima keluar oleh Pengelola Peralatan (%s). Seluruh butir sesuai. Menunggu konfirmasi terima dari peminjam.",
				user.Name),
		); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PeminjamID,
			"serah_terima_menunggu_ttd",
			"Peralatan siap diserahkan",
			fmt.Sprintf("Pengelola sudah memeriksa %s untuk pengajuan %s. Silakan periksa checklist dan konfirmasi penerimaan.",
				nama, p.Kode),
		)
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal menyimpan checklist serah terima")
	}

	return c.responseSerahTerima(ctx, id, adaTS)
}

// batalkanKarenaTS menjalankan cabang butir TS: peminjaman dibatalkan, alat
// ditandai Do Not Use, dan Manager Lab serta peminjam diberi tahu.
func (c *PeminjamanController) batalkanKarenaTS(
	tx *gorm.DB,
	p *models.Peminjaman,
	alat *models.Peralatan,
	user *models.User,
	butirTS []string,
	nama string,
) error {

	alasan := "Pemeriksaan serah terima keluar menemukan ketidaksesuaian: " + strings.Join(butirTS, "; ")

	if err := tx.Model(&models.Peminjaman{}).
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"status":            models.StatusPeminjamanDibatalkanTS,
			"alasan_pembatalan": alasan,
		}).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Peralatan{}).
		Where("id = ?", alat.ID).
		Update("status_alat", "Do Not Use").Error; err != nil {
		return err
	}

	p.Status = models.StatusPeminjamanDibatalkanTS

	if err := c.catatLogbook(tx, p, models.LogbookAksiChecklistKeluar, user.UserID,
		fmt.Sprintf("Pemeriksaan serah terima keluar oleh Pengelola Peralatan (%s). Ditemukan butir TS.", user.Name),
	); err != nil {
		return err
	}

	if err := c.catatLogbook(tx, p, models.LogbookAksiDibatalkanChecklist, user.UserID,
		alasan+". Peralatan ditandai Do Not Use dan perlu ditangani sesuai TLKM13/IK/012.",
	); err != nil {
		return err
	}

	if err := utils.SendNotification(tx, p.ManagerLabID,
		"peralatan_do_not_use",
		"Peralatan ditandai Do Not Use",
		fmt.Sprintf("%s ditandai Do Not Use saat serah terima pengajuan %s. %s. Peralatan perlu ditangani sesuai TLKM13/IK/012.",
			nama, p.Kode, alasan),
	); err != nil {
		return err
	}

	return utils.SendNotification(tx, p.PeminjamID,
		"peminjaman_dibatalkan",
		"Peminjaman dibatalkan",
		fmt.Sprintf("Pengajuan %s untuk %s dibatalkan karena peralatan tidak lolos pemeriksaan saat serah terima.",
			p.Kode, nama),
	)
}

// =====================================================
// LANGKAH 2: PEMINJAM MENGONFIRMASI TERIMA
// PUT /api/peminjaman/:id/konfirmasi-terima
// =====================================================

func (c *PeminjamanController) KonfirmasiTerima(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	// Body boleh kosong karena tanda tangan opsional.
	var req KonfirmasiTerimaRequest
	if len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&req); err != nil {
			return respondPeminjamanError(ctx,
				errPeminjaman(fiber.StatusBadRequest, "Format JSON tidak valid"), "")
		}
	}

	req.TTD = strings.TrimSpace(req.TTD)

	if len(req.TTD) > maksUkuranTTD {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Ukuran tanda tangan terlalu besar"), "")
	}

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		p, alat, err := ambilPeminjamanDanAlat(tx, id)
		if err != nil {
			return err
		}

		if p.PeminjamID != user.UserID {
			return errPeminjaman(fiber.StatusForbidden,
				"Hanya peminjam yang dapat mengonfirmasi penerimaan peralatan")
		}

		if p.Status != models.StatusPeminjamanSiapDiserahkan {
			return errPeminjaman(fiber.StatusConflict,
				"Peminjaman belum siap dikonfirmasi, status saat ini: %s", p.Status)
		}

		if alat.StatusAlat != "Aktif" {
			return errPeminjaman(fiber.StatusConflict,
				"Peralatan berstatus %s sehingga belum dapat diserahkan", alat.StatusAlat)
		}

		var serah models.SerahTerima
		if err := tx.Where("peminjaman_id = ?", p.ID).First(&serah).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPeminjaman(fiber.StatusConflict, "Checklist serah terima belum diisi Pengelola")
			}
			return err
		}

		sekarang := time.Now()

		if err := tx.Model(&models.SerahTerima{}).
			Where("id = ?", serah.ID).
			Updates(map[string]interface{}{
				"diterima_oleh": user.UserID,
				"diterima_at":   sekarang,
				"peminjam_ttd":  req.TTD,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Peminjaman{}).
			Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"status":            models.StatusPeminjamanSedangDipinjam,
				"tgl_keluar_aktual": sekarang,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Peralatan{}).
			Where("id = ?", alat.ID).
			Update("status_alat", "Dipinjam").Error; err != nil {
			return err
		}

		p.Status = models.StatusPeminjamanSedangDipinjam

		if err := c.catatLogbook(tx, p, models.LogbookAksiSerahTerimaKeluar, user.UserID,
			fmt.Sprintf("Peralatan diterima oleh %s. Lokasi penggunaan %s, rencana kembali %s.",
				user.Name, p.LokasiPenggunaan, formatTanggalID(p.RencanaTanggalKembali)),
		); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PengelolaID,
			"peralatan_diserahkan",
			"Peralatan telah diterima peminjam",
			fmt.Sprintf("%s mengonfirmasi penerimaan %s (%s) untuk pengajuan %s. Rencana kembali %s.",
				user.Name, alat.NamaPeralatan, alat.NomorAset, p.Kode, formatTanggalID(p.RencanaTanggalKembali)),
		)
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengonfirmasi penerimaan")
	}

	return c.responseSerahTerima(ctx, id, false)
}

// responseSerahTerima mengirim data pengajuan terbaru beserta gambar tanda tangan.
func (c *PeminjamanController) responseSerahTerima(ctx *fiber.Ctx, id uint64, dibatalkan bool) error {
	result, err := c.Repository.FindByID(id)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Data tersimpan tetapi gagal mengambil data terbaru")
	}

	message := "Checklist tersimpan. Menunggu konfirmasi terima dari peminjam"

	switch {
	case dibatalkan:
		message = "Checklist berisi butir TS. Peminjaman dibatalkan dan peralatan ditandai Do Not Use"
	case result.Status == models.StatusPeminjamanSedangDipinjam:
		message = "Penerimaan dikonfirmasi. Peralatan berstatus Dipinjam"
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    c.buildResponseDetail(result),
	})
}
