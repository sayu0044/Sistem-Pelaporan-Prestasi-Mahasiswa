package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Lecturer struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LecturerID string    `gorm:"type:varchar(20);unique;not null" json:"lecturer_id"`
	Department string    `gorm:"type:varchar(100)" json:"department"`
	CreatedAt  time.Time `json:"created_at"`
}

func (l *Lecturer) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

