package app

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"backend/config"
	"backend/controllers"
	"backend/repositories"
	"backend/routes"
)

func CreateApp() (*fiber.App, error) {

	if err := config.ConnectDatabase(); err != nil {
		return nil, err
	}

	app := fiber.New()

	app.Use(cors.New())

	app.Static("/docs", "./docs")
	app.Static("/static", "./public")

	app.Get("/recaptcha/sitekey", func(c *fiber.Ctx) error {

		site := os.Getenv("SITE_KEY")

		if site == "" {
			return c.Status(
				fiber.StatusInternalServerError,
			).JSON(fiber.Map{
				"success": false,
				"message": "recaptcha site key not configured",
			})
		}

		return c.JSON(fiber.Map{
			"site_key": site,
		})
	})

	// =====================================================
	// REPOSITORY
	// =====================================================

	userRepository :=
		repositories.NewUserRepository(config.DB)

	ruanganRepo :=
		repositories.NewRuanganRepository(config.DB)

	labsRepo :=
		repositories.NewLabsRepository(config.DB)

	kelompokAssetRepo :=
		repositories.NewKelompokAssetRepository(config.DB)

	dokumenPeralatanRepository :=
		repositories.NewDokumenPeralatanRepository(config.DB)

	notificationRepo :=
		repositories.NewNotificationRepository(config.DB)

	// =====================================================
	// CONTROLLER
	// =====================================================

	userController := &controllers.UserController{
		Repository: userRepository,
	}

	labsController := &controllers.LabsController{
		Repository: labsRepo,
	}

	ruanganController := &controllers.RuanganController{
		Repository: ruanganRepo,
	}

	kelompokAssetController :=
		&controllers.KelompokAssetController{
			Repository:     kelompokAssetRepo,
			LabsRepository: labsRepo,
			UserRepository: userRepository,
		}

	dokumenPeralatanController :=
		&controllers.DokumenPeralatanController{
			Repository: dokumenPeralatanRepository,
		}

	// =====================================================
	// ROUTES
	// =====================================================

	routes.UserRoutes(
		app,
		userController,
	)

	routes.LabsRoutes(
		app,
		labsController,
	)

	routes.RuanganRoutes(
		app,
		ruanganController,
	)

	routes.KelompokAssetRoutes(
		app,
		kelompokAssetController,
	)

	routes.DokumenPeralatanRoutes(
		app,
		dokumenPeralatanController,
	)

	routes.NotificationRoutes(
		app,
		&controllers.NotificationController{
			Repo: notificationRepo,
		},
	)

	routes.SetupPeralatanRoutes(
		app,
		config.DB,
		notificationRepo,
	)

	routes.VerifikasiRoutes(
		app,
		config.DB,
		notificationRepo,
	)
	// =====================================================
	// ROOT
	// =====================================================

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Backend API Running",
		})
	})

	return app, nil
}