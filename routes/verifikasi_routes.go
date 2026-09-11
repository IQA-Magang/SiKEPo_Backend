package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func VerifikasiRoutes(
	app *fiber.App,
	controller *controllers.VerifikasiController,
) {
	verifikasi := app.Group(
		"/api/verifikasi",
		middleware.RequireAuth,
	)

	// GET ALL
	verifikasi.Get(
		"/",
		controller.GetVerifikasi,
	)

	// GET BY PERALATAN
	verifikasi.Get(
		"/peralatan/:peralatan_id",
		controller.GetByPeralatan,
	)

	// GET BY ID
	verifikasi.Get(
		"/:id",
		controller.GetVerifikasiByID,
	)

	// CREATE
	// Admin dan Staff dapat membuat verifikasi
	verifikasi.Post(
		"/",
		middleware.RequireRoles("admin", "staff"),
		controller.CreateVerifikasi,
	)

	// APPROVE
	// Hanya Manager
	verifikasi.Put(
		"/:id/approve",
		middleware.RequireRoles("manager"),
		controller.ApproveVerifikasi,
	)

	// DELETE
	// Hanya Admin
	verifikasi.Delete(
		"/:id",
		middleware.RequireRoles("admin"),
		controller.DeleteVerifikasi,
	)
}
