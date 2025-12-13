package service

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type LecturerService interface {
	GetAllLecturers() ([]model.Lecturer, error)
	GetLecturerAdvisees(lecturerID uuid.UUID) ([]model.Student, error)
	HandleGetAllLecturers(c *fiber.Ctx) error
	HandleGetLecturerAdvisees(c *fiber.Ctx) error
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
	studentRepo  repository.StudentRepository
}

func NewLecturerService() LecturerService {
	return &lecturerService{
		lecturerRepo: repository.NewLecturerRepository(),
		studentRepo:  repository.NewStudentRepository(),
	}
}

func (s *lecturerService) GetAllLecturers() ([]model.Lecturer, error) {
	lecturers, err := s.lecturerRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data dosen: %v", err)
	}
	return lecturers, nil
}

func (s *lecturerService) GetLecturerAdvisees(lecturerID uuid.UUID) ([]model.Student, error) {
	// Validasi lecturer exists
	_, err := s.lecturerRepo.FindByID(lecturerID)
	if err != nil {
		return nil, errors.New("dosen tidak ditemukan")
	}

	// Get advisees
	advisees, err := s.lecturerRepo.FindAdvisees(lecturerID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data mahasiswa bimbingan: %v", err)
	}

	return advisees, nil
}

func (s *lecturerService) HandleGetAllLecturers(c *fiber.Ctx) error {
	lecturers, err := s.GetAllLecturers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Format response
	var lecturersData []fiber.Map
	for _, lecturer := range lecturers {
		lecturerData := s.formatLecturerResponse(&lecturer)
		lecturersData = append(lecturersData, lecturerData)
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  lecturersData,
		"total": len(lecturersData),
	})
}

func (s *lecturerService) HandleGetLecturerAdvisees(c *fiber.Ctx) error {
	lecturerIDStr := c.Params("id")
	lecturerID, err := uuid.Parse(lecturerIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Lecturer ID tidak valid",
		})
	}

	advisees, err := s.GetLecturerAdvisees(lecturerID)
	if err != nil {
		if err.Error() == "dosen tidak ditemukan" {
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

	// Format response
	var adviseesData []fiber.Map
	for _, advisee := range advisees {
		adviseeData := s.formatStudentResponse(&advisee)
		adviseesData = append(adviseesData, adviseeData)
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  adviseesData,
		"total": len(adviseesData),
	})
}

// Helper function untuk format lecturer response
func (s *lecturerService) formatLecturerResponse(lecturer *model.Lecturer) fiber.Map {
	response := fiber.Map{
		"id":          lecturer.ID,
		"user_id":     lecturer.UserID,
		"lecturer_id": lecturer.LecturerID,
		"department":   lecturer.Department,
		"created_at":  lecturer.CreatedAt,
	}

	// Add user info jika ada
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

// Helper function untuk format student response (reuse dari student service)
func (s *lecturerService) formatStudentResponse(student *model.Student) fiber.Map {
	response := fiber.Map{
		"id":            student.ID,
		"user_id":       student.UserID,
		"student_id":    student.StudentID,
		"program_study": student.ProgramStudy,
		"academic_year": student.AcademicYear,
		"created_at":    student.CreatedAt,
	}

	// Add user info jika ada
	if student.User.ID != uuid.Nil {
		response["user"] = fiber.Map{
			"id":        student.User.ID,
			"username":  student.User.Username,
			"email":     student.User.Email,
			"full_name": student.User.FullName,
			"role_id":   student.User.RoleID,
		}
	}

	// Add advisor info jika ada
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


