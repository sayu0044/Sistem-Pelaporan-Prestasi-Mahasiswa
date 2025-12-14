package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware_NoPermissions(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("read", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_InvalidPermissionsType(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", "invalid-type")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("read", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_ValidPermission(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", []string{"achievements:read", "achievements:create"})
		c.Locals("role_name", "Student")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("read", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRBACMiddleware_InvalidPermission(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", []string{"achievements:create"})
		c.Locals("role_name", "Student")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("delete", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_AdminRole(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", []string{})
		c.Locals("role_name", "Admin")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("delete", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRBACMiddleware_WildcardPermission(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", []string{"*:*"})
		c.Locals("role_name", "CustomRole")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("delete", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRBACMiddleware_CaseInsensitivePermission(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", []string{"ACHIEVEMENTS:READ"})
		c.Locals("role_name", "Student")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("read", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRBACMiddleware_CaseInsensitiveAdmin(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("permissions", []string{})
		c.Locals("role_name", "ADMINISTRATOR")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware("delete", "achievements"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

