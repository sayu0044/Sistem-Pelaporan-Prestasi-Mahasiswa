package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(username, password string) (string, *model.User, error)
	Register(userData *model.User, password string) (*model.User, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (string, *model.User, error)
	HandleLogin(c *fiber.Ctx) error
	HandleGetMe(c *fiber.Ctx) error
	HandleRefreshToken(c *fiber.Ctx) error
	HandleLogout(c *fiber.Ctx) error
	HandleProfile(c *fiber.Ctx) error
	HandleHealthCheck(c *fiber.Ctx) error
	HandleTest(c *fiber.Ctx) error
}

type authService struct {
	userRepo  repository.UserRepository
	roleRepo  repository.RoleRepository
	jwtSecret string
	jwtExpiry time.Duration
}

type Claims struct {
	UserID     uuid.UUID  `json:"user_id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	RoleID     *uuid.UUID `json:"role_id"`
	RoleName   string     `json:"role_name"`
	Permissions []string  `json:"permissions"` // Format: ["resource:action", "achievements:create", etc.]
	jwt.RegisteredClaims
}

func NewAuthService(jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		userRepo:  repository.NewUserRepository(),
		roleRepo:  repository.NewRoleRepository(),
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *authService) Login(username, password string) (string, *model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", nil, errors.New("username atau password salah")
	}

	if !user.IsActive {
		return "", nil, errors.New("akun tidak aktif")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, errors.New("username atau password salah")
	}

	// Load role dengan permissions untuk dimasukkan ke token
	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindByID(*user.RoleID)
		if err != nil {
			return "", nil, errors.New("gagal memuat data role")
		}
	}

	token, err := s.generateToken(user, role)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) Register(userData *model.User, password string) (*model.User, error) {
	// Check if username exists
	_, err := s.userRepo.FindByUsername(userData.Username)
	if err == nil {
		return nil, errors.New("username sudah digunakan")
	}

	// Check if email exists
	_, err = s.userRepo.FindByEmail(userData.Email)
	if err == nil {
		return nil, errors.New("email sudah digunakan")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userData.PasswordHash = string(hashedPassword)

	// Create user
	err = s.userRepo.Create(userData)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (s *authService) generateToken(user *model.User, role *model.Role) (string, error) {
	// Format permissions ke dalam format "resource:action"
	permissions := []string{}
	roleName := ""
	
	if role != nil {
		roleName = role.Name
		// Check if admin (case-insensitive)
		roleNameLower := strings.ToLower(roleName)
		if strings.Contains(roleNameLower, "admin") {
			// Admin memiliki semua permissions, bisa ditandai dengan "*:*" atau semua permissions
			permissions = append(permissions, "*:*")
		} else {
			// Format permissions sebagai "resource:action"
			for _, perm := range role.Permissions {
				permissionString := strings.ToLower(perm.Resource) + ":" + strings.ToLower(perm.Action)
				permissions = append(permissions, permissionString)
			}
		}
	}

	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		RoleID:      user.RoleID,
		RoleName:    roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) generateRefreshToken(user *model.User, role *model.Role) (string, error) {
	// Format permissions ke dalam format "resource:action"
	permissions := []string{}
	roleName := ""
	
	if role != nil {
		roleName = role.Name
		// Check if admin (case-insensitive)
		roleNameLower := strings.ToLower(roleName)
		if strings.Contains(roleNameLower, "admin") {
			permissions = append(permissions, "*:*")
		} else {
			for _, perm := range role.Permissions {
				permissionString := strings.ToLower(perm.Resource) + ":" + strings.ToLower(perm.Action)
				permissions = append(permissions, permissionString)
			}
		}
	}

	// Refresh token memiliki expiry lebih lama (7 hari)
	refreshExpiry := 7 * 24 * time.Hour

	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		RoleID:      user.RoleID,
		RoleName:    roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token tidak valid")
}

func (s *authService) HandleLogin(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	token, user, err := s.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Load role dengan permissions untuk generate refresh token
	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindByID(*user.RoleID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal memuat data role",
			})
		}
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(user, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal generate refresh token",
		})
	}

	// Format permissions untuk response
	var permissions []string
	var roleName string
	if role != nil {
		roleName = role.Name
		roleNameLower := strings.ToLower(roleName)
		if strings.Contains(roleNameLower, "admin") {
			permissions = append(permissions, "*:*")
		} else {
			for _, perm := range role.Permissions {
				permissionString := strings.ToLower(perm.Resource) + ":" + strings.ToLower(perm.Action)
				permissions = append(permissions, permissionString)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":        token,
			"refreshToken": refreshToken,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"fullName":    user.FullName,
				"role":        roleName,
				"permissions": permissions,
			},
		},
	})
}

func (s *authService) RefreshToken(tokenString string) (string, *model.User, error) {
	// Validate existing token
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", nil, errors.New("token tidak valid")
	}

	// Get user from database
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", nil, errors.New("user tidak ditemukan")
	}

	if !user.IsActive {
		return "", nil, errors.New("akun tidak aktif")
	}

	// Load role dengan permissions untuk generate token baru
	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindByID(*user.RoleID)
		if err != nil {
			return "", nil, errors.New("gagal memuat data role")
		}
	}

	// Generate new token
	newToken, err := s.generateToken(user, role)
	if err != nil {
		return "", nil, err
	}

	return newToken, user, nil
}

func (s *authService) HandleGetMe(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": fiber.Map{
				"user_id":     c.Locals("user_id"),
				"username":    c.Locals("username"),
				"role":        c.Locals("role_name"),
				"permissions": c.Locals("permissions"),
			},
		},
	})
}

func (s *authService) HandleRefreshToken(c *fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Token tidak ditemukan. Pastikan header 'Authorization: Bearer <token>' dikirim",
		})
	}

	// Remove "Bearer " prefix if exists
	token := authHeader
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Token kosong. Format: 'Authorization: Bearer <token>'",
		})
	}

	// Refresh token
	newToken, user, err := s.RefreshToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Load role dengan permissions untuk generate refresh token baru
	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindByID(*user.RoleID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal memuat data role",
			})
		}
	}

	// Generate refresh token baru
	newRefreshToken, err := s.generateRefreshToken(user, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal generate refresh token",
		})
	}

	// Format permissions untuk response
	var permissions []string
	var roleName string
	if role != nil {
		roleName = role.Name
		roleNameLower := strings.ToLower(roleName)
		if strings.Contains(roleNameLower, "admin") {
			permissions = append(permissions, "*:*")
		} else {
			for _, perm := range role.Permissions {
				permissionString := strings.ToLower(perm.Resource) + ":" + strings.ToLower(perm.Action)
				permissions = append(permissions, permissionString)
			}
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":        newToken,
			"refreshToken": newRefreshToken,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"fullName":    user.FullName,
				"role":        roleName,
				"permissions": permissions,
			},
		},
	})
}

func (s *authService) HandleLogout(c *fiber.Ctx) error {
	// JWT is stateless, so logout is mainly handled on client side
	// In a production system, you might want to implement token blacklisting
	// For now, we just return success message
	
	// Get user info from context for logging purposes
	userID := c.Locals("user_id")
	username := c.Locals("username")

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": fiber.Map{
				"user_id":  userID,
				"username": username,
			},
		},
	})
}

func (s *authService) HandleProfile(c *fiber.Ctx) error {
	// Get user ID from context
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User ID tidak ditemukan",
		})
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "User ID tidak valid",
		})
	}

	// Get user from database with full details
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "User tidak ditemukan",
		})
	}

	// Load role dengan permissions
	var role *model.Role
	var permissions []fiber.Map
	var roleData fiber.Map
	if user.RoleID != nil {
		role, err = s.roleRepo.FindByID(*user.RoleID)
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

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"fullName":    user.FullName,
			"role":        roleData,
			"permissions": permissions,
			"isActive":    user.IsActive,
			"createdAt":   user.CreatedAt,
			"updatedAt":   user.UpdatedAt,
		},
	})
}

func (s *authService) HandleHealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "API Sistem Pelaporan Prestasi Mahasiswa",
		"status":  "running",
	})
}

func (s *authService) HandleTest(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Protected route berhasil diakses",
	})
}
