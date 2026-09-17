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

type LabsController struct {
	Repository *repositories.LabsRepository
}

func NewLabsController(
	repository *repositories.LabsRepository,
) *LabsController {
	return &LabsController{
		Repository: repository,
	}
}

// GetAll
// GET /labs
func (c *LabsController) GetAll(ctx *fiber.Ctx) error {
	labs, err := c.Repository.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data labs",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data labs berhasil diambil",
		"data":    labs,
	})
}

// GetByID
// GET /labs/:id
func (c *LabsController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID lab tidak valid",
		})
	}

	labs, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Lab tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data lab",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data lab berhasil diambil",
		"data":    labs,
	})
}

// Create
// POST /labs
func (c *LabsController) Create(ctx *fiber.Ctx) error {
	var input struct {
		NamaLabs  string  `json:"nama_labs"`
		KodeLabs  string  `json:"kode_labs"`
		ManagerID *uint64 `json:"manager_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	input.NamaLabs = strings.TrimSpace(input.NamaLabs)
	input.KodeLabs = strings.TrimSpace(input.KodeLabs)

	// Validasi field wajib
	if input.NamaLabs == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama lab wajib diisi",
		})
	}

	if input.KodeLabs == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Kode lab wajib diisi",
		})
	}

	// Cek kode lab
	exists, err := c.Repository.ExistsByKodeLabs(input.KodeLabs)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memeriksa kode lab",
			"error":   err.Error(),
		})
	}

	if exists {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Kode lab sudah digunakan",
		})
	}

	// Cek nama lab
	exists, err = c.Repository.ExistsByNamaLabs(input.NamaLabs)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memeriksa nama lab",
			"error":   err.Error(),
		})
	}

	if exists {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Nama lab sudah digunakan",
		})
	}

	labs := &models.Labs{
		NamaLabs:  input.NamaLabs,
		KodeLabs:  input.KodeLabs,
		ManagerID: input.ManagerID,
	}

	if err := c.Repository.Create(labs); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat lab",
			"error":   err.Error(),
		})
	}

	// Ambil kembali data beserta Manager
	createdLabs, err := c.Repository.GetByID(labs.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Lab berhasil dibuat tetapi gagal mengambil data",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Lab berhasil dibuat",
		"data":    createdLabs,
	})
}

// Update
// PUT /labs/:id
func (c *LabsController) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID lab tidak valid",
		})
	}

	// Pastikan lab ada
	labs, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Lab tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data lab",
			"error":   err.Error(),
		})
	}

	var input struct {
		NamaLabs  *string `json:"nama_labs"`
		KodeLabs  *string `json:"kode_labs"`
		ManagerID *uint64 `json:"manager_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	updates := make(map[string]interface{})

	// Update nama lab
	if input.NamaLabs != nil {
		namaLabs := strings.TrimSpace(*input.NamaLabs)

		if namaLabs == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Nama lab tidak boleh kosong",
			})
		}

		// Cek jika nama berubah
		if namaLabs != labs.NamaLabs {
			exists, err := c.Repository.ExistsByNamaLabs(namaLabs)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal memeriksa nama lab",
					"error":   err.Error(),
				})
			}

			if exists {
				return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Nama lab sudah digunakan",
				})
			}
		}

		updates["nama_labs"] = namaLabs
	}

	// Update kode lab
	if input.KodeLabs != nil {
		kodeLabs := strings.TrimSpace(*input.KodeLabs)

		if kodeLabs == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Kode lab tidak boleh kosong",
			})
		}

		// Cek jika kode berubah
		if kodeLabs != labs.KodeLabs {
			exists, err := c.Repository.ExistsByKodeLabs(kodeLabs)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal memeriksa kode lab",
					"error":   err.Error(),
				})
			}

			if exists {
				return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Kode lab sudah digunakan",
				})
			}
		}

		updates["kode_labs"] = kodeLabs
	}

	// Update manager
	if input.ManagerID != nil {
		updates["manager_id"] = *input.ManagerID
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
			"message": "Gagal memperbarui lab",
			"error":   err.Error(),
		})
	}

	updatedLabs, err := c.Repository.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Lab berhasil diperbarui tetapi gagal mengambil data terbaru",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Lab berhasil diperbarui",
		"data":    updatedLabs,
	})
}

// Delete
// DELETE /labs/:id
func (c *LabsController) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID lab tidak valid",
		})
	}

	// Pastikan data ada
	_, err = c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Lab tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data lab",
			"error":   err.Error(),
		})
	}

	if err := c.Repository.Delete(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus lab",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Lab berhasil dihapus",
	})
}
