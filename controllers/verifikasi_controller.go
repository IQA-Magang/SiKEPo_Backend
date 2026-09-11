package controllers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"backend/models"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type VerifikasiController struct {
	Repository          *repositories.VerifikasiRepository
	PeralatanRepository *repositories.PeralatanRepository
	UserRepository      *repositories.UserRepository
}

// ========================================
// GET ALL
// ========================================

func (c *VerifikasiController) GetVerifikasi(ctx *fiber.Ctx) error {

	data, err := c.Repository.GetAllVerifikasi()

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data verifikasi",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// ========================================
// GET BY ID
// ========================================

func (c *VerifikasiController) GetVerifikasiByID(ctx *fiber.Ctx) error {

	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID verifikasi tidak valid",
		})
	}

	data, err := c.Repository.GetVerifikasiByID(id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Data verifikasi tidak ditemukan",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data verifikasi",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// ========================================
// GET BY PERALATAN
// ========================================

func (c *VerifikasiController) GetByPeralatan(ctx *fiber.Ctx) error {

	peralatanID, err := strconv.ParseUint(
		ctx.Params("peralatan_id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID peralatan tidak valid",
		})
	}

	data, err := c.Repository.GetByPeralatanID(peralatanID)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil riwayat verifikasi",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// ========================================
// CREATE
// ========================================

func (c *VerifikasiController) CreateVerifikasi(ctx *fiber.Ctx) error {

	type CreateRequest struct {
		IDPeralatan       uint64     `json:"id_peralatan"`
		TanggalVerifikasi *time.Time `json:"tanggal_verifikasi"`
		KodeAktivitas     string     `json:"kode_aktivitas"`
		IDKriteria        *uint64    `json:"id_kriteria"`
		TindakLanjut      string     `json:"tindak_lanjut"`
		Catatan           string     `json:"catatan"`

		HasilVerifikasi struct {
			Identitas    string `json:"identitas"`
			Kelengkapan  string `json:"kelengkapan"`
			Firmware     string `json:"firmware"`
			KondisiFisik string `json:"kondisi_fisik"`
			Segel        string `json:"segel"`
			FungsiAwal   string `json:"fungsi_awal"`
			Metrologi    string `json:"metrologi"`
			Sertifikat   string `json:"sertifikat"`
			Catatan      string `json:"catatan"`
		} `json:"hasil_verifikasi"`
	}

	var request CreateRequest

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	// ========================================
	// VALIDASI PERALATAN
	// ========================================

	if request.IDPeralatan == 0 {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID peralatan wajib diisi",
		})
	}

	peralatan, err := c.PeralatanRepository.FindByID(
		request.IDPeralatan,
	)

	if err != nil || peralatan == nil {
		return ctx.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Peralatan tidak ditemukan",
		})
	}

	// ========================================
	// VALIDASI KODE AKTIVITAS
	// ========================================

	request.KodeAktivitas = strings.TrimSpace(
		request.KodeAktivitas,
	)

	if request.KodeAktivitas == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Kode aktivitas wajib diisi",
		})
	}

	// ========================================
	// TANGGAL
	// ========================================

	tanggal := time.Now()

	if request.TanggalVerifikasi != nil {
		tanggal = *request.TanggalVerifikasi
	}

	// ========================================
	// VALIDASI HASIL
	// ========================================

	hasil := request.HasilVerifikasi

	values := []string{
		hasil.Identitas,
		hasil.Kelengkapan,
		hasil.Firmware,
		hasil.KondisiFisik,
		hasil.Segel,
		hasil.FungsiAwal,
		hasil.Metrologi,
		hasil.Sertifikat,
	}

	for _, value := range values {
		if !isValidHasilVerifikasi(value) {
			return ctx.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Hasil verifikasi hanya boleh S, TS, atau TB",
			})
		}
	}

	// ========================================
	// TENTUKAN KEPUTUSAN
	// ========================================

	keputusan := "Layak"

	for _, value := range values {
		if value == "TS" {
			keputusan = "Tidak Layak"
			break
		}
	}

	// ========================================
	// CATATAN UNTUK TS / TB
	// ========================================

	if (keputusan == "Tidak Layak" ||
		hasTB(values)) &&
		strings.TrimSpace(
			hasil.Catatan,
		) == "" {

		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Catatan wajib diisi apabila terdapat hasil TS atau TB",
		})
	}

	// ========================================
	// CREATE VERIFIKASI
	// ========================================

	verifikasi := models.Verifikasi{
		IDPeralatan:       request.IDPeralatan,
		TanggalVerifikasi: tanggal,
		KodeAktivitas:     request.KodeAktivitas,
		IDKriteria:        request.IDKriteria,
		Keputusan:         nil,
		TindakLanjut:      request.TindakLanjut,
		Catatan:           request.Catatan,
	}

	if err := c.Repository.CreateVerifikasi(
		&verifikasi,
	); err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat verifikasi",
			"error":   err.Error(),
		})
	}

	// ========================================
	// CREATE HASIL VERIFIKASI
	// ========================================

	hasilData := models.HasilVerifikasi{
		IDVerifikasi: verifikasi.IDVerifikasi,

		Identitas:    hasil.Identitas,
		Kelengkapan:  hasil.Kelengkapan,
		Firmware:     hasil.Firmware,
		KondisiFisik: hasil.KondisiFisik,
		Segel:        hasil.Segel,
		FungsiAwal:   hasil.FungsiAwal,
		Metrologi:    hasil.Metrologi,
		Sertifikat:   hasil.Sertifikat,

		Catatan: hasil.Catatan,
	}

	if err := c.Repository.CreateHasilVerifikasi(
		&hasilData,
	); err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Verifikasi dibuat tetapi hasil verifikasi gagal disimpan",
			"error":   err.Error(),
		})
	}

	result, _ := c.Repository.GetVerifikasiByID(
		verifikasi.IDVerifikasi,
	)

	return ctx.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Verifikasi berhasil dibuat",
		"data":    result,
	})
}

