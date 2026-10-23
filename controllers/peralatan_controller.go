package controllers

import (
	"backend/models"
	"backend/repositories"
	"backend/utils"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

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

func (c *PeralatanController) sendPICPeralatanNotification(peralatan *models.Peralatan) {
	if c.DB == nil || peralatan == nil {
		return
	}

	if err := utils.NotifyNewPeralatanCreated(c.DB, peralatan); err != nil {
		log.Printf("Gagal mengirim notifikasi ke PIC peralatan: %v", err)
	}
}

func (c *PeralatanController) sendManagerLabNotification(peralatan *models.Peralatan) {
	if c.DB == nil || peralatan == nil {
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

	if err := utils.NotifyManagerOnVerificationStatusChange(c.DB, peralatan, "menunggu verifikasi"); err != nil {
		log.Printf("Gagal mengirim notifikasi ke manager lab: %v", err)
	}
}

func (c *PeralatanController) NotifyVerificationStatusChanged(peralatan *models.Peralatan, status string) {
	if c.DB == nil || peralatan == nil {
		return
	}

	if err := utils.NotifyManagerOnVerificationStatusChange(c.DB, peralatan, status); err != nil {
		log.Printf("Gagal mengirim notifikasi status verifikasi ke manager: %v", err)
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

func (c *PeralatanController) GetByNomorAset(ctx *fiber.Ctx) error {
	nomorAset := strings.TrimSpace(ctx.Params("nomor_aset"))
	if nomorAset == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Nomor aset peralatan tidak valid",
		})
	}

	peralatan, err := c.Repo.FindByNomorAset(nomorAset)
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

	var detail interface{}
	if c.DB != nil {
		switch peralatan.KategoriPeralatanID {
		case 1:
			var item models.DetailAlatUkur
			err = c.DB.Where("peralatan_id = ?", peralatan.ID).First(&item).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				detail = nil
			} else {
				detail = item
			}
		case 2:
			var item models.DetailAlatBantu
			err = c.DB.Where("peralatan_id = ?", peralatan.ID).First(&item).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				detail = nil
			} else {
				detail = item
			}
		case 3:
			var item models.DetailArtefakAcuan
			err = c.DB.Where("peralatan_id = ?", peralatan.ID).First(&item).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				detail = nil
			} else {
				detail = item
			}
		case 4:
			var item models.DetailKomponenPendukung
			err = c.DB.Where("peralatan_id = ?", peralatan.ID).First(&item).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				detail = nil
			} else {
				detail = item
			}
		default:
			detail = nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal mengambil detail peralatan",
				"error":   err.Error(),
			})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data peralatan berhasil diambil",
		"data": fiber.Map{
			"peralatan": peralatan,
			"detail":    detail,
		},
	})
}

// Create handles POST /api/peralatan
func (c *PeralatanController) Create(ctx *fiber.Ctx) error {
	req := new(models.CreatePeralatanRequest)

	// =====================================================
	// PARSING JSON
	// =====================================================

	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request body tidak valid",
			"error":   err.Error(),
		})
	}

	// =====================================================
	// VALIDASI NAMA
	// =====================================================

	req.NamaPeralatan = strings.TrimSpace(
		req.NamaPeralatan,
	)

	if req.NamaPeralatan == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Nama Peralatan wajib diisi",
		})
	}

	// =====================================================
	// VALIDASI KATEGORI
	// =====================================================

	if req.KategoriPeralatanID < 1 ||
		req.KategoriPeralatanID > 4 {

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Kategori ID tidak valid (harus 1 - 4)",
		})
	}

	// =====================================================
	// VALIDASI KELOMPOK ASSET
	// =====================================================

	if req.KelompokAsetID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Kelompok asset wajib dipilih",
		})
	}

	// =====================================================
	// VALIDASI RUANGAN
	// =====================================================

	if req.RuanganID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Ruangan wajib dipilih",
		})
	}

	// =====================================================
	// CREATE PERALATAN
	// =====================================================

	nomorAset, err := c.Repo.CreatePeralatan(req)

	if err != nil {

		errorMessage := err.Error()

		// Ruangan / Manager Lab belum tersedia
		if strings.Contains(
			errorMessage,
			"ruangan tidak ditemukan",
		) ||
			strings.Contains(
				errorMessage,
				"ruangan belum terhubung",
			) ||
			strings.Contains(
				errorMessage,
				"laboratorium belum memiliki Manager Lab",
			) {

			return ctx.Status(
				fiber.StatusBadRequest,
			).JSON(fiber.Map{
				"status":  "error",
				"message": errorMessage,
			})
		}

		return ctx.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menyimpan data peralatan",
			"error":   errorMessage,
		})
	}

	// =====================================================
	// AMBIL DATA PERALATAN
	// =====================================================

	peralatan, err :=
		c.Repo.FindByNomorAset(nomorAset)

	if err != nil {
		return ctx.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"status":  "error",
			"message": "Peralatan berhasil dibuat tetapi gagal mengambil ID",
			"error":   err.Error(),
		})
	}

	// =====================================================
	// NOTIFIKASI PIC
	// =====================================================

	c.sendPICPeralatanNotification(
		peralatan,
	)

	// =====================================================
	// RESPONSE
	// =====================================================

	return ctx.Status(
		fiber.StatusCreated,
	).JSON(fiber.Map{
		"status":     "success",
		"message":    "Peralatan beserta detail spesifikasinya berhasil ditambahkan",
		"nomor_aset": nomorAset,
		"id":         peralatan.ID,
		"pic_id":     peralatan.PICID,
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

