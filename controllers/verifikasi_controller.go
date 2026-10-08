package controllers

import (
	"errors"

	"fmt"

	"strconv"

	"strings"

	"time"

	"backend/models"

	"backend/repositories"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

// =====================================================

// CONTROLLER

// =====================================================

type VerifikasiController struct {
	Repository *repositories.VerifikasiRepository

	PeralatanRepository repositories.PeralatanRepository

	LogPeninjauanRepository *repositories.LogPeninjauanRepository

	NotificationRepository repositories.NotificationRepository
}

// =====================================================

// CONSTRUCTOR

// =====================================================

func NewVerifikasiController(

	verifikasiRepository *repositories.VerifikasiRepository,

	peralatanRepository repositories.PeralatanRepository,

	logRepository *repositories.LogPeninjauanRepository,

	notificationRepository repositories.NotificationRepository,

) *VerifikasiController {

	return &VerifikasiController{

		Repository: verifikasiRepository,

		PeralatanRepository: peralatanRepository,

		LogPeninjauanRepository: logRepository,

		NotificationRepository: notificationRepository,
	}

}

// =====================================================

// REQUEST CREATE

// =====================================================

type CreateVerifikasiRequest struct {
	IDPeralatan       uint64  `json:"id_peralatan"`
	TanggalVerifikasi string  `json:"tanggal_verifikasi"`
	IDKriteria        *uint64 `json:"id_kriteria"`
	TindakLanjut      string  `json:"tindak_lanjut"`
	Catatan           string  `json:"catatan"`

	HasilVerifikasi struct {
		Identitas     string `json:"identitas"`
		Kelengkapan   string `json:"kelengkapan"`
		Firmware      string `json:"firmware"`
		KondisiFisik  string `json:"kondisi_fisik"`
		Segel         string `json:"segel"`
		FungsiAwal    string `json:"fungsi_awal"`
		Metrologi     string `json:"metrologi"`
		Sertifikat    string `json:"sertifikat"`
		Catatan       string `json:"catatan"`
	} `json:"hasil_verifikasi"`
}

// =====================================================

// SIGNATURE

// =====================================================

type SignatureRequest struct {
	Signature string `json:"signature"`
}

// =====================================================

// REJECT

// =====================================================

type RejectVerifikasiRequest struct {
	Alasan string `json:"alasan"`

	Catatan string `json:"catatan"`
}

// =====================================================

// AUTH USER

// =====================================================

func getAuthenticatedUserID(

	ctx *fiber.Ctx,

) (uint64, error) {

	value := ctx.Locals("user_id")

	if value == nil {

		return 0, errors.New(

			"user tidak terautentikasi",
		)

	}

	switch value := value.(type) {

	case uint64:

		return value, nil

	case uint:

		return uint64(value), nil

	case uint32:

		return uint64(value), nil

	case int:

		return uint64(value), nil

	case int64:

		return uint64(value), nil

	case float64:

		return uint64(value), nil

	case string:

		id, err := strconv.ParseUint(

			value,

			10,

			64,
		)

		if err != nil {

			return 0, errors.New(

				"user ID tidak valid",
			)

		}

		return id, nil

	default:

		return 0, errors.New(

			"user ID tidak valid",
		)

	}

}

// =====================================================

// CEK AKSES ADMIN / MANAGER

// =====================================================

func (c *VerifikasiController) isAdminOrManager(
	ctx *fiber.Ctx,
) bool {
	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return false
	}

	var user models.User
	if err := c.Repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return false
	}
	return user.Role == "admin" || user.Role == "manager"
}

