package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func NotificationRoutes(app *fiber.App, controller *controllers.NotificationController) {
	notifications := app.Group("/api/notifications", middleware.RequireAuth, middleware.RequireRoles("manager"))
	notifications.Get("/user/:user_id", controller.GetByUserID)
	notifications.Patch("/:id/read", controller.MarkRead)
}
