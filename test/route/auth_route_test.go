package route_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/stretchr/testify/assert"
)

type MockAuthServiceForRoute struct {
	LoginFunc       func(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error)
	RefreshTokenFunc func(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error)
}

func (m *MockAuthServiceForRoute) Login(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, username, password)
	}
	return "", "", nil, nil, errors.New("not implemented")
}

func (m *MockAuthServiceForRoute) Register(ctx context.Context, userData *model.User, password string) (*model.User, error) {
	return nil, nil
}

func (m *MockAuthServiceForRoute) ValidateToken(tokenString string) (*service.Claims, error) {
	return nil, nil
}

func (m *MockAuthServiceForRoute) RefreshToken(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, tokenString)
	}
	return "", "", nil, nil, errors.New("not implemented")
}

func (m *MockAuthServiceForRoute) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
	return nil, nil, nil
}

func registerAuthRoutesForTest(app *fiber.App, authService service.AuthService) {
	authPublic := app.Group("/api/v1/auth")
	{
		authPublic.Post("/login", func(c *fiber.Ctx) error {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Permintaan tidak valid",
					"message": "Pastikan body permintaan Anda dalam format JSON yang benar.",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			token, refreshToken, user, role, err := authService.Login(ctx, req.Username, req.Password)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Gagal login",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"error": false,
				"data": fiber.Map{
					"token":        token,
					"refreshToken": refreshToken,
					"user":         user,
					"role":         role,
				},
			})
		})

		authPublic.Post("/refresh", func(c *fiber.Ctx) error {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Token tidak ditemukan",
					"message": "Pastikan header 'Authorization: Bearer <token>' dikirim",
				})
			}

			token := authHeader
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			if token == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Token kosong",
					"message": "Format: 'Authorization: Bearer <token>'",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			newToken, newRefreshToken, user, role, err := authService.RefreshToken(ctx, token)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Gagal refresh token",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"error": false,
				"data": fiber.Map{
					"token":        newToken,
					"refreshToken": newRefreshToken,
					"user":         user,
					"role":         role,
				},
			})
		})
	}
}

func TestAuthRoute_Login_Success(t *testing.T) {
	app := fiber.New()
	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		RoleID:   &roleID,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	mockAuthService := &MockAuthServiceForRoute{
		LoginFunc: func(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error) {
			return "access-token", "refresh-token", user, role, nil
		},
	}

	registerAuthRoutesForTest(app, mockAuthService)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthRoute_Login_InvalidBody(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthServiceForRoute{}

	registerAuthRoutesForTest(app, mockAuthService)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthRoute_Login_InvalidCredentials(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthServiceForRoute{
		LoginFunc: func(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error) {
			return "", "", nil, nil, errors.New("username atau password salah")
		},
	}

	registerAuthRoutesForTest(app, mockAuthService)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRoute_RefreshToken_Success(t *testing.T) {
	app := fiber.New()
	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		RoleID:   &roleID,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	mockAuthService := &MockAuthServiceForRoute{
		RefreshTokenFunc: func(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error) {
			return "new-access-token", "new-refresh-token", user, role, nil
		},
	}

	registerAuthRoutesForTest(app, mockAuthService)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthRoute_RefreshToken_MissingToken(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthServiceForRoute{}

	registerAuthRoutesForTest(app, mockAuthService)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRoute_RefreshToken_InvalidToken(t *testing.T) {
	app := fiber.New()
	mockAuthService := &MockAuthServiceForRoute{
		RefreshTokenFunc: func(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error) {
			return "", "", nil, nil, errors.New("invalid token")
		},
	}

	registerAuthRoutesForTest(app, mockAuthService)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
