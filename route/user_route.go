package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

func RegisterUserRoutes(router fiber.Router, userService service.UserService) {
	users := router.Group("/users")
	{
		users.Get("/", middleware.RBACMiddleware("read", "users"), func(c *fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			users, err := userService.GetAllUsers(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal mengambil data",
					"message": err.Error(),
				})
			}

			var usersData []fiber.Map
			for _, user := range users {
				var roleData fiber.Map
				var permissions []fiber.Map

				if user.RoleID != nil && user.Role.ID != uuid.Nil {
					roleData = fiber.Map{
						"id":          user.Role.ID,
						"name":        user.Role.Name,
						"description": user.Role.Description,
					}

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
				}

				usersData = append(usersData, fiber.Map{
					"id":          user.ID,
					"username":    user.Username,
					"email":       user.Email,
					"full_name":   user.FullName,
					"role_id":     user.RoleID,
					"role":        roleData,
					"permissions": permissions,
					"is_active":   user.IsActive,
					"created_at":  user.CreatedAt,
					"updated_at":  user.UpdatedAt,
				})
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  usersData,
				"total": len(usersData),
			})
		})

		users.Get("/:id", middleware.RBACMiddleware("read", "users"), func(c *fiber.Ctx) error {
			userIDStr := c.Params("id")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "User ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			user, role, err := userService.GetUserByID(ctx, userID)
			if err != nil {
				if err.Error() == "user tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "Gagal mengambil pengguna",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal mengambil data",
					"message": err.Error(),
				})
			}

			var permissions []fiber.Map
			if role != nil && len(role.Permissions) > 0 {
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

			var roleData fiber.Map
			if role != nil {
				roleData = fiber.Map{
					"id":          role.ID,
					"name":        role.Name,
					"description": role.Description,
				}
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data": fiber.Map{
					"id":          user.ID,
					"username":    user.Username,
					"email":       user.Email,
					"full_name":   user.FullName,
					"role_id":     user.RoleID,
					"role":        roleData,
					"permissions": permissions,
					"is_active":   user.IsActive,
					"created_at":  user.CreatedAt,
					"updated_at":  user.UpdatedAt,
				},
			})
		})

		users.Post("/", middleware.RBACMiddleware("create", "users"), func(c *fiber.Ctx) error {
			var req struct {
				Username     string `json:"username"`
				Email        string `json:"email"`
				Password     string `json:"password"`
				FullName     string `json:"full_name"`
				RoleID       string `json:"role_id"`
				IsActive     *bool  `json:"is_active"`
				LecturerID   string `json:"lecturer_id"`
				Department   string `json:"department"`
				StudentID    string `json:"student_id"`
				ProgramStudy string `json:"program_study"`
				AcademicYear string `json:"academic_year"`
				AdvisorID    string `json:"advisor_id"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Pastikan body permintaan Anda dalam format JSON yang benar.",
				})
			}

			if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Username, email, password, dan full_name harus diisi",
				})
			}

			if req.RoleID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "role_id harus diisi",
				})
			}

			roleUUID, err := uuid.Parse(req.RoleID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "role_id tidak valid",
				})
			}

			isActive := true
			if req.IsActive != nil {
				isActive = *req.IsActive
			}

			var advisorUUID *uuid.UUID
			if req.AdvisorID != "" {
				parsedUUID, err := uuid.Parse(req.AdvisorID)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   "Permintaan tidak valid",
						"message": "advisor_id tidak valid",
					})
				}
				advisorUUID = &parsedUUID
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			user, role, err := userService.CreateUser(
				ctx,
				req.Username,
				req.Email,
				req.Password,
				req.FullName,
				roleUUID,
				isActive,
				req.LecturerID,
				req.Department,
				req.StudentID,
				req.ProgramStudy,
				req.AcademicYear,
				advisorUUID,
			)
			if err != nil {
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error":   "Gagal membuat pengguna",
					"message": err.Error(),
				})
			}

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

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"error": false,
				"message": "User berhasil dibuat",
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
				},
			})
		})

		users.Put("/:id", middleware.RBACMiddleware("update", "users"), func(c *fiber.Ctx) error {
			userIDStr := c.Params("id")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "User ID tidak valid",
				})
			}

			var req struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				FullName string `json:"full_name"`
				RoleID   string `json:"role_id"`
				IsActive *bool  `json:"is_active"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Pastikan body permintaan Anda dalam format JSON yang benar.",
				})
			}

			var roleUUID *uuid.UUID
			if req.RoleID != "" {
				parsedUUID, err := uuid.Parse(req.RoleID)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   "Permintaan tidak valid",
						"message": "role_id tidak valid",
					})
				}
				roleUUID = &parsedUUID
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			user, role, err := userService.UpdateUser(ctx, userID, req.Username, req.Email, req.FullName, roleUUID, req.IsActive)
			if err != nil {
				if err.Error() == "user tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "Gagal mengupdate pengguna",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error":   "Gagal mengupdate pengguna",
					"message": err.Error(),
				})
			}

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
				"error": false,
				"message": "User berhasil diupdate",
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
					"updated_at":  user.UpdatedAt,
				},
			})
		})

		users.Delete("/:id", middleware.RBACMiddleware("delete", "users"), func(c *fiber.Ctx) error {
			userIDStr := c.Params("id")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "User ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = userService.DeleteUser(ctx, userID)
			if err != nil {
				if err.Error() == "user tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "Gagal menghapus pengguna",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal menghapus pengguna",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "User berhasil dihapus",
			})
		})

		users.Put("/:id/role", middleware.RBACMiddleware("update", "users"), func(c *fiber.Ctx) error {
			userIDStr := c.Params("id")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "User ID tidak valid",
				})
			}

			var req struct {
				RoleID string `json:"role_id"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Pastikan body permintaan Anda dalam format JSON yang benar.",
				})
			}

			if req.RoleID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "role_id harus diisi",
				})
			}

			roleUUID, err := uuid.Parse(req.RoleID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "role_id tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			user, role, err := userService.UpdateUserRole(ctx, userID, roleUUID)
			if err != nil {
				if err.Error() == "user tidak ditemukan" || err.Error() == "role tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "Gagal mengupdate role",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error":   "Gagal mengupdate role",
					"message": err.Error(),
				})
			}

			var permissions []fiber.Map
			if role != nil && len(role.Permissions) > 0 {
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

			var roleData fiber.Map
			if role != nil {
				roleData = fiber.Map{
					"id":          role.ID,
					"name":        role.Name,
					"description": role.Description,
				}
			}

			return c.JSON(fiber.Map{
				"error": false,
				"message": "Role user berhasil diupdate",
				"data": fiber.Map{
					"id":          user.ID,
					"username":    user.Username,
					"email":       user.Email,
					"full_name":   user.FullName,
					"role_id":     user.RoleID,
					"role":        roleData,
					"permissions": permissions,
					"is_active":   user.IsActive,
					"updated_at":  user.UpdatedAt,
				},
			})
		})
	}
}
