package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/models"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Perpanjangan mengikuti flowchart "Penggunaan dan Perpanjangan Peralatan":
//
//  1. Peminjam mengajukan tanggal kembali baru beserta alasan. Sistem memberi
//     tahu Pengelola dan Manager Lab, lalu mencatat logbook.
//  2. Manager Lab pemilik alat memutuskan. Bila setuju, tanggal kembali
//     berubah, peminjam dan Pengelola diberi tahu. Bila tidak, Manager Lab
//     mengisi alasan dan hanya peminjam yang diberi tahu.
//
// Pengelola tidak ikut memutuskan, sesuai TLKM13/IK/005 yang menyebut Mgr. Lab
// sebagai penyetuju perpanjangan.

type PerpanjanganRequest struct {
	// TanggalKembaliBaru berformat YYYY-MM-DD.
	TanggalKembaliBaru string `json:"tanggal_kembali_baru"`
	Alasan             string `json:"alasan"`
}

type KeputusanPerpanjanganRequest struct {
	// Keputusan berisi "setuju" atau "tolak".
	Keputusan string `json:"keputusan"`

	// Catatan adalah alasan yang wajib diisi saat tolak, dan boleh kosong saat setuju.
	Catatan string `json:"catatan"`

	// TTD wajib saat setuju, berupa data URL PNG. Saat tolak boleh kosong.
	TTD string `json:"ttd"`
}

// =====================================================
// PENGAJUAN
// POST /api/peminjaman/:id/perpanjangan
// =====================================================

func (c *PeminjamanController) AjukanPerpanjangan(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	var req PerpanjanganRequest
	if err := ctx.BodyParser(&req); err != nil {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Format JSON tidak valid"), "")
	}

	req.Alasan = strings.TrimSpace(req.Alasan)
	req.TanggalKembaliBaru = strings.TrimSpace(req.TanggalKembaliBaru)

	if req.TanggalKembaliBaru == "" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "tanggal_kembali_baru wajib diisi"), "")
	}

	tanggalBaru, err := time.ParseInLocation("2006-01-02", req.TanggalKembaliBaru, time.Local)
	if err != nil {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "tanggal_kembali_baru tidak valid, gunakan format YYYY-MM-DD"), "")
	}

	if req.Alasan == "" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Alasan perpanjangan wajib diisi"), "")
	}

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		p, alat, err := ambilPeminjamanDanAlat(tx, id)
		if err != nil {
			return err
		}

		if p.PeminjamID != user.UserID {
			return errPeminjaman(fiber.StatusForbidden,
				"Hanya peminjam yang dapat mengajukan perpanjangan")
		}

		if p.Status != models.StatusPeminjamanSedangDipinjam {
			return errPeminjaman(fiber.StatusConflict,
				"Perpanjangan hanya bisa diajukan pada peminjaman yang sedang berjalan, status saat ini: %s", p.Status)
		}

		// Diajukan paling lambat pada tanggal kembali yang direncanakan.
		if tanggalSaja(time.Now()).After(tanggalSaja(p.RencanaTanggalKembali)) {
			return errPeminjaman(fiber.StatusConflict,
				"Perpanjangan harus diajukan paling lambat pada tanggal kembali yang direncanakan (%s)",
				formatTanggalID(p.RencanaTanggalKembali))
		}

		var menunggu int64
		if err := tx.Model(&models.Perpanjangan{}).
			Where("peminjaman_id = ? AND status = ?", p.ID, models.PerpanjanganMenunggu).
			Count(&menunggu).Error; err != nil {
			return err
		}

		if menunggu > 0 {
			return errPeminjaman(fiber.StatusConflict,
				"Masih ada pengajuan perpanjangan yang menunggu keputusan Manager Lab")
		}

		if !tanggalSaja(tanggalBaru).After(tanggalSaja(p.RencanaTanggalKembali)) {
			return errPeminjaman(fiber.StatusBadRequest,
				"Tanggal kembali baru harus setelah tanggal kembali saat ini (%s)",
				formatTanggalID(p.RencanaTanggalKembali))
		}

		// Tidak boleh melewati jatuh tempo pemeriksaan, dan alat harus masih layak.
		if err := cekKelayakanPeminjaman(tx, alat, tanggalBaru); err != nil {
			return err
		}

		// Tidak boleh bentrok dengan pengajuan lain yang sudah antre sesudah
		// tanggal kembali saat ini.
		bentrok, err := c.Repository.HasJadwalBentrok(tx, p.PeralatanID,
			tanggalSaja(p.RencanaTanggalKembali).AddDate(0, 0, 1), tanggalBaru, p.ID)
		if err != nil {
			return err
		}

		if bentrok {
			return errPeminjaman(fiber.StatusConflict,
				"Tanggal baru bentrok dengan pengajuan lain yang sudah antre untuk peralatan ini")
		}

		perpanjangan := models.Perpanjangan{
			PeminjamanID:   p.ID,
			DiajukanOleh:   user.UserID,
			TglKembaliLama: p.RencanaTanggalKembali,
			TglKembaliBaru: tanggalBaru,
			Alasan:         req.Alasan,
			Status:         models.PerpanjanganMenunggu,
		}

		if err := tx.Create(&perpanjangan).Error; err != nil {
			return err
		}

		if err := c.catatLogbook(tx, p, models.LogbookAksiPerpanjanganDiajukan, user.UserID,
			fmt.Sprintf("Perpanjangan diajukan oleh %s: tanggal kembali dari %s menjadi %s. Alasan: %s.",
				user.Name, formatTanggalID(p.RencanaTanggalKembali), formatTanggalID(tanggalBaru), req.Alasan),
		); err != nil {
			return err
		}

		nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

		if err := utils.SendNotification(tx, p.ManagerLabID,
			"perpanjangan_diajukan", "Pengajuan perpanjangan menunggu keputusan Anda",
			fmt.Sprintf("%s mengajukan perpanjangan %s (pengajuan %s) dari %s menjadi %s. Alasan: %s.",
				user.Name, nama, p.Kode, formatTanggalID(p.RencanaTanggalKembali), formatTanggalID(tanggalBaru), req.Alasan)); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PengelolaID,
			"perpanjangan_diajukan", "Peminjam mengajukan perpanjangan",
			fmt.Sprintf("%s mengajukan perpanjangan %s (pengajuan %s) menjadi %s. Menunggu keputusan Manager Lab.",
				user.Name, nama, p.Kode, formatTanggalID(tanggalBaru)))
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengajukan perpanjangan")
	}

	return c.responsePerpanjangan(ctx, id, "Perpanjangan diajukan. Menunggu keputusan Manager Lab")
}