func (c *VerifikasiController) getAuthenticatedUser(
	ctx *fiber.Ctx,
) (*models.User, error) {
	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := c.Repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Staff hanya dapat mengakses data verifikasi dari Lab yang sama.
func (c *VerifikasiController) canAccessVerifikasi(
	ctx *fiber.Ctx,
	data *models.Verifikasi,
) bool {
	if data == nil {
		return false
	}
	if c.isAdminOrManager(ctx) {
		return true
	}

	user, err := c.getAuthenticatedUser(ctx)
	if err != nil || user.Role != "staff" || !user.Pengelola || user.LabsID == nil || data.IDPeralatan == 0 {
		return false
	}
	return c.isPeralatanInSameLab(data.IDPeralatan, *user.LabsID)
}

// Staff hanya dapat mengakses peralatan dari Lab miliknya.
func (c *VerifikasiController) canAccessPeralatan(
	ctx *fiber.Ctx,
	peralatan *models.Peralatan,
) bool {
	if peralatan == nil {
		return false
	}
	if c.isAdminOrManager(ctx) {
		return true
	}

	user, err := c.getAuthenticatedUser(ctx)
	if err != nil || user.Role != "staff" || !user.Pengelola || user.LabsID == nil {
		return false
	}
	return c.isPeralatanInSameLab(uint64(peralatan.ID), *user.LabsID)
}

func (c *VerifikasiController) canVerifyAsLabPengelola(
	ctx *fiber.Ctx,
	peralatanID uint64,
) bool {
	user, err := c.getAuthenticatedUser(ctx)
	if err != nil || user.Role != "staff" || !user.Pengelola || user.LabsID == nil {
		return false
	}
	return c.isPeralatanInSameLab(peralatanID, *user.LabsID)
}

func (c *VerifikasiController) isPeralatanInSameLab(
	peralatanID uint64,
	labsID uint64,
) bool {
	if peralatanID == 0 || labsID == 0 {
		return false
	}

	var count int64
	err := c.Repository.DB.
		Table("peralatan AS p").
		Joins("INNER JOIN ruangan AS r ON r.id = p.ruangan_id").
		Where("p.id = ?", peralatanID).
		Where("r.labs_id = ?", labsID).
		Count(&count).Error

	return err == nil && count > 0
}

func parseTanggal(

	value string,

) (time.Time, error) {

	value = strings.TrimSpace(value)

	layouts := []string{

		"2006-01-02",

		"2006-01-02 15:04:05",

		"2006-01-02T15:04",

		"2006-01-02T15:04:05",

		time.RFC3339,

		time.RFC3339Nano,
	}

	for _, layout := range layouts {

		var (
			result time.Time

			err error
		)

		if layout == time.RFC3339 ||

			layout == time.RFC3339Nano {

			result, err = time.Parse(

				layout,

				value,
			)

		} else {

			result, err = time.ParseInLocation(

				layout,

				value,

				time.Local,
			)

		}

		if err == nil {

			return result, nil

		}

	}

	return time.Time{}, errors.New(

		"format tanggal tidak dikenali",
	)

}

// =====================================================

// VALIDASI HASIL

// =====================================================

func isValidHasilVerifikasi(

	value string,

) bool {

	value = strings.TrimSpace(value)

	return value == "S" ||

		value == "TS" ||

		value == "TB"

}

// =====================================================

// GET ALL

// =====================================================

// =====================================================
// GET ALL / RIWAYAT VERIFIKASI
// =====================================================

func (c *VerifikasiController) GetVerifikasi(
	ctx *fiber.Ctx,
) error {

	user, err := c.getAuthenticatedUser(ctx)

	if err != nil {
		return ctx.Status(
			fiber.StatusUnauthorized,
		).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// =====================================================
	// QUERY DASAR
	// =====================================================

	query := c.Repository.DB.
		Preload("Peralatan").
		Preload("PICUser").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("verifikasi.deleted_at IS NULL")

	// =====================================================
	// FILTER BERDASARKAN LAB
	// =====================================================
	//
	// Admin
	//   -> Melihat semua Lab
	//
	// Manager
	//   -> Hanya Lab sesuai LabsID
	//
	// Staff
	//   -> Hanya Lab sesuai LabsID
	//
	// =====================================================

	if user.Role == "staff" || user.Role == "manager" {

		if user.LabsID == nil {
			return ctx.Status(
				fiber.StatusForbidden,
			).JSON(fiber.Map{
				"success": false,
				"message": "User belum memiliki Lab",
			})
		}

		query = query.Where(`
			EXISTS (
				SELECT 1
				FROM peralatan AS p
				INNER JOIN ruangan AS r
					ON r.id = p.ruangan_id
				WHERE p.id = verifikasi.id_peralatan
				AND r.labs_id = ?
			)
		`, *user.LabsID)

	} else if user.Role != "admin" {

		return ctx.Status(
			fiber.StatusForbidden,
		).JSON(fiber.Map{
			"success": false,
			"message": "Role tidak memiliki akses ke verifikasi",
		})
	}

	// =====================================================
	// AMBIL DATA
	// =====================================================

	var data []models.Verifikasi

	if err := query.
		Order("verifikasi.id_verifikasi DESC").
		Find(&data).
		Error; err != nil {

		return ctx.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data verifikasi",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// =====================================================
// GET PENGAJUAN MANAGER
// =====================================================

func (c *VerifikasiController) GetPengajuan(
	ctx *fiber.Ctx,
) error {

	user, err := c.getAuthenticatedUser(ctx)

	if err != nil {
		return ctx.Status(
			fiber.StatusUnauthorized,
		).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// =====================================================
	// QUERY DASAR
	// =====================================================

	query := c.Repository.DB.
		Preload("Peralatan").
		Preload("PICUser").
		Preload("HasilVerifikasi").
		Where("verifikasi.status = ?", "Diajukan").
		Where("verifikasi.deleted_at IS NULL")

	// =====================================================
	// FILTER BERDASARKAN LAB
	// =====================================================
	//
	// Admin
	//   -> Semua Lab
	//
	// Manager
	//   -> Lab sesuai LabsID
	//
	// Staff
	//   -> Lab sesuai LabsID
	//
	// =====================================================

	if user.Role == "staff" || user.Role == "manager" {

		if user.LabsID == nil {
			return ctx.Status(
				fiber.StatusForbidden,
			).JSON(fiber.Map{
				"success": false,
				"message": "User belum memiliki Lab",
			})
		}

		query = query.Where(`
			EXISTS (
				SELECT 1
				FROM peralatan AS p
				INNER JOIN ruangan AS r
					ON r.id = p.ruangan_id
				WHERE p.id = verifikasi.id_peralatan
				AND r.labs_id = ?
			)
		`, *user.LabsID)

	} else if user.Role != "admin" {

		return ctx.Status(
			fiber.StatusForbidden,
		).JSON(fiber.Map{
			"success": false,
			"message": "Role tidak memiliki akses ke pengajuan verifikasi",
		})
	}

	// =====================================================
	// AMBIL DATA
	// =====================================================

	var data []models.Verifikasi

	if err := query.
		Order("verifikasi.pic_signed_at DESC").
		Find(&data).
		Error; err != nil {

		return ctx.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil pengajuan",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func (c *VerifikasiController) GetVerifikasiByID(

	ctx *fiber.Ctx,

) error {

	id, err := strconv.ParseUint(

		ctx.Params("id"),

		10,

		64,
	)

	if err != nil || id == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID verifikasi tidak valid",
		})

	}

	data, err :=

		c.Repository.GetVerifikasiByID(id)

	if err != nil {

		if errors.Is(

			err,

			gorm.ErrRecordNotFound,
		) {

			return ctx.Status(

				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Verifikasi tidak ditemukan",
			})

		}

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil verifikasi",

			"error": err.Error(),
		})

	}

	if !c.canAccessVerifikasi(ctx, data) {

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{

			"success": false,

			"message": "Anda tidak memiliki akses ke verifikasi ini",
		})

	}

	return ctx.JSON(fiber.Map{

		"success": true,

		"data": data,
	})

}

