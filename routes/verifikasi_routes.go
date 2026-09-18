package routes

import (
	"backend/controllers"
	"backend/middleware"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func VerifikasiRoutes(
	app *fiber.App,
	db *gorm.DB,
	notificationRepo repositories.NotificationRepository,
) {

	verifikasiRepository :=
		repositories.NewVerifikasiRepository(db)

	peralatanRepository :=
		repositories.NewPeralatanRepository(db)

	logRepository :=
		repositories.NewLogPeninjauanRepository(db)

	controller :=
		controllers.NewVerifikasiController(
			verifikasiRepository,
			peralatanRepository,
			logRepository,
			notificationRepo,
		)

	verifikasi := app.Group(
		"/api/verifikasi",
		middleware.RequireAuth,
	)

	// GET SEMUA
	verifikasi.Get(
		"/",
		controller.GetVerifikasi,
	)

	// GET PENGAJUAN MANAGER
	verifikasi.Get(
		"/pengajuan",
		controller.GetPengajuan,
	)

	// GET LOG
	verifikasi.Get(
		"/log-peninjauan",
		controller.GetLogPeninjauan,
	)

	// GET LOG PERALATAN
	verifikasi.Get(
		"/log-peninjauan/peralatan/:peralatan_id",
		controller.GetLogByPeralatan,
	)

	// GET BY PERALATAN
	verifikasi.Get(
		"/peralatan/:peralatan_id",
		controller.GetByPeralatan,
	)

	// GET BY ID
	verifikasi.Get(
		"/:id",
		controller.GetVerifikasiByID,
	)

	// CREATE DRAFT
	verifikasi.Post(
		"/",
		controller.CreateVerifikasi,
	)

	// PIC SIGN
	verifikasi.Put(
		"/:id/sign-pic",
		controller.SignPIC,
	)

	// MANAGER APPROVE
	verifikasi.Put(
		"/:id/approve",
		controller.ApproveVerifikasi,
	)

	// MANAGER REJECT
	verifikasi.Put(
		"/:id/reject",
		controller.RejectVerifikasi,
	)

	// DELETE
	verifikasi.Delete(
		"/:id",
		controller.DeleteVerifikasi,
	)
}