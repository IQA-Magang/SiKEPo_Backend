package controllers

// import (
// 	"errors"
// 	"strconv"
// 	"time"

// 	"backend/models"
// 	"backend/repositories"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// type DetailPeminjamanController struct {
// 	Repository          *repositories.DetailPeminjamanRepository
// 	PeralatanRepository *repositories.PeralatanRepository
// }

// type DetailPeminjamanRequest struct {
// 	PeminjamanID uint64 `json:"peminjaman_id"`
// 	PeralatanID  uint64 `json:"peralatan_id"`
// 	Jumlah       uint   `json:"jumlah"`
// 	Catatan      string `json:"catatan"`
// }

// type ApprovalRequest struct {
// 	VerificationNote string `json:"verification_note"`
// }

// type KondisiRequest struct {
// 	Kondisi string `json:"kondisi"`
// 	Catatan string `json:"catatan"`
// }

// // ========================================
// // GET ALL
// // ========================================

// func (c *DetailPeminjamanController) GetAll(ctx *fiber.Ctx) error {

// 	data, err := c.Repository.GetAll()

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail peminjaman",
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

// func (c *DetailPeminjamanController) GetByID(ctx *fiber.Ctx) error {

// 	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	data, err := c.Repository.GetByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Detail peminjaman tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"data":    data,
// 	})
// }

// // ========================================
// // GET BY PEMINJAMAN ID
// // ========================================

// func (c *DetailPeminjamanController) GetByPeminjamanID(
// 	ctx *fiber.Ctx,
// ) error {

// 	peminjamanID, err := strconv.ParseUint(
// 		ctx.Params("peminjaman_id"),
// 		10,
// 		64,
// 	)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID peminjaman tidak valid",
// 		})
// 	}

// 	data, err := c.Repository.GetByPeminjamanID(peminjamanID)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"data":    data,
// 	})
// }

// // ========================================
// // CREATE
// // ========================================

// func (c *DetailPeminjamanController) Create(
// 	ctx *fiber.Ctx,
// ) error {

// 	var req DetailPeminjamanRequest

// 	if err := ctx.BodyParser(&req); err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 		})
// 	}

// 	if req.PeminjamanID == 0 {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "peminjaman_id wajib diisi",
// 		})
// 	}

// 	if req.PeralatanID == 0 {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "peralatan_id wajib diisi",
// 		})
// 	}

// 	if req.Jumlah < 1 {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Jumlah minimal 1",
// 		})
// 	}

// 	// ========================================
// 	// CEK PERALATAN
// 	// ========================================

// 	// peralatan, err := c.PeralatanRepository.FindByID(req.PeralatanID)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan tidak ditemukan",
// 		})
// 	}

// 	// ========================================
// 	// CEK STATUS PERALATAN
// 	// ========================================

// 	if peralatan.StatusKelayakan != "aktif" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan belum berstatus aktif dan tidak dapat dipinjam",
// 		})
// 	}

// 	// ========================================
// 	// CEK KONDISI PERALATAN
// 	// ========================================

// 	if peralatan.Kondisi == "tidak_sesuai" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan dalam kondisi tidak sesuai dan tidak dapat dipinjam",
// 		})
// 	}

// 	// ========================================
// 	// STATUS DEFAULT PENDING
// 	// ========================================

// 	data := models.DetailPeminjaman{
// 		PeminjamanID: req.PeminjamanID,
// 		PeralatanID:  req.PeralatanID,
// 		Jumlah:       req.Jumlah,
// 		Status:       "pending",
// 		Catatan:      req.Catatan,
// 	}

// 	err = c.Repository.Create(&data)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal membuat detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, err := c.Repository.GetByID(data.ID)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail berhasil dibuat tetapi gagal mengambil data",
// 		})
// 	}

// 	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
// 		"success": true,
// 		"message": "Detail peminjaman berhasil dibuat",
// 		"data":    result,
// 	})
// }

// // ========================================
// // UPDATE
// // ========================================

// func (c *DetailPeminjamanController) Update(
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
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	existing, err := c.Repository.GetByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Detail peminjaman tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail peminjaman",
// 		})
// 	}

