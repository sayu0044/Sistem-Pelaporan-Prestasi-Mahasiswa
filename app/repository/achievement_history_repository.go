package repository

import (
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/database"
	"gorm.io/gorm"
)

type AchievementHistoryRepository interface {
	Create(history *model.AchievementHistory) error
	FindByAchievementRefID(achievementRefID uuid.UUID) ([]model.AchievementHistory, error)
	FindByMongoAchievementID(mongoID string) ([]model.AchievementHistory, error)
}

type achievementHistoryRepository struct {
	db *gorm.DB
}

func NewAchievementHistoryRepository() AchievementHistoryRepository {
	return &achievementHistoryRepository{
		db: database.DB,
	}
}

func (r *achievementHistoryRepository) Create(history *model.AchievementHistory) error {
	return r.db.Create(history).Error
}

func (r *achievementHistoryRepository) FindByAchievementRefID(achievementRefID uuid.UUID) ([]model.AchievementHistory, error) {
	var histories []model.AchievementHistory
	err := r.db.Preload("ChangedByUser").Preload("AchievementRef").
		Where("achievement_ref_id = ?", achievementRefID).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

func (r *achievementHistoryRepository) FindByMongoAchievementID(mongoID string) ([]model.AchievementHistory, error) {
	var histories []model.AchievementHistory
	err := r.db.Preload("ChangedByUser").Preload("AchievementRef").
		Where("mongo_achievement_id = ?", mongoID).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

