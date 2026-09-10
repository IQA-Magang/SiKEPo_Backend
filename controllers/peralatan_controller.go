package controllers

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"backend/models"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
)

type PeralatanController struct {
	Repository        *repositories.PeralatanRepository
	RuanganRepository *repositories.RuanganRepository
	UserRepository    *repositories.UserRepository
}

type PeralatanRequest struct {
	RuanganID         uint64     `json:"ruangan_id"`
	PICID             *uint64    `json:"pic_id"`
	NomorAset         string     `json:"nomor_aset"`
	NamaPeralatan     string     `json:"nama_peralatan"`
	Merk              string     `json:"merk"`
	Model             string     `json:"model"`
	NomorSeri         string     `json:"nomor_seri"`
	Jumlah            *uint      `json:"jumlah"`
	KategoriPeralatan string     `json:"kategori_peralatan"`
	Kondisi           string     `json:"kondisi"`
	StatusKelayakan   string     `json:"status_kelayakan"`
	Metode            string     `json:"metode"`
	JenisPakai        string     `json:"jenis_pakai"`
	VerifiedAt        *time.Time `json:"verified_at"`
	VerificationNote  string     `json:"verification_note"`
	VerifiedBy        *uint64    `json:"verified_by"`
}

func valid(value string, values ...string) bool {
	for _, item := range values {
		if value == item {
			return true
		}
	}
	return false
}

func (c *PeralatanController) Create(ctx *fiber.Ctx) error {
	var req PeralatanRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validasi gagal",
			"errors":  formatBindingError(err),
		})
	}

	if req.NomorAset == "" || req.NamaPeralatan == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validasi gagal",
			"errors": fiber.Map{
				"nomor_aset":     "Nomor aset wajib diisi",
				"nama_peralatan": "Nama peralatan wajib diisi",
			},
		})
	}

	// Validasi Enum termasuk kategori_peralatan
	if !valid(req.KategoriPeralatan, "peralatan", "Peralatan bantu", "referensi uji", "golden sample", "Komponen pendukung") ||
		!valid(req.Kondisi, "", "sesuai", "tidak_sesuai", "tidak_berlaku") ||
		!valid(req.StatusKelayakan, "", "pending", "aktif", "ditolak", "tidak_aktif") ||
		!valid(req.Metode, "", "internal", "eksternal") ||
		!valid(req.JenisPakai, "", "habis_pakai", "tidak_habis_pakai") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nilai pilihan (enum) tidak valid",
		})
	}
	exists, err := c.Repository.IsNomorAsetExists(req.NomorAset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan pada server"})
	}
	if exists {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "nomor_aset sudah digunakan"})
	}
	if _, err = c.RuanganRepository.GetByID(req.RuanganID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Ruangan tidak ditemukan"})
	}
	if req.PICID != nil {
		if _, err = c.UserRepository.GetUserByID(*req.PICID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "PIC tidak ditemukan", "errors": fiber.Map{"pic_id": "User dengan pic_id tersebut tidak ditemukan"}})
		}
	}
	if req.VerifiedBy != nil {
		if _, err = c.UserRepository.GetUserByID(*req.VerifiedBy); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Verifikator tidak ditemukan", "errors": fiber.Map{"verified_by": "User dengan verified_by tersebut tidak ditemukan"}})
		}
	}
	userID, ok := ctx.Locals("user_id").(float64)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "User tidak terautentikasi"})
	}
	jumlah := uint(1)
	if req.Jumlah != nil {
		jumlah = *req.Jumlah
	}
	item := models.Peralatan{RuanganID: req.RuanganID, PICID: req.PICID, NomorAset: req.NomorAset, NamaPeralatan: req.NamaPeralatan, Merk: req.Merk, Model: req.Model, NomorSeri: req.NomorSeri, Jumlah: jumlah, KategoriPeralatan: req.KategoriPeralatan, Kondisi: req.Kondisi, StatusKelayakan: req.StatusKelayakan, Metode: req.Metode, JenisPakai: req.JenisPakai, InputBy: uint64(userID), VerifiedAt: req.VerifiedAt, VerificationNote: req.VerificationNote, VerifiedBy: req.VerifiedBy}
	if item.Kondisi == "" {
		item.Kondisi = "sesuai"
	}
	if item.StatusKelayakan == "" {
		item.StatusKelayakan = "pending"
	}
	if err := c.Repository.Create(&item); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan peralatan"})
	}
	result, _ := c.Repository.FindByID(item.ID)
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": result})
}

