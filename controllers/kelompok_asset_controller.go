package controllers

import (
	"errors"
	"strconv"
	"strings"

	"backend/models"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

type KelompokAssetController struct {
	Repository *repositories.KelompokAssetRepository

	LabsRepository *repositories.LabsRepository

	UserRepository *repositories.UserRepository
}

// =====================================================
// REQUEST
// =====================================================

type KelompokAssetRequest struct {
	LabID uint64 `json:"lab_id"`

	PICID uint64 `json:"pic_id"`

	Kode string `json:"kode"`

	Nama string `json:"nama"`
}

// =====================================================
// CREATE
// POST /api/v1/kelompok-asset
// =====================================================

func (ctrl *KelompokAssetController) Create(
	c *fiber.Ctx,
) error {

	var req KelompokAssetRequest

	// =========================
	// PARSE BODY
	// =========================

	if err := c.BodyParser(&req); err != nil {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Format request tidak valid",

			"error": err.Error(),
		})
	}

	// =========================
	// TRIM
	// =========================

	req.Kode = strings.TrimSpace(req.Kode)

	req.Nama = strings.TrimSpace(req.Nama)

	// =========================
	// VALIDATION
	// =========================

	if req.LabID == 0 {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Lab wajib dipilih",
		})
	}

	if req.PICID == 0 {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "PIC wajib dipilih",
		})
	}

	if req.Kode == "" {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Kode wajib diisi",
		})
	}

	if req.Nama == "" {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Nama wajib diisi",
		})
	}

	// =========================
	// CHECK LAB
	// =========================

	_, err := ctrl.LabsRepository.GetByID(req.LabID)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return c.Status(
				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Lab tidak ditemukan",
			})
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal memeriksa Lab",

			"error": err.Error(),
		})
	}

	// =========================
	// CHECK PIC
	// =========================

	_, err = ctrl.UserRepository.GetUserByID(req.PICID)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return c.Status(
				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "PIC tidak ditemukan",
			})
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal memeriksa PIC",

			"error": err.Error(),
		})
	}

	// =========================
	// CHECK DUPLICATE
	// =========================

	exists, err := ctrl.Repository.IsExists(
		req.LabID,
		req.Kode,
	)

	if err != nil {

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal memeriksa data duplicate",

			"error": err.Error(),
		})
	}

	if exists {

		return c.Status(
			fiber.StatusConflict,
		).JSON(fiber.Map{

			"success": false,

			"message": "Kode kelompok asset sudah digunakan pada Lab tersebut",
		})
	}

	// =========================
	// CREATE DATA
	// =========================

	data := &models.KelompokAsset{

		LabID: req.LabID,

		PICID: req.PICID,

		Kode: req.Kode,

		Nama: req.Nama,
	}

	if err := ctrl.Repository.Create(data); err != nil {

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal membuat kelompok asset",

			"error": err.Error(),
		})
	}

	// =========================
	// GET CREATED DATA
	// =========================

	result, err := ctrl.Repository.FindByID(data.ID)

	if err != nil {

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Data berhasil dibuat tetapi gagal mengambil data",

			"error": err.Error(),
		})
	}

	return c.Status(
		fiber.StatusCreated,
	).JSON(fiber.Map{

		"success": true,

		"message": "Kelompok asset berhasil dibuat",

		"data": result,
	})
}

// =====================================================
// GET ALL
// GET /api/v1/kelompok-asset
// =====================================================

func (ctrl *KelompokAssetController) GetAll(
	c *fiber.Ctx,
) error {

	search := strings.TrimSpace(
		c.Query("search"),
	)

	// =========================
	// LAB FILTER
	// =========================

	var labID uint64

	labIDString := c.Query("lab_id")

	if labIDString != "" {

		value, err := strconv.ParseUint(
			labIDString,
			10,
			64,
		)

		if err != nil {

			return c.Status(
				fiber.StatusBadRequest,
			).JSON(fiber.Map{

				"success": false,

				"message": "lab_id tidak valid",
			})
		}

		labID = value
	}

	// =========================
	// PIC FILTER
	// =========================

	var picID uint64

	picIDString := c.Query("pic_id")

	if picIDString != "" {

		value, err := strconv.ParseUint(
			picIDString,
			10,
			64,
		)

		if err != nil {

			return c.Status(
				fiber.StatusBadRequest,
			).JSON(fiber.Map{

				"success": false,

				"message": "pic_id tidak valid",
			})
		}

		picID = value
	}

	// =========================
	// GET DATA
	// =========================

	data, err := ctrl.Repository.FindAll(
		search,
		labID,
		picID,
	)

	if err != nil {

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil data kelompok asset",

			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{

		"success": true,

		"message": "Data kelompok asset berhasil diambil",

		"data": data,
	})
}

// =====================================================
// GET BY ID
// GET /api/v1/kelompok-asset/:id
// =====================================================

func (ctrl *KelompokAssetController) GetByID(
	c *fiber.Ctx,
) error {

	idString := c.Params("id")

	id, err := strconv.ParseUint(
		idString,
		10,
		64,
	)

	if err != nil {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID tidak valid",
		})
	}

	// =========================
	// FIND
	// =========================

	data, err := ctrl.Repository.FindByID(id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return c.Status(
				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Kelompok asset tidak ditemukan",
			})
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil kelompok asset",

			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{

		"success": true,

		"message": "Data kelompok asset berhasil diambil",

		"data": data,
	})
}

