package route_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/route"
	"github.com/stretchr/testify/assert"
)

type MockReportServiceForRoute struct {
	GetStatisticsFunc        func(ctx context.Context, userID uuid.UUID) (*service.StatisticsResponse, error)
	GetStudentStatisticsFunc func(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*service.StudentStatisticsResponse, error)
}

func (m *MockReportServiceForRoute) GetStatistics(ctx context.Context, userID uuid.UUID) (*service.StatisticsResponse, error) {
	if m.GetStatisticsFunc != nil {
		return m.GetStatisticsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockReportServiceForRoute) GetStudentStatistics(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*service.StudentStatisticsResponse, error) {
	if m.GetStudentStatisticsFunc != nil {
		return m.GetStudentStatisticsFunc(ctx, userID, studentID)
	}
	return nil, nil
}

func setupAuthMiddlewareForReport(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("username", "testuser")
		c.Locals("permissions", []string{"achievements:read"})
		c.Locals("role_name", "Admin")
		return c.Next()
	})
}

func TestReportRoute_GetStatistics_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForReport(app)

	stats := &service.StatisticsResponse{
		TotalByType: map[string]int64{
			"competition": 5,
			"academic":    3,
		},
		TotalByPeriod: []service.PeriodStatResponse{
			{Period: "2024-01", Count: 2},
			{Period: "2024-02", Count: 6},
		},
		TopStudents: []service.TopStudentStatResponse{
			{StudentID: "1234567890", TotalPoints: 100.0, TotalAchievements: 8},
		},
		CompetitionLevelDistribution: map[string]int64{
			"national": 3,
			"local":    2,
		},
	}

	mockReportService := &MockReportServiceForRoute{
		GetStatisticsFunc: func(ctx context.Context, userID uuid.UUID) (*service.StatisticsResponse, error) {
			return stats, nil
		},
	}

	route.RegisterReportRoutes(app, mockReportService)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReportRoute_GetStatistics_UserNotFound(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("username", "testuser")
		c.Locals("permissions", []string{"achievements:read"})
		c.Locals("role_name", "Admin")
		return c.Next()
	})

	mockReportService := &MockReportServiceForRoute{
		GetStatisticsFunc: func(ctx context.Context, userID uuid.UUID) (*service.StatisticsResponse, error) {
			return nil, errors.New("user tidak ditemukan")
		},
	}

	route.RegisterReportRoutes(app, mockReportService)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestReportRoute_GetStudentStatistics_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForReport(app)

	studentID := uuid.New()
	stats := &service.StudentStatisticsResponse{
		StudentID:                    studentID.String(),
		StudentName:                 "Test Student",
		TotalByType:                  map[string]int64{"competition": 5},
		TotalByPeriod:                []service.PeriodStatResponse{{Period: "2024-01", Count: 5}},
		CompetitionLevelDistribution: map[string]int64{"national": 5},
		TotalPoints:                  100.0,
		TotalAchievements:            5,
	}

	mockReportService := &MockReportServiceForRoute{
		GetStudentStatisticsFunc: func(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*service.StudentStatisticsResponse, error) {
			return stats, nil
		},
	}

	route.RegisterReportRoutes(app, mockReportService)

	req := httptest.NewRequest("GET", "/reports/student/"+studentID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReportRoute_GetStudentStatistics_InvalidUUID(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForReport(app)

	mockReportService := &MockReportServiceForRoute{}
	route.RegisterReportRoutes(app, mockReportService)

	req := httptest.NewRequest("GET", "/reports/student/invalid-uuid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestReportRoute_GetStudentStatistics_StudentNotFound(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForReport(app)

	mockReportService := &MockReportServiceForRoute{
		GetStudentStatisticsFunc: func(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*service.StudentStatisticsResponse, error) {
			return nil, errors.New("mahasiswa tidak ditemukan")
		},
	}

	route.RegisterReportRoutes(app, mockReportService)

	req := httptest.NewRequest("GET", "/reports/student/"+uuid.New().String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

