package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/service"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/middleware"
)

// RegisterUserRoutes mendaftarkan route untuk user management
func RegisterUserRoutes(router fiber.Router, userService service.UserService) {
	users := router.Group("/users")
	{
		// GET /api/v1/users - List semua users
		// Requires: read users permission
		users.Get("/", middleware.RBACMiddleware("read", "users"), userService.HandleGetAllUsers)

		// GET /api/v1/users/:id - Detail user
		// Requires: read users permission
		users.Get("/:id", middleware.RBACMiddleware("read", "users"), userService.HandleGetUserByID)

		// POST /api/v1/users - Create user
		// Requires: create users permission
		users.Post("/", middleware.RBACMiddleware("create", "users"), userService.HandleCreateUser)

		// PUT /api/v1/users/:id - Update user
		// Requires: update users permission
		users.Put("/:id", middleware.RBACMiddleware("update", "users"), userService.HandleUpdateUser)

		// DELETE /api/v1/users/:id - Delete user
		// Requires: delete users permission
		users.Delete("/:id", middleware.RBACMiddleware("delete", "users"), userService.HandleDeleteUser)

		// PUT /api/v1/users/:id/role - Update user role
		// Requires: update users permission
		users.Put("/:id/role", middleware.RBACMiddleware("update", "users"), userService.HandleUpdateUserRole)
	}
}