// =====================================================

// GET BY PERALATAN

// =====================================================

func (c *VerifikasiController) GetByPeralatan(

	ctx *fiber.Ctx,

) error {

	peralatanID, err := strconv.ParseUint(

		ctx.Params("peralatan_id"),

		10,

		64,
	)

	if err != nil || peralatanID == 0 {

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{

			"success": false,

			"message": "ID peralatan tidak valid",
		})

	}

	peralatan, err := c.PeralatanRepository.FindByID(uint(peralatanID))

	if err != nil || peralatan == nil {

		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{

			"success": false,

			"message": "Peralatan tidak ditemukan",
		})

	}

	if !c.canAccessPeralatan(ctx, peralatan) {

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{

			"success": false,

			"message": "Anda tidak memiliki akses ke peralatan dari Lab ini",
		})

	}

	data, err := c.Repository.GetByPeralatanID(peralatanID)

	if err != nil {

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil riwayat verifikasi",

			"error": err.Error(),
		})

	}

	return ctx.JSON(fiber.Map{

		"success": true,

		"data": data,
	})

}

// =====================================================

// CREATE DRAFT

// =====================================================

func (c *VerifikasiController) CreateVerifikasi(

	ctx *fiber.Ctx,

) error {

	userID, err :=

		getAuthenticatedUserID(ctx)

	if err != nil {

		return ctx.Status(

			fiber.StatusUnauthorized,
		).JSON(fiber.Map{

			"success": false,

			"message": err.Error(),
		})

	}

	var request CreateVerifikasiRequest

	if err := ctx.BodyParser(

		&request,
	); err != nil {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Format JSON tidak valid",

			"error": err.Error(),
		})

	}

	if request.IDPeralatan == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID peralatan wajib diisi",
		})

	}

	peralatan, err :=

		c.PeralatanRepository.FindByID(

			uint(request.IDPeralatan),
		)

	if err != nil ||

		peralatan == nil {

		return ctx.Status(

			fiber.StatusNotFound,
		).JSON(fiber.Map{

			"success": false,

			"message": "Peralatan tidak ditemukan",
		})

	}

	if !c.canVerifyAsLabPengelola(ctx, uint64(peralatan.ID)) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya pengelola staff dari Lab peralatan yang dapat membuat verifikasi",
		})
	}

	// =====================================================
	// KODE AKTIVITAS OTOMATIS
	// =====================================================

	kodeAktivitas := strings.TrimSpace(
		peralatan.KodeAktivitas,
	)

	if kodeAktivitas == "" {
		kodeAktivitas = "A1"
	}

	tanggal := time.Now()

	if request.TanggalVerifikasi != "" {

		tanggal, err =

			parseTanggal(

				request.TanggalVerifikasi,
			)

		if err != nil {

			return ctx.Status(

				fiber.StatusBadRequest,
			).JSON(fiber.Map{

				"success": false,

				"message": err.Error(),
			})

		}

	}

	hasil := request.HasilVerifikasi

	values := []string{

		hasil.Identitas,

		hasil.Kelengkapan,

		hasil.Firmware,

		hasil.KondisiFisik,

		hasil.Segel,

		hasil.FungsiAwal,

		hasil.Metrologi,

		hasil.Sertifikat,
	}

	for _, value := range values {

		if !isValidHasilVerifikasi(value) {

			return ctx.Status(

				fiber.StatusBadRequest,
			).JSON(fiber.Map{

				"success": false,

				"message": "Nilai hasil verifikasi hanya boleh S, TS, atau TB",

				"nilai": value,
			})

		}

	}

	var verifikasi models.Verifikasi

	err = c.Repository.DB.Transaction(

		func(tx *gorm.DB) error {

			verifikasi = models.Verifikasi{
			IDPeralatan:       request.IDPeralatan,
			TanggalVerifikasi: tanggal,
			KodeAktivitas:     kodeAktivitas,
			IDKriteria:        request.IDKriteria,
			Status:            "Draft",
			TindakLanjut:      request.TindakLanjut,
			PICID:             &userID,
			Catatan:           request.Catatan,
		}

			if err :=

				c.Repository.CreateVerifikasi(

					tx,

					&verifikasi,
				); err != nil {

				return err

			}

			hasilData :=

				&models.HasilVerifikasi{

					IDVerifikasi: verifikasi.IDVerifikasi,

					Identitas: hasil.Identitas,

					Kelengkapan: hasil.Kelengkapan,

					Firmware: hasil.Firmware,

					KondisiFisik: hasil.KondisiFisik,

					Segel: hasil.Segel,

					FungsiAwal: hasil.FungsiAwal,

					Metrologi: hasil.Metrologi,

					Sertifikat: hasil.Sertifikat,

					Catatan: hasil.Catatan,
				}

			if err :=

				c.Repository.CreateHasilVerifikasi(

					tx,

					hasilData,
				); err != nil {

				return err

			}

			return tx.
				Model(&models.Peralatan{}).
				Where(

					"id = ?",

					request.IDPeralatan,
				).
				Updates(map[string]interface{}{

					"status_alat": "Karantina",

					"status_verifikasi": "Draft",
				}).
				Error

		},
	)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal membuat draft verifikasi",

			"error": err.Error(),
		})

	}

	result, err :=

		c.Repository.GetVerifikasiByID(

			verifikasi.IDVerifikasi,
		)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Draft berhasil dibuat tetapi gagal mengambil data",

			"error": err.Error(),
		})

	}

	return ctx.Status(

		fiber.StatusCreated,
	).JSON(fiber.Map{

		"success": true,

		"message": "Draft verifikasi berhasil dibuat",

		"data": result,
	})

}

