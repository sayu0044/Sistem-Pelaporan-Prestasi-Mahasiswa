package route

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB, mongoDB *mongo.Database, jwtSecret string, jwtExpiry time.Duration) {
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	achievementRepo := repository.NewAchievementRepository(db, mongoDB)
	historyRepo := repository.NewAchievementHistoryRepository(db)

	authService := service.NewAuthService(userRepo, roleRepo, jwtSecret, jwtExpiry)
	userService := service.NewUserService(userRepo, roleRepo, lecturerRepo, studentRepo, authService)
	achievementService := service.NewAchievementService(achievementRepo, historyRepo, studentRepo, lecturerRepo, userRepo, roleRepo)
	studentService := service.NewStudentService(studentRepo, lecturerRepo, achievementRepo)
	lecturerService := service.NewLecturerService(lecturerRepo, studentRepo)
	reportService := service.NewReportService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "API Sistem Pelaporan Prestasi Mahasiswa",
			"status":  "running",
		})
	})

	authPublic := app.Group("/api/v1/auth")
	{
		authPublic.Post("/login", func(c *fiber.Ctx) error {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Pastikan body permintaan Anda dalam format JSON yang benar.",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			token, refreshToken, user, role, err := authService.Login(ctx, req.Username, req.Password)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Gagal login",
					"message": err.Error(),
				})
			}

			var permissions []string
			var roleName string
			if role != nil {
				roleName = role.Name
				roleNameLower := strings.ToLower(roleName)
				if strings.Contains(roleNameLower, "admin") {
					permissions = append(permissions, "*:*")
				} else {
					for _, perm := range role.Permissions {
						permissionString := strings.ToLower(perm.Resource) + ":" + strings.ToLower(perm.Action)
						permissions = append(permissions, permissionString)
					}
				}
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"error": false,
				"data": fiber.Map{
					"token":        token,
					"refreshToken": refreshToken,
					"user": fiber.Map{
						"id":          user.ID,
						"username":    user.Username,
						"fullName":    user.FullName,
						"role":        roleName,
						"permissions": permissions,
					},
				},
			})
		})

		authPublic.Post("/refresh", func(c *fiber.Ctx) error {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Token tidak ditemukan",
					"message": "Pastikan header 'Authorization: Bearer <token>' dikirim",
				})
			}

			token := authHeader
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			if token == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Token kosong",
					"message": "Format: 'Authorization: Bearer <token>'",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			newToken, newRefreshToken, user, role, err := authService.RefreshToken(ctx, token)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Gagal refresh token",
					"message": err.Error(),
				})
			}

			var permissions []string
			var roleName string
			if role != nil {
				roleName = role.Name
				roleNameLower := strings.ToLower(roleName)
				if strings.Contains(roleNameLower, "admin") {
					permissions = append(permissions, "*:*")
				} else {
					for _, perm := range role.Permissions {
						permissionString := strings.ToLower(perm.Resource) + ":" + strings.ToLower(perm.Action)
						permissions = append(permissions, permissionString)
					}
				}
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data": fiber.Map{
					"token":        newToken,
					"refreshToken": newRefreshToken,
					"user": fiber.Map{
						"id":          user.ID,
						"username":    user.Username,
						"fullName":    user.FullName,
						"role":        roleName,
						"permissions": permissions,
					},
				},
			})
		})
	}

	api := app.Group("/api", middleware.JWTMiddleware(authService))
	{
		api.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"message": "Protected route berhasil diakses",
			})
		})

		v1 := api.Group("/v1")
		{
			auth := v1.Group("/auth")
			{
				auth.Get("/me", func(c *fiber.Ctx) error {
					return c.JSON(fiber.Map{
						"error": false,
						"data": fiber.Map{
							"user": fiber.Map{
								"user_id":     c.Locals("user_id"),
								"username":    c.Locals("username"),
								"role":        c.Locals("role_name"),
								"permissions": c.Locals("permissions"),
							},
						},
					})
				})

				auth.Get("/profile", func(c *fiber.Ctx) error {
					userIDInterface := c.Locals("user_id")
					if userIDInterface == nil {
						return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
							"error":   "User ID tidak ditemukan",
							"message": "User ID tidak ditemukan",
						})
					}

					userID, ok := userIDInterface.(uuid.UUID)
					if !ok {
						return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
							"error":   "User ID tidak valid",
							"message": "User ID tidak valid",
						})
					}

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					user, role, err := authService.GetProfile(ctx, userID)
					if err != nil {
						if err.Error() == "user tidak ditemukan" {
							return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
								"error":   "User tidak ditemukan",
								"message": err.Error(),
							})
						}
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error":   "Gagal mengambil data",
							"message": err.Error(),
						})
					}

					var permissions []fiber.Map
					var roleData fiber.Map
					if role != nil {
						roleData = fiber.Map{
							"id":          role.ID,
							"name":        role.Name,
							"description": role.Description,
						}

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
					}

					return c.JSON(fiber.Map{
						"error": false,
						"data": fiber.Map{
							"id":          user.ID,
							"username":    user.Username,
							"fullName":    user.FullName,
							"role":        roleData,
							"permissions": permissions,
							"isActive":    user.IsActive,
							"createdAt":   user.CreatedAt,
							"updatedAt":   user.UpdatedAt,
						},
					})
				})

				auth.Post("/logout", func(c *fiber.Ctx) error {
					userID := c.Locals("user_id")
					username := c.Locals("username")

					return c.JSON(fiber.Map{
						"error": false,
						"data": fiber.Map{
							"user": fiber.Map{
								"user_id":  userID,
								"username": username,
							},
						},
					})
				})
			}

			RegisterUserRoutes(v1, userService)
			RegisterAchievementRoutes(v1, achievementService)
			RegisterStudentRoutes(v1, studentService)
			RegisterLecturerRoutes(v1, lecturerService)
			RegisterReportRoutes(v1, reportService)
		}
	}
}
