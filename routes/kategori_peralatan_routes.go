package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func KategoriPeralatanRoutes(app *fiber.App, controller *controllers.KategoriPeralatanController) {
	api := app.Group("/api/kategori-peralatan", middleware.RequireAuth)

	api.Get("/", controller.GetAll)
	api.Get("/:id", controller.GetByID)

	admin := app.Group("/api/kategori-peralatan", middleware.RequireAuth, middleware.RequireRoles("admin"))
	admin.Post("/", controller.Create)
	admin.Put("/:id", controller.Update)
	admin.Delete("/:id", controller.Delete)
}
