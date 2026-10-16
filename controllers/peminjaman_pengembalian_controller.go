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
)

// Pengembalian berjalan satu langkah, karena Pengelola dan peminjam bertemu
// langsung saat alat diserahkan kembali:
//
//  1. Pengelola mengisi kolom Kembali pada lembar F/010 dengan butir yang sama
//     seperti saat keluar, mencatat kejadian yang disampaikan peminjam, lalu
//     keduanya menandatangani.
//  2. Sistem menentukan tindak lanjut. Bila ada kejadian, ada butir TS, atau
//     alat kembali setelah jatuh tempo pemeriksaan, alat menjadi Do Not Use.
//     Bila tidak, alat masuk Karantina dan menunggu verifikasi ulang (IK/003).
//
// Verifikasi ulangnya dikerjakan Pengelola di halaman verifikasi yang sudah ada.

type PengembalianRequest struct {
	// LampiranA berisi tujuh butir kolom Kembali, berurutan seperti saat keluar.
	LampiranA []SerahTerimaItemRequest `json:"lampiran_a"`

	// AdaKejadian bernilai true bila peminjam menyampaikan kejadian selama
	// peminjaman, misalnya terjatuh, basah, atau rusak.
	AdaKejadian    bool   `json:"ada_kejadian"`
	UraianKejadian string `json:"uraian_kejadian"`

	Catatan string `json:"catatan"`

	// Tanda tangan keduanya wajib, data URL PNG.
	TTDPengelola string `json:"ttd_pengelola"`
	TTDPeminjam  string `json:"ttd_peminjam"`
}

// =====================================================
// PENGEMBALIAN
// POST /api/peminjaman/:id/pengembalian
// =====================================================

