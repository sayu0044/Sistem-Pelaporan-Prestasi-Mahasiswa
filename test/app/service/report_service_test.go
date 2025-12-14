package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/stretchr/testify/assert"
)

func TestReportService_GetStatistics(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:     userID,
		RoleID: &roleID,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}
	studentID := uuid.New()
	student := &model.Student{
		ID:       studentID,
		UserID:   userID,
		StudentID: "1234567890",
		User:     *user,
	}

	stats := &repository.AchievementStatistics{
		TotalByType: map[string]int64{
			"competition": 5,
			"academic":    3,
		},
		TotalByPeriod: []repository.PeriodStat{
			{Period: "2024-01", Count: 2},
			{Period: "2024-02", Count: 6},
		},
		TopStudents: []repository.TopStudentStat{
			{StudentID: studentID.String(), TotalPoints: 100.0, TotalAchievements: 8},
		},
		CompetitionLevelDistribution: map[string]int64{
			"national": 3,
			"local":    2,
		},
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMocks func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository)
		wantErr    bool
	}{
		{
			name:   "Success get statistics for student",
			userID: userID,
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					GetAchievementStatisticsFunc: func(ctx context.Context, studentIDs []string) (*repository.AchievementStatistics, error) {
						return stats, nil
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return user, nil
					},
				}
				roleRepo := &MockRoleRepository{
					FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
						return role, nil
					},
				}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name:   "Error user not found",
			userID: uuid.New(),
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{}
				studentRepo := &MockStudentRepository{}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return nil, errors.New("not found")
					},
				}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo := tt.setupMocks()
			reportService := service.NewReportService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

			result, err := reportService.GetStatistics(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestReportService_GetStudentStatistics(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:     userID,
		RoleID: &roleID,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}
	studentID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    userID,
		StudentID: "1234567890",
		User:      *user,
	}

	stats := &repository.AchievementStatistics{
		TotalByType: map[string]int64{
			"competition": 5,
		},
		TotalByPeriod: []repository.PeriodStat{
			{Period: "2024-01", Count: 5},
		},
		CompetitionLevelDistribution: map[string]int64{
			"national": 5,
		},
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		studentID  uuid.UUID
		setupMocks func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository)
		wantErr    bool
	}{
		{
			name:      "Success get student statistics",
			userID:    userID,
			studentID: studentID,
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					GetAchievementStatisticsFunc: func(ctx context.Context, studentIDs []string) (*repository.AchievementStatistics, error) {
						return stats, nil
					},
					FindAchievementsByStudentIDFunc: func(ctx context.Context, studentID string) ([]model.Achievement, error) {
						return []model.Achievement{}, nil
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return student, nil
					},
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return user, nil
					},
				}
				roleRepo := &MockRoleRepository{
					FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
						return role, nil
					},
				}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name:      "Error student not found",
			userID:    userID,
			studentID: uuid.New(),
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{}
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return nil, errors.New("not found")
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return user, nil
					},
				}
				roleRepo := &MockRoleRepository{
					FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
						return role, nil
					},
				}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo := tt.setupMocks()
			reportService := service.NewReportService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

			result, err := reportService.GetStudentStatistics(ctx, tt.userID, tt.studentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

