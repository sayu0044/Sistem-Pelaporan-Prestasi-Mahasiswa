package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error)
	Register(ctx context.Context, userData *model.User, password string) (*model.User, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error)
}

type authService struct {
	userRepo  repository.UserRepository
	roleRepo  repository.RoleRepository
	jwtSecret string
	jwtExpiry time.Duration
}

type Claims struct {
	UserID      uuid.UUID  `json:"user_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	RoleID      *uuid.UUID `json:"role_id"`
	RoleName    string     `json:"role_name"`
	Permissions []string   `json:"permissions"` // Format: ["resource:action", "achievements:create", etc.]
	jwt.RegisteredClaims
}

func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, string, *model.User, *model.Role, error) {
	user, err := s.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		return "", "", nil, nil, errors.New("username atau password salah")
	}

	if !user.IsActive {
		return "", "", nil, nil, errors.New("akun tidak aktif")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", nil, nil, errors.New("username atau password salah")
	}

	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindRoleByID(ctx, *user.RoleID)
		if err != nil {
			return "", "", nil, nil, errors.New("gagal memuat data role")
		}
	}

	token, err := s.generateToken(user, role)
	if err != nil {
		return "", "", nil, nil, err
	}

	refreshToken, err := s.generateRefreshToken(user, role)
	if err != nil {
		return "", "", nil, nil, errors.New("gagal generate refresh token")
	}

	return token, refreshToken, user, role, nil
}

func (s *authService) Register(ctx context.Context, userData *model.User, password string) (*model.User, error) {
	_, err := s.userRepo.FindUserByUsername(ctx, userData.Username)
	if err == nil {
		return nil, errors.New("username sudah digunakan")
	}

	_, err = s.userRepo.FindUserByEmail(ctx, userData.Email)
	if err == nil {
		return nil, errors.New("email sudah digunakan")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userData.PasswordHash = string(hashedPassword)

	err = s.userRepo.CreateUser(ctx, userData)
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

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindRoleByID(ctx, *user.RoleID)
		if err != nil {
			return nil, nil, errors.New("gagal memuat data role")
		}
	}

	return user, role, nil
}

func (s *authService) RefreshToken(ctx context.Context, tokenString string) (string, string, *model.User, *model.Role, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", "", nil, nil, errors.New("token tidak valid")
	}

	user, err := s.userRepo.FindUserByID(ctx, claims.UserID)
	if err != nil {
		return "", "", nil, nil, errors.New("user tidak ditemukan")
	}

	if !user.IsActive {
		return "", "", nil, nil, errors.New("akun tidak aktif")
	}

	var role *model.Role
	if user.RoleID != nil {
		role, err = s.roleRepo.FindRoleByID(ctx, *user.RoleID)
		if err != nil {
			return "", "", nil, nil, errors.New("gagal memuat data role")
		}
	}

	newToken, err := s.generateToken(user, role)
	if err != nil {
		return "", "", nil, nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(user, role)
	if err != nil {
		return "", "", nil, nil, errors.New("gagal generate refresh token")
	}

	return newToken, newRefreshToken, user, role, nil
}
