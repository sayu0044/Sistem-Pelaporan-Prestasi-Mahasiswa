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
	studentService := service.NewStudentService()
	lecturerService := service.NewLecturerService()

	// Health check
	app.Get("/", authService.HandleHealthCheck)

	// Public auth routes (login tidak perlu JWT) - HARUS didefinisikan SEBELUM protected routes
	authPublic := app.Group("/api/v1/auth")
	{
		authPublic.Post("/login", authService.HandleLogin)
		authPublic.Post("/refresh", authService.HandleRefreshToken) // Refresh token (perlu token lama)
	}

	// API routes (protected)
	api := app.Group("/api", middleware.JWTMiddleware(jwtSecret, jwtExpiry))
	{
		api.Get("/test", authService.HandleTest)

		// V1 API routes
		v1 := api.Group("/v1")
		{
			// Auth routes (protected) - menggunakan /api/v1/auth
			auth := v1.Group("/auth")
			{
				auth.Get("/me", authService.HandleGetMe)
				auth.Get("/profile", authService.HandleProfile)
				auth.Post("/logout", authService.HandleLogout)
			}

			// User routes
			RegisterUserRoutes(v1, userService)

			// Achievement routes
			RegisterAchievementRoutes(v1, achievementService)

			// Student routes
			RegisterStudentRoutes(v1, studentService)

			// Lecturer routes
			RegisterLecturerRoutes(v1, lecturerService)
		}
	}
}
