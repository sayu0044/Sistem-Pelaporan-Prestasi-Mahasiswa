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

type MockAuthService struct {
	RegisterFunc func(ctx context.Context, userData *model.User, password string) (*model.User, error)
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error) {
	return "", "", nil, nil, nil
}

func (m *MockAuthService) Register(ctx context.Context, userData *model.User, password string) (*model.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, userData, password)
	}
	return nil, nil
}

func (m *MockAuthService) ValidateToken(tokenString string) (*service.Claims, error) {
	return nil, nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error) {
	return "", "", nil, nil, nil
}

func (m *MockAuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
	return nil, nil, nil
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		username   string
		email      string
		password   string
		fullName   string
		roleID     uuid.UUID
		setupMocks func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService)
		wantErr    bool
	}{
		{
			name:     "Success create user",
			username: "newuser",
			email:    "newuser@example.com",
			password: "password123",
			fullName: "New User",
			roleID:   uuid.New(),
			setupMocks: func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService) {
				roleID := uuid.New()
				role := &model.Role{
					ID:   roleID,
					Name: "Student",
				}
				user := &model.User{
					ID:       uuid.New(),
					Username: "newuser",
					Email:    "newuser@example.com",
					FullName: "New User",
					RoleID:   &roleID,
				}

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
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByLecturerIDFunc: func(ctx context.Context, lecturerID string) (*model.Lecturer, error) {
						return nil, errors.New("not found")
					},
				}
				studentRepo := &MockStudentRepository{
					FindStudentByStudentIDFunc: func(ctx context.Context, studentID string) (*model.Student, error) {
						return nil, errors.New("not found")
					},
					CreateStudentFunc: func(ctx context.Context, student *model.Student) error {
						return nil
					},
				}
				authService := &MockAuthService{
					RegisterFunc: func(ctx context.Context, userData *model.User, password string) (*model.User, error) {
						if userData.ID == uuid.Nil {
							userData.ID = uuid.New()
						}
						return userData, nil
					},
				}
				return userRepo, roleRepo, lecturerRepo, studentRepo, authService
			},
			wantErr: false,
		},
		{
			name:     "Error role not found",
			username: "newuser",
			email:    "newuser@example.com",
			password: "password123",
			fullName: "New User",
			roleID:   uuid.New(),
			setupMocks: func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService) {
				userRepo := &MockUserRepository{}
				roleRepo := &MockRoleRepository{
					FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
						return nil, errors.New("role not found")
					},
				}
				lecturerRepo := &MockLecturerRepository{}
				studentRepo := &MockStudentRepository{}
				authService := &MockAuthService{}
				return userRepo, roleRepo, lecturerRepo, studentRepo, authService
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo, lecturerRepo, studentRepo, authService := tt.setupMocks()
			userService := service.NewUserService(userRepo, roleRepo, lecturerRepo, studentRepo, authService)

			studentID := ""
			if tt.name == "Success create user" {
				studentID = "1234567890"
			}
			result, role, err := userService.CreateUser(ctx, tt.username, tt.email, tt.password, tt.fullName, tt.roleID, true, "", "", studentID, "", "", nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if role != nil {
					assert.NotNil(t, role)
				}
			}
		})
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	ctx := context.Background()

	user1 := model.User{
		ID:       uuid.New(),
		Username: "user1",
		Email:    "user1@example.com",
		FullName: "User 1",
	}
	user2 := model.User{
		ID:       uuid.New(),
		Username: "user2",
		Email:    "user2@example.com",
		FullName: "User 2",
	}

	userRepo := &MockUserRepository{
		FindAllUsersFunc: func(ctx context.Context) ([]model.User, error) {
			return []model.User{user1, user2}, nil
		},
	}
	roleRepo := &MockRoleRepository{}
	lecturerRepo := &MockLecturerRepository{}
	studentRepo := &MockStudentRepository{}
	authService := &MockAuthService{}

	userService := service.NewUserService(userRepo, roleRepo, lecturerRepo, studentRepo, authService)

	users, err := userService.GetAllUsers(ctx)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	roleID := uuid.New()
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		RoleID:   &roleID,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMocks func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService)
		wantErr    bool
	}{
		{
			name:   "Success get user by ID",
			userID: user.ID,
			setupMocks: func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService) {
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
				lecturerRepo := &MockLecturerRepository{}
				studentRepo := &MockStudentRepository{}
				authService := &MockAuthService{}
				return userRepo, roleRepo, lecturerRepo, studentRepo, authService
			},
			wantErr: false,
		},
		{
			name:   "Error user not found",
			userID: uuid.New(),
			setupMocks: func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService) {
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return nil, errors.New("not found")
					},
				}
				roleRepo := &MockRoleRepository{}
				lecturerRepo := &MockLecturerRepository{}
				studentRepo := &MockStudentRepository{}
				authService := &MockAuthService{}
				return userRepo, roleRepo, lecturerRepo, studentRepo, authService
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo, lecturerRepo, studentRepo, authService := tt.setupMocks()
			userService := service.NewUserService(userRepo, roleRepo, lecturerRepo, studentRepo, authService)

			result, resultRole, err := userService.GetUserByID(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, user.ID, result.ID)
				if resultRole != nil {
					assert.NotNil(t, resultRole)
				}
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMocks func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService)
		wantErr    bool
	}{
		{
			name:   "Success delete user",
			userID: userID,
			setupMocks: func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService) {
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return user, nil
					},
					DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
						return nil
					},
				}
				roleRepo := &MockRoleRepository{}
				lecturerRepo := &MockLecturerRepository{}
				studentRepo := &MockStudentRepository{}
				authService := &MockAuthService{}
				return userRepo, roleRepo, lecturerRepo, studentRepo, authService
			},
			wantErr: false,
		},
		{
			name:   "Error user not found",
			userID: uuid.New(),
			setupMocks: func() (repository.UserRepository, repository.RoleRepository, repository.LecturerRepository, repository.StudentRepository, service.AuthService) {
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return nil, errors.New("not found")
					},
				}
				roleRepo := &MockRoleRepository{}
				lecturerRepo := &MockLecturerRepository{}
				studentRepo := &MockStudentRepository{}
				authService := &MockAuthService{}
				return userRepo, roleRepo, lecturerRepo, studentRepo, authService
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo, lecturerRepo, studentRepo, authService := tt.setupMocks()
			userService := service.NewUserService(userRepo, roleRepo, lecturerRepo, studentRepo, authService)

			err := userService.DeleteUser(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockLecturerRepository struct {
	CreateLecturerFunc          func(ctx context.Context, lecturer *model.Lecturer) error
	FindLecturerByIDFunc        func(ctx context.Context, id uuid.UUID) (*model.Lecturer, error)
	FindLecturerByUserIDFunc    func(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error)
	FindLecturerByLecturerIDFunc func(ctx context.Context, lecturerID string) (*model.Lecturer, error)
	FindAllLecturersFunc        func(ctx context.Context) ([]model.Lecturer, error)
	UpdateLecturerFunc          func(ctx context.Context, lecturer *model.Lecturer) error
	DeleteLecturerFunc          func(ctx context.Context, id uuid.UUID) error
	FindAdviseesFunc            func(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error)
}

func (m *MockLecturerRepository) CreateLecturer(ctx context.Context, lecturer *model.Lecturer) error {
	if m.CreateLecturerFunc != nil {
		return m.CreateLecturerFunc(ctx, lecturer)
	}
	return nil
}

func (m *MockLecturerRepository) FindLecturerByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
	if m.FindLecturerByIDFunc != nil {
		return m.FindLecturerByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockLecturerRepository) FindLecturerByUserID(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error) {
	if m.FindLecturerByUserIDFunc != nil {
		return m.FindLecturerByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("not found")
}

func (m *MockLecturerRepository) FindLecturerByLecturerID(ctx context.Context, lecturerID string) (*model.Lecturer, error) {
	if m.FindLecturerByLecturerIDFunc != nil {
		return m.FindLecturerByLecturerIDFunc(ctx, lecturerID)
	}
	return nil, errors.New("not found")
}

func (m *MockLecturerRepository) FindAllLecturers(ctx context.Context) ([]model.Lecturer, error) {
	if m.FindAllLecturersFunc != nil {
		return m.FindAllLecturersFunc(ctx)
	}
	return nil, nil
}

func (m *MockLecturerRepository) UpdateLecturer(ctx context.Context, lecturer *model.Lecturer) error {
	if m.UpdateLecturerFunc != nil {
		return m.UpdateLecturerFunc(ctx, lecturer)
	}
	return nil
}

func (m *MockLecturerRepository) DeleteLecturer(ctx context.Context, id uuid.UUID) error {
	if m.DeleteLecturerFunc != nil {
		return m.DeleteLecturerFunc(ctx, id)
	}
	return nil
}

func (m *MockLecturerRepository) FindAdvisees(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error) {
	if m.FindAdviseesFunc != nil {
		return m.FindAdviseesFunc(ctx, lecturerID)
	}
	return nil, nil
}

type MockStudentRepository struct {
	CreateStudentFunc          func(ctx context.Context, student *model.Student) error
	FindStudentByIDFunc        func(ctx context.Context, id uuid.UUID) (*model.Student, error)
	FindStudentByUserIDFunc    func(ctx context.Context, userID uuid.UUID) (*model.Student, error)
	FindStudentByStudentIDFunc func(ctx context.Context, studentID string) (*model.Student, error)
	FindAllStudentsFunc        func(ctx context.Context) ([]model.Student, error)
	UpdateStudentFunc          func(ctx context.Context, student *model.Student) error
	DeleteStudentFunc          func(ctx context.Context, id uuid.UUID) error
}

func (m *MockStudentRepository) CreateStudent(ctx context.Context, student *model.Student) error {
	if m.CreateStudentFunc != nil {
		return m.CreateStudentFunc(ctx, student)
	}
	return nil
}

func (m *MockStudentRepository) FindStudentByID(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	if m.FindStudentByIDFunc != nil {
		return m.FindStudentByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockStudentRepository) FindStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
	if m.FindStudentByUserIDFunc != nil {
		return m.FindStudentByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("not found")
}

func (m *MockStudentRepository) FindStudentByStudentID(ctx context.Context, studentID string) (*model.Student, error) {
	if m.FindStudentByStudentIDFunc != nil {
		return m.FindStudentByStudentIDFunc(ctx, studentID)
	}
	return nil, errors.New("not found")
}

func (m *MockStudentRepository) FindAllStudents(ctx context.Context) ([]model.Student, error) {
	if m.FindAllStudentsFunc != nil {
		return m.FindAllStudentsFunc(ctx)
	}
	return nil, nil
}

func (m *MockStudentRepository) UpdateStudent(ctx context.Context, student *model.Student) error {
	if m.UpdateStudentFunc != nil {
		return m.UpdateStudentFunc(ctx, student)
	}
	return nil
}

func (m *MockStudentRepository) DeleteStudent(ctx context.Context, id uuid.UUID) error {
	if m.DeleteStudentFunc != nil {
		return m.DeleteStudentFunc(ctx, id)
	}
	return nil
}

