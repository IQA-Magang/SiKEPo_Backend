package routes

import (
	"backend/controllers"
	"backend/middleware"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PeminjamanRoutes mendaftarkan endpoint peminjaman peralatan.
//
// Semua endpoint hanya membutuhkan login. Hak akses tiap langkah, yaitu staff
// peminjam atau penyetuju yang ditetapkan, dicek di controller karena
// bergantung pada data pengajuan, bukan sekadar role.
func PeminjamanRoutes(app *fiber.App, db *gorm.DB) {
	controller := controllers.NewPeminjamanController(
		repositories.NewPeminjamanRepository(db),
		repositories.NewLogbookRepository(db),
	)

	peminjaman := app.Group("/api/peminjaman", middleware.RequireAuth)

	peminjaman.Get("/", controller.GetAll)
	peminjaman.Post("/", controller.Create)
	peminjaman.Get("/:id", controller.GetByID)
	peminjaman.Put("/:id/keputusan", controller.Keputusan)
}
