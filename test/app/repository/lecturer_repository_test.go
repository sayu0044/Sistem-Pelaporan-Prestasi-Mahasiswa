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

func setupLecturerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "cgo") {
			t.Skip("Skipping test: CGO is disabled. Run tests with CGO_ENABLED=1")
		}
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Lecturer{}, &model.Student{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewLecturerRepository(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)

	assert.NotNil(t, repo)
}

func TestLecturerRepository_CreateLecturer(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(user)

	tests := []struct {
		name     string
		lecturer *model.Lecturer
		wantErr  bool
	}{
		{
			name: "Success create lecturer",
			lecturer: &model.Lecturer{
				UserID:     user.ID,
				LecturerID: "L001",
				Department: "Computer Science",
			},
			wantErr: false,
		},
		{
			name: "Error duplicate lecturer_id",
			lecturer: &model.Lecturer{
				UserID:     user.ID,
				LecturerID: "L001",
				Department: "IT",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateLecturer(ctx, tt.lecturer)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.lecturer.ID)
			}
		})
	}
}

func TestLecturerRepository_FindLecturerByID(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(user)

	lecturer := &model.Lecturer{
		UserID:     user.ID,
		LecturerID: "L001",
		Department: "Computer Science",
	}
	repo.CreateLecturer(ctx, lecturer)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "Success find lecturer by ID",
			id:      lecturer.ID,
			wantErr: false,
		},
		{
			name:    "Error lecturer not found",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindLecturerByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, lecturer.ID, found.ID)
			}
		})
	}
}

func TestLecturerRepository_FindLecturerByUserID(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(user)

	lecturer := &model.Lecturer{
		UserID:     user.ID,
		LecturerID: "L001",
		Department: "Computer Science",
	}
	repo.CreateLecturer(ctx, lecturer)

	found, err := repo.FindLecturerByUserID(ctx, user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, lecturer.ID, found.ID)
}

func TestLecturerRepository_FindLecturerByLecturerID(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(user)

	lecturer := &model.Lecturer{
		UserID:     user.ID,
		LecturerID: "L001",
		Department: "Computer Science",
	}
	repo.CreateLecturer(ctx, lecturer)

	found, err := repo.FindLecturerByLecturerID(ctx, "L001")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, lecturer.LecturerID, found.LecturerID)
}

func TestLecturerRepository_FindAllLecturers(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user1 := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash1",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	user2 := &model.User{
		Username:     "lecturer2",
		Email:        "lecturer2@example.com",
		PasswordHash: "hash2",
		FullName:     "Lecturer 2",
		IsActive:     true,
	}
	db.Create(user1)
	db.Create(user2)

	lecturer1 := &model.Lecturer{
		UserID:     user1.ID,
		LecturerID: "L001",
		Department: "CS",
	}
	lecturer2 := &model.Lecturer{
		UserID:     user2.ID,
		LecturerID: "L002",
		Department: "IT",
	}

	repo.CreateLecturer(ctx, lecturer1)
	repo.CreateLecturer(ctx, lecturer2)

	lecturers, err := repo.FindAllLecturers(ctx)

	assert.NoError(t, err)
	assert.Len(t, lecturers, 2)
}

func TestLecturerRepository_UpdateLecturer(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(user)

	lecturer := &model.Lecturer{
		UserID:     user.ID,
		LecturerID: "L001",
		Department: "Computer Science",
	}
	repo.CreateLecturer(ctx, lecturer)

	lecturer.Department = "Information Technology"
	err := repo.UpdateLecturer(ctx, lecturer)

	assert.NoError(t, err)

	updated, _ := repo.FindLecturerByID(ctx, lecturer.ID)
	assert.Equal(t, "Information Technology", updated.Department)
}

func TestLecturerRepository_DeleteLecturer(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(user)

	lecturer := &model.Lecturer{
		UserID:     user.ID,
		LecturerID: "L001",
		Department: "Computer Science",
	}
	repo.CreateLecturer(ctx, lecturer)

	err := repo.DeleteLecturer(ctx, lecturer.ID)
	assert.NoError(t, err)

	_, err = repo.FindLecturerByID(ctx, lecturer.ID)
	assert.Error(t, err)
}

func TestLecturerRepository_FindAdvisees(t *testing.T) {
	db := setupLecturerTestDB(t)
	repo := repository.NewLecturerRepository(db)
	ctx := context.Background()

	lecturerUser := &model.User{
		Username:     "lecturer1",
		Email:        "lecturer1@example.com",
		PasswordHash: "hash",
		FullName:     "Lecturer 1",
		IsActive:     true,
	}
	db.Create(lecturerUser)

	lecturer := &model.Lecturer{
		UserID:     lecturerUser.ID,
		LecturerID: "L001",
		Department: "Computer Science",
	}
	repo.CreateLecturer(ctx, lecturer)

	studentUser1 := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash1",
		FullName:     "Student 1",
		IsActive:     true,
	}
	studentUser2 := &model.User{
		Username:     "student2",
		Email:        "student2@example.com",
		PasswordHash: "hash2",
		FullName:     "Student 2",
		IsActive:     true,
	}
	db.Create(studentUser1)
	db.Create(studentUser2)

	student1 := &model.Student{
		UserID:       studentUser1.ID,
		StudentID:    "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
		AdvisorID:    &lecturer.ID,
	}
	student2 := &model.Student{
		UserID:       studentUser2.ID,
		StudentID:    "0987654321",
		ProgramStudy: "IT",
		AcademicYear: "2024",
		AdvisorID:    &lecturer.ID,
	}
	db.Create(student1)
	db.Create(student2)

	advisees, err := repo.FindAdvisees(ctx, lecturer.ID)

	assert.NoError(t, err)
	assert.Len(t, advisees, 2)
}

