package repository

import (
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/database"
	"gorm.io/gorm"
)

type StudentRepository interface {
	Create(student *model.Student) error
	FindByID(id uuid.UUID) (*model.Student, error)
	FindByUserID(userID uuid.UUID) (*model.Student, error)
	FindByStudentID(studentID string) (*model.Student, error)
	FindAll() ([]model.Student, error)
	Update(student *model.Student) error
	Delete(id uuid.UUID) error
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository() StudentRepository {
	return &studentRepository{
		db: database.DB,
	}
}

func (r *studentRepository) Create(student *model.Student) error {
	return r.db.Create(student).Error
}

func (r *studentRepository) FindByID(id uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.Preload("User").Preload("Advisor").Preload("Advisor.User").Where("id = ?", id).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.Preload("User").Preload("Advisor").Preload("Advisor.User").Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindByStudentID(studentID string) (*model.Student, error) {
	var student model.Student
	err := r.db.Preload("User").Preload("Advisor").Preload("Advisor.User").Where("student_id = ?", studentID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindAll() ([]model.Student, error) {
	var students []model.Student
	err := r.db.Preload("User").Preload("Advisor").Preload("Advisor.User").Find(&students).Error
	return students, err
}

func (r *studentRepository) Update(student *model.Student) error {
	return r.db.Save(student).Error
}

func (r *studentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Student{}, id).Error
}
