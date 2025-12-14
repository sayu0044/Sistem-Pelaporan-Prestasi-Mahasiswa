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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockAchievementRepository struct {
	CreateAchievementFunc            func(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error)
	FindAchievementByIDFunc           func(ctx context.Context, id string) (*model.Achievement, error)
	FindAchievementsByStudentIDFunc   func(ctx context.Context, studentID string) ([]model.Achievement, error)
	UpdateAchievementFunc            func(ctx context.Context, id string, achievement *model.Achievement) error
	SoftDeleteAchievementFunc        func(ctx context.Context, id string) error
	CreateReferenceFunc              func(ctx context.Context, reference *model.AchievementReference) error
	UpdateReferenceFunc              func(ctx context.Context, reference *model.AchievementReference) error
	FindReferenceByIDFunc            func(ctx context.Context, id uuid.UUID) (*model.AchievementReference, error)
	FindReferenceByMongoIDFunc       func(ctx context.Context, mongoID string) (*model.AchievementReference, error)
	FindReferencesByStudentIDsFunc   func(ctx context.Context, studentIDs []uuid.UUID) ([]model.AchievementReference, error)
	FindReferencesWithPaginationFunc  func(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]model.AchievementReference, int64, error)
	DeleteReferenceFunc              func(ctx context.Context, id uuid.UUID) error
	GetAchievementStatisticsFunc     func(ctx context.Context, studentIDs []string) (*repository.AchievementStatistics, error)
}

func (m *MockAchievementRepository) CreateAchievement(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error) {
	if m.CreateAchievementFunc != nil {
		return m.CreateAchievementFunc(ctx, achievement)
	}
	return nil, nil
}

func (m *MockAchievementRepository) FindAchievementByID(ctx context.Context, id string) (*model.Achievement, error) {
	if m.FindAchievementByIDFunc != nil {
		return m.FindAchievementByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockAchievementRepository) FindAchievementsByStudentID(ctx context.Context, studentID string) ([]model.Achievement, error) {
	if m.FindAchievementsByStudentIDFunc != nil {
		return m.FindAchievementsByStudentIDFunc(ctx, studentID)
	}
	return nil, nil
}

func (m *MockAchievementRepository) UpdateAchievement(ctx context.Context, id string, achievement *model.Achievement) error {
	if m.UpdateAchievementFunc != nil {
		return m.UpdateAchievementFunc(ctx, id, achievement)
	}
	return nil
}

func (m *MockAchievementRepository) SoftDeleteAchievement(ctx context.Context, id string) error {
	if m.SoftDeleteAchievementFunc != nil {
		return m.SoftDeleteAchievementFunc(ctx, id)
	}
	return nil
}

func (m *MockAchievementRepository) CreateReference(ctx context.Context, reference *model.AchievementReference) error {
	if m.CreateReferenceFunc != nil {
		return m.CreateReferenceFunc(ctx, reference)
	}
	return nil
}

func (m *MockAchievementRepository) UpdateReference(ctx context.Context, reference *model.AchievementReference) error {
	if m.UpdateReferenceFunc != nil {
		return m.UpdateReferenceFunc(ctx, reference)
	}
	return nil
}

func (m *MockAchievementRepository) FindReferenceByID(ctx context.Context, id uuid.UUID) (*model.AchievementReference, error) {
	if m.FindReferenceByIDFunc != nil {
		return m.FindReferenceByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockAchievementRepository) FindReferenceByMongoID(ctx context.Context, mongoID string) (*model.AchievementReference, error) {
	if m.FindReferenceByMongoIDFunc != nil {
		return m.FindReferenceByMongoIDFunc(ctx, mongoID)
	}
	return nil, errors.New("not found")
}

func (m *MockAchievementRepository) FindReferencesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID) ([]model.AchievementReference, error) {
	if m.FindReferencesByStudentIDsFunc != nil {
		return m.FindReferencesByStudentIDsFunc(ctx, studentIDs)
	}
	return nil, nil
}

func (m *MockAchievementRepository) FindReferencesWithPagination(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]model.AchievementReference, int64, error) {
	if m.FindReferencesWithPaginationFunc != nil {
		return m.FindReferencesWithPaginationFunc(ctx, studentIDs, page, limit)
	}
	return nil, 0, nil
}

func (m *MockAchievementRepository) DeleteReference(ctx context.Context, id uuid.UUID) error {
	if m.DeleteReferenceFunc != nil {
		return m.DeleteReferenceFunc(ctx, id)
	}
	return nil
}

func (m *MockAchievementRepository) GetAchievementStatistics(ctx context.Context, studentIDs []string) (*repository.AchievementStatistics, error) {
	if m.GetAchievementStatisticsFunc != nil {
		return m.GetAchievementStatisticsFunc(ctx, studentIDs)
	}
	return nil, nil
}

func TestStudentService_GetAllStudents(t *testing.T) {
	ctx := context.Background()

	student1 := model.Student{
		ID:          uuid.New(),
		StudentID:   "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
	}
	student2 := model.Student{
		ID:          uuid.New(),
		StudentID:   "0987654321",
		ProgramStudy: "IT",
		AcademicYear: "2024",
	}

	studentRepo := &MockStudentRepository{
		FindAllStudentsFunc: func(ctx context.Context) ([]model.Student, error) {
			return []model.Student{student1, student2}, nil
		},
	}
	lecturerRepo := &MockLecturerRepository{}
	achievementRepo := &MockAchievementRepository{}

	studentService := service.NewStudentService(studentRepo, lecturerRepo, achievementRepo)

	students, err := studentService.GetAllStudents(ctx)

	assert.NoError(t, err)
	assert.Len(t, students, 2)
}