// =====================================================

// SIGN PIC + NOTIFICATION MANAGER

// =====================================================

func (c *VerifikasiController) SignPIC(

	ctx *fiber.Ctx,

) error {

	userID, err :=

		getAuthenticatedUserID(ctx)

	if err != nil {

		return ctx.Status(

			fiber.StatusUnauthorized,
		).JSON(fiber.Map{

			"success": false,

			"message": err.Error(),
		})

	}

	id, err := strconv.ParseUint(

		ctx.Params("id"),

		10,

		64,
	)

	if err != nil || id == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID verifikasi tidak valid",
		})

	}

	var request SignatureRequest

	if err := ctx.BodyParser(

		&request,
	); err != nil {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Format signature tidak valid",
		})

	}

	request.Signature =

		strings.TrimSpace(

			request.Signature,
		)

	if request.Signature == "" {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Tanda tangan PIC wajib diisi",
		})

	}

	data, err :=

		c.Repository.GetVerifikasiByID(id)

	if err != nil {

		if errors.Is(

			err,

			gorm.ErrRecordNotFound,
		) {

			return ctx.Status(

				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Verifikasi tidak ditemukan",
			})

		}

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil verifikasi",

			"error": err.Error(),
		})

	}

	if data.Peralatan == nil || !c.canVerifyAsLabPengelola(ctx, uint64(data.IDPeralatan)) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya pengelola staff dari Lab peralatan yang dapat menandatangani",
		})
	}

	if data.Status != "Draft" {

		return ctx.Status(

			fiber.StatusConflict,
		).JSON(fiber.Map{

			"success": false,

			"message": "Verifikasi sudah diajukan atau diproses",
		})

	}

	err = c.Repository.DB.Transaction(

		func(tx *gorm.DB) error {

			if err :=

				c.Repository.SignPIC(

					tx,

					id,

					userID,

					request.Signature,
				); err != nil {

				return err

			}

			if err :=

				tx.
					Model(&models.Peralatan{}).
					Where(

						"id = ?",

						data.IDPeralatan,
					).
					Update(

						"status_verifikasi",

						"Diajukan",
					).
					Error; err != nil {

				return err

			}

			// =================================================

			// AMBIL MANAGER LAB

			// =================================================

			var ruangan models.Ruangan

			if err :=

				tx.
					Preload("Labs").
					First(

						&ruangan,
					).Error; err != nil {

				return err

			}

			if ruangan.Labs == nil ||

				ruangan.Labs.ManagerID == nil {

				return nil

			}

			managerID :=

				*ruangan.Labs.ManagerID

			// =================================================

			// NOTIFIKASI MANAGER

			// =================================================

			notification :=

				&models.Notification{

					UserID: managerID,

					Type: "verification_submitted",

					Title: "Verifikasi peralatan diajukan",

					Message: fmt.Sprintf(

						"Verifikasi peralatan %s (%s) telah ditandatangani PIC dan diajukan untuk ditinjau.",

						data.Peralatan.NamaPeralatan,

						data.Peralatan.NomorAset,
					),
				}

			return c.NotificationRepository.Create(

				notification,
			)

		},
	)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengajukan verifikasi",

			"error": err.Error(),
		})

	}

	result, err :=

		c.Repository.GetVerifikasiByID(id)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Verifikasi berhasil diajukan tetapi gagal mengambil data",

			"error": err.Error(),
		})

	}

	return ctx.JSON(fiber.Map{

		"success": true,

		"message": "Verifikasi berhasil ditandatangani PIC dan diajukan kepada Manager",

		"data": result,
	})

}

