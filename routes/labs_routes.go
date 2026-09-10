package routes

import (
	"github.com/gofiber/fiber/v2"

	"backend/controllers"
	"backend/middleware"
)

func LabsRoutes(
	app *fiber.App,
	controller *controllers.LabsController,
) {
	// Public / authenticated routes
	labs := app.Group("/api/v1/labs", middleware.RequireAuth)

	labs.Get("/", controller.GetAll)
	labs.Get("/:id", controller.GetByID)

	// Protected admin routes
	admin := labs.Group(
		"/",
		middleware.RequireAuth,
		middleware.RequireRoles("admin"),
	)

	admin.Post("/", controller.Create)
	admin.Put("/:id", controller.Update)
	admin.Delete("/:id", controller.Delete)
}
