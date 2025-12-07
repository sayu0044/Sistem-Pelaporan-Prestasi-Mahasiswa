package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type UserService interface {
	CreateUser(username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error)
	UpdateUser(userID uuid.UUID, username, email, fullName string, roleID *uuid.UUID, isActive *bool) (*model.User, *model.Role, error)
	GetAllUsers() ([]model.User, error)
	GetUserByID(userID uuid.UUID) (*model.User, *model.Role, error)
	DeleteUser(userID uuid.UUID) error
	UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) (*model.User, *model.Role, error)
	HandleCreateUser(c *fiber.Ctx) error
	HandleUpdateUser(c *fiber.Ctx) error
	HandleGetAllUsers(c *fiber.Ctx) error
	HandleGetUserByID(c *fiber.Ctx) error
	HandleDeleteUser(c *fiber.Ctx) error
	HandleUpdateUserRole(c *fiber.Ctx) error
}

type userService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	lecturerRepo repository.LecturerRepository
	studentRepo  repository.StudentRepository
	authService  AuthService
}

func NewUserService(authService AuthService) UserService {
	return &userService{
		userRepo:     repository.NewUserRepository(),
		roleRepo:     repository.NewRoleRepository(),
		lecturerRepo: repository.NewLecturerRepository(),
		studentRepo:  repository.NewStudentRepository(),
		authService:  authService,
	}
}

func (s *userService) CreateUser(username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error) {
	// Cek apakah role ada dan load permissions-nya
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return nil, nil, errors.New("role tidak ditemukan")
	}

	// Parse user data
	user := &model.User{
		Username: username,
		Email:    email,
		FullName: fullName,
		RoleID:   &roleID,
		IsActive: isActive,
	}

	// Register user (akan hash password)
	registeredUser, err := s.authService.Register(user, password)
	if err != nil {
		return nil, nil, err
	}

	// Auto-create Lecturer jika role adalah dosen (case-insensitive check)
	roleNameLower := strings.ToLower(role.Name)
	if strings.Contains(roleNameLower, "dosen") {
		if lecturerID == "" {
			return nil, nil, errors.New("lecturer_id harus diisi untuk role dosen")
		}

		// Cek apakah lecturer_id sudah digunakan
		existingLecturer, err := s.lecturerRepo.FindByLecturerID(lecturerID)
		if err == nil && existingLecturer != nil {
			return nil, nil, errors.New("lecturer_id sudah digunakan")
		}

		lecturer := &model.Lecturer{
			UserID:     registeredUser.ID,
			LecturerID: lecturerID,
			Department: department,
		}

		if err := s.lecturerRepo.Create(lecturer); err != nil {
			return nil, nil, errors.New("gagal membuat data lecturer: " + err.Error())
		}
	}

	// Auto-create Student jika role adalah mahasiswa (case-insensitive check)
	if strings.Contains(roleNameLower, "mahasiswa") || strings.Contains(roleNameLower, "student") {
		if studentID == "" {
			return nil, nil, errors.New("student_id harus diisi untuk role mahasiswa")
		}

		// Cek apakah student_id sudah digunakan
		existingStudent, err := s.studentRepo.FindByStudentID(studentID)
		if err == nil && existingStudent != nil {
			return nil, nil, errors.New("student_id sudah digunakan")
		}

		// Validasi advisor_id jika diberikan
		if advisorID != nil {
			_, err := s.lecturerRepo.FindByID(*advisorID)
			if err != nil {
				return nil, nil, errors.New("advisor_id tidak ditemukan di tabel lecturers")
			}
		}

		student := &model.Student{
			UserID:       registeredUser.ID,
			StudentID:    studentID,
			ProgramStudy: programStudy,
			AcademicYear: academicYear,
			AdvisorID:    advisorID,
		}

		if err := s.studentRepo.Create(student); err != nil {
			return nil, nil, errors.New("gagal membuat data student: " + err.Error())
		}
	}

	// Load user untuk mendapatkan data lengkap dengan role dan permissions
	userWithRole, err := s.userRepo.FindByID(registeredUser.ID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data user")
	}

	return userWithRole, role, nil
}

func (s *userService) GetAllUsers() ([]model.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) GetUserByID(userID uuid.UUID) (*model.User, *model.Role, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

	// Load role dengan permissions
	role, err := s.roleRepo.FindByID(*user.RoleID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data role")
	}

	return user, role, nil
}

