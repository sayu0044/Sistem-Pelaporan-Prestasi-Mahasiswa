package service

import (
	"errors"
	"time"

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
}

type authService struct {
	userRepo  repository.UserRepository
	roleRepo  repository.RoleRepository
	jwtSecret string
	jwtExpiry time.Duration
}

type Claims struct {
	UserID   uuid.UUID  `json:"user_id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	RoleID   *uuid.UUID `json:"role_id"`
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

	token, err := s.generateToken(user)
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

func (s *authService) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RoleID:   user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
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
