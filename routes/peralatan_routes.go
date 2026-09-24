package routes

import (
	"backend/controllers"
	"backend/middleware"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupPeralatanRoutes mendaftarkan endpoint untuk modul peralatan
func SetupPeralatanRoutes(app *fiber.App, db *gorm.DB, notificationRepo repositories.NotificationRepository) {
	// 1. Inisialisasi Repository dan Controller
	peralatanRepo := repositories.NewPeralatanRepository(db)
	peralatanController := controllers.NewPeralatanController(peralatanRepo)
	peralatanController.NotificationRepo = notificationRepo
	peralatanController.DB = db

	// 2. Grouping Route API
	api := app.Group("/api/peralatan", middleware.RequireAuth)

	// 3. Daftarkan Endpoint POST
	api.Get("/", peralatanController.GetAll)
	api.Post("/", peralatanController.Create, middleware.RequireAdminOrStaffPIC())
	api.Post("/:id/foto", peralatanController.UploadFoto, middleware.RequireAdminOrStaffPIC())
	api.Get("/:id/qr", peralatanController.GenerateQRCode, middleware.RequireAdminOrStaffPIC())
}
