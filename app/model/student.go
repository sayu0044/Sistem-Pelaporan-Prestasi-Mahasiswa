package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Student struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	StudentID   string    `gorm:"type:varchar(20);unique;not null" json:"student_id"`
	ProgramStudy string   `gorm:"type:varchar(100)" json:"program_study"`
	AcademicYear string   `gorm:"type:varchar(10)" json:"academic_year"`
	AdvisorID   *uuid.UUID `gorm:"type:uuid" json:"advisor_id"`
	Advisor     Lecturer  `gorm:"foreignKey:AdvisorID" json:"advisor,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Student) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

