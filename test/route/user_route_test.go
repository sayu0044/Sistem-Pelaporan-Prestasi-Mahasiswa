package route_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/route"
	"github.com/stretchr/testify/assert"
)

type MockUserServiceForRoute struct {
	GetAllUsersFunc func(ctx context.Context) ([]model.User, error)
	GetUserByIDFunc func(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error)
	CreateUserFunc  func(ctx context.Context, username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error)
	UpdateUserFunc  func(ctx context.Context, userID uuid.UUID, username, email, fullName string, roleID *uuid.UUID, isActive *bool) (*model.User, *model.Role, error)
	DeleteUserFunc  func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockUserServiceForRoute) GetAllUsers(ctx context.Context) ([]model.User, error) {
	if m.GetAllUsersFunc != nil {
		return m.GetAllUsersFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserServiceForRoute) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, nil, errors.New("not found")
}

func (m *MockUserServiceForRoute) CreateUser(ctx context.Context, username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, username, email, password, fullName, roleID, isActive, lecturerID, department, studentID, programStudy, academicYear, advisorID)
	}
	return nil, nil, nil
}

func (m *MockUserServiceForRoute) UpdateUser(ctx context.Context, userID uuid.UUID, username, email, fullName string, roleID *uuid.UUID, isActive *bool) (*model.User, *model.Role, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, userID, username, email, fullName, roleID, isActive)
	}
	return nil, nil, nil
}

func (m *MockUserServiceForRoute) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserServiceForRoute) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) (*model.User, *model.Role, error) {
	return nil, nil, nil
}

func setupAuthMiddleware(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("username", "testuser")
		c.Locals("permissions", []string{"users:read", "users:create", "users:update", "users:delete"})
		c.Locals("role_name", "Admin")
		return c.Next()
	})
}

func TestUserRoute_GetAllUsers_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddleware(app)

	user1 := model.User{
		ID:       uuid.New(),
		Username: "user1",
		Email:    "user1@example.com",
		FullName: "User 1",
	}
	user2 := model.User{
		ID:       uuid.New(),
		Username: "user2",
		Email:    "user2@example.com",
		FullName: "User 2",
	}

	mockUserService := &MockUserServiceForRoute{
		GetAllUsersFunc: func(ctx context.Context) ([]model.User, error) {
			return []model.User{user1, user2}, nil
		},
	}

	route.RegisterUserRoutes(app, mockUserService)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserRoute_GetUserByID_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddleware(app)

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

	mockUserService := &MockUserServiceForRoute{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, *model.Role, error) {
			return user, role, nil
		},
	}

	route.RegisterUserRoutes(app, mockUserService)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserRoute_GetUserByID_InvalidUUID(t *testing.T) {
	app := fiber.New()
	setupAuthMiddleware(app)

	mockUserService := &MockUserServiceForRoute{}
	route.RegisterUserRoutes(app, mockUserService)

	req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUserRoute_GetUserByID_NotFound(t *testing.T) {
	app := fiber.New()
	setupAuthMiddleware(app)

	mockUserService := &MockUserServiceForRoute{
		GetUserByIDFunc: func(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
			return nil, nil, errors.New("user tidak ditemukan")
		},
	}

	route.RegisterUserRoutes(app, mockUserService)

	req := httptest.NewRequest("GET", "/users/"+uuid.New().String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestUserRoute_CreateUser_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddleware(app)

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "newuser",
		Email:    "newuser@example.com",
		FullName: "New User",
		RoleID:   &roleID,
	}
	role := &model.Role{
		ID:   roleID,
		Name: "Student",
	}

	mockUserService := &MockUserServiceForRoute{
		CreateUserFunc: func(ctx context.Context, username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error) {
			return user, role, nil
		},
	}

	route.RegisterUserRoutes(app, mockUserService)

	reqBody := map[string]interface{}{
		"username":  "newuser",
		"email":     "newuser@example.com",
		"password":  "password123",
		"full_name": "New User",
		"role_id":   roleID.String(),
		"is_active": true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestUserRoute_DeleteUser_Success(t *testing.T) {
	app := fiber.New()
	setupAuthMiddleware(app)

	userID := uuid.New()

	mockUserService := &MockUserServiceForRoute{
		GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, *model.Role, error) {
			return &model.User{ID: id}, nil, nil
		},
		DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	route.RegisterUserRoutes(app, mockUserService)

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

