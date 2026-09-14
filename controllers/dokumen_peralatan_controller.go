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

type DokumenPeralatanController struct {
	Repository *repositories.DokumenPeralatanRepository
}

func NewDokumenPeralatanController(
	repository *repositories.DokumenPeralatanRepository,
) *DokumenPeralatanController {
	return &DokumenPeralatanController{
		Repository: repository,
	}
}

// GetAll
// GET /api/dokumen-peralatan
func (c *DokumenPeralatanController) GetAll(ctx *fiber.Ctx) error {
	dokumen, err := c.Repository.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data dokumen peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data dokumen peralatan berhasil diambil",
		"data":    dokumen,
	})
}

// GetByID
// GET /api/dokumen-peralatan/:id
func (c *DokumenPeralatanController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID dokumen tidak valid",
		})
	}

	dokumen, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Dokumen tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil dokumen",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dokumen berhasil diambil",
		"data":    dokumen,
	})
}

// GetByPeralatanID
// GET /api/dokumen-peralatan/peralatan/:peralatan_id
func (c *DokumenPeralatanController) GetByPeralatanID(ctx *fiber.Ctx) error {
	peralatanID, err := strconv.ParseUint(
		ctx.Params("peralatan_id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID peralatan tidak valid",
		})
	}

	dokumen, err := c.Repository.GetByPeralatanID(peralatanID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil dokumen peralatan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dokumen peralatan berhasil diambil",
		"data":    dokumen,
	})
}

// Create
// POST /api/dokumen-peralatan
func (c *DokumenPeralatanController) Create(ctx *fiber.Ctx) error {
	var input struct {
		NamaDokumen string `json:"nama_dokumen"`
		PathDokumen string `json:"path_dokumen"`
		PeralatanID uint64 `json:"peralatan_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	input.NamaDokumen = strings.TrimSpace(input.NamaDokumen)
	input.PathDokumen = strings.TrimSpace(input.PathDokumen)

	if input.NamaDokumen == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama dokumen wajib diisi",
		})
	}

	if input.PathDokumen == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Path dokumen wajib diisi",
		})
	}

	if input.PeralatanID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Peralatan ID wajib diisi",
		})
	}

	// Cek nama dokumen
	exists, err := c.Repository.ExistsByNamaDokumen(
		input.NamaDokumen,
	)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memeriksa nama dokumen",
			"error":   err.Error(),
		})
	}

	if exists {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Nama dokumen sudah digunakan",
		})
	}

	dokumen := &models.DokumenPeralatan{
		NamaDokumen: input.NamaDokumen,
		PathDokumen: input.PathDokumen,
		PeralatanID: input.PeralatanID,
	}

	if err := c.Repository.Create(dokumen); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat dokumen",
			"error":   err.Error(),
		})
	}

	createdDokumen, err := c.Repository.GetByID(dokumen.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Dokumen berhasil dibuat tetapi gagal mengambil data",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Dokumen berhasil dibuat",
		"data":    createdDokumen,
	})
}

// Update
// PUT /api/dokumen-peralatan/:id
func (c *DokumenPeralatanController) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID dokumen tidak valid",
		})
	}

	dokumen, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Dokumen tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil dokumen",
			"error":   err.Error(),
		})
	}

	var input struct {
		NamaDokumen *string `json:"nama_dokumen"`
		PathDokumen *string `json:"path_dokumen"`
		PeralatanID *uint64 `json:"peralatan_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	updates := make(map[string]interface{})

	if input.NamaDokumen != nil {
		namaDokumen := strings.TrimSpace(*input.NamaDokumen)

		if namaDokumen == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Nama dokumen tidak boleh kosong",
			})
		}

		if namaDokumen != dokumen.NamaDokumen {
			exists, err := c.Repository.ExistsByNamaDokumen(
				namaDokumen,
			)

			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal memeriksa nama dokumen",
					"error":   err.Error(),
				})
			}

			if exists {
				return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Nama dokumen sudah digunakan",
				})
			}
		}

		updates["nama_dokumen"] = namaDokumen
	}

	if input.PathDokumen != nil {
		pathDokumen := strings.TrimSpace(*input.PathDokumen)

		if pathDokumen == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Path dokumen tidak boleh kosong",
			})
		}

		updates["path_dokumen"] = pathDokumen
	}

	if input.PeralatanID != nil {
		if *input.PeralatanID == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Peralatan ID tidak valid",
			})
		}

		updates["peralatan_id"] = *input.PeralatanID
	}

	if len(updates) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tidak ada data yang diubah",
		})
	}

	if err := c.Repository.Update(id, updates); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui dokumen",
			"error":   err.Error(),
		})
	}

	updatedDokumen, err := c.Repository.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Dokumen berhasil diperbarui tetapi gagal mengambil data terbaru",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dokumen berhasil diperbarui",
		"data":    updatedDokumen,
	})
}

// Delete
// DELETE /api/dokumen-peralatan/:id
func (c *DokumenPeralatanController) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID dokumen tidak valid",
		})
	}

	_, err = c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Dokumen tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil dokumen",
			"error":   err.Error(),
		})
	}

	if err := c.Repository.Delete(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus dokumen",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Dokumen berhasil dihapus",
	})
}
