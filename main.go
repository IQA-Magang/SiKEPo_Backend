package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"backend/config"
	"backend/controllers"
	"backend/repositories"
	"backend/routes"
)

func main() {

	// =========================
	// DATABASE
	// =========================
	config.ConnectDatabase()

	// =========================
	// FIBER
	// =========================
	app := fiber.New()

	// =========================
	// CORS
	// =========================
	app.Use(cors.New())

	// =========================
	// STATIC (simple frontend for testing)
	// Serve files under /static (e.g. /static/login.html)
	// =========================
	app.Static("/static", "./public")

	// expose site key for the client login page
	app.Get("/recaptcha/sitekey", func(c *fiber.Ctx) error {
		site := os.Getenv("SITE_KEY")
		if site == "" {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "recaptcha site key not configured",
			})
		}

		return c.JSON(fiber.Map{"site_key": site})
	})

	// =========================
	// REPOSITORY
	// =========================
	userRepository := repositories.NewUserRepository(config.DB)

	// =========================
	// CONTROLLER
	// =========================
	userController := &controllers.UserController{
		Repository: userRepository,
	}

	// =========================
	// ROUTES
	// =========================
	routes.UserRoutes(app, userController)

	// ROOT API
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Backend API Running",
		})
	})

	// PORT
	port := os.Getenv("APP_PORT")

	if port == "" {
		port = "5000"
	}
	// RUN SERVER
	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}
