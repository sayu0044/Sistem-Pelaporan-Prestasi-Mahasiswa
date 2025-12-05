package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/route"
)

// SetupApp membuat fiber instance, middleware, dan register route
func SetupApp() *fiber.App {
	// Parse JWT expiry
	jwtExpiry, _ := time.ParseDuration(JWTExpiry)
	app := fiber.New(fiber.Config{
		AppName: "Sistem Pelaporan Prestasi Mahasiswa",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Use(logger.New(logger.Config{
		Output: LoggerWriter,
	}))

	// Register routes
	route.RegisterRoutes(app, JWTSecret, jwtExpiry)

	return app
}
