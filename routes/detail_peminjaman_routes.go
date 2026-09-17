package routes

// import (
// 	"backend/controllers"
// 	"backend/middleware"

// 	"github.com/gofiber/fiber/v2"
// )

// func DetailPeminjamanRoutes(
// 	app *fiber.App,
// 	controller *controllers.DetailPeminjamanController,
// ) {

// 	detail := app.Group(
// 		"/api/detail-peminjaman",
// 		middleware.RequireAuth,
// 	)

// 	// ========================================
// 	// GET
// 	// ========================================

// 	detail.Get(
// 		"/",
// 		controller.GetAll,
// 	)

// 	// HARUS DITARUH SEBELUM /:id
// 	detail.Get(
// 		"/peminjaman/:peminjaman_id",
// 		controller.GetByPeminjamanID,
// 	)

// 	detail.Get(
// 		"/:id",
// 		controller.GetByID,
// 	)

// 	// ========================================
// 	// CREATE
// 	// ========================================

// 	detail.Post(
// 		"/",
// 		middleware.RequireRoles("admin", "staff", "manager"),
// 		controller.Create,
// 	)

// 	// ========================================
// 	// UPDATE
// 	// ========================================

// 	detail.Put(
// 		"/:id",
// 		middleware.RequireRoles("admin", "staff", "manager"),
// 		controller.Update,
// 	)

// 	// ========================================
// 	// APPROVAL
// 	// ========================================

// 	detail.Put(
// 		"/:id/approve",
// 		middleware.RequireRoles("manager", "staff"),
// 		controller.Approve,
// 	)

// 	detail.Put(
// 		"/:id/reject",
// 		middleware.RequireRoles("manager", "staff"),
// 		controller.Reject,
// 	)

// 	// ========================================
// 	// KONDISI SAAT PINJAM
// 	// ========================================

// 	detail.Put(
// 		"/:id/kondisi-pinjam",
// 		middleware.RequireRoles("admin", "staff", "manager"),
// 		controller.SetKondisiPinjam,
// 	)

// 	// ========================================
// 	// KONDISI SAAT KEMBALI
// 	// ========================================

// 	detail.Put(
// 		"/:id/kondisi-kembali",
// 		middleware.RequireRoles("admin", "staff", "manager"),
// 		controller.SetKondisiKembali,
// 	)

// 	// ========================================
// 	// DELETE
// 	// ========================================

// 	detail.Delete(
// 		"/:id",
// 		middleware.RequireRoles("admin", "manager"),
// 		controller.Delete,
// 	)
// }
