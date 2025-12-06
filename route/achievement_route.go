package route

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
)

// RegisterAchievementRoutes mendaftarkan route untuk achievement management
func RegisterAchievementRoutes(router fiber.Router, achievementService service.AchievementService) {
	v1 := router.Group("/v1")
	achievements := v1.Group("/achievements")
	{
		// GET /api/v1/achievements - List (filtered by role)
		achievements.Get("/", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			// Get pagination params
			page, _ := strconv.Atoi(c.Query("page", "1"))
			limit, _ := strconv.Atoi(c.Query("limit", "10"))
			status := c.Query("status", "")

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.GetAchievements(ctx, userID, page, limit, status)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  result,
			})
		})

		// GET /api/v1/achievements/:id - Detail
		achievements.Get("/:id", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.GetAchievementByID(ctx, userID, achievementID)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  result,
			})
		})

		// POST /api/v1/achievements - Create (Mahasiswa) dengan support file upload
		achievements.Post("/", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			// Parse form data (support multipart/form-data untuk file upload)
			var req service.CreateAchievementRequest

			// Cek apakah request adalah multipart/form-data atau application/json
			contentType := c.Get("Content-Type")
			if strings.Contains(contentType, "multipart/form-data") {
				// Parse dari form data
				req.AchievementType = model.AchievementType(c.FormValue("achievement_type"))
				req.Title = c.FormValue("title")
				req.Description = c.FormValue("description")

				// Parse details dari JSON string (jika dikirim sebagai string)
				detailsStr := c.FormValue("details")
				if detailsStr != "" {
					if err := json.Unmarshal([]byte(detailsStr), &req.Details); err != nil {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error":   true,
							"message": "Invalid details format. Details harus berupa JSON string",
						})
					}
				}

				// Parse tags dari comma-separated string atau JSON array
				tagsStr := c.FormValue("tags")
				if tagsStr != "" {
					if err := json.Unmarshal([]byte(tagsStr), &req.Tags); err != nil {
						// Jika bukan JSON, coba split by comma
						req.Tags = strings.Split(tagsStr, ",")
					}
				}

				// Parse points
				if pointsStr := c.FormValue("points"); pointsStr != "" {
					if points, err := strconv.ParseFloat(pointsStr, 64); err == nil {
						req.Points = points
					}
				}
			} else {
				// Parse dari JSON body
				if err := c.BodyParser(&req); err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   true,
						"message": "Invalid request body",
					})
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Create achievement
			result, err := achievementService.CreateAchievement(ctx, userID, &req)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			// Handle file uploads jika ada
			if strings.Contains(contentType, "multipart/form-data") {
				form, err := c.MultipartForm()
				if err == nil && form.File != nil {
					files := form.File["files"] // Support multiple files dengan key "files"
					if len(files) == 0 {
						files = form.File["file"] // Fallback ke single file dengan key "file"
					}

					for _, file := range files {
						// Validate file
						if file.Size > 10*1024*1024 { // 10MB limit
							continue // Skip file yang terlalu besar
						}

						allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx"}
						ext := strings.ToLower(filepath.Ext(file.Filename))
						allowed := false
						for _, allowedExt := range allowedExtensions {
							if ext == allowedExt {
								allowed = true
								break
							}
						}
						if !allowed {
							continue // Skip file yang tidak diizinkan
						}

						// Create upload directory
						uploadDir := "uploads/achievements"
						if err := os.MkdirAll(uploadDir, 0755); err != nil {
							continue
						}

						// Generate unique filename
						filename := result.ID + "_" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + file.Filename
						filePath := filepath.Join(uploadDir, filename)

						// Save file
						if err := c.SaveFile(file, filePath); err != nil {
							continue
						}

						// Update achievement dengan attachment
						_, err = achievementService.UploadAttachment(ctx, userID, result.ID, filePath)
						if err != nil {
							// Rollback: delete file
							os.Remove(filePath)
						}
					}

					// Reload achievement dengan attachments terbaru
					result, _ = achievementService.GetAchievementByID(ctx, userID, result.ID)
				}
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"error":   false,
				"message": "Achievement berhasil dibuat",
				"data":    result,
			})
		})

		// PUT /api/v1/achievements/:id - Update (Mahasiswa)
		achievements.Put("/:id", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			var req service.UpdateAchievementRequest
			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid request body",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.UpdateAchievement(ctx, userID, achievementID, &req)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Achievement berhasil diupdate",
				"data":    result,
			})
		})

		// DELETE /api/v1/achievements/:id - Delete (Mahasiswa)
		achievements.Delete("/:id", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = achievementService.DeleteAchievement(ctx, userID, achievementID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Achievement berhasil dihapus",
			})
		})

		// POST /api/v1/achievements/:id/submit - Submit for verification
		achievements.Post("/:id/submit", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.SubmitAchievement(ctx, userID, achievementID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Achievement berhasil disubmit untuk verifikasi",
				"data":    result,
			})
		})

		// POST /api/v1/achievements/:id/verify - Verify (Dosen Wali)
		achievements.Post("/:id/verify", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.VerifyAchievement(ctx, userID, achievementID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Achievement berhasil diverifikasi",
				"data":    result,
			})
		})

		// POST /api/v1/achievements/:id/reject - Reject (Dosen Wali)
		achievements.Post("/:id/reject", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			var req struct {
				RejectionNote string `json:"rejection_note"`
			}
			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid request body",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.RejectAchievement(ctx, userID, achievementID, req.RejectionNote)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Achievement ditolak",
				"data":    result,
			})
		})

		// GET /api/v1/achievements/:id/history - Status history
		achievements.Get("/:id/history", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := achievementService.GetAchievementHistory(ctx, userID, achievementID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  result,
			})
		})

		// POST /api/v1/achievements/:id/attachments - Upload files
		achievements.Post("/:id/attachments", func(c *fiber.Ctx) error {
			userID, err := getUserIDFromContext(c)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			achievementID := c.Params("id")
			if achievementID == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "achievement ID harus diisi",
				})
			}

			// Get file from form
			file, err := c.FormFile("file")
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "File tidak ditemukan. Gunakan form field 'file'",
				})
			}

			// Validate file
			if file.Size > 10*1024*1024 { // 10MB limit
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Ukuran file maksimal 10MB",
				})
			}

			allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx"}
			ext := strings.ToLower(filepath.Ext(file.Filename))
			allowed := false
			for _, allowedExt := range allowedExtensions {
				if ext == allowedExt {
					allowed = true
					break
				}
			}

			if !allowed {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Tipe file tidak diizinkan. Hanya PDF, JPG, PNG, DOC, DOCX",
				})
			}

			// Create upload directory
			uploadDir := "uploads/achievements"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": "Gagal membuat directory",
				})
			}

			// Generate unique filename
			filename := achievementID + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ext
			filePath := filepath.Join(uploadDir, filename)

			// Save file
			if err := c.SaveFile(file, filePath); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": "Gagal menyimpan file: " + err.Error(),
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Update achievement with attachment
			result, err := achievementService.UploadAttachment(ctx, userID, achievementID, filePath)
			if err != nil {
				// Rollback: delete file
				os.Remove(filePath)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "File berhasil diupload",
				"data": fiber.Map{
					"file_path": result,
				},
			})
		})
	}
}

// Helper function to get user ID from context
func getUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User ID tidak ditemukan")
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User ID tidak valid")
	}

	return userID, nil
}
