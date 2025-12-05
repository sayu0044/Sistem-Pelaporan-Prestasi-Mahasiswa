package route

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterAuthRoutes mendaftarkan route untuk authentication
func RegisterAuthRoutes(app *fiber.App, authService service.AuthService, jwtSecret string, jwtExpiry time.Duration) {
	auth := app.Group("/api/auth")
	{
		auth.Post("/login", func(c *fiber.Ctx) error {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid request body",
				})
			}

			token, user, err := authService.Login(req.Username, req.Password)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Login berhasil",
				"token":   token,
				"user": fiber.Map{
					"id":        user.ID,
					"username":  user.Username,
					"email":     user.Email,
					"full_name": user.FullName,
					"role_id":   user.RoleID,
				},
			})
		})

		// Protected route untuk test JWT
		auth.Get("/me", middleware.JWTMiddleware(jwtSecret, jwtExpiry), func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"error": false,
				"user": fiber.Map{
					"user_id":  c.Locals("user_id"),
					"username": c.Locals("username"),
					"email":    c.Locals("email"),
					"role_id":  c.Locals("role_id"),
				},
			})
		})
	}
}