// =====================================================

// APPROVE MANAGER

// =====================================================

func (c *VerifikasiController) ApproveVerifikasi(

	ctx *fiber.Ctx,

) error {

	if !c.isAdminOrManager(ctx) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya Admin atau Manager yang dapat menyetujui verifikasi",
		})
	}

	managerID, err :=

		getAuthenticatedUserID(ctx)

	if err != nil {

		return ctx.Status(

			fiber.StatusUnauthorized,
		).JSON(fiber.Map{

			"success": false,

			"message": err.Error(),
		})

	}

	id, err := strconv.ParseUint(

		ctx.Params("id"),

		10,

		64,
	)

	if err != nil || id == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID verifikasi tidak valid",
		})

	}

	var request SignatureRequest

	if err := ctx.BodyParser(

		&request,
	); err != nil {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Format signature tidak valid",
		})

	}

	request.Signature =

		strings.TrimSpace(

			request.Signature,
		)

	if request.Signature == "" {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Tanda tangan Manager wajib diisi",
		})

	}

	data, err :=

		c.Repository.GetVerifikasiByID(id)

	if err != nil {

		if errors.Is(

			err,

			gorm.ErrRecordNotFound,
		) {

			return ctx.Status(

				fiber.StatusNotFound,
			).JSON(fiber.Map{

				"success": false,

				"message": "Verifikasi tidak ditemukan",
			})

		}

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil verifikasi",

			"error": err.Error(),
		})

	}

	if data.Status != "Diajukan" {

		return ctx.Status(

			fiber.StatusConflict,
		).JSON(fiber.Map{

			"success": false,

			"message": "Verifikasi belum diajukan oleh PIC",
		})

	}

	if data.PICSignature == "" {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "PIC belum menandatangani verifikasi",
		})

	}

	if len(data.HasilVerifikasi) == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Hasil verifikasi belum tersedia",
		})

	}

	hasil := data.HasilVerifikasi[0]

	values := []string{

		hasil.Identitas,

		hasil.Kelengkapan,

		hasil.Firmware,

		hasil.KondisiFisik,

		hasil.Segel,

		hasil.FungsiAwal,

		hasil.Metrologi,

		hasil.Sertifikat,
	}

	for _, value := range values {

		if value == "TS" {

			return ctx.Status(

				fiber.StatusBadRequest,
			).JSON(fiber.Map{

				"success": false,

				"message": "Verifikasi memiliki hasil TS sehingga belum dapat disetujui",
			})

		}

	}

	err = c.Repository.DB.Transaction(

		func(tx *gorm.DB) error {

			if err :=

				c.Repository.ApproveVerifikasi(

					tx,

					id,

					managerID,

					request.Signature,
				); err != nil {

				return err

			}

			return tx.
				Model(&models.Peralatan{}).
				Where(

					"id = ?",

					data.IDPeralatan,
				).
				Updates(map[string]interface{}{

					"status_alat": "Aktif",

					"status_verifikasi": "Disetujui",
				}).
				Error

		},
	)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal menyetujui verifikasi",

			"error": err.Error(),
		})

	}

	result, err :=

		c.Repository.GetVerifikasiByID(id)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Verifikasi berhasil disetujui tetapi gagal mengambil data",

			"error": err.Error(),
		})

	}

	return ctx.JSON(fiber.Map{

		"success": true,

		"message": "Verifikasi disetujui dan peralatan masuk inventaris",

		"data": result,
	})

}

