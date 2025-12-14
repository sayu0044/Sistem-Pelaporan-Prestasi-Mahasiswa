package route_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/route"
	"github.com/stretchr/testify/assert"
)

type MockLecturerServiceForRoute struct {
	GetAllLecturersFunc   func(ctx context.Context) ([]model.Lecturer, error)
	GetLecturerAdviseesFunc func(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error)
}

func (m *MockLecturerServiceForRoute) GetAllLecturers(ctx context.Context) ([]model.Lecturer, error) {
	if m.GetAllLecturersFunc != nil {
		return m.GetAllLecturersFunc(ctx)
	}
	return nil, nil
}

func (m *MockLecturerServiceForRoute) GetLecturerAdvisees(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error) {
	if m.GetLecturerAdviseesFunc != nil {
		return m.GetLecturerAdviseesFunc(ctx, lecturerID)
	}
	return nil, nil
}

func setupAuthMiddlewareForLecturer(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("username", "testuser")
		c.Locals("permissions", []string{"lecturers:read"})
		c.Locals("role_name", "Admin")
		return c.Next()
	})
}

func TestLecturerRoute_GetAllLecturers_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForLecturer(app)

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

	mockLecturerService := &MockLecturerServiceForRoute{
		GetAllLecturersFunc: func(ctx context.Context) ([]model.Lecturer, error) {
			return []model.Lecturer{lecturer1, lecturer2}, nil
		},
	}

	route.RegisterLecturerRoutes(app, mockLecturerService)

	req := httptest.NewRequest("GET", "/lecturers", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLecturerRoute_GetLecturerAdvisees_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForLecturer(app)

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

	mockLecturerService := &MockLecturerServiceForRoute{
		GetLecturerAdviseesFunc: func(ctx context.Context, id uuid.UUID) ([]model.Student, error) {
			return []model.Student{student1, student2}, nil
		},
	}

	route.RegisterLecturerRoutes(app, mockLecturerService)

	req := httptest.NewRequest("GET", "/lecturers/"+lecturerID.String()+"/advisees", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLecturerRoute_GetLecturerAdvisees_InvalidUUID(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForLecturer(app)

	mockLecturerService := &MockLecturerServiceForRoute{}
	route.RegisterLecturerRoutes(app, mockLecturerService)

	req := httptest.NewRequest("GET", "/lecturers/invalid-uuid/advisees", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLecturerRoute_GetLecturerAdvisees_LecturerNotFound(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForLecturer(app)

	mockLecturerService := &MockLecturerServiceForRoute{
		GetLecturerAdviseesFunc: func(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error) {
			return nil, errors.New("dosen tidak ditemukan")
		},
	}

	route.RegisterLecturerRoutes(app, mockLecturerService)

	req := httptest.NewRequest("GET", "/lecturers/"+uuid.New().String()+"/advisees", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

