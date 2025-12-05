package route

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterRoutes mendaftarkan semua route aplikasi
func RegisterRoutes(app *fiber.App, jwtSecret string, jwtExpiry time.Duration) {
	// Initialize services
	authService := service.NewAuthService(jwtSecret, jwtExpiry)
	userService := service.NewUserService(authService)

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

		// V1 routes
		v1 := api.Group("/v1")
		{
			// Users routes
			users := v1.Group("/users")
			{
				// GET /api/v1/users - Get all users
				users.Get("/", func(c *fiber.Ctx) error {
					users, err := userService.GetAllUsers()
					if err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error":   true,
							"message": err.Error(),
						})
					}

					// Format response
					var formattedUsers []fiber.Map
					for _, user := range users {
						var permissions []fiber.Map
						if len(user.Role.Permissions) > 0 {
							for _, perm := range user.Role.Permissions {
								permissions = append(permissions, fiber.Map{
									"id":          perm.ID,
									"name":        perm.Name,
									"resource":    perm.Resource,
									"action":      perm.Action,
									"description": perm.Description,
								})
							}
						}

						formattedUsers = append(formattedUsers, fiber.Map{
							"id":        user.ID,
							"username":  user.Username,
							"email":     user.Email,
							"full_name": user.FullName,
							"role_id":   user.RoleID,
							"role": fiber.Map{
								"id":          user.Role.ID,
								"name":        user.Role.Name,
								"description": user.Role.Description,
							},
							"permissions": permissions,
							"is_active":   user.IsActive,
							"created_at":  user.CreatedAt,
							"updated_at":  user.UpdatedAt,
						})
					}

					return c.JSON(fiber.Map{
						"error":   false,
						"message": "Berhasil mengambil data users",
						"data":    formattedUsers,
					})
				})

				// GET /api/v1/users/:id - Get user by ID
				users.Get("/:id", func(c *fiber.Ctx) error {
					userIDStr := c.Params("id")
					userID, err := uuid.Parse(userIDStr)
					if err != nil {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error":   true,
							"message": "User ID tidak valid",
						})
					}

					user, role, err := userService.GetUserByID(userID)
					if err != nil {
						if err.Error() == "user tidak ditemukan" {
							return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
								"error":   true,
								"message": err.Error(),
							})
						}
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error":   true,
							"message": err.Error(),
						})
					}

					// Format permissions
					var permissions []fiber.Map
					if len(role.Permissions) > 0 {
						for _, perm := range role.Permissions {
							permissions = append(permissions, fiber.Map{
								"id":          perm.ID,
								"name":        perm.Name,
								"resource":    perm.Resource,
								"action":      perm.Action,
								"description": perm.Description,
							})
						}
					}

					return c.JSON(fiber.Map{
						"error":   false,
						"message": "Berhasil mengambil data user",
						"data": fiber.Map{
							"id":        user.ID,
							"username":  user.Username,
							"email":     user.Email,
							"full_name": user.FullName,
							"role_id":   user.RoleID,
							"role": fiber.Map{
								"id":          role.ID,
								"name":        role.Name,
								"description": role.Description,
							},
							"permissions": permissions,
							"is_active":   user.IsActive,
							"created_at":  user.CreatedAt,
							"updated_at":  user.UpdatedAt,
						},
					})
				})

				// DELETE /api/v1/users/:id - Delete user by ID
				users.Delete("/:id", func(c *fiber.Ctx) error {
					userIDStr := c.Params("id")
					userID, err := uuid.Parse(userIDStr)
					if err != nil {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error":   true,
							"message": "User ID tidak valid",
						})
					}

					err = userService.DeleteUser(userID)
					if err != nil {
						if err.Error() == "user tidak ditemukan" {
							return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
								"error":   true,
								"message": err.Error(),
							})
						}
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error":   true,
							"message": err.Error(),
						})
					}

					return c.JSON(fiber.Map{
						"error":   false,
						"message": "User berhasil dihapus",
					})
				})
			}
		}

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
	}
}
