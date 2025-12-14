package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	FindUserByUsernameFunc func(ctx context.Context, username string) (*model.User, error)
	FindUserByEmailFunc     func(ctx context.Context, email string) (*model.User, error)
	FindUserByIDFunc        func(ctx context.Context, id uuid.UUID) (*model.User, error)
	CreateUserFunc          func(ctx context.Context, user *model.User) error
	FindAllUsersFunc        func(ctx context.Context) ([]model.User, error)
	UpdateUserFunc          func(ctx context.Context, user *model.User) error
	DeleteUserFunc          func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.FindUserByIDFunc != nil {
		return m.FindUserByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.FindUserByUsernameFunc != nil {
		return m.FindUserByUsernameFunc(ctx, username)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.FindUserByEmailFunc != nil {
		return m.FindUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) FindAllUsers(ctx context.Context) ([]model.User, error) {
	if m.FindAllUsersFunc != nil {
		return m.FindAllUsersFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

type MockRoleRepository struct {
	FindRoleByIDFunc   func(ctx context.Context, id uuid.UUID) (*model.Role, error)
	FindRoleByNameFunc func(ctx context.Context, name string) (*model.Role, error)
	FindAllRolesFunc   func(ctx context.Context) ([]model.Role, error)
}

func (m *MockRoleRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	if m.FindRoleByIDFunc != nil {
		return m.FindRoleByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockRoleRepository) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	if m.FindRoleByNameFunc != nil {
		return m.FindRoleByNameFunc(ctx, name)
	}
	return nil, errors.New("not found")
}

func (m *MockRoleRepository) FindAllRoles(ctx context.Context) ([]model.Role, error) {
	if m.FindAllRolesFunc != nil {
		return m.FindAllRolesFunc(ctx)
	}
	return nil, nil
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	jwtExpiry := 24 * time.Hour

	tests := []struct {
		name        string
		username    string
		password    string
		setupMocks  func() (repository.UserRepository, repository.RoleRepository)
		wantErr     bool
		checkTokens bool
	}{
		{
			name:     "Success login",
			username: "testuser",
			password: "password123",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				roleID := uuid.New()
				user := &model.User{
					ID:           uuid.New(),
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					FullName:     "Test User",
					RoleID:       &roleID,
					IsActive:     true,
				}
				role := &model.Role{
					ID:   roleID,
					Name: "Student",
				}

				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return user, nil
					},
				}
				roleRepo := &MockRoleRepository{
					FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
						return role, nil
					},
				}
				return userRepo, roleRepo
			},
			wantErr:     false,
			checkTokens: true,
		},
		{
			name:     "Error user not found",
			username: "nonexistent",
			password: "password123",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return nil, errors.New("not found")
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr:     true,
			checkTokens: false,
		},
		{
			name:     "Error wrong password",
			username: "testuser",
			password: "wrongpassword",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &model.User{
					ID:           uuid.New(),
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					FullName:     "Test User",
					IsActive:     true,
				}

				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return user, nil
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr:     true,
			checkTokens: false,
		},
		{
			name:     "Error inactive account",
			username: "testuser",
			password: "password123",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &model.User{
					ID:           uuid.New(),
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					FullName:     "Test User",
					IsActive:     false,
				}

				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return user, nil
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr:     true,
			checkTokens: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo := tt.setupMocks()
			authService := service.NewAuthService(userRepo, roleRepo, jwtSecret, jwtExpiry)

			token, refreshToken, user, role, err := authService.Login(ctx, tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				if tt.checkTokens {
					assert.NotEmpty(t, token)
					assert.NotEmpty(t, refreshToken)
				}
				assert.NotNil(t, user)
				if role != nil {
					assert.NotNil(t, role)
				}
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	jwtExpiry := 24 * time.Hour

	tests := []struct {
		name       string
		userData   *model.User
		password   string
		setupMocks func() (repository.UserRepository, repository.RoleRepository)
		wantErr    bool
	}{
		{
			name: "Success register",
			userData: &model.User{
				Username: "newuser",
				Email:    "newuser@example.com",
				FullName: "New User",
			},
			password: "password123",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return nil, errors.New("not found")
					},
					FindUserByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
						return nil, errors.New("not found")
					},
					CreateUserFunc: func(ctx context.Context, user *model.User) error {
						if user.ID == uuid.Nil {
							user.ID = uuid.New()
						}
						return nil
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name: "Error duplicate username",
			userData: &model.User{
				Username: "existinguser",
				Email:    "newuser@example.com",
				FullName: "New User",
			},
			password: "password123",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return &model.User{Username: "existinguser"}, nil
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr: true,
		},
		{
			name: "Error duplicate email",
			userData: &model.User{
				Username: "newuser",
				Email:    "existing@example.com",
				FullName: "New User",
			},
			password: "password123",
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				userRepo := &MockUserRepository{
					FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
						return nil, errors.New("not found")
					},
					FindUserByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
						return &model.User{Email: "existing@example.com"}, nil
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo := tt.setupMocks()
			authService := service.NewAuthService(userRepo, roleRepo, jwtSecret, jwtExpiry)

			result, err := authService.Register(ctx, tt.userData, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEqual(t, uuid.Nil, result.ID)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	jwtExpiry := 24 * time.Hour

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	roleID := uuid.New()
	user := &model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Test User",
		RoleID:       &roleID,
		IsActive:     true,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	userRepo := &MockUserRepository{
		FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
			return user, nil
		},
	}
	roleRepo := &MockRoleRepository{
		FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
			return role, nil
		},
	}

	authService := service.NewAuthService(userRepo, roleRepo, jwtSecret, jwtExpiry)

	token, _, _, _, _ := authService.Login(ctx, "testuser", "password123")

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Success validate token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "Error invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "Error empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, user.ID, claims.UserID)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	jwtExpiry := 24 * time.Hour

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	roleID := uuid.New()
	user := &model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Test User",
		RoleID:       &roleID,
		IsActive:     true,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	userRepo := &MockUserRepository{
		FindUserByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
			return user, nil
		},
		FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return user, nil
		},
	}
	roleRepo := &MockRoleRepository{
		FindRoleByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
			return role, nil
		},
	}

	authService := service.NewAuthService(userRepo, roleRepo, jwtSecret, jwtExpiry)

	_, refreshToken, _, _, _ := authService.Login(ctx, "testuser", "password123")

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		checkTokens bool
	}{
		{
			name:        "Success refresh token",
			token:       refreshToken,
			wantErr:     false,
			checkTokens: true,
		},
		{
			name:        "Error invalid token",
			token:       "invalid-token",
			wantErr:     true,
			checkTokens: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, newRefreshToken, user, role, err := authService.RefreshToken(ctx, tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, newToken)
				assert.Empty(t, newRefreshToken)
			} else {
				assert.NoError(t, err)
				if tt.checkTokens {
					assert.NotEmpty(t, newToken)
					assert.NotEmpty(t, newRefreshToken)
				}
				assert.NotNil(t, user)
				assert.NotNil(t, role)
			}
		})
	}
}

func TestAuthService_GetProfile(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	jwtExpiry := 24 * time.Hour

	roleID := uuid.New()
	user := &model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		FullName:     "Test User",
		RoleID:       &roleID,
		IsActive:     true,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMocks func() (repository.UserRepository, repository.RoleRepository)
		wantErr    bool
	}{
		{
			name:   "Success get profile",
			userID: user.ID,
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
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
				return userRepo, roleRepo
			},
			wantErr: false,
		},
		{
			name:   "Error user not found",
			userID: uuid.New(),
			setupMocks: func() (repository.UserRepository, repository.RoleRepository) {
				userRepo := &MockUserRepository{
					FindUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						return nil, errors.New("not found")
					},
				}
				roleRepo := &MockRoleRepository{}
				return userRepo, roleRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo := tt.setupMocks()
			authService := service.NewAuthService(userRepo, roleRepo, jwtSecret, jwtExpiry)

			resultUser, resultRole, err := authService.GetProfile(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resultUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultUser)
				assert.Equal(t, user.ID, resultUser.ID)
				if resultRole != nil {
					assert.NotNil(t, resultRole)
				}
			}
		})
	}
}
