package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"gorm.io/gorm"
)

type AchievementHistoryRepository interface {
	CreateHistory(ctx context.Context, history *model.AchievementHistory) error
	FindHistoriesByAchievementRefID(ctx context.Context, achievementRefID uuid.UUID) ([]model.AchievementHistory, error)
	FindHistoriesByMongoAchievementID(ctx context.Context, mongoID string) ([]model.AchievementHistory, error)
}

type achievementHistoryRepository struct {
	db *gorm.DB
}

func NewAchievementHistoryRepository(db *gorm.DB) AchievementHistoryRepository {
	return &achievementHistoryRepository{
		db: db,
	}
}

func (r *achievementHistoryRepository) CreateHistory(ctx context.Context, history *model.AchievementHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *achievementHistoryRepository) FindHistoriesByAchievementRefID(ctx context.Context, achievementRefID uuid.UUID) ([]model.AchievementHistory, error) {
	var histories []model.AchievementHistory
	err := r.db.WithContext(ctx).Preload("ChangedByUser").Preload("AchievementRef").
		Where("achievement_ref_id = ?", achievementRefID).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

func (r *achievementHistoryRepository) FindHistoriesByMongoAchievementID(ctx context.Context, mongoID string) ([]model.AchievementHistory, error) {
	var histories []model.AchievementHistory
	err := r.db.WithContext(ctx).Preload("ChangedByUser").Preload("AchievementRef").
		Where("mongo_achievement_id = ?", mongoID).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

