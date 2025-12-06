package route

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterRoutes mendaftarkan semua route aplikasi
func RegisterRoutes(app *fiber.App, jwtSecret string, jwtExpiry time.Duration) {
	// Initialize services
	authService := service.NewAuthService(jwtSecret, jwtExpiry)
	userService := service.NewUserService(authService)
	achievementService := service.NewAchievementService()

	// Health check
	app.Get("/", authService.HandleHealthCheck)

	// Auth routes
	auth := app.Group("/api/auth")
	{
		auth.Post("/login", authService.HandleLogin)
		auth.Get("/me", middleware.JWTMiddleware(jwtSecret, jwtExpiry), authService.HandleGetMe)
	}

	// API routes (protected)
	api := app.Group("/api", middleware.JWTMiddleware(jwtSecret, jwtExpiry))
	{
		api.Get("/test", authService.HandleTest)

		// Admin routes
		admin := api.Group("/admin")
		{
			// Admin Users Management
			users := admin.Group("/users")
			{
				users.Post("/", userService.HandleCreateUser)
				users.Put("/:id", userService.HandleUpdateUser)
			}
		}

		// Achievement routes
		RegisterAchievementRoutes(api, achievementService)
	}
}
