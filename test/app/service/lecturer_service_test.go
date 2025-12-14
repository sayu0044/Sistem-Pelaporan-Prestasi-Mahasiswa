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

func TestLecturerService_GetAllLecturers(t *testing.T) {
	ctx := context.Background()

	lecturer1 := model.Lecturer{
		ID:          uuid.New(),
		LecturerID:  "L001",
		Department: "Computer Science",
	}
	lecturer2 := model.Lecturer{
		ID:          uuid.New(),
		LecturerID:  "L002",
		Department: "Information Technology",
	}

	lecturerRepo := &MockLecturerRepository{
		FindAllLecturersFunc: func(ctx context.Context) ([]model.Lecturer, error) {
			return []model.Lecturer{lecturer1, lecturer2}, nil
		},
	}
	studentRepo := &MockStudentRepository{}

	lecturerService := service.NewLecturerService(lecturerRepo, studentRepo)

	lecturers, err := lecturerService.GetAllLecturers(ctx)

	assert.NoError(t, err)
	assert.Len(t, lecturers, 2)
}

func TestLecturerService_GetLecturerAdvisees(t *testing.T) {
	ctx := context.Background()

	lecturerID := uuid.New()
	student1 := model.Student{
		ID:          uuid.New(),
		StudentID:   "1234567890",
		ProgramStudy: "CS",
		AcademicYear: "2024",
		AdvisorID:   &lecturerID,
	}
	student2 := model.Student{
		ID:          uuid.New(),
		StudentID:   "0987654321",
		ProgramStudy: "IT",
		AcademicYear: "2024",
		AdvisorID:   &lecturerID,
	}

	tests := []struct {
		name       string
		lecturerID uuid.UUID
		setupMocks func() (repository.LecturerRepository, repository.StudentRepository)
		wantErr    bool
	}{
		{
			name:       "Success get advisees",
			lecturerID: lecturerID,
			setupMocks: func() (repository.LecturerRepository, repository.StudentRepository) {
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
						return &model.Lecturer{ID: lecturerID}, nil
					},
					FindAdviseesFunc: func(ctx context.Context, id uuid.UUID) ([]model.Student, error) {
						return []model.Student{student1, student2}, nil
					},
				}
				studentRepo := &MockStudentRepository{}
				return lecturerRepo, studentRepo
			},
			wantErr: false,
		},
		{
			name:       "Error lecturer not found",
			lecturerID: uuid.New(),
			setupMocks: func() (repository.LecturerRepository, repository.StudentRepository) {
				lecturerRepo := &MockLecturerRepository{
					FindLecturerByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
						return nil, errors.New("not found")
					},
				}
				studentRepo := &MockStudentRepository{}
				return lecturerRepo, studentRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lecturerRepo, studentRepo := tt.setupMocks()
			lecturerService := service.NewLecturerService(lecturerRepo, studentRepo)

			advisees, err := lecturerService.GetLecturerAdvisees(ctx, tt.lecturerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, advisees)
			} else {
				assert.NoError(t, err)
				assert.Len(t, advisees, 2)
			}
		})
	}
}

