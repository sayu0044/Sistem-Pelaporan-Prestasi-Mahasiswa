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

func RegisterStudentRoutes(router fiber.Router, studentService service.StudentService) {
	students := router.Group("/students")
	{
		students.Get("/", middleware.RBACMiddleware("read", "students"), func(c *fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			students, err := studentService.GetAllStudents(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal mengambil data",
					"message": err.Error(),
				})
			}

			var studentsData []fiber.Map
			for _, student := range students {
				studentData := formatStudentResponse(&student)
				studentsData = append(studentsData, studentData)
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data":  studentsData,
				"total": len(studentsData),
			})
		})

		students.Get("/:id", middleware.RBACMiddleware("read", "students"), func(c *fiber.Ctx) error {
			studentIDStr := c.Params("id")
			studentID, err := uuid.Parse(studentIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Student ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			student, err := studentService.GetStudentByID(ctx, studentID)
			if err != nil {
				if err.Error() == "mahasiswa tidak ditemukan" {
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

			return c.JSON(fiber.Map{
				"error": false,
				"data":  formatStudentResponse(student),
			})
		})

		students.Get("/:id/achievements", middleware.RBACMiddleware("read", "students"), func(c *fiber.Ctx) error {
			studentIDStr := c.Params("id")
			studentID, err := uuid.Parse(studentIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Student ID tidak valid",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			achievements, err := studentService.GetStudentAchievements(ctx, studentID)
			if err != nil {
				if err.Error() == "mahasiswa tidak ditemukan" {
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

			return c.JSON(fiber.Map{
				"error": false,
				"data":  achievements,
				"total": len(achievements),
			})
		})

		students.Put("/:id/advisor", middleware.RBACMiddleware("update", "students"), func(c *fiber.Ctx) error {
			studentIDStr := c.Params("id")
			studentID, err := uuid.Parse(studentIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Student ID tidak valid",
				})
			}

			var req struct {
				AdvisorID string `json:"advisor_id"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Pastikan body permintaan Anda dalam format JSON yang benar.",
				})
			}

			var advisorID *uuid.UUID
			if req.AdvisorID != "" {
				parsedUUID, err := uuid.Parse(req.AdvisorID)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   "Permintaan tidak valid",
						"message": "Advisor ID tidak valid",
					})
				}
				advisorID = &parsedUUID
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			student, err := studentService.UpdateStudentAdvisor(ctx, studentID, advisorID)
			if err != nil {
				if err.Error() == "mahasiswa tidak ditemukan" || err.Error() == "dosen wali tidak ditemukan" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error":   "Gagal mengupdate data",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Gagal mengupdate data",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error":   false,
				"message": "Dosen wali berhasil diupdate",
				"data":    formatStudentResponse(student),
			})
		})
	}
}

func formatStudentResponse(student *model.Student) fiber.Map {
	response := fiber.Map{
		"id":            student.ID,
		"user_id":       student.UserID,
		"student_id":    student.StudentID,
		"program_study": student.ProgramStudy,
		"academic_year": student.AcademicYear,
		"created_at":    student.CreatedAt,
	}

	if student.User.ID != uuid.Nil {
		response["user"] = fiber.Map{
			"id":        student.User.ID,
			"username":  student.User.Username,
			"email":     student.User.Email,
			"full_name": student.User.FullName,
			"role_id":   student.User.RoleID,
		}
	}

	if student.AdvisorID != nil {
		response["advisor_id"] = student.AdvisorID
		if student.Advisor.ID != uuid.Nil {
			response["advisor"] = fiber.Map{
				"id":           student.Advisor.ID,
				"lecturer_id":  student.Advisor.LecturerID,
				"department":   student.Advisor.Department,
				"created_at":   student.Advisor.CreatedAt,
			}
			if student.Advisor.User.ID != uuid.Nil {
				response["advisor"].(fiber.Map)["user"] = fiber.Map{
					"id":        student.Advisor.User.ID,
					"username":  student.Advisor.User.Username,
					"email":     student.Advisor.User.Email,
					"full_name": student.Advisor.User.FullName,
				}
			}
		}
	}

	return response
}
