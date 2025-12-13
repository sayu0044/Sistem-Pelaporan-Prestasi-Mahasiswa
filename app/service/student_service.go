package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type StudentService interface {
	GetAllStudents() ([]model.Student, error)
	GetStudentByID(studentID uuid.UUID) (*model.Student, error)
	GetStudentAchievements(ctx context.Context, studentID uuid.UUID) ([]AchievementResponse, error)
	UpdateStudentAdvisor(studentID uuid.UUID, advisorID *uuid.UUID) (*model.Student, error)
	HandleGetAllStudents(c *fiber.Ctx) error
	HandleGetStudentByID(c *fiber.Ctx) error
	HandleGetStudentAchievements(c *fiber.Ctx) error
	HandleUpdateStudentAdvisor(c *fiber.Ctx) error
}

type studentService struct {
	studentRepo       repository.StudentRepository
	lecturerRepo      repository.LecturerRepository
	achievementRepo   repository.AchievementRepository
}

func NewStudentService() StudentService {
	return &studentService{
		studentRepo:     repository.NewStudentRepository(),
		lecturerRepo:    repository.NewLecturerRepository(),
		achievementRepo: repository.NewAchievementRepository(),
	}
}

func (s *studentService) GetAllStudents() ([]model.Student, error) {
	students, err := s.studentRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data mahasiswa: %v", err)
	}
	return students, nil
}

func (s *studentService) GetStudentByID(studentID uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}
	return student, nil
}

func (s *studentService) GetStudentAchievements(ctx context.Context, studentID uuid.UUID) ([]AchievementResponse, error) {
	// Validasi student exists
	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}

	// Get achievements dari MongoDB
	achievements, err := s.achievementRepo.FindAchievementsByStudentID(ctx, student.ID.String())
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data prestasi: %v", err)
	}

	// Get references dari PostgreSQL
	references, err := s.achievementRepo.FindReferencesByStudentIDs([]uuid.UUID{student.ID})
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data reference: %v", err)
	}

	// Create map untuk lookup reference by mongo ID
	refMap := make(map[string]*model.AchievementReference)
	for i := range references {
		ref := &references[i]
		refMap[ref.MongoAchievementID] = ref
	}

	// Map achievements ke response format
	var result []AchievementResponse
	for _, achievement := range achievements {
		reference := refMap[achievement.ID.Hex()]
		response := s.mapToAchievementResponse(ctx, &achievement, reference, student)
		result = append(result, *response)
	}

	return result, nil
}

func (s *studentService) UpdateStudentAdvisor(studentID uuid.UUID, advisorID *uuid.UUID) (*model.Student, error) {
	// Get student
	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}

	// Validasi advisor jika diberikan
	if advisorID != nil {
		_, err := s.lecturerRepo.FindByID(*advisorID)
		if err != nil {
			return nil, errors.New("dosen wali tidak ditemukan")
		}
	}

	// Update advisor
	student.AdvisorID = advisorID

	// Save update
	if err := s.studentRepo.Update(student); err != nil {
		return nil, fmt.Errorf("gagal mengupdate dosen wali: %v", err)
	}

	// Reload student dengan data terbaru
	updatedStudent, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, errors.New("gagal memuat data mahasiswa")
	}

	return updatedStudent, nil
}

func (s *studentService) HandleGetAllStudents(c *fiber.Ctx) error {
	students, err := s.GetAllStudents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Format response
	var studentsData []fiber.Map
	for _, student := range students {
		studentData := s.formatStudentResponse(&student)
		studentsData = append(studentsData, studentData)
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  studentsData,
		"total": len(studentsData),
	})
}

func (s *studentService) HandleGetStudentByID(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Student ID tidak valid",
		})
	}

	student, err := s.GetStudentByID(studentID)
	if err != nil {
		if err.Error() == "mahasiswa tidak ditemukan" {
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
		"data":  s.formatStudentResponse(student),
	})
}

func (s *studentService) HandleGetStudentAchievements(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Student ID tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	achievements, err := s.GetStudentAchievements(ctx, studentID)
	if err != nil {
		if err.Error() == "mahasiswa tidak ditemukan" {
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
		"data":  achievements,
		"total": len(achievements),
	})
}

func (s *studentService) HandleUpdateStudentAdvisor(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Student ID tidak valid",
		})
	}

	var req struct {
		AdvisorID string `json:"advisor_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Parse advisor_id (bisa null jika kosong)
	var advisorID *uuid.UUID
	if req.AdvisorID != "" {
		parsedUUID, err := uuid.Parse(req.AdvisorID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Advisor ID tidak valid",
			})
		}
		advisorID = &parsedUUID
	}

	student, err := s.UpdateStudentAdvisor(studentID, advisorID)
	if err != nil {
		if err.Error() == "mahasiswa tidak ditemukan" || err.Error() == "dosen wali tidak ditemukan" {
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
		"error":   false,
		"message": "Dosen wali berhasil diupdate",
		"data":    s.formatStudentResponse(student),
	})
}

// Helper function untuk format student response
func (s *studentService) formatStudentResponse(student *model.Student) fiber.Map {
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

// Helper function untuk map achievement response (reuse dari achievement service)
func (s *studentService) mapToAchievementResponse(ctx context.Context, achievement *model.Achievement, reference *model.AchievementReference, student *model.Student) *AchievementResponse {
	var studentInfo *StudentInfo
	if student != nil {
		studentInfo = &StudentInfo{
			ID:           student.ID.String(),
			StudentID:    student.StudentID,
			FullName:     student.User.FullName,
			ProgramStudy: student.ProgramStudy,
			AcademicYear: student.AcademicYear,
		}
	}

	var refInfo *AchievementReferenceInfo
	if reference != nil {
		var verifiedBy *string
		if reference.VerifiedBy != nil {
			verifiedByStr := reference.VerifiedBy.String()
			verifiedBy = &verifiedByStr
		}

		refInfo = &AchievementReferenceInfo{
			ID:            reference.ID.String(),
			Status:        reference.Status,
			SubmittedAt:   reference.SubmittedAt,
			VerifiedAt:    reference.VerifiedAt,
			VerifiedBy:    verifiedBy,
			RejectionNote: reference.RejectionNote,
		}
	}

	return &AchievementResponse{
		ID:              achievement.ID.Hex(),
		StudentID:       achievement.StudentID,
		Student:         studentInfo,
		AchievementType: achievement.AchievementType,
		Title:           achievement.Title,
		Description:     achievement.Description,
		Details:         achievement.Details,
		Attachments:     achievement.Attachments,
		Tags:            achievement.Tags,
		Points:          achievement.Points,
		Status:          achievement.Status,
		Reference:       refInfo,
		CreatedAt:       achievement.CreatedAt,
		UpdatedAt:       achievement.UpdatedAt,
	}
}

