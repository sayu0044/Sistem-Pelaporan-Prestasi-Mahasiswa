package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RBACMiddleware creates a middleware that checks if the user has the required permission
// This middleware should be used AFTER JWTMiddleware to ensure permissions are available in context
// Permissions are read from JWT token (not from database) for better performance
func RBACMiddleware(action, resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get permissions from context (set by JWT middleware from token)
		permissionsInterface := c.Locals("permissions")
		if permissionsInterface == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Anda tidak memiliki izin untuk mengakses resource ini",
				"required_permission": fiber.Map{
					"action":   action,
					"resource": resource,
				},
			})
		}

		// Convert permissions to []string
		permissions, ok := permissionsInterface.([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Anda tidak memiliki izin untuk mengakses resource ini",
				"required_permission": fiber.Map{
					"action":   action,
					"resource": resource,
				},
			})
		}

		// Get role_name from context for admin check
		roleNameInterface := c.Locals("role_name")
		roleName := ""
		if roleNameInterface != nil {
			if name, ok := roleNameInterface.(string); ok {
				roleName = name
			}
		}

		// Check if user is admin (case-insensitive)
		// Admin memiliki akses penuh (permissions contains "*:*")
		roleNameLower := strings.ToLower(roleName)
		if strings.Contains(roleNameLower, "admin") {
			return c.Next()
		}

		// Check for wildcard permission
		for _, perm := range permissions {
			if perm == "*:*" {
				return c.Next()
			}
		}

		// Format required permission as "resource:action"
		requiredPermission := strings.ToLower(resource) + ":" + strings.ToLower(action)

		// Check if user has the required permission
		hasPermission := false
		for _, perm := range permissions {
			if strings.ToLower(perm) == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Anda tidak memiliki izin untuk mengakses resource ini",
				"required_permission": fiber.Map{
					"action":   action,
					"resource": resource,
				},
			})
		}

		// Permission granted, continue to next handler
		return c.Next()
	}
}

