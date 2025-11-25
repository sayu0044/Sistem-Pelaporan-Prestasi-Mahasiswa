package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementStatus string

const (
	StatusDraft     AchievementStatus = "draft"
	StatusSubmitted AchievementStatus = "submitted"
	StatusVerified  AchievementStatus = "verified"
	StatusRejected  AchievementStatus = "rejected"
)

type AchievementReference struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	StudentID          uuid.UUID         `gorm:"type:uuid;not null" json:"student_id"`
	Student            Student           `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	MongoAchievementID string            `gorm:"type:varchar(24);not null" json:"mongo_achievement_id"`
	Status             AchievementStatus `gorm:"type:achievement_status;default:'draft'" json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at,omitempty"`
	VerifiedAt         *time.Time        `json:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID        `gorm:"type:uuid" json:"verified_by,omitempty"`
	Verifier           User              `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
	RejectionNote      string            `gorm:"type:text" json:"rejection_note,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

func (a *AchievementReference) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
