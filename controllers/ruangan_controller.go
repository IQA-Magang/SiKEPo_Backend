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

type RuanganController struct {
	Repository *repositories.RuanganRepository
}

func NewRuanganController(
	repository *repositories.RuanganRepository,
) *RuanganController {
	return &RuanganController{
		Repository: repository,
	}
}

// GetAll
// GET /ruangan
func (c *RuanganController) GetAll(ctx *fiber.Ctx) error {
	ruangan, err := c.Repository.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data ruangan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data ruangan berhasil diambil",
		"data":    ruangan,
	})
}

// GetByID
// GET /ruangan/:id
func (c *RuanganController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID ruangan tidak valid",
		})
	}

	ruangan, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Ruangan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data ruangan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data ruangan berhasil diambil",
		"data":    ruangan,
	})
}

// GetByLabsID
// GET /ruangan/labs/:labs_id
func (c *RuanganController) GetByLabsID(ctx *fiber.Ctx) error {
	labsID, err := strconv.ParseUint(
		ctx.Params("labs_id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Labs ID tidak valid",
		})
	}

	ruangan, err := c.Repository.GetByLabsID(labsID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data ruangan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data ruangan berhasil diambil",
		"data":    ruangan,
	})
}

// GetByPICUserID
// GET /ruangan/pic/:pic_user_id
func (c *RuanganController) GetByPICUserID(ctx *fiber.Ctx) error {
	picUserID, err := strconv.ParseUint(
		ctx.Params("pic_user_id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "PIC user ID tidak valid",
		})
	}

	ruangan, err := c.Repository.GetByPICUserID(picUserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data ruangan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data ruangan berhasil diambil",
		"data":    ruangan,
	})
}

// Create
// POST /ruangan
func (c *RuanganController) Create(ctx *fiber.Ctx) error {
	var input struct {
		NamaRuangan   string  `json:"nama_ruangan"`
		KodeRuangan   string  `json:"kode_ruangan"`
		LantaiRuangan string  `json:"lantai_ruangan"`
		LabsID        *uint64 `json:"labs_id"`
		PICUserID     *uint64 `json:"pic_user_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	input.NamaRuangan = strings.TrimSpace(input.NamaRuangan)
	input.KodeRuangan = strings.TrimSpace(input.KodeRuangan)
	input.LantaiRuangan = strings.TrimSpace(input.LantaiRuangan)

	if input.NamaRuangan == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama ruangan wajib diisi",
		})
	}

	if input.KodeRuangan == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Kode ruangan wajib diisi",
		})
	}

	if input.LantaiRuangan == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Lantai ruangan wajib diisi",
		})
	}

	// Cek kode ruangan
	exists, err := c.Repository.ExistsByKodeRuangan(
		input.KodeRuangan,
	)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memeriksa kode ruangan",
			"error":   err.Error(),
		})
	}

	if exists {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Kode ruangan sudah digunakan",
		})
	}

	// Cek nama ruangan
	exists, err = c.Repository.ExistsByNamaRuangan(
		input.NamaRuangan,
	)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memeriksa nama ruangan",
			"error":   err.Error(),
		})
	}

	if exists {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Nama ruangan sudah digunakan",
		})
	}

	ruangan := &models.Ruangan{
		NamaRuangan:   input.NamaRuangan,
		KodeRuangan:   input.KodeRuangan,
		LantaiRuangan: input.LantaiRuangan,
		LabsID:        input.LabsID,
		PICUserID:     input.PICUserID,
	}

	if err := c.Repository.Create(ruangan); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat ruangan",
			"error":   err.Error(),
		})
	}

	createdRuangan, err := c.Repository.GetByID(ruangan.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Ruangan berhasil dibuat tetapi gagal mengambil data",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Ruangan berhasil dibuat",
		"data":    createdRuangan,
	})
}

// Update
// PUT /ruangan/:id
func (c *RuanganController) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID ruangan tidak valid",
		})
	}

	// Pastikan ruangan ada
	ruangan, err := c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Ruangan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data ruangan",
			"error":   err.Error(),
		})
	}

	var input struct {
		NamaRuangan   *string `json:"nama_ruangan"`
		KodeRuangan   *string `json:"kode_ruangan"`
		LantaiRuangan *string `json:"lantai_ruangan"`
		LabsID        *uint64 `json:"labs_id"`
		PICUserID     *uint64 `json:"pic_user_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	updates := make(map[string]interface{})

	// Update nama ruangan
	if input.NamaRuangan != nil {
		namaRuangan := strings.TrimSpace(*input.NamaRuangan)

		if namaRuangan == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Nama ruangan tidak boleh kosong",
			})
		}

		if namaRuangan != ruangan.NamaRuangan {
			exists, err := c.Repository.ExistsByNamaRuangan(
				namaRuangan,
			)

			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal memeriksa nama ruangan",
					"error":   err.Error(),
				})
			}

			if exists {
				return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Nama ruangan sudah digunakan",
				})
			}
		}

		updates["nama_ruangan"] = namaRuangan
	}

	// Update kode ruangan
	if input.KodeRuangan != nil {
		kodeRuangan := strings.TrimSpace(*input.KodeRuangan)

		if kodeRuangan == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Kode ruangan tidak boleh kosong",
			})
		}

		if kodeRuangan != ruangan.KodeRuangan {
			exists, err := c.Repository.ExistsByKodeRuangan(
				kodeRuangan,
			)

			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal memeriksa kode ruangan",
					"error":   err.Error(),
				})
			}

			if exists {
				return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Kode ruangan sudah digunakan",
				})
			}
		}

		updates["kode_ruangan"] = kodeRuangan
	}

	if input.LantaiRuangan != nil {
		lantaiRuangan := strings.TrimSpace(*input.LantaiRuangan)

		if lantaiRuangan == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Lantai ruangan tidak boleh kosong",
			})
		}

		updates["lantai_ruangan"] = lantaiRuangan
	}

	// Update lab
	if input.LabsID != nil {
		updates["labs_id"] = *input.LabsID
	}

	// Update PIC
	if input.PICUserID != nil {
		updates["pic_user_id"] = *input.PICUserID
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
			"message": "Gagal memperbarui ruangan",
			"error":   err.Error(),
		})
	}

	updatedRuangan, err := c.Repository.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Ruangan berhasil diperbarui tetapi gagal mengambil data terbaru",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Ruangan berhasil diperbarui",
		"data":    updatedRuangan,
	})
}

// Delete
// DELETE /ruangan/:id
func (c *RuanganController) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID ruangan tidak valid",
		})
	}

	// Pastikan data ada
	_, err = c.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Ruangan tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data ruangan",
			"error":   err.Error(),
		})
	}

	if err := c.Repository.Delete(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus ruangan",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Ruangan berhasil dihapus",
	})
}
