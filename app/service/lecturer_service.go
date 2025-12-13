package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
)

type LecturerService interface {
	GetAllLecturers(ctx context.Context) ([]model.Lecturer, error)
	GetLecturerAdvisees(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error)
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
	studentRepo  repository.StudentRepository
}

func NewLecturerService(
	lecturerRepo repository.LecturerRepository,
	studentRepo repository.StudentRepository,
) LecturerService {
	return &lecturerService{
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
	}
}

func (s *lecturerService) GetAllLecturers(ctx context.Context) ([]model.Lecturer, error) {
	lecturers, err := s.lecturerRepo.FindAllLecturers(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data dosen: %v", err)
	}
	return lecturers, nil
}

func (s *lecturerService) GetLecturerAdvisees(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error) {
	_, err := s.lecturerRepo.FindLecturerByID(ctx, lecturerID)
	if err != nil {
		return nil, errors.New("dosen tidak ditemukan")
	}

	advisees, err := s.lecturerRepo.FindAdvisees(ctx, lecturerID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data mahasiswa bimbingan: %v", err)
	}

	return advisees, nil
}
