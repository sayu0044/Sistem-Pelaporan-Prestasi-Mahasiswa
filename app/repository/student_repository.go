package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"gorm.io/gorm"
)

type StudentRepository interface {
	CreateStudent(ctx context.Context, student *model.Student) error
	FindStudentByID(ctx context.Context, id uuid.UUID) (*model.Student, error)
	FindStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error)
	FindStudentByStudentID(ctx context.Context, studentID string) (*model.Student, error)
	FindAllStudents(ctx context.Context) ([]model.Student, error)
	UpdateStudent(ctx context.Context, student *model.Student) error
	DeleteStudent(ctx context.Context, id uuid.UUID) error
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{
		db: db,
	}
}

func (r *studentRepository) CreateStudent(ctx context.Context, student *model.Student) error {
	return r.db.WithContext(ctx).Create(student).Error
}

func (r *studentRepository) FindStudentByID(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.WithContext(ctx).Preload("User").Preload("Advisor").Preload("Advisor.User").Where("id = ?", id).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.WithContext(ctx).Preload("User").Preload("Advisor").Preload("Advisor.User").Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindStudentByStudentID(ctx context.Context, studentID string) (*model.Student, error) {
	var student model.Student
	err := r.db.WithContext(ctx).Preload("User").Preload("Advisor").Preload("Advisor.User").Where("student_id = ?", studentID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindAllStudents(ctx context.Context) ([]model.Student, error) {
	var students []model.Student
	err := r.db.WithContext(ctx).Preload("User").Preload("Advisor").Preload("Advisor.User").Find(&students).Error
	return students, err
}

func (r *studentRepository) UpdateStudent(ctx context.Context, student *model.Student) error {
	return r.db.WithContext(ctx).Model(&model.Student{}).Where("id = ?", student.ID).Update("advisor_id", student.AdvisorID).Error
}

func (r *studentRepository) DeleteStudent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Student{}, id).Error
}
