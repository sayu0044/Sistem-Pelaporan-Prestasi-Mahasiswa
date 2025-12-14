package middleware_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
	"github.com/stretchr/testify/assert"
)

type MockAuthService struct {
	ValidateTokenFunc func(tokenString string) (*service.Claims, error)
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error) {
	return "", "", nil, nil, nil
}

func (m *MockAuthService) Register(ctx context.Context, userData *model.User, password string) (*model.User, error) {
	return nil, nil
}

func (m *MockAuthService) ValidateToken(tokenString string) (*service.Claims, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(tokenString)
	}
	return nil, errors.New("invalid token")
}

func (m *MockAuthService) RefreshToken(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error) {
	return "", "", nil, nil, nil
}

func (m *MockAuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
	return nil, nil, nil
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthService{}

	app.Use(middleware.JWTMiddleware(mockAuthService))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_EmptyToken(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthService{}

	app.Use(middleware.JWTMiddleware(mockAuthService))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(tokenString string) (*service.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}

	app.Use(middleware.JWTMiddleware(mockAuthService))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	app := fiber.New()
	userID := uuid.New()
	roleID := uuid.New()
	claims := &service.Claims{
		UserID:      userID,
		Username:    "testuser",
		Email:       "test@example.com",
		RoleID:      &roleID,
		RoleName:    "Student",
		Permissions: []string{"achievements:read"},
	}

	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(tokenString string) (*service.Claims, error) {
			return claims, nil
		},
	}

	app.Use(middleware.JWTMiddleware(mockAuthService))
	app.Get("/test", func(c *fiber.Ctx) error {
		userIDFromContext := c.Locals("user_id")
		usernameFromContext := c.Locals("username")
		assert.Equal(t, userID, userIDFromContext)
		assert.Equal(t, "testuser", usernameFromContext)
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestJWTMiddleware_TokenWithoutBearerPrefix(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(tokenString string) (*service.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}

	app.Use(middleware.JWTMiddleware(mockAuthService))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "token-without-bearer")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