func TestStudentService_GetStudentByID(t *testing.T) {
	ctx := context.Background()

	studentID := uuid.New()
	userID := uuid.New()
	student := &model.Student{
		ID:          studentID,
		UserID:      userID,
		StudentID:   "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
	}

	tests := []struct {
		name       string
		studentID  uuid.UUID
		setupMocks func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository)
		wantErr    bool
	}{
		{
			name:     "Success get student by ID",
			studentID: studentID,
			setupMocks: func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository) {
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				achievementRepo := &MockAchievementRepository{}
				return studentRepo, lecturerRepo, achievementRepo
			},
			wantErr: false,
		},
		{
			name:     "Error student not found",
			studentID: uuid.New(),
			setupMocks: func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository) {
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return nil, errors.New("not found")
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				achievementRepo := &MockAchievementRepository{}
				return studentRepo, lecturerRepo, achievementRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			studentRepo, lecturerRepo, achievementRepo := tt.setupMocks()
			studentService := service.NewStudentService(studentRepo, lecturerRepo, achievementRepo)

			result, err := studentService.GetStudentByID(ctx, tt.studentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, student.ID, result.ID)
			}
		})
	}
}

func TestStudentService_GetStudentAchievements(t *testing.T) {
	ctx := context.Background()

	studentID := uuid.New()
	userID := uuid.New()
	student := &model.Student{
		ID:          studentID,
		UserID:      userID,
		StudentID:   "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
		User: model.User{
			ID:       userID,
			FullName: "Test Student",
		},
	}

	achievement1 := model.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       studentID.String(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Achievement 1",
		Status:          model.StatusDraft,
	}
	achievement2 := model.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       studentID.String(),
		AchievementType: model.AchievementTypeAcademic,
		Title:           "Achievement 2",
		Status:          model.StatusSubmitted,
	}

	reference1 := model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievement1.ID.Hex(),
		Status:             model.StatusDraft,
	}
	reference2 := model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievement2.ID.Hex(),
		Status:             model.StatusSubmitted,
	}

	studentRepo := &MockStudentRepository{
		FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
			return student, nil
		},
	}
	lecturerRepo := &MockLecturerRepository{}
	achievementRepo := &MockAchievementRepository{
		FindAchievementsByStudentIDFunc: func(ctx context.Context, studentID string) ([]model.Achievement, error) {
			return []model.Achievement{achievement1, achievement2}, nil
		},
		FindReferencesByStudentIDsFunc: func(ctx context.Context, studentIDs []uuid.UUID) ([]model.AchievementReference, error) {
			return []model.AchievementReference{reference1, reference2}, nil
		},
	}

	studentService := service.NewStudentService(studentRepo, lecturerRepo, achievementRepo)

	achievements, err := studentService.GetStudentAchievements(ctx, studentID)

	assert.NoError(t, err)
	assert.Len(t, achievements, 2)
}

func TestStudentService_UpdateStudentAdvisor(t *testing.T) {
	ctx := context.Background()

	studentID := uuid.New()
	advisorID := uuid.New()
	student := &model.Student{
		ID:          studentID,
		StudentID:   "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
	}
	updatedStudent := &model.Student{
		ID:          studentID,
		StudentID:   "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
		AdvisorID:   &advisorID,
	}

	tests := []struct {
		name       string
		studentID  uuid.UUID
		advisorID  *uuid.UUID
		setupMocks func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository)
		wantErr    bool
	}{
		{
			name:     "Success update advisor",
			studentID: studentID,
			advisorID: &advisorID,
			setupMocks: func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository) {
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						if id == studentID {
							return student, nil
						}
						return updatedStudent, nil
					},
					UpdateStudentFunc: func(ctx context.Context, student *model.Student) error {
						return nil
					},
				}
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
						return &model.Lecturer{ID: advisorID}, nil
					},
				}
				achievementRepo := &MockAchievementRepository{}
				return studentRepo, lecturerRepo, achievementRepo
			},
			wantErr: false,
		},
		{
			name:     "Error student not found",
			studentID: uuid.New(),
			advisorID: &advisorID,
			setupMocks: func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository) {
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return nil, errors.New("not found")
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				achievementRepo := &MockAchievementRepository{}
				return studentRepo, lecturerRepo, achievementRepo
			},
			wantErr: true,
		},
		{
			name:     "Error advisor not found",
			studentID: studentID,
			advisorID: &advisorID,
			setupMocks: func() (repository.StudentRepository, repository.LecturerRepository, repository.AchievementRepository) {
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
						return nil, errors.New("not found")
					},
				}
				achievementRepo := &MockAchievementRepository{}
				return studentRepo, lecturerRepo, achievementRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			studentRepo, lecturerRepo, achievementRepo := tt.setupMocks()
			studentService := service.NewStudentService(studentRepo, lecturerRepo, achievementRepo)

			result, err := studentService.UpdateStudentAdvisor(ctx, tt.studentID, tt.advisorID)

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

