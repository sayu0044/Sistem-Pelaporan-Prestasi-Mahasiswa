package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterStudentRoutes mendaftarkan route untuk student management
func RegisterStudentRoutes(router fiber.Router, studentService service.StudentService) {
	students := router.Group("/students")
	{
		// GET /api/v1/students - List semua mahasiswa
		// Requires: read students permission
		students.Get("/", middleware.RBACMiddleware("read", "students"), studentService.HandleGetAllStudents)

		// GET /api/v1/students/:id - Detail mahasiswa
		// Requires: read students permission
		students.Get("/:id", middleware.RBACMiddleware("read", "students"), studentService.HandleGetStudentByID)

		// GET /api/v1/students/:id/achievements - Prestasi mahasiswa
		// Requires: read students permission
		students.Get("/:id/achievements", middleware.RBACMiddleware("read", "students"), studentService.HandleGetStudentAchievements)

		// PUT /api/v1/students/:id/advisor - Update dosen wali
		// Requires: update students permission
		students.Put("/:id/advisor", middleware.RBACMiddleware("update", "students"), studentService.HandleUpdateStudentAdvisor)
	}
}
