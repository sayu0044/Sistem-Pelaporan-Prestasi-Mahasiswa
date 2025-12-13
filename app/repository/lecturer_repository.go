package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"gorm.io/gorm"
)

type LecturerRepository interface {
	CreateLecturer(ctx context.Context, lecturer *model.Lecturer) error
	FindLecturerByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error)
	FindLecturerByUserID(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error)
	FindLecturerByLecturerID(ctx context.Context, lecturerID string) (*model.Lecturer, error)
	FindAllLecturers(ctx context.Context) ([]model.Lecturer, error)
	UpdateLecturer(ctx context.Context, lecturer *model.Lecturer) error
	DeleteLecturer(ctx context.Context, id uuid.UUID) error
	FindAdvisees(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error)
}

type lecturerRepository struct {
	db *gorm.DB
}

func NewLecturerRepository(db *gorm.DB) LecturerRepository {
	return &lecturerRepository{
		db: db,
	}
}

func (r *lecturerRepository) CreateLecturer(ctx context.Context, lecturer *model.Lecturer) error {
	return r.db.WithContext(ctx).Create(lecturer).Error
}

func (r *lecturerRepository) FindLecturerByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindLecturerByUserID(ctx context.Context, userID uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindLecturerByLecturerID(ctx context.Context, lecturerID string) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.WithContext(ctx).Preload("User").Where("lecturer_id = ?", lecturerID).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindAllLecturers(ctx context.Context) ([]model.Lecturer, error) {
	var lecturers []model.Lecturer
	err := r.db.WithContext(ctx).Preload("User").Find(&lecturers).Error
	return lecturers, err
}

func (r *lecturerRepository) UpdateLecturer(ctx context.Context, lecturer *model.Lecturer) error {
	return r.db.WithContext(ctx).Save(lecturer).Error
}

func (r *lecturerRepository) DeleteLecturer(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Lecturer{}, id).Error
}

func (r *lecturerRepository) FindAdvisees(ctx context.Context, lecturerID uuid.UUID) ([]model.Student, error) {
	var students []model.Student
	err := r.db.WithContext(ctx).Preload("User").Where("advisor_id = ?", lecturerID).Find(&students).Error
	return students, err
}
