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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockAchievementServiceForRoute struct {
	GetAchievementsFunc    func(ctx context.Context, userID uuid.UUID, page, limit int, status string) (*service.AchievementListResponse, error)
	GetAchievementByIDFunc func(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error)
	CreateAchievementFunc  func(ctx context.Context, userID uuid.UUID, req *service.CreateAchievementRequest) (*service.AchievementResponse, error)
	UpdateAchievementFunc  func(ctx context.Context, userID uuid.UUID, achievementID string, req *service.UpdateAchievementRequest) (*service.AchievementResponse, error)
	DeleteAchievementFunc  func(ctx context.Context, userID uuid.UUID, achievementID string) error
	SubmitAchievementFunc  func(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error)
	VerifyAchievementFunc  func(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error)
	RejectAchievementFunc  func(ctx context.Context, userID uuid.UUID, achievementID string, rejectionNote string) (*service.AchievementResponse, error)
}

func (m *MockAchievementServiceForRoute) GetAchievements(ctx context.Context, userID uuid.UUID, page, limit int, status string) (*service.AchievementListResponse, error) {
	if m.GetAchievementsFunc != nil {
		return m.GetAchievementsFunc(ctx, userID, page, limit, status)
	}
	return nil, nil
}

func (m *MockAchievementServiceForRoute) GetAchievementByID(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error) {
	if m.GetAchievementByIDFunc != nil {
		return m.GetAchievementByIDFunc(ctx, userID, achievementID)
	}
	return nil, errors.New("not found")
}

func (m *MockAchievementServiceForRoute) CreateAchievement(ctx context.Context, userID uuid.UUID, req *service.CreateAchievementRequest) (*service.AchievementResponse, error) {
	if m.CreateAchievementFunc != nil {
		return m.CreateAchievementFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockAchievementServiceForRoute) UpdateAchievement(ctx context.Context, userID uuid.UUID, achievementID string, req *service.UpdateAchievementRequest) (*service.AchievementResponse, error) {
	if m.UpdateAchievementFunc != nil {
		return m.UpdateAchievementFunc(ctx, userID, achievementID, req)
	}
	return nil, nil
}

func (m *MockAchievementServiceForRoute) DeleteAchievement(ctx context.Context, userID uuid.UUID, achievementID string) error {
	if m.DeleteAchievementFunc != nil {
		return m.DeleteAchievementFunc(ctx, userID, achievementID)
	}
	return nil
}

func (m *MockAchievementServiceForRoute) SubmitAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error) {
	if m.SubmitAchievementFunc != nil {
		return m.SubmitAchievementFunc(ctx, userID, achievementID)
	}
	return nil, nil
}

func (m *MockAchievementServiceForRoute) VerifyAchievement(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error) {
	if m.VerifyAchievementFunc != nil {
		return m.VerifyAchievementFunc(ctx, userID, achievementID)
	}
	return nil, nil
}

func (m *MockAchievementServiceForRoute) RejectAchievement(ctx context.Context, userID uuid.UUID, achievementID string, rejectionNote string) (*service.AchievementResponse, error) {
	if m.RejectAchievementFunc != nil {
		return m.RejectAchievementFunc(ctx, userID, achievementID, rejectionNote)
	}
	return nil, nil
}

func (m *MockAchievementServiceForRoute) GetAchievementHistory(ctx context.Context, userID uuid.UUID, achievementID string) ([]service.AchievementHistoryResponse, error) {
	return nil, nil
}

func (m *MockAchievementServiceForRoute) UploadAttachment(ctx context.Context, userID uuid.UUID, achievementID string, filePath string) (string, error) {
	return "", nil
}

func setupAuthMiddlewareForAchievement(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("username", "testuser")
		c.Locals("permissions", []string{"achievements:read", "achievements:create", "achievements:update", "achievements:delete"})
		c.Locals("role_name", "Student")
		return c.Next()
	})
}

func TestAchievementRoute_GetAchievements_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForAchievement(app)

	achievementID := primitive.NewObjectID()
	response := &service.AchievementListResponse{
		Data: []service.AchievementResponse{
			{
				ID:              achievementID.Hex(),
				AchievementType: model.AchievementTypeCompetition,
				Title:           "Test Achievement",
				Status:          model.StatusDraft,
			},
		},
		Page:       1,
		Limit:      10,
		Total:      1,
		TotalPages: 1,
	}

	mockAchievementService := &MockAchievementServiceForRoute{
		GetAchievementsFunc: func(ctx context.Context, userID uuid.UUID, page, limit int, status string) (*service.AchievementListResponse, error) {
			return response, nil
		},
	}

	route.RegisterAchievementRoutes(app, mockAchievementService)

	req := httptest.NewRequest("GET", "/achievements", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAchievementRoute_GetAchievementByID_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForAchievement(app)

	achievementID := primitive.NewObjectID()
	response := &service.AchievementResponse{
		ID:              achievementID.Hex(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Test Achievement",
		Status:          model.StatusDraft,
	}

	mockAchievementService := &MockAchievementServiceForRoute{
		GetAchievementByIDFunc: func(ctx context.Context, userID uuid.UUID, id string) (*service.AchievementResponse, error) {
			return response, nil
		},
	}

	route.RegisterAchievementRoutes(app, mockAchievementService)

	req := httptest.NewRequest("GET", "/achievements/"+achievementID.Hex(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAchievementRoute_GetAchievementByID_NotFound(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForAchievement(app)

	mockAchievementService := &MockAchievementServiceForRoute{
		GetAchievementByIDFunc: func(ctx context.Context, userID uuid.UUID, achievementID string) (*service.AchievementResponse, error) {
			return nil, errors.New("prestasi tidak ditemukan")
		},
	}

	route.RegisterAchievementRoutes(app, mockAchievementService)

	req := httptest.NewRequest("GET", "/achievements/"+primitive.NewObjectID().Hex(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestAchievementRoute_SubmitAchievement_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForAchievement(app)

	achievementID := primitive.NewObjectID()
	response := &service.AchievementResponse{
		ID:              achievementID.Hex(),
		AchievementType: model.AchievementTypeCompetition,
		Title:           "Test Achievement",
		Status:          model.StatusSubmitted,
	}

	mockAchievementService := &MockAchievementServiceForRoute{
		SubmitAchievementFunc: func(ctx context.Context, userID uuid.UUID, id string) (*service.AchievementResponse, error) {
			return response, nil
		},
	}

	route.RegisterAchievementRoutes(app, mockAchievementService)

	req := httptest.NewRequest("POST", "/achievements/"+achievementID.Hex()+"/submit", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAchievementRoute_DeleteAchievement_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddlewareForAchievement(app)

	achievementID := primitive.NewObjectID()

	mockAchievementService := &MockAchievementServiceForRoute{
		DeleteAchievementFunc: func(ctx context.Context, userID uuid.UUID, id string) error {
			return nil
		},
	}

	route.RegisterAchievementRoutes(app, mockAchievementService)

	req := httptest.NewRequest("DELETE", "/achievements/"+achievementID.Hex(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

