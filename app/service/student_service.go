package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type StudentService interface {
	GetAllStudents(ctx context.Context) ([]model.Student, error)
	GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error)
	GetStudentAchievements(ctx context.Context, studentID uuid.UUID) ([]AchievementResponse, error)
	UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID *uuid.UUID) (*model.Student, error)
}

type studentService struct {
	studentRepo     repository.StudentRepository
	lecturerRepo    repository.LecturerRepository
	achievementRepo repository.AchievementRepository
}

func NewStudentService(
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	achievementRepo repository.AchievementRepository,
) StudentService {
	return &studentService{
		studentRepo:     studentRepo,
		lecturerRepo:    lecturerRepo,
		achievementRepo: achievementRepo,
	}
}

func (s *studentService) GetAllStudents(ctx context.Context) ([]model.Student, error) {
	students, err := s.studentRepo.FindAllStudents(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data mahasiswa: %v", err)
	}
	return students, nil
}

func (s *studentService) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.FindStudentByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}
	return student, nil
}

func (s *studentService) GetStudentAchievements(ctx context.Context, studentID uuid.UUID) ([]AchievementResponse, error) {
	student, err := s.studentRepo.FindStudentByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}

	achievements, err := s.achievementRepo.FindAchievementsByStudentID(ctx, student.ID.String())
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data prestasi: %v", err)
	}

	references, err := s.achievementRepo.FindReferencesByStudentIDs(ctx, []uuid.UUID{student.ID})
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data reference: %v", err)
	}

	refMap := make(map[string]*model.AchievementReference)
	for i := range references {
		ref := &references[i]
		refMap[ref.MongoAchievementID] = ref
	}

	var result []AchievementResponse
	for _, achievement := range achievements {
		reference := refMap[achievement.ID.Hex()]
		response := s.mapToAchievementResponse(ctx, &achievement, reference, student)
		result = append(result, *response)
	}

	return result, nil
}

func (s *studentService) UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID *uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.FindStudentByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}

	if advisorID != nil {
		_, err := s.lecturerRepo.FindLecturerByID(ctx, *advisorID)
		if err != nil {
			return nil, errors.New("dosen wali tidak ditemukan")
		}
	}

	student.AdvisorID = advisorID

	if err := s.studentRepo.UpdateStudent(ctx, student); err != nil {
		return nil, fmt.Errorf("gagal mengupdate dosen wali: %v", err)
	}

	updatedStudent, err := s.studentRepo.FindStudentByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("gagal memuat data mahasiswa")
	}

	return updatedStudent, nil
}

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
