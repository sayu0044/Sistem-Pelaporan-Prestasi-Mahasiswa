package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
)

// JWTMiddleware untuk validasi JWT token
func JWTMiddleware(jwtSecret string, jwtExpiry time.Duration) fiber.Handler {
	authService := service.NewAuthService(jwtSecret, jwtExpiry)

	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token tidak ditemukan",
			})
		}

		// Remove "Bearer " prefix if exists
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token tidak valid",
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("role_id", claims.RoleID)

		return c.Next()
	}
}

