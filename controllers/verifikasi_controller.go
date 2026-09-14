package controllers

// import (
// 	"errors"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"backend/models"
// 	"backend/repositories"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// type VerifikasiController struct {
// 	Repository          *repositories.VerifikasiRepository
// 	PeralatanRepository *repositories.PeralatanRepository
// 	UserRepository      *repositories.UserRepository
// }

// // ========================================
// // GET ALL
// // ========================================

// func (c *VerifikasiController) GetVerifikasi(ctx *fiber.Ctx) error {

// 	data, err := c.Repository.GetAllVerifikasi()

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"data":    data,
// 	})
// }

// // ========================================
// // GET BY ID
// // ========================================

// func (c *VerifikasiController) GetVerifikasiByID(ctx *fiber.Ctx) error {

// 	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID verifikasi tidak valid",
// 		})
// 	}

// 	data, err := c.Repository.GetVerifikasiByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Data verifikasi tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"data":    data,
// 	})
// }

// // ========================================
// // GET BY PERALATAN
// // ========================================

// func (c *VerifikasiController) GetByPeralatan(ctx *fiber.Ctx) error {

// 	peralatanID, err := strconv.ParseUint(
// 		ctx.Params("peralatan_id"),
// 		10,
// 		64,
// 	)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID peralatan tidak valid",
// 		})
// 	}

// 	data, err := c.Repository.GetByPeralatanID(peralatanID)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil riwayat verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"data":    data,
// 	})
// }

// // ========================================
// // CREATE VERIFIKASI
// // ========================================

// func (c *VerifikasiController) CreateVerifikasi(ctx *fiber.Ctx) error {

// 	// ========================================
// 	// REQUEST
// 	// ========================================

// 	type CreateRequest struct {
// 		IDPeralatan       uint64  `json:"id_peralatan"`
// 		TanggalVerifikasi string  `json:"tanggal_verifikasi"`
// 		KodeAktivitas     string  `json:"kode_aktivitas"`
// 		IDKriteria        *uint64 `json:"id_kriteria"`
// 		TindakLanjut      string  `json:"tindak_lanjut"`
// 		Catatan           string  `json:"catatan"`

// 		HasilVerifikasi struct {
// 			Identitas    string `json:"identitas"`
// 			Kelengkapan  string `json:"kelengkapan"`
// 			Firmware     string `json:"firmware"`
// 			KondisiFisik string `json:"kondisi_fisik"`
// 			Segel        string `json:"segel"`
// 			FungsiAwal   string `json:"fungsi_awal"`
// 			Metrologi    string `json:"metrologi"`
// 			Sertifikat   string `json:"sertifikat"`
// 			Catatan      string `json:"catatan"`
// 		} `json:"hasil_verifikasi"`
// 	}

// 	var request CreateRequest

// 	// ========================================
// 	// PARSE JSON
// 	// ========================================

// 	if err := ctx.BodyParser(&request); err != nil {

// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ========================================
// 	// VALIDASI ID PERALATAN
// 	// ========================================

// 	if request.IDPeralatan == 0 {

// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID peralatan wajib diisi",
// 		})
// 	}

// 	peralatan, err := c.PeralatanRepository.FindByID(
// 		request.IDPeralatan,
// 	)

// 	if err != nil || peralatan == nil {

// 		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan tidak ditemukan",
// 		})
// 	}

// 	// ========================================
// 	// VALIDASI KODE AKTIVITAS
// 	// ========================================

// 	request.KodeAktivitas = strings.TrimSpace(
// 		request.KodeAktivitas,
// 	)

// 	if request.KodeAktivitas == "" {

// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Kode aktivitas wajib diisi",
// 		})
// 	}

// 	// ========================================
// 	// PARSE TANGGAL
// 	// ========================================

// 	tanggal := time.Now()

// 	tanggalInput := strings.TrimSpace(
// 		request.TanggalVerifikasi,
// 	)

// 	if tanggalInput != "" {

// 		tanggalParsed, err := parseTanggal(tanggalInput)

// 		if err != nil {

