package routes

import (
	"github.com/gofiber/fiber/v2"

	"backend/controllers"
	"backend/middleware"
)

func PeralatanRoutes(app *fiber.App, ctrl *controllers.PeralatanController) {
	api := app.Group("/api/v1/peralatan", middleware.RequireAuth)

	api.Post("/", ctrl.Create)
	api.Get("/", ctrl.GetAll)
	api.Get("/:id", ctrl.GetByID)
	api.Put("/:id", ctrl.Update)
	api.Delete("/:id", ctrl.Delete)
}