func (s *userService) DeleteUser(userID uuid.UUID) error {
	// Cek apakah user ada
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// Hapus user
	err = s.userRepo.Delete(userID)
	if err != nil {
		return errors.New("gagal menghapus user")
	}

	return nil
}

func (s *userService) UpdateUser(userID uuid.UUID, username, email, fullName string, roleID *uuid.UUID, isActive *bool) (*model.User, *model.Role, error) {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

	// Update fields jika diberikan
	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	if fullName != "" {
		user.FullName = fullName
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	// Update role jika diberikan
	if roleID != nil {
		// Cek apakah role ada
		_, err := s.roleRepo.FindByID(*roleID)
		if err != nil {
			return nil, nil, errors.New("role tidak ditemukan")
		}

		user.RoleID = roleID
	}

	// Update user
	if err := s.userRepo.Update(user); err != nil {
		return nil, nil, errors.New("gagal mengupdate user")
	}

	// Reload user dengan role dan permissions
	updatedUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data user")
	}

	// Load role dengan permissions
	role, err := s.roleRepo.FindByID(*updatedUser.RoleID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data role")
	}

	return updatedUser, role, nil
}

func (s *userService) UpdateUserRole(userID uuid.UUID, roleID uuid.UUID) (*model.User, *model.Role, error) {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

	// Cek apakah role ada
	_, err = s.roleRepo.FindByID(roleID)
	if err != nil {
		return nil, nil, errors.New("role tidak ditemukan")
	}

	// Update role
	user.RoleID = &roleID

	// Update user
	if err := s.userRepo.Update(user); err != nil {
		return nil, nil, errors.New("gagal mengupdate role user")
	}

	// Reload user dengan role dan permissions
	updatedUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data user")
	}

	// Load role dengan permissions
	updatedRole, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data role")
	}

	return updatedUser, updatedRole, nil
}

