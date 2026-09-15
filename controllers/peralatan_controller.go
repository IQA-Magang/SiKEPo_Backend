package controllers

import (
	"backend/models"
	"backend/repositories"
	"backend/utils"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type PeralatanController struct {
	Repo             repositories.PeralatanRepository
	DB               *gorm.DB
	NotificationRepo repositories.NotificationRepository
}

func NewPeralatanController(repo repositories.PeralatanRepository) *PeralatanController {
	return &PeralatanController{Repo: repo}
}

func (c *PeralatanController) sendManagerLabNotification(peralatan *models.Peralatan) {
	if c.DB == nil || c.NotificationRepo == nil || peralatan == nil {
		return
	}

	var room models.Ruangan
	if err := c.DB.Preload("Labs").First(&room, peralatan.RuanganID).Error; err != nil {
		log.Printf("Gagal mengambil ruangan untuk notifikasi manager lab: %v", err)
		return
	}

	if room.Labs == nil || room.Labs.ManagerID == nil {
		return
	}

	notification := &models.Notification{
		UserID:  *room.Labs.ManagerID,
		Type:    "equipment_added",
		Title:   "Peralatan baru ditambahkan",
		Message: fmt.Sprintf("Peralatan %s (%s) telah ditambahkan ke ruangan %s.", peralatan.NamaPeralatan, peralatan.NomorAset, room.NamaRuangan),
	}

	if err := c.NotificationRepo.Create(notification); err != nil {
		log.Printf("Gagal mengirim notifikasi ke manager lab: %v", err)
	}
}

// GetAll handles GET /api/peralatan.
func (c *PeralatanController) GetAll(ctx *fiber.Ctx) error {
	peralatan, err := c.Repo.FindAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil daftar peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   peralatan,
	})
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
	peralatan, err := c.Repo.FindByNomorAset(nomorAset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Peralatan berhasil dibuat tetapi gagal mengambil ID",
			"error":   err.Error(),
		})
	}

	c.sendManagerLabNotification(peralatan)

	// 3. Response Berhasil
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":     "success",
		"message":    "Peralatan beserta detail spesifikasinya berhasil ditambahkan",
		"nomor_aset": nomorAset,
		"id":         peralatan.ID,
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

// UploadFoto handles POST /api/peralatan/:id/foto.
func (c *PeralatanController) UploadFoto(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID peralatan tidak valid",
		})
	}

	if _, err := c.Repo.FindByID(uint(id)); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Peralatan tidak ditemukan",
		})
	}

	file, err := ctx.FormFile("foto")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "File foto wajib diisi",
		})
	}

	path, err := utils.SaveUploadedFile(file, "peralatan", fmt.Sprintf("peralatan-%d", id))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menyimpan foto",
			"error":   err.Error(),
		})
	}

	if err := c.Repo.UpdateFoto(uint(id), path); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menyimpan path foto",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"foto":   path,
	})
}