func (c *PeminjamanController) Pengembalian(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	var req PengembalianRequest
	if err := ctx.BodyParser(&req); err != nil {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Format JSON tidak valid"), "")
	}

	req.UraianKejadian = strings.TrimSpace(req.UraianKejadian)
	req.Catatan = strings.TrimSpace(req.Catatan)
	req.TTDPengelola = strings.TrimSpace(req.TTDPengelola)
	req.TTDPeminjam = strings.TrimSpace(req.TTDPeminjam)

	if err := validasiTTD(req.TTDPengelola, "Pengelola"); err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	if err := validasiTTD(req.TTDPeminjam, "peminjam"); err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	if req.AdaKejadian && req.UraianKejadian == "" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Uraian kejadian wajib diisi bila ada kejadian"), "")
	}

	if !req.AdaKejadian {
		req.UraianKejadian = ""
	}

	items, butirTS, err := validasiLembarSerahTerima(req.LampiranA)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		p, alat, err := ambilPeminjamanDanAlat(tx, id)
		if err != nil {
			return err
		}

		if p.PengelolaID != user.UserID {
			return errPeminjaman(fiber.StatusForbidden,
				"Hanya Pengelola Peralatan yang ditetapkan yang dapat menerima pengembalian")
		}

		if p.Status != models.StatusPeminjamanSedangDipinjam {
			return errPeminjaman(fiber.StatusConflict,
				"Pengembalian hanya bisa dicatat pada peminjaman yang sedang berjalan, status saat ini: %s", p.Status)
		}

		var serah models.SerahTerima
		if err := tx.
			Preload("Items").
			Where("peminjaman_id = ?", p.ID).
			First(&serah).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPeminjaman(fiber.StatusConflict, "Lembar serah terima keluar tidak ditemukan")
			}
			return err
		}

		if len(serah.Items) != len(items) {
			return errPeminjaman(fiber.StatusConflict,
				"Lembar serah terima keluar memakai format lama, hubungi admin")
		}

		sekarang := time.Now()

		// ---------- tindak lanjut ----------

		alasan := []string{}

		if req.AdaKejadian {
			alasan = append(alasan, "terjadi kejadian selama peminjaman: "+req.UraianKejadian)
		}

		if len(butirTS) > 0 {
			alasan = append(alasan, "kondisi tidak sesuai saat kembali ("+strings.Join(butirTS, "; ")+")")
		}

		jatuhTempo, label, _, err := detailKelayakan(tx, alat)
		if err != nil {
			return err
		}

		if jatuhTempo != nil && tanggalSaja(sekarang).After(tanggalSaja(*jatuhTempo)) {
			alasan = append(alasan, fmt.Sprintf("alat kembali setelah melewati jatuh tempo %s (%s)",
				label, formatTanggalID(*jatuhTempo)))
		}

		tindakLanjut := models.TindakLanjutPerluVerifikasi
		if len(alasan) > 0 {
			tindakLanjut = models.TindakLanjutDoNotUse
		}

		alasanText := strings.Join(alasan, "; ")

		// ---------- simpan kolom Kembali pada lembar ----------

		hasilPerNomor := map[int]models.SerahTerimaItem{}
		for _, item := range items {
			hasilPerNomor[item.NomorButir] = item
		}

		for _, lama := range serah.Items {
			baru, ada := hasilPerNomor[lama.NomorButir]
			if !ada {
				return errPeminjaman(fiber.StatusConflict,
					"Butir %d pada lembar keluar tidak ditemukan pada data kembali", lama.NomorButir)
			}

			if err := tx.Model(&models.SerahTerimaItem{}).
				Where("id = ?", lama.ID).
				Updates(map[string]interface{}{
					"hasil_kembali":      baru.Hasil,
					"keterangan_kembali": baru.Keterangan,
				}).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&models.SerahTerima{}).
			Where("id = ?", serah.ID).
			Updates(map[string]interface{}{
				"kembali_diterima_oleh": user.UserID,
				"kembali_at":            sekarang,
				"kembali_pengelola_ttd": req.TTDPengelola,
				"kembali_peminjam_ttd":  req.TTDPeminjam,
				"ada_kejadian":          req.AdaKejadian,
				"uraian_kejadian":       req.UraianKejadian,
				"kembali_ada_ts":        len(butirTS) > 0,
				"kembali_catatan":       req.Catatan,
				"tindak_lanjut":         tindakLanjut,
				"alasan_tindak_lanjut":  alasanText,
			}).Error; err != nil {
			return err
		}

		// ---------- tutup peminjaman dan ubah status alat ----------

		if err := tx.Model(&models.Peminjaman{}).
			Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"status":             models.StatusPeminjamanDikembalikan,
				"tgl_kembali_aktual": sekarang,
			}).Error; err != nil {
			return err
		}

		p.Status = models.StatusPeminjamanDikembalikan

		// Perpanjangan yang masih menunggu tidak relevan lagi karena alat sudah kembali.
		if err := tx.Model(&models.Perpanjangan{}).
			Where("peminjaman_id = ? AND status = ?", p.ID, models.PerpanjanganMenunggu).
			Updates(map[string]interface{}{
				"status":            models.PerpanjanganDibatalkan,
				"catatan_keputusan": "Alat sudah dikembalikan sebelum perpanjangan diputuskan",
			}).Error; err != nil {
			return err
		}

		perubahanAlat := map[string]interface{}{"status_alat": "Do Not Use"}
		if tindakLanjut == models.TindakLanjutPerluVerifikasi {
			perubahanAlat = map[string]interface{}{
				"status_alat":       "Karantina",
				"status_verifikasi": "Belum Diverifikasi",
			}
		}

		if err := tx.Model(&models.Peralatan{}).
			Where("id = ?", alat.ID).
			Updates(perubahanAlat).Error; err != nil {
			return err
		}

		// ---------- logbook ----------

		var peminjam models.User
		if err := tx.Where("user_id = ?", p.PeminjamID).First(&peminjam).Error; err != nil {
			return err
		}

		ket := fmt.Sprintf("Peralatan dikembalikan oleh %s dan diterima Pengelola Peralatan (%s).",
			peminjam.Name, user.Name)

		if tanggalSaja(sekarang).After(tanggalSaja(p.RencanaTanggalKembali)) {
			hari := int(tanggalSaja(sekarang).Sub(tanggalSaja(p.RencanaTanggalKembali)).Hours() / 24)
			ket += fmt.Sprintf(" Terlambat %d hari dari rencana kembali %s.", hari, formatTanggalID(p.RencanaTanggalKembali))
		}

		if req.AdaKejadian {
			ket += " Kejadian: " + req.UraianKejadian + "."
		}

		if err := c.catatLogbook(tx, p, models.LogbookAksiPengembalian, user.UserID, ket); err != nil {
			return err
		}

		nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

		if tindakLanjut == models.TindakLanjutDoNotUse {
			if err := c.catatLogbook(tx, p, models.LogbookAksiDihentikanPengembalian, user.UserID,
				"Peralatan ditandai Do Not Use karena "+alasanText+". Perlu ditangani sesuai TLKM13/IK/012."); err != nil {
				return err
			}

			if err := utils.SendNotification(tx, p.ManagerLabID,
				"peralatan_do_not_use", "Peralatan ditandai Do Not Use",
				fmt.Sprintf("%s ditandai Do Not Use saat pengembalian pengajuan %s karena %s. Peralatan perlu ditangani sesuai TLKM13/IK/012.",
					nama, p.Kode, alasanText)); err != nil {
				return err
			}

			return utils.SendNotification(tx, p.PeminjamID,
				"peralatan_dikembalikan", "Pengembalian peralatan dicatat",
				fmt.Sprintf("Pengembalian %s untuk pengajuan %s sudah dicatat. Peralatan dihentikan penggunaannya karena %s.",
					nama, p.Kode, alasanText))
		}

		if err := c.catatLogbook(tx, p, models.LogbookAksiPerluVerifikasi, user.UserID,
			"Peralatan masuk Karantina dan perlu diverifikasi ulang sesuai TLKM13/IK/003 sebelum dipakai kembali."); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PeminjamID,
			"peralatan_dikembalikan", "Pengembalian peralatan dicatat",
			fmt.Sprintf("Pengembalian %s untuk pengajuan %s sudah dicatat. Peralatan akan diverifikasi ulang oleh Pengelola.",
				nama, p.Kode))
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal menyimpan pengembalian")
	}

	result, err := c.Repository.FindByID(id)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Pengembalian tersimpan tetapi gagal mengambil data terbaru")
	}

	message := "Pengembalian dicatat. Peralatan masuk Karantina dan perlu diverifikasi ulang"
	if result.SerahTerima != nil && result.SerahTerima.TindakLanjut == models.TindakLanjutDoNotUse {
		message = "Pengembalian dicatat. Peralatan ditandai Do Not Use"
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    c.buildResponseDetail(result),
	})
}
