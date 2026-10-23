package app

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"backend/config"
	"backend/controllers"
	"backend/repositories"
	"backend/routes"
	"backend/utils"
)

func CreateApp() (*fiber.App, error) {

	// =====================================================
	// DATABASE
	// =====================================================

	if err := config.ConnectDatabase(); err != nil {
		return nil, err
	}

	// =====================================================
	// FIBER
	// =====================================================

	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     os.Stdout,
		Done: func(c *fiber.Ctx, _ []byte) {
			if c.Response().StatusCode() < fiber.StatusBadRequest {
				return
			}

			var response struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			_ = json.Unmarshal(c.Response().Body(), &response)
			detail := strings.TrimSpace(response.Error)
			if detail == "" {
				detail = strings.TrimSpace(response.Message)
			}
			if len(detail) > 500 {
				detail = detail[:500] + "..."
			}

			log.Printf(
				"HTTP error response | status=%d | %s %s | %s",
				c.Response().StatusCode(),
				c.Method(),
				c.Path(),
				detail,
			)
		},
	}))
	app.Use(cors.New())

	// =====================================================
	// STATIC FILE
	// =====================================================

	app.Static("/docs", utils.UploadRoot())
	app.Static("/static", "./public")

	// =====================================================
	// RECAPTCHA
	// =====================================================

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

	kategoriPeralatanRepo :=
		repositories.NewKategoriPeralatanRepository(config.DB)

	kategoriPeralatanController :=
		controllers.NewKategoriPeralatanController(
			kategoriPeralatanRepo,
		)

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

	routes.KategoriPeralatanRoutes(
		app,
		kategoriPeralatanController,
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
	routes.SetupJadwalPemeriksaanBerkalaRoutes(
		app,
		config.DB,
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
