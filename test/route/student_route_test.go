package route_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/route"
	"github.com/stretchr/testify/assert"
)

type MockStudentServiceForRoute struct {
	GetAllStudentsFunc        func(ctx context.Context) ([]model.Student, error)
	GetStudentByIDFunc        func(ctx context.Context, studentID uuid.UUID) (*model.Student, error)
	GetStudentAchievementsFunc func(ctx context.Context, studentID uuid.UUID) ([]service.AchievementResponse, error)
	UpdateStudentAdvisorFunc   func(ctx context.Context, studentID uuid.UUID, advisorID *uuid.UUID) (*model.Student, error)
}

func (m *MockStudentServiceForRoute) GetAllStudents(ctx context.Context) ([]model.Student, error) {
	if m.GetAllStudentsFunc != nil {
		return m.GetAllStudentsFunc(ctx)
	}
	return nil, nil
}

func (m *MockStudentServiceForRoute) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
	if m.GetStudentByIDFunc != nil {
		return m.GetStudentByIDFunc(ctx, studentID)
	}
	return nil, errors.New("not found")
}

func (m *MockStudentServiceForRoute) GetStudentAchievements(ctx context.Context, studentID uuid.UUID) ([]service.AchievementResponse, error) {
	if m.GetStudentAchievementsFunc != nil {
		return m.GetStudentAchievementsFunc(ctx, studentID)
	}
	return nil, nil
}

func (m *MockStudentServiceForRoute) UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID *uuid.UUID) (*model.Student, error) {
	if m.UpdateStudentAdvisorFunc != nil {
		return m.UpdateStudentAdvisorFunc(ctx, studentID, advisorID)
	}
	return nil, nil
}

func setupAuthMiddlewareForStudent(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("username", "testuser")
		c.Locals("permissions", []string{"students:read"})
		c.Locals("role_name", "Admin")
		return c.Next()
	})
}

func TestStudentRoute_GetAllStudents_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForStudent(app)

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

	mockStudentService := &MockStudentServiceForRoute{
		GetAllStudentsFunc: func(ctx context.Context) ([]model.Student, error) {
			return []model.Student{student1, student2}, nil
		},
	}

	route.RegisterStudentRoutes(app, mockStudentService)

	req := httptest.NewRequest("GET", "/students", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestStudentRoute_GetStudentByID_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForStudent(app)

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

	mockStudentService := &MockStudentServiceForRoute{
		GetStudentByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Student, error) {
			return student, nil
		},
	}

	route.RegisterStudentRoutes(app, mockStudentService)

	req := httptest.NewRequest("GET", "/students/"+studentID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestStudentRoute_GetStudentByID_InvalidUUID(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForStudent(app)

	mockStudentService := &MockStudentServiceForRoute{}
	route.RegisterStudentRoutes(app, mockStudentService)

	req := httptest.NewRequest("GET", "/students/invalid-uuid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestStudentRoute_GetStudentByID_NotFound(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForStudent(app)

	mockStudentService := &MockStudentServiceForRoute{
		GetStudentByIDFunc: func(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
			return nil, errors.New("mahasiswa tidak ditemukan")
		},
	}

	route.RegisterStudentRoutes(app, mockStudentService)

	req := httptest.NewRequest("GET", "/students/"+uuid.New().String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

