package routes

import (
	"github.com/gofiber/fiber/v2"

	"backend/controllers"
	"backend/middleware"
)

func RuanganRoutes(
	app *fiber.App,
	controller *controllers.RuanganController,
) {
	// Public / authenticated routes
	ruangan := app.Group("/api/v1/ruangan", middleware.RequireAuth)

	ruangan.Get("/", controller.GetAll)

	// Harus diletakkan sebelum /:id
	ruangan.Get("/labs/:labs_id", controller.GetByLabsID)
	ruangan.Get("/pic/:pic_user_id", controller.GetByPICUserID)

	ruangan.Get("/:id", controller.GetByID)

	// Protected admin routes
	admin := ruangan.Group(
		"/",
		middleware.RequireAuth,
		middleware.RequireRoles("admin"),
	)

	admin.Post("/", controller.Create)
	admin.Put("/:id", controller.Update)
	admin.Delete("/:id", controller.Delete)
}