// =====================================================

// REJECT MANAGER

// =====================================================

func (c *VerifikasiController) RejectVerifikasi(

	ctx *fiber.Ctx,

) error {

	if !c.isAdminOrManager(ctx) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya Admin atau Manager yang dapat menolak verifikasi",
		})
	}

	managerID, err :=

		getAuthenticatedUserID(ctx)

	if err != nil {

		return ctx.Status(

			fiber.StatusUnauthorized,
		).JSON(fiber.Map{

			"success": false,

			"message": err.Error(),
		})

	}

	id, err := strconv.ParseUint(

		ctx.Params("id"),

		10,

		64,
	)

	if err != nil || id == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID verifikasi tidak valid",
		})

	}

	var request RejectVerifikasiRequest

	if err := ctx.BodyParser(

		&request,
	); err != nil {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Format request tidak valid",
		})

	}

	request.Alasan =

		strings.TrimSpace(

			request.Alasan,
		)

	if request.Alasan == "" {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "Alasan penolakan wajib diisi",
		})

	}

	data, err :=

		c.Repository.GetVerifikasiByID(id)

	if err != nil {

		return ctx.Status(

			fiber.StatusNotFound,
		).JSON(fiber.Map{

			"success": false,

			"message": "Verifikasi tidak ditemukan",
		})

	}

	if data.Status != "Diajukan" {

		return ctx.Status(

			fiber.StatusConflict,
		).JSON(fiber.Map{

			"success": false,

			"message": "Verifikasi belum diajukan atau sudah diproses",
		})

	}

	err = c.Repository.DB.Transaction(

		func(tx *gorm.DB) error {

			if err :=

				c.Repository.RejectVerifikasi(

					tx,

					id,

					managerID,

					request.Catatan,
				); err != nil {

				return err

			}

			if err :=

				tx.
					Model(&models.Peralatan{}).
					Where(

						"id = ?",

						data.IDPeralatan,
					).
					Updates(map[string]interface{}{

						"status_alat": "Karantina",

						"status_verifikasi": "Ditolak",
					}).
					Error; err != nil {

				return err

			}

			logData :=

				&models.LogPeninjauanPeralatan{

					IDPeralatan: data.IDPeralatan,

					IDVerifikasi: data.IDVerifikasi,

					IDManager: managerID,

					Status: "Ditolak",

					Alasan: request.Alasan,

					Catatan: request.Catatan,
				}

			return c.LogPeninjauanRepository.Create(

				tx,

				logData,
			)

		},
	)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal menolak verifikasi",

			"error": err.Error(),
		})

	}

	result, _ :=

		c.Repository.GetVerifikasiByID(id)

	return ctx.JSON(fiber.Map{

		"success": true,

		"message": "Verifikasi ditolak dan masuk log peninjauan",

		"data": result,
	})

}

