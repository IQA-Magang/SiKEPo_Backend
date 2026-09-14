package controllers

import (
	"backend/models"
	"backend/repositories"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type PeralatanController struct {
	Repo repositories.PeralatanRepository
}

func NewPeralatanController(repo repositories.PeralatanRepository) *PeralatanController {
	return &PeralatanController{Repo: repo}
}

// Create handles POST /api/peralatan
func (c *PeralatanController) Create(ctx *fiber.Ctx) error {
	req := new(models.CreatePeralatanRequest)

	// 1. Parsing JSON Body ke Struct Request
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request body tidak valid",
			"error":   err.Error(),
		})
	}

	// (Opsional) Lakukan validasi manual sederhana jika diperlukan
	if req.NamaPeralatan == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Nama Peralatan wajib diisi",
		})
	}

	if req.KategoriPeralatanID < 1 || req.KategoriPeralatanID > 4 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Kategori ID tidak valid (harus 1 - 4)",
		})
	}

	// 2. Eksekusi Repository untuk Simpan Data
	nomorAset, err := c.Repo.CreatePeralatan(req)
	if err != nil {
		// Pengecekan apakah error karena nomor aset duplikat (tergantung driver DB, ini error umum)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menyimpan data peralatan",
			"error":   err.Error(),
		})
	}

	// 3. Response Berhasil
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":     "success",
		"message":    "Peralatan beserta detail spesifikasinya berhasil ditambahkan",
		"nomor_aset": nomorAset,
	})
}

// GenerateQRCode handles GET /api/peralatan/:id/qr.
func (c *PeralatanController) GenerateQRCode(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID peralatan tidak valid",
		})
	}

	peralatan, err := c.Repo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Peralatan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data peralatan",
			"error":   err.Error(),
		})
	}

	qr, err := qrcode.Encode(peralatan.NomorAset, qrcode.Medium, 256)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membuat QR code",
			"error":   err.Error(),
		})
	}

	ctx.Set(fiber.HeaderContentType, "image/png")
	return ctx.Send(qr)
}
