package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
)

func JWTMiddleware(authService service.AuthService) fiber.Handler {

	return func(c *fiber.Ctx) error {
		// Cek Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token tidak ditemukan. Pastikan header 'Authorization: Bearer <token>' dikirim",
			})
		}

		// Remove "Bearer " prefix if exists
		token := authHeader
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else if len(token) > 0 {
			// Jika tidak ada prefix "Bearer ", coba langsung gunakan token
			// Tapi lebih baik tetap cek apakah ada prefix
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token kosong. Format: 'Authorization: Bearer <token>'",
			})
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token tidak valid: " + err.Error(),
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("role_id", claims.RoleID)
		c.Locals("role_name", claims.RoleName)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}