func (c *PeralatanController) GetAll(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	ruanganID, _ := strconv.ParseUint(ctx.Query("ruangan_id"), 10, 64)
	picID, _ := strconv.ParseUint(ctx.Query("pic_id"), 10, 64)
	list, total, err := c.Repository.FindAll(repositories.PeralatanQueryParams{Page: page, Limit: limit, Search: ctx.Query("search"), RuanganID: ruanganID, PICID: picID, StatusKelayakan: ctx.Query("status_kelayakan")})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengambil data peralatan"})
	}
	return ctx.JSON(fiber.Map{"success": true, "data": list, "meta": fiber.Map{"page": page, "limit": limit, "total_data": total}})
}

func (c *PeralatanController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	item, err := c.Repository.FindByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data peralatan tidak ditemukan"})
	}
	return ctx.JSON(fiber.Map{"success": true, "data": item})
}

func (c *PeralatanController) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	if _, err = c.Repository.FindByID(id); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data peralatan tidak ditemukan"})
	}

	var req PeralatanRequest
	if err = ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validasi gagal",
			"errors":  formatBindingError(err),
		})
	}

	// 1. Validasi Enum (jika diisi)
	if !valid(req.KategoriPeralatan, "", "peralatan", "Peralatan bantu", "referensi uji", "golden sample", "Komponen pendukung") ||
		!valid(req.Kondisi, "", "sesuai", "tidak_sesuai", "tidak_berlaku") ||
		!valid(req.StatusKelayakan, "", "pending", "aktif", "ditolak", "tidak_aktif") ||
		!valid(req.Metode, "", "internal", "eksternal") ||
		!valid(req.JenisPakai, "", "habis_pakai", "tidak_habis_pakai") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nilai enum tidak valid"})
	}

	updates := map[string]interface{}{}

	// 2. Validasi Ruangan ke Database jika di-update
	if req.RuanganID > 0 {
		if _, err = c.RuanganRepository.GetByID(req.RuanganID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Ruangan tidak ditemukan"})
		}
		updates["ruangan_id"] = req.RuanganID
	}

	// 3. Validasi PIC ke Database jika di-update
	if req.PICID != nil {
		if _, err = c.UserRepository.GetUserByID(*req.PICID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "PIC tidak ditemukan"})
		}
		updates["pic_id"] = req.PICID
	}

	// 4. Validasi Verifikator ke Database jika di-update
	if req.VerifiedBy != nil {
		if _, err = c.UserRepository.GetUserByID(*req.VerifiedBy); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Verifikator tidak ditemukan"})
		}
		updates["verified_by"] = req.VerifiedBy
	}

	if req.NomorAset != "" {
		updates["nomor_aset"] = req.NomorAset
	}
	if req.NamaPeralatan != "" {
		updates["nama_peralatan"] = req.NamaPeralatan
	}
	if req.Merk != "" {
		updates["merk"] = req.Merk
	}
	if req.Model != "" {
		updates["model"] = req.Model
	}
	if req.NomorSeri != "" {
		updates["nomor_seri"] = req.NomorSeri
	}
	if req.Jumlah != nil {
		updates["jumlah"] = *req.Jumlah
	}
	if req.KategoriPeralatan != "" {
		updates["kategori_peralatan"] = req.KategoriPeralatan
	}
	if req.Kondisi != "" {
		updates["kondisi"] = req.Kondisi
	}
	if req.StatusKelayakan != "" {
		updates["status_kelayakan"] = req.StatusKelayakan
	}
	if req.Metode != "" {
		updates["metode"] = req.Metode
	}
	if req.JenisPakai != "" {
		updates["jenis_pakai"] = req.JenisPakai
	}
	if req.VerifiedAt != nil {
		updates["verified_at"] = req.VerifiedAt
	}
	if req.VerificationNote != "" {
		updates["verification_note"] = req.VerificationNote
	}

	if err = c.Repository.Update(id, updates); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui peralatan",
			"error":   err.Error(),
		})
	}

	item, _ := c.Repository.FindByID(id)
	return ctx.JSON(fiber.Map{"success": true, "data": item})
}

func (c *PeralatanController) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if err = c.Repository.Delete(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus peralatan"})
	}
	return ctx.JSON(fiber.Map{"success": true, "message": "Peralatan berhasil dihapus"})
}

func formatBindingError(err error) map[string]string {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		switch typeErr.Field {
		case "jumlah":
			return map[string]string{"jumlah": "Jumlah harus berupa angka bulat positif minimal 1"}
		case "ruangan_id", "pic_id", "verified_by":
			return map[string]string{typeErr.Field: "ID harus berupa angka bulat"}
		default:
			return map[string]string{typeErr.Field: "Tipe data tidak sesuai"}
		}
	}
	return map[string]string{"json": "Format JSON tidak valid atau rusak"}
}
