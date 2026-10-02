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

	manager := users.Group("/manager", middleware.RequireAuth, middleware.RequireRoles("manager"))
	manager.Get("/", controller.GetUsers)
	manager.Get(":id", controller.GetUserByID)
	manager.Put(":id", controller.UpdateUser)

	// Protected admin routes must follow static manager routes so /manager is not
	// matched as the admin /:id route.
	admin := users.Group("/", middleware.RequireAuth, middleware.RequireRoles("admin"))
	admin.Post("/", controller.CreateUser)
	admin.Put(":id", controller.UpdateUser)
	admin.Delete(":id", controller.DeleteUser)
	admin.Get("/", controller.GetUsers)
	admin.Get(":id", controller.GetUserByID)
}