// =====================================================
// KEPUTUSAN MANAGER LAB
// PUT /api/peminjaman/:id/perpanjangan/:perpanjangan_id/keputusan
// =====================================================

func (c *PeminjamanController) KeputusanPerpanjangan(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	perpanjanganID, err := strconv.ParseUint(ctx.Params("perpanjangan_id"), 10, 64)
	if err != nil || perpanjanganID == 0 {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "ID perpanjangan tidak valid"), "")
	}

	var req KeputusanPerpanjanganRequest
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
	ttd := strings.TrimSpace(req.TTD)

	if !setuju && catatan == "" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Alasan penolakan wajib diisi"), "")
	}

	// Tanda tangan wajib saat setuju. Saat tolak boleh kosong, tetapi bila
	// diisi harus valid.
	if setuju || ttd != "" {
		if err := validasiTTD(ttd, "Manager Lab"); err != nil {
			return respondPeminjamanError(ctx, err, "")
		}
	}

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		p, alat, err := ambilPeminjamanDanAlat(tx, id)
		if err != nil {
			return err
		}

		if p.ManagerLabID != user.UserID {
			return errPeminjaman(fiber.StatusForbidden,
				"Hanya Manager Lab pemilik peralatan yang dapat memutuskan perpanjangan")
		}

		var perpanjangan models.Perpanjangan
		if err := tx.
			Where("id = ? AND peminjaman_id = ?", perpanjanganID, p.ID).
			First(&perpanjangan).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPeminjaman(fiber.StatusNotFound, "Pengajuan perpanjangan tidak ditemukan")
			}
			return err
		}

		if perpanjangan.Status != models.PerpanjanganMenunggu {
			return errPeminjaman(fiber.StatusConflict,
				"Pengajuan perpanjangan sudah diproses, status saat ini: %s", perpanjangan.Status)
		}

		if p.Status != models.StatusPeminjamanSedangDipinjam {
			return errPeminjaman(fiber.StatusConflict,
				"Peminjaman tidak lagi berjalan, status saat ini: %s", p.Status)
		}

		sekarang := time.Now()
		nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

		perubahan := map[string]interface{}{
			"diputuskan_oleh":   user.UserID,
			"diputuskan_at":     sekarang,
			"catatan_keputusan": catatan,
		}

		if ttd != "" {
			perubahan["ttd"] = ttd
		}

		if !setuju {
			perubahan["status"] = models.PerpanjanganDitolak

			if err := tx.Model(&models.Perpanjangan{}).
				Where("id = ?", perpanjangan.ID).
				Updates(perubahan).Error; err != nil {
				return err
			}

			if err := c.catatLogbook(tx, p, models.LogbookAksiPerpanjanganDitolak, user.UserID,
				fmt.Sprintf("Perpanjangan ditolak oleh Manager Lab (%s). Tanggal kembali tetap %s. Alasan: %s.",
					user.Name, formatTanggalID(perpanjangan.TglKembaliLama), catatan),
			); err != nil {
				return err
			}

			return utils.SendNotification(tx, p.PeminjamID,
				"perpanjangan_ditolak", "Perpanjangan ditolak",
				fmt.Sprintf("Perpanjangan %s (pengajuan %s) ditolak Manager Lab. Alasan: %s. Alat tetap harus dikembalikan pada %s.",
					nama, p.Kode, catatan, formatTanggalID(perpanjangan.TglKembaliLama)))
		}

		// Kelayakan dan jadwal dicek ulang, karena bisa berubah sejak diajukan.
		if err := cekKelayakanPeminjaman(tx, alat, perpanjangan.TglKembaliBaru); err != nil {
			return err
		}

		bentrok, err := c.Repository.HasJadwalBentrok(tx, p.PeralatanID,
			tanggalSaja(p.RencanaTanggalKembali).AddDate(0, 0, 1), perpanjangan.TglKembaliBaru, p.ID)
		if err != nil {
			return err
		}

		if bentrok {
			return errPeminjaman(fiber.StatusConflict,
				"Tanggal baru bentrok dengan pengajuan lain yang sudah antre untuk peralatan ini")
		}

		perubahan["status"] = models.PerpanjanganDisetujui

		if err := tx.Model(&models.Perpanjangan{}).
			Where("id = ?", perpanjangan.ID).
			Updates(perubahan).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Peminjaman{}).
			Where("id = ?", p.ID).
			Update("rencana_tanggal_kembali", perpanjangan.TglKembaliBaru).Error; err != nil {
			return err
		}

		ket := fmt.Sprintf("Perpanjangan disetujui oleh Manager Lab (%s). Tanggal kembali dari %s menjadi %s.",
			user.Name, formatTanggalID(perpanjangan.TglKembaliLama), formatTanggalID(perpanjangan.TglKembaliBaru))
		if catatan != "" {
			ket += " Catatan: " + catatan + "."
		}

		if err := c.catatLogbook(tx, p, models.LogbookAksiPerpanjanganDisetujui, user.UserID, ket); err != nil {
			return err
		}

		pesan := fmt.Sprintf("Perpanjangan %s (pengajuan %s) disetujui Manager Lab. Tanggal kembali menjadi %s.",
			nama, p.Kode, formatTanggalID(perpanjangan.TglKembaliBaru))

		if err := utils.SendNotification(tx, p.PeminjamID,
			"perpanjangan_disetujui", "Perpanjangan disetujui", pesan); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PengelolaID,
			"perpanjangan_disetujui", "Perpanjangan disetujui, tanggal kembali berubah", pesan)
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal menyimpan keputusan perpanjangan")
	}

	message := "Perpanjangan ditolak"
	if setuju {
		message = "Perpanjangan disetujui. Tanggal kembali diperbarui"
	}

	return c.responsePerpanjangan(ctx, id, message)
}

// responsePerpanjangan mengirim data pengajuan terbaru beserta riwayat perpanjangan.
func (c *PeminjamanController) responsePerpanjangan(ctx *fiber.Ctx, id uint64, message string) error {
	result, err := c.Repository.FindByID(id)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Data tersimpan tetapi gagal mengambil data terbaru")
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    c.buildResponseDetail(result),
	})
}
