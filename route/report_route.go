package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

func RegisterReportRoutes(router fiber.Router, reportService service.ReportService) {
	reports := router.Group("/reports")
	{
		reports.Get("/statistics", middleware.RBACMiddleware("read", "achievements"), func(c *fiber.Ctx) error {
			userIDInterface := c.Locals("user_id")
			if userIDInterface == nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "User ID tidak ditemukan",
				})
			}

			userID, ok := userIDInterface.(uuid.UUID)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "User ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			statistics, err := reportService.GetStatistics(ctx, userID)
			if err != nil {
				if err.Error() == "user tidak ditemukan" || err.Error() == "role tidak ditemukan" {
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
				"error": false,
				"data":  statistics,
			})
		})

		reports.Get("/student/:id", middleware.RBACMiddleware("read", "achievements"), func(c *fiber.Ctx) error {
			userIDInterface := c.Locals("user_id")
			if userIDInterface == nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "User ID tidak ditemukan",
				})
			}

			userID, ok := userIDInterface.(uuid.UUID)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "User ID tidak valid",
				})
			}

			studentIDParam := c.Params("id")
			studentID, err := uuid.Parse(studentIDParam)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Student ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			statistics, err := reportService.GetStudentStatistics(ctx, userID, studentID)
			if err != nil {
				if err.Error() == "mahasiswa tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   true,
						"message": err.Error(),
					})
				}
				if err.Error() == "anda hanya dapat melihat statistik prestasi sendiri" || 
				   err.Error() == "anda bukan dosen wali dari mahasiswa ini" {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
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
				"error": false,
				"data":  statistics,
			})
		})
	}
}

