package repository_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "cgo") {
			t.Skip("Skipping test: CGO is disabled. Run tests with CGO_ENABLED=1")
		}
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Role{}, &model.Permission{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	assert.NotNil(t, repo)
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *model.User
		wantErr bool
	}{
		{
			name: "Success create user",
			user: &model.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashedpassword",
				FullName:     "Test User",
				IsActive:     true,
			},
			wantErr: false,
		},
		{
			name: "Error duplicate username",
			user: &model.User{
				Username:     "testuser",
				Email:        "test2@example.com",
				PasswordHash: "hashedpassword",
				FullName:     "Test User 2",
				IsActive:     true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateUser(ctx, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.user.ID)
			}
		})
	}
}

func TestUserRepository_FindUserByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		IsActive:     true,
	}
	repo.CreateUser(ctx, user)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "Success find user by ID",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "Error user not found",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindUserByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, user.ID, found.ID)
			}
		})
	}
}

func TestUserRepository_FindUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		IsActive:     true,
	}
	repo.CreateUser(ctx, user)

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "Success find user by username",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "Error user not found",
			username: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindUserByUsername(ctx, tt.username)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, user.Username, found.Username)
			}
		})
	}
}

func TestUserRepository_FindUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		IsActive:     true,
	}
	repo.CreateUser(ctx, user)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "Success find user by email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Error user not found",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindUserByEmail(ctx, tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, user.Email, found.Email)
			}
		})
	}
}

func TestUserRepository_FindAllUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user1 := &model.User{
		Username:     "user1",
		Email:        "user1@example.com",
		PasswordHash: "hash1",
		FullName:     "User 1",
		IsActive:     true,
	}
	user2 := &model.User{
		Username:     "user2",
		Email:        "user2@example.com",
		PasswordHash: "hash2",
		FullName:     "User 2",
		IsActive:     true,
	}

	repo.CreateUser(ctx, user1)
	repo.CreateUser(ctx, user2)

	users, err := repo.FindAllUsers(ctx)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		IsActive:     true,
	}
	repo.CreateUser(ctx, user)

	user.FullName = "Updated Name"
	err := repo.UpdateUser(ctx, user)

	assert.NoError(t, err)

	updated, _ := repo.FindUserByID(ctx, user.ID)
	assert.Equal(t, "Updated Name", updated.FullName)
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		IsActive:     true,
	}
	repo.CreateUser(ctx, user)

	err := repo.DeleteUser(ctx, user.ID)
	assert.NoError(t, err)

	_, err = repo.FindUserByID(ctx, user.ID)
	assert.Error(t, err)
}
