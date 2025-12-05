package repository

import (
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/database"
	"gorm.io/gorm"
)

type LecturerRepository interface {
	Create(lecturer *model.Lecturer) error
	FindByID(id uuid.UUID) (*model.Lecturer, error)
	FindByUserID(userID uuid.UUID) (*model.Lecturer, error)
	FindByLecturerID(lecturerID string) (*model.Lecturer, error)
	FindAll() ([]model.Lecturer, error)
	Update(lecturer *model.Lecturer) error
	Delete(id uuid.UUID) error
	FindAdvisees(lecturerID uuid.UUID) ([]model.Student, error)
}

type lecturerRepository struct {
	db *gorm.DB
}

func NewLecturerRepository() LecturerRepository {
	return &lecturerRepository{
		db: database.DB,
	}
}

func (r *lecturerRepository) Create(lecturer *model.Lecturer) error {
	return r.db.Create(lecturer).Error
}

func (r *lecturerRepository) FindByID(id uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.Preload("User").Where("id = ?", id).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindByLecturerID(lecturerID string) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.Preload("User").Where("lecturer_id = ?", lecturerID).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindAll() ([]model.Lecturer, error) {
	var lecturers []model.Lecturer
	err := r.db.Preload("User").Find(&lecturers).Error
	return lecturers, err
}

func (r *lecturerRepository) Update(lecturer *model.Lecturer) error {
	return r.db.Save(lecturer).Error
}

func (r *lecturerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Lecturer{}, id).Error
}

func (r *lecturerRepository) FindAdvisees(lecturerID uuid.UUID) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Preload("User").Where("advisor_id = ?", lecturerID).Find(&students).Error
	return students, err
}