// 	// Detail yang sudah diproses tidak boleh diubah sembarangan
// 	if existing.Status != "pending" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail peminjaman yang sudah diproses tidak dapat diubah",
// 		})
// 	}

// 	var req DetailPeminjamanRequest

// 	if err := ctx.BodyParser(&req); err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 		})
// 	}

// 	if req.PeminjamanID == 0 {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "peminjaman_id wajib diisi",
// 		})
// 	}

// 	if req.PeralatanID == 0 {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "peralatan_id wajib diisi",
// 		})
// 	}

// 	if req.Jumlah < 1 {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Jumlah minimal 1",
// 		})
// 	}

// 	peralatan, err := c.PeralatanRepository.FindByID(req.PeralatanID)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan tidak ditemukan",
// 		})
// 	}

// 	if peralatan.StatusKelayakan != "aktif" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan belum berstatus aktif",
// 		})
// 	}

// 	if peralatan.Kondisi == "tidak_sesuai" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan tidak dapat dipinjam karena kondisinya tidak sesuai",
// 		})
// 	}

// 	data := models.DetailPeminjaman{
// 		PeminjamanID: req.PeminjamanID,
// 		PeralatanID:  req.PeralatanID,
// 		Jumlah:       req.Jumlah,
// 		Catatan:      req.Catatan,
// 	}

// 	err = c.Repository.Update(id, &data)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal memperbarui detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, _ := c.Repository.GetByID(id)

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Detail peminjaman berhasil diperbarui",
// 		"data":    result,
// 	})
// }

// // ========================================
// // APPROVE
// // ========================================

// func (c *DetailPeminjamanController) Approve(
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
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	detail, err := c.Repository.GetByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Detail peminjaman tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail peminjaman",
// 		})
// 	}

// 	if detail.Status != "pending" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail peminjaman sudah diproses",
// 		})
// 	}

// 	// Ambil peralatan terbaru
// 	peralatan, err := c.PeralatanRepository.FindByID(detail.PeralatanID)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan tidak ditemukan",
// 		})
// 	}

// 	// SRS FR-M6-05
// 	if peralatan.StatusKelayakan != "aktif" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peminjaman tidak dapat disetujui karena peralatan belum aktif",
// 		})
// 	}

// 	if peralatan.Kondisi == "tidak_sesuai" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peminjaman tidak dapat disetujui karena kondisi peralatan tidak sesuai",
// 		})
// 	}

// 	// user_id dari JWT
// 	userIDValue := ctx.Locals("user_id")

// 	var verifiedBy uint64

// 	switch v := userIDValue.(type) {
// 	case float64:
// 		verifiedBy = uint64(v)
// 	case uint64:
// 		verifiedBy = v
// 	case int:
// 		verifiedBy = uint64(v)
// 	case string:
// 		verifiedBy, err = strconv.ParseUint(v, 10, 64)
// 		if err != nil {
// 			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 				"success": false,
// 				"message": "User ID pada token tidak valid",
// 			})
// 		}
// 	default:
// 		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"success": false,
// 			"message": "User ID tidak ditemukan pada token",
// 		})
// 	}

// 	var req ApprovalRequest

// 	if err := ctx.BodyParser(&req); err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 		})
// 	}

// 	err = c.Repository.Approve(
// 		id,
// 		verifiedBy,
// 		req.VerificationNote,
// 	)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal menyetujui detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, _ := c.Repository.GetByID(id)

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Detail peminjaman berhasil disetujui",
// 		"data":    result,
// 	})
// }

// // ========================================
// // REJECT
// // ========================================