// =====================================================
// UPDATE
// PUT /api/v1/kelompok-asset/:id
// =====================================================

func (ctrl *KelompokAssetController) Update(
	c *fiber.Ctx,
) error {

	idString := c.Params("id")

	id, err := strconv.ParseUint(
		idString,
		10,
		64,
	)

	if err != nil {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID tidak valid",
		})
	}

	// =========================
	// FIND EXISTING
	// =========================

	existing, err := ctrl.Repository.FindByID(id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return c.Status(
				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Kelompok asset tidak ditemukan",
			})
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil data",

			"error": err.Error(),
		})
	}

	// =========================
	// PARSE REQUEST
	// =========================

	var req KelompokAssetRequest

	if err := c.BodyParser(&req); err != nil {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Format request tidak valid",

			"error": err.Error(),
		})
	}

	req.Kode = strings.TrimSpace(req.Kode)

	req.Nama = strings.TrimSpace(req.Nama)

	// =========================
	// DEFAULT VALUE
	// =========================

	if req.LabID == 0 {
		req.LabID = existing.LabID
	}

	if req.PICID == 0 {
		req.PICID = existing.PICID
	}

	if req.Kode == "" {
		req.Kode = existing.Kode
	}

	// =========================
	// CHECK LAB
	// =========================

	if req.LabID != existing.LabID {

		_, err = ctrl.LabsRepository.GetByID(
			req.LabID,
		)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				return c.Status(
					fiber.StatusNotFound,
				).JSON(fiber.Map{

					"success": false,

					"message": "Lab tidak ditemukan",
				})
			}

			return c.Status(
				fiber.StatusInternalServerError,
			).JSON(fiber.Map{

				"success": false,

				"message": "Gagal memeriksa Lab",

				"error": err.Error(),
			})
		}
	}

	// =========================
	// CHECK PIC
	// =========================

	if req.PICID != existing.PICID {

		_, err = ctrl.UserRepository.GetUserByID(
			req.PICID,
		)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				return c.Status(
					fiber.StatusNotFound,
				).JSON(fiber.Map{

					"success": false,

					"message": "PIC tidak ditemukan",
				})
			}

			return c.Status(
				fiber.StatusInternalServerError,
			).JSON(fiber.Map{

				"success": false,

				"message": "Gagal memeriksa PIC",

				"error": err.Error(),
			})
		}
	}

	// =========================
	// CHECK DUPLICATE
	// =========================

	if req.LabID != existing.LabID ||
		req.Kode != existing.Kode {

		exists, err := ctrl.Repository.IsExists(
			req.LabID,
			req.Kode,
			id,
		)

		if err != nil {

			return c.Status(
				fiber.StatusInternalServerError,
			).JSON(fiber.Map{

				"success": false,

				"message": "Gagal memeriksa duplicate",

				"error": err.Error(),
			})
		}

		if exists {

			return c.Status(
				fiber.StatusConflict,
			).JSON(fiber.Map{

				"success": false,

				"message": "Kode kelompok asset sudah digunakan pada Lab tersebut",
			})
		}
	}

	// =========================
	// UPDATE DATA
	// =========================

	updates := map[string]interface{}{}

	if req.LabID != existing.LabID {

		updates["lab_id"] = req.LabID
	}

	if req.PICID != existing.PICID {

		updates["pic_id"] = req.PICID
	}

	if req.Kode != existing.Kode {

		updates["kode"] = req.Kode
	}

	if req.Nama != "" &&
		req.Nama != existing.Nama {

		updates["nama"] = req.Nama
	}

	// =========================
	// NO CHANGE
	// =========================

	if len(updates) == 0 {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Tidak ada perubahan data",
		})
	}

	// =========================
	// UPDATE
	// =========================

	if err := ctrl.Repository.Update(
		id,
		updates,
	); err != nil {

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal memperbarui kelompok asset",

			"error": err.Error(),
		})
	}

	// =========================
	// GET UPDATED DATA
	// =========================

	result, err := ctrl.Repository.FindByID(id)

	if err != nil {

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Data berhasil diperbarui tetapi gagal mengambil data",

			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{

		"success": true,

		"message": "Kelompok asset berhasil diperbarui",

		"data": result,
	})
}

// =====================================================
// DELETE
// DELETE /api/v1/kelompok-asset/:id
// =====================================================

func (ctrl *KelompokAssetController) Delete(
	c *fiber.Ctx,
) error {

	idString := c.Params("id")

	id, err := strconv.ParseUint(
		idString,
		10,
		64,
	)

	if err != nil {

		return c.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID tidak valid",
		})
	}

	// =========================
	// CHECK DATA
	// =========================

	_, err = ctrl.Repository.FindByID(id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return c.Status(
				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Kelompok asset tidak ditemukan",
			})
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil data",

			"error": err.Error(),
		})
	}

	// =========================
	// DELETE
	// =========================

	if err := ctrl.Repository.Delete(id); err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			return c.Status(
				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Kelompok asset tidak ditemukan",
			})
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal menghapus kelompok asset",

			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{

		"success": true,

		"message": "Kelompok asset berhasil dihapus",
	})
}