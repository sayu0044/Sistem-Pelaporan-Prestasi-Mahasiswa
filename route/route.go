package route

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterRoutes mendaftarkan semua route aplikasi
func RegisterRoutes(app *fiber.App, jwtSecret string, jwtExpiry time.Duration) {
	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "API Sistem Pelaporan Prestasi Mahasiswa",
			"status":  "running",
		})
	})

	// Auth routes
	authService := service.NewAuthService(jwtSecret, jwtExpiry)
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

		auth.Post("/register", func(c *fiber.Ctx) error {
			var req struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				Password string `json:"password"`
				FullName string `json:"full_name"`
				RoleID   string `json:"role_id,omitempty"` // Optional, bisa kosong
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid request body",
				})
			}

			// Validasi required fields
			if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Username, email, password, dan full_name harus diisi",
				})
			}

			// Parse user data dari request
			user := &model.User{
				Username: req.Username,
				Email:    req.Email,
				FullName: req.FullName,
				IsActive: true,
			}

			// Set role_id jika diberikan
			if req.RoleID != "" {
				roleUUID, err := uuid.Parse(req.RoleID)
				if err == nil {
					user.RoleID = &roleUUID
				}
			} else {
				// Auto-assign role "Mahasiswa" sebagai default jika role_id tidak diisi
				roleRepo := repository.NewRoleRepository()
				role, err := roleRepo.FindByName("Mahasiswa")
				if err == nil {
					user.RoleID = &role.ID
				}
				// Jika role "Mahasiswa" tidak ditemukan, role_id akan tetap null
			}

			// Register user
			registeredUser, err := authService.Register(user, req.Password)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"error":   false,
				"message": "Registrasi berhasil",
				"user": fiber.Map{
					"id":        registeredUser.ID,
					"username":  registeredUser.Username,
					"email":     registeredUser.Email,
					"full_name": registeredUser.FullName,
					"role_id":   registeredUser.RoleID,
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

	// API routes (protected)
	api := app.Group("/api", middleware.JWTMiddleware(jwtSecret, jwtExpiry))
	{
		api.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"message": "Protected route berhasil diakses",
			})
		})
	}
}
