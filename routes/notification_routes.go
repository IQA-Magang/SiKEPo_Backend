package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

// NotificationRoutes mendaftarkan endpoint notifikasi. Semua role yang login boleh
// memakainya, tetapi controller memastikan user hanya mengakses notifikasi miliknya.
func NotificationRoutes(app *fiber.App, controller *controllers.NotificationController) {
	notifications := app.Group("/api/notifications", middleware.RequireAuth)
	notifications.Get("/user/:user_id", controller.GetByUserID)
	notifications.Patch("/:id/read", controller.MarkRead)
}