// ========================================
// APPROVE / SAHKAN
// ========================================

func (c *VerifikasiController) ApproveVerifikasi(
	ctx *fiber.Ctx,
) error {

	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID verifikasi tidak valid",
		})
	}

	data, err := c.Repository.GetVerifikasiByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Data verifikasi tidak ditemukan",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil verifikasi",
		})
	}

	// Tidak boleh mengubah verifikasi yang sudah disahkan
	if data.VerifiedBy != nil {
		return ctx.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "Verifikasi sudah disahkan dan tidak dapat diubah",
		})
	}

	// ========================================
	// AMBIL USER DARI JWT
	// ========================================

	userIDValue := ctx.Locals("user_id")

	if userIDValue == nil {
		return ctx.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "User tidak terautentikasi",
		})
	}

	var userID uint64

	switch value := userIDValue.(type) {

	case float64:
		userID = uint64(value)

	case uint64:
		userID = value

	case int:
		userID = uint64(value)

	default:
		return ctx.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "User ID tidak valid",
		})
	}

	// ========================================
	// APPROVE
	// ========================================

	if err := c.Repository.ApproveVerifikasi(
		id,
		userID,
	); err != nil {

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengesahkan verifikasi",
			"error":   err.Error(),
		})
	}

	result, _ := c.Repository.GetVerifikasiByID(id)

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Verifikasi berhasil disahkan",
		"data":    result,
	})
}

// ========================================
// DELETE
// ========================================

func (c *VerifikasiController) DeleteVerifikasi(
	ctx *fiber.Ctx,
) error {

	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID verifikasi tidak valid",
		})
	}

	data, err := c.Repository.GetVerifikasiByID(id)

	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Data verifikasi tidak ditemukan",
		})
	}

	// Verifikasi yang sudah disahkan tidak boleh dihapus
	if data.VerifiedBy != nil {
		return ctx.Status(409).JSON(fiber.Map{
			"success": false,
			"message": "Verifikasi yang sudah disahkan tidak dapat dihapus",
		})
	}

	if err := c.Repository.DeleteVerifikasi(id); err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus verifikasi",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Verifikasi berhasil dihapus",
	})
}

// ========================================
// VALIDASI S / TS / TB
// ========================================

func isValidHasilVerifikasi(value string) bool {

	switch value {
	case "S", "TS", "TB":
		return true

	default:
		return false
	}
}

// ========================================
// CEK TB
// ========================================

func hasTB(values []string) bool {

	for _, value := range values {
		if value == "TB" {
			return true
		}
	}

	return false
}
