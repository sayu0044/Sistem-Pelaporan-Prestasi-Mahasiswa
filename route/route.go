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
			// Auth routes (protected)
			auth := v1.Group("/auth")
			{
				auth.Get("/me", authService.HandleGetMe)
				auth.Get("/profile", authService.HandleProfile)
				auth.Post("/logout", authService.HandleLogout)
			}

			// Users Management Routes
			// Requires: read/create/update/delete users permissions
			users := v1.Group("/users")
			{
				users.Get("/", middleware.RBACMiddleware("read", "users"), userService.HandleGetAllUsers)
				users.Get("/:id", middleware.RBACMiddleware("read", "users"), userService.HandleGetUserByID)
				users.Post("/", middleware.RBACMiddleware("create", "users"), userService.HandleCreateUser)
				users.Put("/:id", middleware.RBACMiddleware("update", "users"), userService.HandleUpdateUser)
				users.Delete("/:id", middleware.RBACMiddleware("delete", "users"), userService.HandleDeleteUser)
				users.Put("/:id/role", middleware.RBACMiddleware("update", "users"), userService.HandleUpdateUserRole)
			}

			// Achievement routes
			RegisterAchievementRoutes(v1, achievementService)
		}
	}
}
