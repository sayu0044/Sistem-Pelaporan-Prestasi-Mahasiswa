package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

func RegisterLecturerRoutes(router fiber.Router, lecturerService service.LecturerService) {
	lecturers := router.Group("/lecturers")
	{
		lecturers.Get("/", middleware.RBACMiddleware("read", "lecturers"), func(c *fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			lecturers, err := lecturerService.GetAllLecturers(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal mengambil data",
					"message": err.Error(),
				})
			}

			var lecturersData []fiber.Map
			for _, lecturer := range lecturers {
				lecturerData := formatLecturerResponse(&lecturer)
				lecturersData = append(lecturersData, lecturerData)
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  lecturersData,
				"total": len(lecturersData),
			})
		})

		lecturers.Get("/:id/advisees", middleware.RBACMiddleware("read", "lecturers"), func(c *fiber.Ctx) error {
			lecturerIDStr := c.Params("id")
			lecturerID, err := uuid.Parse(lecturerIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Lecturer ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			advisees, err := lecturerService.GetLecturerAdvisees(ctx, lecturerID)
			if err != nil {
				if err.Error() == "dosen tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "Gagal mengambil data",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal mengambil data",
					"message": err.Error(),
				})
			}

			var adviseesData []fiber.Map
			for _, advisee := range advisees {
				adviseeData := formatStudentResponse(&advisee)
				adviseesData = append(adviseesData, adviseeData)
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  adviseesData,
				"total": len(adviseesData),
			})
		})
	}
}

func formatLecturerResponse(lecturer *model.Lecturer) fiber.Map {
	response := fiber.Map{
		"id":          lecturer.ID,
		"user_id":     lecturer.UserID,
		"lecturer_id": lecturer.LecturerID,
		"department":  lecturer.Department,
		"created_at":  lecturer.CreatedAt,
	}

	if lecturer.User.ID != uuid.Nil {
		response["user"] = fiber.Map{
			"id":        lecturer.User.ID,
			"username":  lecturer.User.Username,
			"email":     lecturer.User.Email,
			"full_name": lecturer.User.FullName,
			"role_id":   lecturer.User.RoleID,
		}
	}

	return response
}
