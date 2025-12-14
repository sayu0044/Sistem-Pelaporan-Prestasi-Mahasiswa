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

func setupRoleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "cgo") {
			t.Skip("Skipping test: CGO is disabled. Run tests with CGO_ENABLED=1")
		}
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&model.Role{}, &model.Permission{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewRoleRepository(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := repository.NewRoleRepository(db)

	assert.NotNil(t, repo)
}

func TestRoleRepository_FindRoleByID(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{
		Name:        "Admin",
		Description: "Administrator role",
	}
	db.Create(role)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "Success find role by ID",
			id:      role.ID,
			wantErr: false,
		},
		{
			name:    "Error role not found",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindRoleByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, role.ID, found.ID)
			}
		})
	}
}

func TestRoleRepository_FindRoleByName(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role := &model.Role{
		Name:        "Admin",
		Description: "Administrator role",
	}
	db.Create(role)

	tests := []struct {
		name    string
		roleName string
		wantErr bool
	}{
		{
			name:     "Success find role by name",
			roleName: "Admin",
			wantErr:  false,
		},
		{
			name:     "Error role not found",
			roleName: "Nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindRoleByName(ctx, tt.roleName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, role.Name, found.Name)
			}
		})
	}
}

func TestRoleRepository_FindAllRoles(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := repository.NewRoleRepository(db)
	ctx := context.Background()

	role1 := &model.Role{
		Name:        "Admin",
		Description: "Administrator",
	}
	role2 := &model.Role{
		Name:        "Student",
		Description: "Student role",
	}
	db.Create(role1)
	db.Create(role2)

	roles, err := repo.FindAllRoles(ctx)

	assert.NoError(t, err)
	assert.Len(t, roles, 2)
}

