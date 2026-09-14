package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func KelompokAssetRoutes(
	app *fiber.App,
	ctrl *controllers.KelompokAssetController,
) {

	api := app.Group(
		"/api/v1/kelompok-asset",
		middleware.RequireAuth,
	)

	// CREATE
	api.Post("/", ctrl.Create)

	// GET ALL
	api.Get("/", ctrl.GetAll)

	// GET BY ID
	api.Get("/:id", ctrl.GetByID)

	// UPDATE
	api.Put("/:id", ctrl.Update)

	// DELETE
	api.Delete("/:id", ctrl.Delete)
}