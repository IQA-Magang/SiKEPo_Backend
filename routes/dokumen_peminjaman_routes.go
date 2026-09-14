package routes

import (
	"github.com/gofiber/fiber/v2"

	"backend/controllers"
	"backend/middleware"
)

func DokumenPeralatanRoutes(
	app *fiber.App,
	controller *controllers.DokumenPeralatanController,
) {
	// Public / authenticated routes
	dokumen := app.Group("/api/dokumen-peralatan", middleware.RequireAuth)

	dokumen.Get("/", controller.GetAll)

	// Harus sebelum /:id
	dokumen.Get(
		"/peralatan/:peralatan_id",
		controller.GetByPeralatanID,
	)

	dokumen.Get("/:id", controller.GetByID)

	// Protected admin routes
	admin := dokumen.Group(
		"/",
		middleware.RequireAuth,
		middleware.RequireRoles("admin"),
	)

	admin.Post("/", controller.Create)
	admin.Put("/:id", controller.Update)
	admin.Delete("/:id", controller.Delete)
}