// 			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Format tanggal_verifikasi tidak valid",
// 				"format_yang_diterima": []string{
// 					"2026-09-11T08:00",
// 					"2026-09-11T08:00:00",
// 					"2026-09-11T08:00:00+07:00",
// 					"2026-09-11",
// 					"11/09/2026",
// 					"11/09/2026 08:00",
// 					"2026-09-11 08:00",
// 				},
// 				"nilai_diterima": tanggalInput,
// 				"error":          err.Error(),
// 			})
// 		}

// 		tanggal = tanggalParsed
// 	}

// 	// ========================================
// 	// VALIDASI HASIL VERIFIKASI
// 	// ========================================

// 	hasil := request.HasilVerifikasi

// 	values := []string{
// 		hasil.Identitas,
// 		hasil.Kelengkapan,
// 		hasil.Firmware,
// 		hasil.KondisiFisik,
// 		hasil.Segel,
// 		hasil.FungsiAwal,
// 		hasil.Metrologi,
// 		hasil.Sertifikat,
// 	}

// 	for _, value := range values {

// 		if !isValidHasilVerifikasi(value) {

// 			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Nilai hasil verifikasi hanya boleh S, TS, atau TB",
// 				"nilai":   value,
// 			})
// 		}
// 	}

// 	// ========================================
// 	// CREATE VERIFIKASI
// 	// ========================================

// 	verifikasi := models.Verifikasi{
// 		IDPeralatan:       request.IDPeralatan,
// 		TanggalVerifikasi: tanggal,
// 		KodeAktivitas:     request.KodeAktivitas,
// 		IDKriteria:        request.IDKriteria,
// 		Keputusan:         nil,
// 		TindakLanjut:      request.TindakLanjut,
// 		Catatan:           request.Catatan,
// 	}

// 	if err := c.Repository.CreateVerifikasi(
// 		&verifikasi,
// 	); err != nil {

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal menyimpan data verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ========================================
// 	// CREATE HASIL VERIFIKASI
// 	// ========================================

// 	hasilData := models.HasilVerifikasi{
// 		IDVerifikasi: verifikasi.IDVerifikasi,

// 		Identitas:    hasil.Identitas,
// 		Kelengkapan:  hasil.Kelengkapan,
// 		Firmware:     hasil.Firmware,
// 		KondisiFisik: hasil.KondisiFisik,
// 		Segel:        hasil.Segel,
// 		FungsiAwal:   hasil.FungsiAwal,
// 		Metrologi:    hasil.Metrologi,
// 		Sertifikat:   hasil.Sertifikat,

// 		Catatan: hasil.Catatan,
// 	}

// 	if err := c.Repository.CreateHasilVerifikasi(
// 		&hasilData,
// 	); err != nil {

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Data verifikasi berhasil dibuat, tetapi hasil verifikasi gagal disimpan",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ========================================
// 	// AMBIL DATA TERBARU
// 	// ========================================

// 	result, err := c.Repository.GetVerifikasiByID(
// 		verifikasi.IDVerifikasi,
// 	)

// 	if err != nil {

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Verifikasi berhasil dibuat tetapi gagal mengambil data",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ========================================
// 	// RESPONSE
// 	// ========================================

// 	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
// 		"success": true,
// 		"message": "Verifikasi berhasil dibuat",
// 		"data":    result,
// 	})
// }

// // ========================================
// // APPROVE VERIFIKASI
// // ========================================

// func (c *VerifikasiController) ApproveVerifikasi(
// 	ctx *fiber.Ctx,
// ) error {

// 	id, err := strconv.ParseUint(
// 		ctx.Params("id"),
// 		10,
// 		64,
// 	)

// 	if err != nil {

// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID verifikasi tidak valid",
// 		})
// 	}

// 	data, err := c.Repository.GetVerifikasiByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {

// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Data verifikasi tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ========================================
// 	// CEK SUDAH DIVERIFIKASI
// 	// ========================================

// 	if data.VerifiedBy != nil {

// 		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Verifikasi sudah disahkan",
// 		})
// 	}

// 	// ========================================
// 	// AMBIL USER ID DARI JWT
// 	// ========================================

// 	userIDValue := ctx.Locals("user_id")

// 	if userIDValue == nil {

// 		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"success": false,
// 			"message": "User tidak terautentikasi",
// 		})
// 	}

// 	var userID uint64

// 	switch value := userIDValue.(type) {

// 	case uint64:
// 		userID = value

// 	case uint:
// 		userID = uint64(value)

// 	case uint32:
// 		userID = uint64(value)

