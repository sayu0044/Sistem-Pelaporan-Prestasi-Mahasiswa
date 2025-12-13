package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterLecturerRoutes mendaftarkan route untuk lecturer management
func RegisterLecturerRoutes(router fiber.Router, lecturerService service.LecturerService) {
	lecturers := router.Group("/lecturers")
	{
		// GET /api/v1/lecturers - List semua dosen
		// Requires: read lecturers permission
		lecturers.Get("/", middleware.RBACMiddleware("read", "lecturers"), lecturerService.HandleGetAllLecturers)

		// GET /api/v1/lecturers/:id/advisees - Mahasiswa bimbingan dosen
		// Requires: read lecturers permission
		lecturers.Get("/:id/advisees", middleware.RBACMiddleware("read", "lecturers"), lecturerService.HandleGetLecturerAdvisees)
	}
}
