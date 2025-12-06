package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AchievementHistory model untuk audit trail di PostgreSQL
type AchievementHistory struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	AchievementRefID   uuid.UUID         `gorm:"type:uuid;not null" json:"achievement_ref_id"`
	AchievementRef     AchievementReference `gorm:"foreignKey:AchievementRefID" json:"achievement_ref,omitempty"`
	MongoAchievementID string            `gorm:"type:varchar(24);not null" json:"mongo_achievement_id"`
	OldStatus          *AchievementStatus `gorm:"type:achievement_status" json:"old_status,omitempty"` // Nullable untuk status awal
	NewStatus          AchievementStatus  `gorm:"type:achievement_status;not null" json:"new_status"`
	ChangedBy          uuid.UUID         `gorm:"type:uuid;not null" json:"changed_by"`
	ChangedByUser      User              `gorm:"foreignKey:ChangedBy" json:"changed_by_user,omitempty"`
	Notes              string            `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
}

func (a *AchievementHistory) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

