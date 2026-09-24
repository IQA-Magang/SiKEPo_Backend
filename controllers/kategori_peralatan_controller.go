package controllers

import (
	"backend/models"
	"backend/repositories"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type KategoriPeralatanController struct {
	Repository *repositories.KategoriPeralatanRepository
}

func NewKategoriPeralatanController(repository *repositories.KategoriPeralatanRepository) *KategoriPeralatanController {
	return &KategoriPeralatanController{Repository: repository}
}

func (c *KategoriPeralatanController) GetAll(ctx *fiber.Ctx) error {
	data, err := c.Repository.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data kategori peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data kategori peralatan berhasil diambil",
		"data":    data,
	})
}

func (c *KategoriPeralatanController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID kategori peralatan tidak valid",
		})
	}

	data, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Kategori peralatan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data kategori peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data kategori peralatan berhasil diambil",
		"data":    data,
	})
}

func (c *KategoriPeralatanController) Create(ctx *fiber.Ctx) error {
	var input struct {
		NamaKategori string `json:"nama_kategori"`
		Description  string `json:"description"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	input.NamaKategori = strings.TrimSpace(input.NamaKategori)
	input.Description = strings.TrimSpace(input.Description)

	if input.NamaKategori == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama kategori wajib diisi",
		})
	}

	exists, err := c.Repository.ExistsByNama(input.NamaKategori)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memeriksa nama kategori",
			"error":   err.Error(),
		})
	}
	if exists {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Nama kategori sudah digunakan",
		})
	}

	data := &models.KategoriPeralatan{
		NamaKategori: input.NamaKategori,
		Description:  input.Description,
	}

	if err := c.Repository.Create(data); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat kategori peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Kategori peralatan berhasil dibuat",
		"data":    data,
	})
}

func (c *KategoriPeralatanController) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID kategori peralatan tidak valid",
		})
	}

	if _, err := c.Repository.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Kategori peralatan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data kategori peralatan",
			"error":   err.Error(),
		})
	}

	var input struct {
		NamaKategori *string `json:"nama_kategori"`
		Description  *string `json:"description"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	updates := map[string]interface{}{}
	if input.NamaKategori != nil {
		trimmed := strings.TrimSpace(*input.NamaKategori)
		if trimmed == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Nama kategori tidak boleh kosong",
			})
		}
		updates["nama_kategori"] = trimmed
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}

	if len(updates) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tidak ada field yang berubah",
		})
	}

	if err := c.Repository.Update(id, updates); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui kategori peralatan",
			"error":   err.Error(),
		})
	}

	updated, err := c.Repository.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Kategori peralatan berhasil diperbarui tetapi gagal mengambil data terbaru",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Kategori peralatan berhasil diperbarui",
		"data":    updated,
	})
}

func (c *KategoriPeralatanController) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID kategori peralatan tidak valid",
		})
	}

	if _, err := c.Repository.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Kategori peralatan tidak ditemukan",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data kategori peralatan",
			"error":   err.Error(),
		})
	}

	if err := c.Repository.Delete(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus kategori peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Kategori peralatan berhasil dihapus",
	})
}
