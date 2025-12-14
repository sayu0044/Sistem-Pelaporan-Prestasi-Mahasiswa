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

func setupStudentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "cgo") {
			t.Skip("Skipping test: CGO is disabled. Run tests with CGO_ENABLED=1")
		}
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Student{}, &model.Lecturer{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewStudentRepository(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)

	assert.NotNil(t, repo)
}

func TestStudentRepository_CreateStudent(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash",
		FullName:     "Student 1",
		IsActive:     true,
	}
	db.Create(user)

	tests := []struct {
		name    string
		student *model.Student
		wantErr bool
	}{
		{
			name: "Success create student",
			student: &model.Student{
				UserID:      user.ID,
				StudentID:   "1234567890",
				ProgramStudy: "Computer Science",
				AcademicYear: "2024",
			},
			wantErr: false,
		},
		{
			name: "Error duplicate student_id",
			student: &model.Student{
				UserID:      user.ID,
				StudentID:   "1234567890",
				ProgramStudy: "Computer Science",
				AcademicYear: "2024",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateStudent(ctx, tt.student)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.student.ID)
			}
		})
	}
}

func TestStudentRepository_FindStudentByID(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash",
		FullName:     "Student 1",
		IsActive:     true,
	}
	db.Create(user)

	student := &model.Student{
		UserID:       user.ID,
		StudentID:    "1234567890",
		ProgramStudy: "Computer Science",
		AcademicYear: "2024",
	}
	repo.CreateStudent(ctx, student)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "Success find student by ID",
			id:      student.ID,
			wantErr: false,
		},
		{
			name:    "Error student not found",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindStudentByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, student.ID, found.ID)
			}
		})
	}
}

func TestStudentRepository_FindStudentByUserID(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash",
		FullName:     "Student 1",
		IsActive:     true,
	}
	db.Create(user)

	student := &model.Student{
		UserID:       user.ID,
		StudentID:    "1234567890",
		ProgramStudy: "Computer Science",
		AcademicYear: "2024",
	}
	repo.CreateStudent(ctx, student)

	found, err := repo.FindStudentByUserID(ctx, user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, student.ID, found.ID)
}

func TestStudentRepository_FindStudentByStudentID(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash",
		FullName:     "Student 1",
		IsActive:     true,
	}
	db.Create(user)

	student := &model.Student{
		UserID:       user.ID,
		StudentID:    "1234567890",
		ProgramStudy: "Computer Science",
		AcademicYear: "2024",
	}
	repo.CreateStudent(ctx, student)

	found, err := repo.FindStudentByStudentID(ctx, "1234567890")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, student.StudentID, found.StudentID)
}

func TestStudentRepository_FindAllStudents(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user1 := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash1",
		FullName:     "Student 1",
		IsActive:     true,
	}
	user2 := &model.User{
		Username:     "student2",
		Email:        "student2@example.com",
		PasswordHash: "hash2",
		FullName:     "Student 2",
		IsActive:     true,
	}
	db.Create(user1)
	db.Create(user2)

	student1 := &model.Student{
		UserID:       user1.ID,
		StudentID:    "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
	}
	student2 := &model.Student{
		UserID:       user2.ID,
		StudentID:    "0987654321",
		ProgramStudy: "IT",
		AcademicYear: "2024",
	}

	repo.CreateStudent(ctx, student1)
	repo.CreateStudent(ctx, student2)

	students, err := repo.FindAllStudents(ctx)

	assert.NoError(t, err)
	assert.Len(t, students, 2)
}

func TestStudentRepository_UpdateStudent(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash",
		FullName:     "Student 1",
		IsActive:     true,
	}
	db.Create(user)

	lecturer := &model.Lecturer{
		UserID:     user.ID,
		LecturerID: "L001",
		Department: "CS",
	}
	db.Create(lecturer)

	student := &model.Student{
		UserID:       user.ID,
		StudentID:    "1234567890",
		ProgramStudy: "Computer Science",
		AcademicYear: "2024",
	}
	repo.CreateStudent(ctx, student)

	student.AdvisorID = &lecturer.ID
	err := repo.UpdateStudent(ctx, student)

	assert.NoError(t, err)

	updated, _ := repo.FindStudentByID(ctx, student.ID)
	assert.Equal(t, lecturer.ID, *updated.AdvisorID)
}

func TestStudentRepository_DeleteStudent(t *testing.T) {
	db := setupStudentTestDB(t)
	repo := repository.NewStudentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "student1",
		Email:        "student1@example.com",
		PasswordHash: "hash",
		FullName:     "Student 1",
		IsActive:     true,
	}
	db.Create(user)

	student := &model.Student{
		UserID:       user.ID,
		StudentID:    "1234567890",
		ProgramStudy: "Computer Science",
		AcademicYear: "2024",
	}
	repo.CreateStudent(ctx, student)

	err := repo.DeleteStudent(ctx, student.ID)
	assert.NoError(t, err)

	_, err = repo.FindStudentByID(ctx, student.ID)
	assert.Error(t, err)
}

