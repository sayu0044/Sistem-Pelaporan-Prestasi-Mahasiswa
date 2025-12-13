package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type ReportService interface {
	GetStatistics(ctx context.Context, userID uuid.UUID) (*StatisticsResponse, error)
	GetStudentStatistics(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*StudentStatisticsResponse, error)
}

type reportService struct {
	achievementRepo repository.AchievementRepository
	studentRepo     repository.StudentRepository
	lecturerRepo    repository.LecturerRepository
	userRepo        repository.UserRepository
	roleRepo        repository.RoleRepository
}

func NewReportService(
	achievementRepo repository.AchievementRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) ReportService {
	return &reportService{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
		lecturerRepo:    lecturerRepo,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
	}
}

type StatisticsResponse struct {
	TotalByType                  map[string]int64         `json:"total_by_type"`
	TotalByPeriod                []PeriodStatResponse     `json:"total_by_period"`
	TopStudents                  []TopStudentStatResponse `json:"top_students"`
	CompetitionLevelDistribution map[string]int64         `json:"competition_level_distribution"`
}

type PeriodStatResponse struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

type TopStudentStatResponse struct {
	StudentID         string  `json:"student_id"`
	StudentName       string  `json:"student_name,omitempty"`
	TotalPoints       float64 `json:"total_points"`
	TotalAchievements int64   `json:"total_achievements"`
}

type StudentStatisticsResponse struct {
	StudentID                    string               `json:"student_id"`
	StudentName                  string               `json:"student_name"`
	TotalByType                  map[string]int64     `json:"total_by_type"`
	TotalByPeriod                []PeriodStatResponse `json:"total_by_period"`
	CompetitionLevelDistribution map[string]int64     `json:"competition_level_distribution"`
	TotalPoints                  float64              `json:"total_points"`
	TotalAchievements            int64                `json:"total_achievements"`
}

func (s *reportService) checkRole(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	if user.RoleID == nil {
		return "", errors.New("user tidak memiliki role")
	}

	role, err := s.roleRepo.FindRoleByID(ctx, *user.RoleID)
	if err != nil {
		return "", errors.New("role tidak ditemukan")
	}

	roleNameLower := strings.ToLower(role.Name)
	if strings.Contains(roleNameLower, "mahasiswa") || strings.Contains(roleNameLower, "student") {
		return "student", nil
	}
	if strings.Contains(roleNameLower, "dosen") || strings.Contains(roleNameLower, "lecturer") {
		return "lecturer", nil
	}
	if strings.Contains(roleNameLower, "admin") {
		return "admin", nil
	}

	return "", errors.New("role tidak dikenali")
}

func (s *reportService) getStudentIDsByRole(ctx context.Context, userID uuid.UUID) ([]string, error) {
	role, err := s.checkRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	var studentIDs []string

	if role == "student" {
		student, err := s.studentRepo.FindStudentByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("user tidak ditemukan sebagai mahasiswa")
		}
		studentIDs = []string{student.ID.String()}
	} else if role == "lecturer" {
		lecturer, err := s.lecturerRepo.FindLecturerByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("user tidak ditemukan sebagai dosen")
		}

		advisees, err := s.lecturerRepo.FindAdvisees(ctx, lecturer.ID)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat mahasiswa bimbingan: %v", err)
		}

		for _, advisee := range advisees {
			studentIDs = append(studentIDs, advisee.ID.String())
		}
	} else {
		allStudents, err := s.studentRepo.FindAllStudents(ctx)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat students: %v", err)
		}

		for _, student := range allStudents {
			studentIDs = append(studentIDs, student.ID.String())
		}
	}

	return studentIDs, nil
}

func (s *reportService) GetStatistics(ctx context.Context, userID uuid.UUID) (*StatisticsResponse, error) {
	studentIDs, err := s.getStudentIDsByRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.achievementRepo.GetAchievementStatistics(ctx, studentIDs)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil statistik: %v", err)
	}

	response := &StatisticsResponse{
		TotalByType:                  stats.TotalByType,
		TotalByPeriod:                []PeriodStatResponse{},
		TopStudents:                  []TopStudentStatResponse{},
		CompetitionLevelDistribution: stats.CompetitionLevelDistribution,
	}

	for _, period := range stats.TotalByPeriod {
		response.TotalByPeriod = append(response.TotalByPeriod, PeriodStatResponse{
			Period: period.Period,
			Count:  period.Count,
		})
	}

	for _, topStudent := range stats.TopStudents {
		studentUUID, err := uuid.Parse(topStudent.StudentID)
		if err != nil {
			continue
		}

		student, err := s.studentRepo.FindStudentByID(ctx, studentUUID)
		if err != nil {
			continue
		}

		response.TopStudents = append(response.TopStudents, TopStudentStatResponse{
			StudentID:         topStudent.StudentID,
			StudentName:       student.User.FullName,
			TotalPoints:       topStudent.TotalPoints,
			TotalAchievements: topStudent.TotalAchievements,
		})
	}

	return response, nil
}

func (s *reportService) GetStudentStatistics(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*StudentStatisticsResponse, error) {
	role, err := s.checkRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	student, err := s.studentRepo.FindStudentByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("mahasiswa tidak ditemukan")
	}

	if role == "student" {
		currentStudent, err := s.studentRepo.FindStudentByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("user tidak ditemukan sebagai mahasiswa")
		}
		if currentStudent.ID != studentID {
			return nil, errors.New("anda hanya dapat melihat statistik prestasi sendiri")
		}
	} else if role == "lecturer" {
		lecturer, err := s.lecturerRepo.FindLecturerByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("user tidak ditemukan sebagai dosen")
		}
		if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return nil, errors.New("anda bukan dosen wali dari mahasiswa ini")
		}
	}

	stats, err := s.achievementRepo.GetAchievementStatistics(ctx, []string{studentID.String()})
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil statistik: %v", err)
	}

	totalPoints := float64(0)
	totalAchievements := int64(0)

	achievements, err := s.achievementRepo.FindAchievementsByStudentID(ctx, studentID.String())
	if err == nil {
		for _, achievement := range achievements {
			totalPoints += achievement.Points
			totalAchievements++
		}
	}

	response := &StudentStatisticsResponse{
		StudentID:                    studentID.String(),
		StudentName:                  student.User.FullName,
		TotalByType:                  stats.TotalByType,
		TotalByPeriod:                []PeriodStatResponse{},
		CompetitionLevelDistribution: stats.CompetitionLevelDistribution,
		TotalPoints:                  totalPoints,
		TotalAchievements:            totalAchievements,
	}

	for _, period := range stats.TotalByPeriod {
		response.TotalByPeriod = append(response.TotalByPeriod, PeriodStatResponse{
			Period: period.Period,
			Count:  period.Count,
		})
	}

	return response, nil
}
