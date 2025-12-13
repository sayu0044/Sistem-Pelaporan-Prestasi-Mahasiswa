package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, username, email, fullName string, roleID *uuid.UUID, isActive *bool) (*model.User, *model.Role, error)
	GetAllUsers(ctx context.Context) ([]model.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) (*model.User, *model.Role, error)
}

type userService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	lecturerRepo repository.LecturerRepository
	studentRepo  repository.StudentRepository
	authService  AuthService
}

func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	lecturerRepo repository.LecturerRepository,
	studentRepo repository.StudentRepository,
	authService AuthService,
) UserService {
	return &userService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
		authService:  authService,
	}
}

func (s *userService) CreateUser(ctx context.Context, username, email, password, fullName string, roleID uuid.UUID, isActive bool, lecturerID, department, studentID, programStudy, academicYear string, advisorID *uuid.UUID) (*model.User, *model.Role, error) {
	role, err := s.roleRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, nil, errors.New("role tidak ditemukan")
	}

	user := &model.User{
		Username: username,
		Email:    email,
		FullName: fullName,
		RoleID:   &roleID,
		IsActive: isActive,
	}

	registeredUser, err := s.authService.Register(ctx, user, password)
	if err != nil {
		return nil, nil, err
	}

	roleNameLower := strings.ToLower(role.Name)
	if strings.Contains(roleNameLower, "dosen") {
		if lecturerID == "" {
			return nil, nil, errors.New("lecturer_id harus diisi untuk role dosen")
		}

		existingLecturer, err := s.lecturerRepo.FindLecturerByLecturerID(ctx, lecturerID)
		if err == nil && existingLecturer != nil {
			return nil, nil, errors.New("lecturer_id sudah digunakan")
		}

		lecturer := &model.Lecturer{
			UserID:     registeredUser.ID,
			LecturerID: lecturerID,
			Department: department,
		}

		if err := s.lecturerRepo.CreateLecturer(ctx, lecturer); err != nil {
			return nil, nil, errors.New("gagal membuat data lecturer: " + err.Error())
		}
	}

	if strings.Contains(roleNameLower, "mahasiswa") || strings.Contains(roleNameLower, "student") {
		if studentID == "" {
			return nil, nil, errors.New("student_id harus diisi untuk role mahasiswa")
		}

		existingStudent, err := s.studentRepo.FindStudentByStudentID(ctx, studentID)
		if err == nil && existingStudent != nil {
			return nil, nil, errors.New("student_id sudah digunakan")
		}

		if advisorID != nil {
			_, err := s.lecturerRepo.FindLecturerByID(ctx, *advisorID)
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

		if err := s.studentRepo.CreateStudent(ctx, student); err != nil {
			return nil, nil, errors.New("gagal membuat data student: " + err.Error())
		}
	}

	userWithRole, err := s.userRepo.FindUserByID(ctx, registeredUser.ID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data user")
	}

	return userWithRole, role, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	users, err := s.userRepo.FindAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, *model.Role, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

	role, err := s.roleRepo.FindRoleByID(ctx, *user.RoleID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data role")
	}

	return user, role, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	err = s.userRepo.DeleteUser(ctx, userID)
	if err != nil {
		return errors.New("gagal menghapus user")
	}

	return nil
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, username, email, fullName string, roleID *uuid.UUID, isActive *bool) (*model.User, *model.Role, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

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

	if roleID != nil {
		_, err := s.roleRepo.FindRoleByID(ctx, *roleID)
		if err != nil {
			return nil, nil, errors.New("role tidak ditemukan")
		}

		user.RoleID = roleID
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, nil, errors.New("gagal mengupdate user")
	}

	updatedUser, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data user")
	}

	role, err := s.roleRepo.FindRoleByID(ctx, *updatedUser.RoleID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data role")
	}

	return updatedUser, role, nil
}

func (s *userService) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) (*model.User, *model.Role, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("user tidak ditemukan")
	}

	_, err = s.roleRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, nil, errors.New("role tidak ditemukan")
	}

	user.RoleID = &roleID

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, nil, errors.New("gagal mengupdate role user")
	}

	updatedUser, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data user")
	}

	updatedRole, err := s.roleRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, nil, errors.New("gagal memuat data role")
	}

	return updatedUser, updatedRole, nil
}
