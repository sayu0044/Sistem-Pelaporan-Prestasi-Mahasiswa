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

func TestAchievementService_CreateAchievement(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	studentID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    userID,
		StudentID: "1234567890",
		User: model.User{
			ID:       userID,
			FullName: "Test Student",
		},
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		req        *service.CreateAchievementRequest
		setupMocks func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository)
		wantErr    bool
	}{
		{
			name:   "Success create achievement",
			userID: userID,
			req: &service.CreateAchievementRequest{
				AchievementType: model.AchievementTypeCompetition,
				Title:           "Test Achievement",
				Description:     "Test Description",
				Details:         model.AchievementDetails{},
				Tags:            []string{"test"},
				Points:          10.0,
			},
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					CreateAchievementFunc: func(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error) {
						achievement.ID = primitive.NewObjectID()
						return achievement, nil
					},
					CreateReferenceFunc: func(ctx context.Context, reference *model.AchievementReference) error {
						return nil
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name:   "Error not a student",
			userID: userID,
			req: &service.CreateAchievementRequest{
				AchievementType: model.AchievementTypeCompetition,
				Title:           "Test Achievement",
				Description:     "Test Description",
			},
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{}
				studentRepo := &MockStudentRepository{
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return nil, errors.New("not found")
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: true,
		},
		{
			name:   "Error empty title",
			userID: userID,
			req: &service.CreateAchievementRequest{
				AchievementType: model.AchievementTypeCompetition,
				Title:           "",
				Description:     "Test Description",
			},
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{}
				studentRepo := &MockStudentRepository{
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo := tt.setupMocks()
			achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

			result, err := achievementService.CreateAchievement(ctx, tt.userID, tt.req)

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

func TestAchievementService_SubmitAchievement(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	studentID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    userID,
		StudentID: "1234567890",
		User: model.User{
			ID:       userID,
			FullName: "Test Student",
		},
	}

	achievementID := primitive.NewObjectID().Hex()
	objectID, _ := primitive.ObjectIDFromHex(achievementID)
	achievement := &model.Achievement{
		ID:              objectID,
		StudentID:       studentID.String(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Test Achievement",
		Description:     "Test Description",
		Status:          model.StatusDraft,
	}
	reference := &model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievementID,
		Status:             model.StatusDraft,
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		achievementID string
		setupMocks    func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository)
		wantErr       bool
	}{
		{
			name:          "Success submit achievement",
			userID:        userID,
			achievementID: achievementID,
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					FindAchievementByIDFunc: func(ctx context.Context, id string) (*model.Achievement, error) {
						return achievement, nil
					},
					UpdateAchievementFunc: func(ctx context.Context, id string, ach *model.Achievement) error {
						return nil
					},
					FindReferenceByMongoIDFunc: func(ctx context.Context, mongoID string) (*model.AchievementReference, error) {
						return reference, nil
					},
					UpdateReferenceFunc: func(ctx context.Context, ref *model.AchievementReference) error {
						return nil
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name:          "Error achievement not found",
			userID:        userID,
			achievementID: achievementID,
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					FindAchievementByIDFunc: func(ctx context.Context, id string) (*model.Achievement, error) {
						return nil, errors.New("not found")
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo := tt.setupMocks()
			achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

			result, err := achievementService.SubmitAchievement(ctx, tt.userID, tt.achievementID)

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

func TestAchievementService_VerifyAchievement(t *testing.T) {
	ctx := context.Background()

	lecturerID := uuid.New()
	lecturerUserID := uuid.New()
	lecturer := &model.Lecturer{
		ID:         lecturerID,
		UserID:     lecturerUserID,
		LecturerID: "L001",
	}

	studentID := uuid.New()
	studentUserID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    studentUserID,
		StudentID: "1234567890",
		AdvisorID: &lecturerID,
		User: model.User{
			ID:       studentUserID,
			FullName: "Test Student",
		},
	}

	achievementID := primitive.NewObjectID().Hex()
	objectID, _ := primitive.ObjectIDFromHex(achievementID)
	achievement := &model.Achievement{
		ID:              objectID,
		StudentID:       studentID.String(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Test Achievement",
		Status:          model.StatusSubmitted,
	}
	reference := &model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievementID,
		Status:             model.StatusSubmitted,
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		achievementID string
		setupMocks    func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository)
		wantErr       bool
	}{
		{
			name:          "Success verify achievement",
			userID:        lecturerUserID,
			achievementID: achievementID,
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					FindAchievementByIDFunc: func(ctx context.Context, id string) (*model.Achievement, error) {
						return achievement, nil
					},
					UpdateAchievementFunc: func(ctx context.Context, id string, ach *model.Achievement) error {
						return nil
					},
					FindReferenceByMongoIDFunc: func(ctx context.Context, mongoID string) (*model.AchievementReference, error) {
						return reference, nil
					},
					UpdateReferenceFunc: func(ctx context.Context, ref *model.AchievementReference) error {
						return nil
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error) {
						return lecturer, nil
					},
				}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name:          "Error not a lecturer",
			userID:        uuid.New(),
			achievementID: achievementID,
			setupMocks: func() (repository.AchievementRepository, repository.StudentRepository, repository.LecturerRepository, repository.UserRepository, repository.RoleRepository) {
				achievementRepo := &MockAchievementRepository{
					FindAchievementByIDFunc: func(ctx context.Context, id string) (*model.Achievement, error) {
						return achievement, nil
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
						return student, nil
					},
				}
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error) {
						return nil, errors.New("not found")
					},
				}
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{}
				return achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo := tt.setupMocks()
			achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

			result, err := achievementService.VerifyAchievement(ctx, tt.userID, tt.achievementID)

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

func TestAchievementService_RejectAchievement(t *testing.T) {
	ctx := context.Background()

	lecturerID := uuid.New()
	lecturerUserID := uuid.New()
	lecturer := &model.Lecturer{
		ID:         lecturerID,
		UserID:     lecturerUserID,
		LecturerID: "L001",
	}

	studentID := uuid.New()
	studentUserID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    studentUserID,
		StudentID: "1234567890",
		AdvisorID: &lecturerID,
		User: model.User{
			ID:       studentUserID,
			FullName: "Test Student",
		},
	}

	achievementID := primitive.NewObjectID().Hex()
	objectID, _ := primitive.ObjectIDFromHex(achievementID)
	achievement := &model.Achievement{
		ID:              objectID,
		StudentID:       studentID.String(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Test Achievement",
		Status:          model.StatusSubmitted,
	}
	reference := &model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievementID,
		Status:             model.StatusSubmitted,
	}

	achievementRepo := &MockAchievementRepository{
		FindAchievementByIDFunc: func(ctx context.Context, id string) (*model.Achievement, error) {
			return achievement, nil
		},
		UpdateAchievementFunc: func(ctx context.Context, id string, ach *model.Achievement) error {
			return nil
		},
		FindReferenceByMongoIDFunc: func(ctx context.Context, mongoID string) (*model.AchievementReference, error) {
			return reference, nil
		},
		UpdateReferenceFunc: func(ctx context.Context, ref *model.AchievementReference) error {
			return nil
		},
	}
	studentRepo := &MockStudentRepository{
		FindStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
			return student, nil
		},
	}
	lecturerRepo := &MockLecturerRepository{
		FindLecturerByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error) {
			return lecturer, nil
		},
	}
	userRepo := &MockUserRepository{}
	roleRepo := &MockRoleRepository{}

	achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

	result, err := achievementService.RejectAchievement(ctx, lecturerUserID, achievementID, "Rejection note")

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAchievementService_GetAchievements(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	studentID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    userID,
		StudentID: "1234567890",
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
	reference1 := model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievement1.ID.Hex(),
		Status:             model.StatusDraft,
	}

	roleID := uuid.New()
	userRepo := &MockUserRepository{
		FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return &model.User{ID: userID, RoleID: &roleID}, nil
		},
	}
	roleRepo := &MockRoleRepository{
		FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
			if id == roleID {
				return &model.Role{ID: roleID, Name: "Student"}, nil
			}
			return nil, errors.New("role tidak ditemukan")
		},
	}
	studentRepo := &MockStudentRepository{
		FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
			return student, nil
		},
	}
	lecturerRepo := &MockLecturerRepository{}
	achievementRepo := &MockAchievementRepository{
		FindAchievementsByStudentIDFunc: func(ctx context.Context, studentID string) ([]model.Achievement, error) {
			return []model.Achievement{achievement1}, nil
		},
		FindReferencesByStudentIDsFunc: func(ctx context.Context, studentIDs []uuid.UUID) ([]model.AchievementReference, error) {
			return []model.AchievementReference{reference1}, nil
		},
	}

	achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

	result, err := achievementService.GetAchievements(ctx, userID, 1, 10, "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)
}

func TestAchievementService_DeleteAchievement(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	studentID := uuid.New()
	student := &model.Student{
		ID:        studentID,
		UserID:    userID,
		StudentID: "1234567890",
	}

	achievementID := primitive.NewObjectID().Hex()
	objectID, _ := primitive.ObjectIDFromHex(achievementID)
	achievement := &model.Achievement{
		ID:              objectID,
		StudentID:       studentID.String(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Test Achievement",
		Status:          model.StatusDraft,
	}
	reference := &model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievementID,
		Status:             model.StatusDraft,
	}

	achievementRepo := &MockAchievementRepository{
		FindAchievementByIDFunc: func(ctx context.Context, id string) (*model.Achievement, error) {
			return achievement, nil
		},
		SoftDeleteAchievementFunc: func(ctx context.Context, id string) error {
			return nil
		},
		FindReferenceByMongoIDFunc: func(ctx context.Context, mongoID string) (*model.AchievementReference, error) {
			return reference, nil
		},
		DeleteReferenceFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	studentRepo := &MockStudentRepository{
		FindStudentByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
			return student, nil
		},
	}
	lecturerRepo := &MockLecturerRepository{}
	userRepo := &MockUserRepository{}
	roleRepo := &MockRoleRepository{}

	achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo, userRepo, roleRepo)

	err := achievementService.DeleteAchievement(ctx, userID, achievementID)

	assert.NoError(t, err)
}