// =====================================================
// GET LOG
// =====================================================

func (c *VerifikasiController) GetLogPeninjauan(
	ctx *fiber.Ctx,
) error {
	user, err := c.getAuthenticatedUser(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	var data []models.LogPeninjauanPeralatan

	query := c.Repository.DB.
		Model(&models.LogPeninjauanPeralatan{})

	// =====================================================
	// FILTER BERDASARKAN ROLE
	// =====================================================

	if user.Role == "staff" {

		// Staff wajib memiliki Lab
		if user.LabsID == nil {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Staff belum memiliki Lab",
			})
		}

		// Staff hanya dapat melihat log
		// peralatan yang berada di Lab miliknya
		query = query.Where(
			"id_peralatan IN (?)",
			c.Repository.DB.
				Table("peralatan AS p").
				Select("p.id").
				Joins(
					"INNER JOIN ruangan AS r ON r.id = p.ruangan_id",
				).
				Where(
					"r.labs_id = ?",
					*user.LabsID,
				),
		)

	} else if user.Role != "admin" && user.Role != "manager" {

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Role tidak memiliki akses ke log peninjauan",
		})
	}

	// =====================================================
	// AMBIL DATA LOG
	// =====================================================
	//
	// Primary key tabel log_peninjauan_peralatan:
	// id_log
	//
	// BUKAN:
	// id
	//

	if err := query.
		Order("id_log DESC").
		Find(&data).Error; err != nil {

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil log peninjauan",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func (c *VerifikasiController) GetLogByPeralatan(

	ctx *fiber.Ctx,

) error {

	peralatanID, err := strconv.ParseUint(

		ctx.Params("peralatan_id"),

		10,

		64,
	)

	if err != nil || peralatanID == 0 {

		return ctx.Status(

			fiber.StatusBadRequest,
		).JSON(fiber.Map{

			"success": false,

			"message": "ID peralatan tidak valid",
		})

	}

	peralatan, err := c.PeralatanRepository.FindByID(uint(peralatanID))

	if err != nil || peralatan == nil {

		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{

			"success": false,

			"message": "Peralatan tidak ditemukan",
		})

	}

	if !c.canAccessPeralatan(ctx, peralatan) {

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{

			"success": false,

			"message": "Anda tidak memiliki akses ke peralatan dari Lab ini",
		})

	}

	data, err :=

		c.LogPeninjauanRepository.GetByPeralatan(

			peralatanID,
		)

	if err != nil {

		return ctx.Status(

			fiber.StatusInternalServerError,
		).JSON(fiber.Map{

			"success": false,

			"message": "Gagal mengambil log peninjauan",

			"error": err.Error(),
		})

	}

	return ctx.JSON(fiber.Map{

		"success": true,

		"data": data,
	})

}