func (s *userService) HandleCreateUser(c *fiber.Ctx) error {
	// Permission check sudah dilakukan oleh RBAC middleware
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive *bool  `json:"is_active"`
		// Lecturer fields
		LecturerID string `json:"lecturer_id"`
		Department string `json:"department"`
		// Student fields
		StudentID    string `json:"student_id"`
		ProgramStudy string `json:"program_study"`
		AcademicYear string `json:"academic_year"`
		AdvisorID    string `json:"advisor_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Validasi required fields
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Username, email, password, dan full_name harus diisi",
		})
	}

	// Validasi role_id
	if req.RoleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "role_id harus diisi",
		})
	}

	// Parse role_id
	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "role_id tidak valid",
		})
	}

	// Set is_active default ke true jika tidak diisi
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Parse advisor_id jika diberikan
	var advisorUUID *uuid.UUID
	if req.AdvisorID != "" {
		parsedUUID, err := uuid.Parse(req.AdvisorID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "advisor_id tidak valid",
			})
		}
		advisorUUID = &parsedUUID
	}

	// Create user via service
	user, role, err := s.CreateUser(req.Username, req.Email, req.Password, req.FullName, roleUUID, isActive, req.LecturerID, req.Department, req.StudentID, req.ProgramStudy, req.AcademicYear, advisorUUID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Format permissions
	var permissions []fiber.Map
	if len(role.Permissions) > 0 {
		for _, perm := range role.Permissions {
			permissions = append(permissions, fiber.Map{
				"id":          perm.ID,
				"name":        perm.Name,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"description": perm.Description,
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "User berhasil dibuat",
		"user": fiber.Map{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"full_name": user.FullName,
			"role_id":   user.RoleID,
			"role": fiber.Map{
				"id":          role.ID,
				"name":        role.Name,
				"description": role.Description,
			},
			"permissions": permissions,
			"is_active":   user.IsActive,
			"created_at":  user.CreatedAt,
		},
	})
}

func (s *userService) HandleUpdateUser(c *fiber.Ctx) error {
	// Permission check sudah dilakukan oleh RBAC middleware
	// Get user ID from URL parameter
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "User ID tidak valid",
		})
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Parse role_id jika diberikan
	var roleUUID *uuid.UUID
	if req.RoleID != "" {
		parsedUUID, err := uuid.Parse(req.RoleID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "role_id tidak valid",
			})
		}
		roleUUID = &parsedUUID
	}

	// Update user via service
	user, role, err := s.UpdateUser(userID, req.Username, req.Email, req.FullName, roleUUID, req.IsActive)
	if err != nil {
		// Handle different error types
		if err.Error() == "user tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		if err.Error() == "role tidak ditemukan" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Format permissions
	var permissions []fiber.Map
	if len(role.Permissions) > 0 {
		for _, perm := range role.Permissions {
			permissions = append(permissions, fiber.Map{
				"id":          perm.ID,
				"name":        perm.Name,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"description": perm.Description,
			})
		}
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "User berhasil diupdate",
		"user": fiber.Map{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"full_name": user.FullName,
			"role_id":   user.RoleID,
			"role": fiber.Map{
				"id":          role.ID,
				"name":        role.Name,
				"description": role.Description,
			},
			"permissions": permissions,
			"is_active":   user.IsActive,
			"updated_at":  user.UpdatedAt,
		},
	})
}

func (s *userService) HandleGetAllUsers(c *fiber.Ctx) error {
	// Permission check sudah dilakukan oleh RBAC middleware
	users, err := s.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Gagal mengambil data users: " + err.Error(),
		})
	}

	// Format response dengan role dan permissions
	var usersData []fiber.Map
	for _, user := range users {
		// Load role dengan permissions jika role_id ada
		var roleData fiber.Map
		var permissions []fiber.Map

		if user.RoleID != nil {
			role, err := s.roleRepo.FindByID(*user.RoleID)
			if err == nil && role != nil {
				roleData = fiber.Map{
					"id":          role.ID,
					"name":        role.Name,
					"description": role.Description,
				}

				// Format permissions
				if len(role.Permissions) > 0 {
					for _, perm := range role.Permissions {
						permissions = append(permissions, fiber.Map{
							"id":          perm.ID,
							"name":        perm.Name,
							"resource":    perm.Resource,
							"action":      perm.Action,
							"description": perm.Description,
						})
					}
				}
			}
		}

		usersData = append(usersData, fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"full_name":   user.FullName,
			"role_id":     user.RoleID,
			"role":        roleData,
			"permissions": permissions,
			"is_active":   user.IsActive,
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  usersData,
		"total": len(usersData),
	})
}

func (s *userService) HandleGetUserByID(c *fiber.Ctx) error {
	// Permission check sudah dilakukan oleh RBAC middleware
	// Get user ID from URL parameter
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "User ID tidak valid",
		})
	}

	// Get user via service
	user, role, err := s.GetUserByID(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Format permissions
	var permissions []fiber.Map
	if role != nil && len(role.Permissions) > 0 {
		for _, perm := range role.Permissions {
			permissions = append(permissions, fiber.Map{
				"id":          perm.ID,
				"name":        perm.Name,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"description": perm.Description,
			})
		}
	}

	var roleData fiber.Map
	if role != nil {
		roleData = fiber.Map{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
		}
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"full_name":   user.FullName,
			"role_id":     user.RoleID,
			"role":        roleData,
			"permissions": permissions,
			"is_active":   user.IsActive,
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
		},
	})
}

func (s *userService) HandleDeleteUser(c *fiber.Ctx) error {
	// Permission check sudah dilakukan oleh RBAC middleware
	// Get user ID from URL parameter
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "User ID tidak valid",
		})
	}

	// Delete user via service
	err = s.DeleteUser(userID)
	if err != nil {
		if err.Error() == "user tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "User berhasil dihapus",
	})
}

func (s *userService) HandleUpdateUserRole(c *fiber.Ctx) error {
	// Permission check sudah dilakukan oleh RBAC middleware
	// Get user ID from URL parameter
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "User ID tidak valid",
		})
	}

	var req struct {
		RoleID string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Validasi role_id
	if req.RoleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "role_id harus diisi",
		})
	}

	// Parse role_id
	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "role_id tidak valid",
		})
	}

	// Update user role via service
	user, role, err := s.UpdateUserRole(userID, roleUUID)
	if err != nil {
		// Handle different error types
		if err.Error() == "user tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		if err.Error() == "role tidak ditemukan" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Format permissions
	var permissions []fiber.Map
	if role != nil && len(role.Permissions) > 0 {
		for _, perm := range role.Permissions {
			permissions = append(permissions, fiber.Map{
				"id":          perm.ID,
				"name":        perm.Name,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"description": perm.Description,
			})
		}
	}

	var roleData fiber.Map
	if role != nil {
		roleData = fiber.Map{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
		}
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Role user berhasil diupdate",
		"data": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"full_name":   user.FullName,
			"role_id":     user.RoleID,
			"role":        roleData,
			"permissions": permissions,
			"is_active":   user.IsActive,
			"updated_at":  user.UpdatedAt,
		},
	})
}
