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

type LogbookController struct {
	Repository          *repositories.LogbookRepository
	PeralatanRepository repositories.PeralatanRepository
}

func NewLogbookController(
	repository *repositories.LogbookRepository,
	peralatanRepository repositories.PeralatanRepository,
) *LogbookController {

	return &LogbookController{
		Repository:          repository,
		PeralatanRepository: peralatanRepository,
	}
}

// logbookUser hanya memuat id dan nama pelaku.
type logbookUser struct {
	UserID uint64 `json:"user_id"`
	Name   string `json:"name"`
}

// logbookItem adalah satu baris logbook pada response.
type logbookItem struct {
	ID          uint64       `json:"id"`
	Jenis       string       `json:"jenis"`
	Aksi        string       `json:"aksi"`
	Judul       string       `json:"judul"`
	Status      string       `json:"status"`
	Keterangan  string       `json:"keterangan"`
	ReferensiID *uint64      `json:"referensi_id"`
	User        *logbookUser `json:"user"`
	CreatedAt   time.Time    `json:"created_at"`
}

// =====================================================
// GET BY PERALATAN
// GET /api/peralatan/:id/logbook?jenis=peminjaman
// =====================================================

func (c *LogbookController) GetByPeralatan(ctx *fiber.Ctx) error {

	peralatanID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || peralatanID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID peralatan tidak valid",
		})
	}

	jenis := strings.ToLower(strings.TrimSpace(ctx.Query("jenis")))
	if jenis != "" && !models.IsValidLogbookJenis(jenis) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "jenis tidak valid",
		})
	}

	if _, err := c.PeralatanRepository.FindByID(uint(peralatanID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Peralatan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data peralatan",
			"error":   err.Error(),
		})
	}

	logs, err := c.Repository.FindByPeralatan(peralatanID, jenis)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil logbook peralatan",
			"error":   err.Error(),
		})
	}

	// Slice dibuat dengan make agar hasil kosong menjadi [] dan bukan null.
	items := make([]logbookItem, 0, len(logs))

	for _, entry := range logs {
		item := logbookItem{
			ID:          entry.ID,
			Jenis:       entry.Jenis,
			Aksi:        entry.Aksi,
			Judul:       models.LogbookJudul(entry.Aksi),
			Status:      entry.Status,
			Keterangan:  entry.Keterangan,
			ReferensiID: entry.ReferensiID,
			CreatedAt:   entry.CreatedAt,
		}

		if entry.User != nil {
			item.User = &logbookUser{
				UserID: entry.User.UserID,
				Name:   entry.User.Name,
			}
		}

		items = append(items, item)
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    items,
	})
}