// 	case int:
// 		userID = uint64(value)

// 	case int64:
// 		userID = uint64(value)

// 	case float64:
// 		userID = uint64(value)

// 	case string:
// 		parsed, err := strconv.ParseUint(value, 10, 64)

// 		if err != nil {

// 			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 				"success": false,
// 				"message": "User ID tidak valid",
// 			})
// 		}

// 		userID = parsed

// 	default:

// 		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"success": false,
// 			"message": "User ID tidak valid",
// 		})
// 	}

// 	// ========================================
// 	// APPROVE
// 	// ========================================

// 	if err := c.Repository.ApproveVerifikasi(
// 		id,
// 		userID,
// 	); err != nil {

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengesahkan verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, err := c.Repository.GetVerifikasiByID(id)

// 	if err != nil {

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Verifikasi berhasil disahkan tetapi gagal mengambil data terbaru",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Verifikasi berhasil disahkan",
// 		"data":    result,
// 	})
// }

// // ========================================
// // DELETE
// // ========================================

// func (c *VerifikasiController) DeleteVerifikasi(
// 	ctx *fiber.Ctx,
// ) error {

// 	id, err := strconv.ParseUint(
// 		ctx.Params("id"),
// 		10,
// 		64,
// 	)

// 	if err != nil {

// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID verifikasi tidak valid",
// 		})
// 	}

// 	data, err := c.Repository.GetVerifikasiByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {

// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Data verifikasi tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil data verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ========================================
// 	// CEK SUDAH DISAHKAN
// 	// ========================================

// 	if data.VerifiedBy != nil {

// 		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Verifikasi yang sudah disahkan tidak dapat dihapus",
// 		})
// 	}

// 	// ========================================
// 	// DELETE
// 	// ========================================

// 	if err := c.Repository.DeleteVerifikasi(id); err != nil {

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal menghapus verifikasi",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Verifikasi berhasil dihapus",
// 	})
// }

// // ========================================
// // PARSE TANGGAL
// // ========================================

// func parseTanggal(value string) (time.Time, error) {

// 	value = strings.TrimSpace(value)

// 	// ========================================
// 	// FORMAT YANG DIDUKUNG
// 	// ========================================

// 	layouts := []string{

// 		// datetime-local HTML
// 		"2006-01-02T15:04",

// 		// datetime dengan detik
// 		"2006-01-02T15:04:05",

// 		// datetime dengan milisecond
// 		"2006-01-02T15:04:05.000",

// 		// RFC3339
// 		time.RFC3339,

// 		// RFC3339 dengan nano
// 		time.RFC3339Nano,

// 		// format database
// 		"2006-01-02 15:04",

// 		"2006-01-02 15:04:05",

// 		// hanya tanggal
// 		"2006-01-02",

// 		// format Indonesia
// 		"02/01/2006",

// 		"02/01/2006 15:04",

// 		"02/01/2006 15:04:05",

// 		// format Indonesia menggunakan -
// 		"02-01-2006",

// 		"02-01-2006 15:04",

// 		"02-01-2006 15:04:05",
// 	}

// 	// ========================================
// 	// COBA SEMUA FORMAT
// 	// ========================================

// 	for _, layout := range layouts {

// 		// Untuk format yang memiliki timezone
// 		if layout == time.RFC3339 ||
// 			layout == time.RFC3339Nano {

// 			result, err := time.Parse(layout, value)

// 			if err == nil {
// 				return result, nil
// 			}

// 			continue
// 		}

// 		// Format tanpa timezone
// 		result, err := time.ParseInLocation(
// 			layout,
// 			value,
// 			time.Local,
// 		)

// 		if err == nil {
// 			return result, nil
// 		}
// 	}

// 	// ========================================
// 	// SEMUA FORMAT GAGAL
// 	// ========================================

// 	return time.Time{}, errors.New(
// 		"format tanggal tidak dikenali",
// 	)
// }

// // ========================================
// // VALIDASI HASIL VERIFIKASI
// // ========================================

// func isValidHasilVerifikasi(value string) bool {

// 	value = strings.TrimSpace(value)

// 	switch value {

// 	case "S":
// 		return true

// 	case "TS":
// 		return true

// 	case "TB":
// 		return true

// 	default:
// 		return false
// 	}
// }
