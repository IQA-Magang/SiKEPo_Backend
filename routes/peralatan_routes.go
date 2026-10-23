package routes

import (
	"backend/controllers"
	"backend/middleware"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupPeralatanRoutes mendaftarkan endpoint untuk modul peralatan
func SetupPeralatanRoutes(
	app *fiber.App,
	db *gorm.DB,
	notificationRepo repositories.NotificationRepository,
) {
	// =====================================================
	// REPOSITORY
	// =====================================================

	peralatanRepo :=
		repositories.NewPeralatanRepository(db)

	// =====================================================
	// CONTROLLER
	// =====================================================

	peralatanController :=
		controllers.NewPeralatanController(
			peralatanRepo,
		)

	peralatanController.NotificationRepo =
		notificationRepo

	peralatanController.DB =
		db

	// =====================================================
	// GUEST
	// =====================================================

	guest :=
		app.Group("/api/peralatan")

	guest.Get(
		"/:nomor_aset",
		peralatanController.GetByNomorAset,
	)

	// =====================================================
	// AUTHENTICATED
	// =====================================================

	api :=
		app.Group(
			"/api/peralatan",
			middleware.RequireAuth,
		)

	// =====================================================
	// GET ALL
	// =====================================================

	api.Get(
		"/",
		peralatanController.GetAll,
	)

	// =====================================================
	// CREATE
	// =====================================================

	api.Post(
		"/",
		peralatanController.Create,
		middleware.RequireAdminOrStaffPengelola(),
	)

	// =====================================================
	// UPLOAD FOTO
	// =====================================================

	api.Post(
		"/:id/foto",
		peralatanController.UploadFoto,
		middleware.RequireAdminOrStaffPengelola(),
	)

	// =====================================================
	// QR CODE
	// =====================================================

	api.Get(
		"/:id/qr",
		peralatanController.GenerateQRCode,
		middleware.RequireAdminOrStaffPengelola(),
	)
}