// =====================================================

// DELETE

// =====================================================

func (c *VerifikasiController) DeleteVerifikasi(
	ctx *fiber.Ctx,
) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID verifikasi tidak valid"})
	}

	data, err := c.Repository.GetVerifikasiByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Verifikasi tidak ditemukan"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengambil verifikasi", "error": err.Error()})
	}

	if !c.canAccessVerifikasi(ctx, data) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Anda tidak memiliki akses ke verifikasi dari Lab ini"})
	}

	if err := c.Repository.DeleteVerifikasi(id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return ctx.JSON(fiber.Map{"success": true, "message": "Verifikasi berhasil dihapus"})
}

// =====================================================
// GET HISTORI VERIFIKASI PERALATAN
// =====================================================
//
// Semua user yang sudah login dapat melihat histori
// verifikasi suatu peralatan.
//
// Tidak menggunakan filter LabsID.
// Tidak menggunakan canAccessPeralatan().
//
// =====================================================

func (c *VerifikasiController) GetHistoriVerifikasi(
	ctx *fiber.Ctx,
) error {

	// =====================================================
	// PARSE ID PERALATAN
	// =====================================================

	peralatanID, err := strconv.ParseUint(
		ctx.Params("peralatan_id"),
		10,
		64,
	)

	if err != nil || peralatanID == 0 {

		return ctx.Status(
			fiber.StatusBadRequest,
		).JSON(fiber.Map{
			"success": false,
			"message": "ID peralatan tidak valid",
		})
	}

	// =====================================================
	// CEK PERALATAN
	// =====================================================

	peralatan, err :=
		c.PeralatanRepository.FindByID(
			uint(peralatanID),
		)
	if err != nil {
		return ctx.Status(
			fiber.StatusNotFound,
		).JSON(fiber.Map{
			"success": false,
			"message": "Peralatan tidak ditemukan",
		})
	}

	// =====================================================
	// AMBIL HISTORI
	// =====================================================

	data, total, err :=
		c.Repository.GetHistoriByPeralatan(
			peralatanID,
		)

	if err != nil {

		return ctx.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil histori verifikasi",
			"error":   err.Error(),
		})
	}

	// =====================================================
	// RESPONSE
	// =====================================================

	return ctx.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"peralatan":        peralatan,
			"total_verifikasi": total,
			"histori":          data,
		},
	})
}