// func (c *DetailPeminjamanController) Reject(
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
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	detail, err := c.Repository.GetByID(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Detail peminjaman tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mengambil detail peminjaman",
// 		})
// 	}

// 	if detail.Status != "pending" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail peminjaman sudah diproses",
// 		})
// 	}

// 	var req ApprovalRequest

// 	if err := ctx.BodyParser(&req); err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 		})
// 	}

// 	if req.VerificationNote == "" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Catatan penolakan wajib diisi",
// 		})
// 	}

// 	userIDValue := ctx.Locals("user_id")

// 	var verifiedBy uint64

// 	switch v := userIDValue.(type) {
// 	case float64:
// 		verifiedBy = uint64(v)
// 	case uint64:
// 		verifiedBy = v
// 	case int:
// 		verifiedBy = uint64(v)
// 	case string:
// 		verifiedBy, err = strconv.ParseUint(v, 10, 64)
// 		if err != nil {
// 			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 				"success": false,
// 				"message": "User ID pada token tidak valid",
// 			})
// 		}
// 	default:
// 		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"success": false,
// 			"message": "User ID tidak ditemukan pada token",
// 		})
// 	}

// 	err = c.Repository.Reject(
// 		id,
// 		verifiedBy,
// 		req.VerificationNote,
// 	)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal menolak detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, _ := c.Repository.GetByID(id)

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Detail peminjaman berhasil ditolak",
// 		"data":    result,
// 	})
// }

// // ========================================
// // KONDISI SAAT PINJAM
// // ========================================

// func (c *DetailPeminjamanController) SetKondisiPinjam(
// 	ctx *fiber.Ctx,
// ) error {

// 	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	detail, err := c.Repository.GetByID(id)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail peminjaman tidak ditemukan",
// 		})
// 	}

// 	if detail.Status != "approved" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Peralatan belum disetujui untuk dipinjam",
// 		})
// 	}

// 	var req KondisiRequest

// 	if err := ctx.BodyParser(&req); err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 		})
// 	}

// 	if !isValidKondisi(req.Kondisi) {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Kondisi harus baik, rusak_ringan, atau rusak_berat",
// 		})
// 	}

// 	err = c.Repository.SetKondisiPinjam(
// 		id,
// 		req.Kondisi,
// 		req.Catatan,
// 	)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mencatat kondisi saat pinjam",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, _ := c.Repository.GetByID(id)

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Kondisi saat pinjam berhasil dicatat",
// 		"data":    result,
// 	})
// }

// // ========================================
// // KONDISI SAAT KEMBALI
// // ========================================

// func (c *DetailPeminjamanController) SetKondisiKembali(
// 	ctx *fiber.Ctx,
// ) error {

// 	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	detail, err := c.Repository.GetByID(id)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail peminjaman tidak ditemukan",
// 		})
// 	}

// 	if detail.Status != "approved" {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Detail peminjaman belum disetujui",
// 		})
// 	}

// 	var req KondisiRequest

// 	if err := ctx.BodyParser(&req); err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Format JSON tidak valid",
// 		})
// 	}

// 	if !isValidKondisi(req.Kondisi) {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Kondisi harus baik, rusak_ringan, atau rusak_berat",
// 		})
// 	}

// 	err = c.Repository.SetKondisiKembali(
// 		id,
// 		req.Kondisi,
// 		req.Catatan,
// 	)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal mencatat kondisi saat kembali",
// 			"error":   err.Error(),
// 		})
// 	}

// 	result, _ := c.Repository.GetByID(id)

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Kondisi saat kembali berhasil dicatat",
// 		"data":    result,
// 	})
// }

// // ========================================
// // DELETE
// // ========================================

// func (c *DetailPeminjamanController) Delete(
// 	ctx *fiber.Ctx,
// ) error {

// 	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

// 	if err != nil {
// 		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "ID detail peminjaman tidak valid",
// 		})
// 	}

// 	err = c.Repository.Delete(id)

// 	if err != nil {

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"success": false,
// 				"message": "Detail peminjaman tidak ditemukan",
// 			})
// 		}

// 		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Gagal menghapus detail peminjaman",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return ctx.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Detail peminjaman berhasil dihapus",
// 	})
// }

// // ========================================
// // VALIDASI KONDISI
// // ========================================

// func isValidKondisi(kondisi string) bool {

// 	switch kondisi {
// 	case "baik", "rusak_ringan", "rusak_berat":
// 		return true
// 	default:
// 		return false
// 	}
// }

// // Hindari unused import ketika menggunakan time pada project tertentu.
// var _ = time.Now
