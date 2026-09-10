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

	// ========================================
	// GET ALL
	// ========================================

	verifikasi.Get(
		"/",
		controller.GetVerifikasi,
	)

	// ========================================
	// GET BY PERALATAN
	// HARUS SEBELUM /:id
	// ========================================

	verifikasi.Get(
		"/peralatan/:peralatan_id",
		controller.GetByPeralatan,
	)

	// ========================================
	// GET BY ID
	// ========================================

	verifikasi.Get(
		"/:id",
		controller.GetVerifikasiByID,
	)

	// ========================================
	// CREATE
	// Staff PIC
	// ========================================

	verifikasi.Post(
		"/",
		middleware.RequireRoles("staff"),
		controller.CreateVerifikasi,
	)

	// ========================================
	// APPROVE
	// MANAGER
	// ========================================

	verifikasi.Put(
		"/:id/approve",
		middleware.RequireRoles("manager"),
		controller.ApproveVerifikasi,
	)

	// ========================================
	// DELETE
	// Admin
	// ========================================

	verifikasi.Delete(
		"/:id",
		middleware.RequireRoles("admin"),
		controller.DeleteVerifikasi,
	)
}