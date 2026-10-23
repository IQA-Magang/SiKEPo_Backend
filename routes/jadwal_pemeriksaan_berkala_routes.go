package routes

import (
	"backend/controllers"
	"backend/middleware"
	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupJadwalPemeriksaanBerkalaRoutes(
	app *fiber.App,
	db *gorm.DB,
) {
	repo := repositories.NewJadwalPemeriksaanBerkalaRepository(db)
	controller := controllers.NewJadwalPemeriksaanBerkalaController(repo)

	api := app.Group(
		"/api/jadwal-pemeriksaan-berkala",
		middleware.RequireAuth,
	)

	api.Post("/", controller.Create)
	api.Get("/", controller.FindAll)
	api.Get("/:id", controller.FindByID)
	api.Put("/:id/realisasi", controller.UpdateRealisasi)
}
