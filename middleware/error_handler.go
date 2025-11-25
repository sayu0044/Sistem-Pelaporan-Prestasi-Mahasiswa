package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler untuk menangani error secara global
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Log error (using standard log since fiber.App doesn't have Logger method)
	// Error sudah di-handle oleh fiber secara default

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}

