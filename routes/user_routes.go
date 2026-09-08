package routes

import (
	"github.com/gofiber/fiber/v2"

	"backend/controllers"
	"backend/middleware"
)

func UserRoutes(
	app *fiber.App,
	controller *controllers.UserController,
) {

	// Public routes
	users := app.Group("/api/users")
	users.Post("/login", controller.Login)
	users.Get("/", controller.GetUsers)
	users.Get(":id", controller.GetUserByID)

	// Protected admin routes
	admin := users.Group("/", middleware.RequireAuth, middleware.RequireRoles("admin"))
	admin.Post("/", controller.CreateUser)
	admin.Put(":id", controller.UpdateUser)
	admin.Delete(":id", controller.DeleteUser)
}
