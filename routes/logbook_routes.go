package routes

import (
	"backend/controllers"
	"backend/middleware"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// LogbookRoutes mendaftarkan endpoint logbook peralatan.
//
// Logbook hanya bisa dibaca. Barisnya ditulis oleh modul lain, misalnya
// peminjaman, di dalam transaksinya sendiri.
func LogbookRoutes(app *fiber.App, db *gorm.DB) {
	controller := controllers.NewLogbookController(
		repositories.NewLogbookRepository(db),
		repositories.NewPeralatanRepository(db),
	)

	app.Get(
		"/api/peralatan/:id/logbook",
		middleware.RequireAuth,
		controller.GetByPeralatan,
	)
}
