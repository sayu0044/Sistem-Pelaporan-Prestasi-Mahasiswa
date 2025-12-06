package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	// MongoDB operations
	CreateAchievement(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error)
	FindAchievementByID(ctx context.Context, id string) (*model.Achievement, error)
	FindAchievementsByStudentID(ctx context.Context, studentID string) ([]model.Achievement, error)
	UpdateAchievement(ctx context.Context, id string, achievement *model.Achievement) error
	SoftDeleteAchievement(ctx context.Context, id string) error

	// PostgreSQL operations
	CreateReference(reference *model.AchievementReference) error
	UpdateReference(reference *model.AchievementReference) error
	FindReferenceByID(id uuid.UUID) (*model.AchievementReference, error)
	FindReferenceByMongoID(mongoID string) (*model.AchievementReference, error)
	FindReferencesByStudentIDs(studentIDs []uuid.UUID) ([]model.AchievementReference, error)
	FindReferencesWithPagination(studentIDs []uuid.UUID, page, limit int) ([]model.AchievementReference, int64, error)
	DeleteReference(id uuid.UUID) error
}

type achievementRepository struct {
	mongoCollection *mongo.Collection
	db              *gorm.DB
}

func NewAchievementRepository() AchievementRepository {
	return &achievementRepository{
		mongoCollection: database.MongoDB.Collection("achievements"),
		db:              database.DB,
	}
}

// MongoDB operations

func (r *achievementRepository) CreateAchievement(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error) {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	result, err := r.mongoCollection.InsertOne(ctx, achievement)
	if err != nil {
		return nil, err
	}

	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return achievement, nil
}

func (r *achievementRepository) FindAchievementByID(ctx context.Context, id string) (*model.Achievement, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var achievement model.Achievement
	filter := bson.M{
		"_id":       objectID,
		"deletedAt": bson.M{"$exists": false},
	}

	err = r.mongoCollection.FindOne(ctx, filter).Decode(&achievement)
	if err != nil {
		return nil, err
	}

	return &achievement, nil
}

func (r *achievementRepository) FindAchievementsByStudentID(ctx context.Context, studentID string) ([]model.Achievement, error) {
	filter := bson.M{
		"studentId": studentID,
		"deletedAt": bson.M{"$exists": false},
	}

	cursor, err := r.mongoCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []model.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

func (r *achievementRepository) UpdateAchievement(ctx context.Context, id string, achievement *model.Achievement) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	achievement.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"achievementType": achievement.AchievementType,
			"title":           achievement.Title,
			"description":     achievement.Description,
			"details":         achievement.Details,
			"attachments":     achievement.Attachments,
			"tags":            achievement.Tags,
			"points":          achievement.Points,
			"status":          achievement.Status,
			"updatedAt":       achievement.UpdatedAt,
		},
	}

	filter := bson.M{
		"_id":       objectID,
		"deletedAt": bson.M{"$exists": false},
	}

	_, err = r.mongoCollection.UpdateOne(ctx, filter, update)
	return err
}

func (r *achievementRepository) SoftDeleteAchievement(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deletedAt": now,
			"updatedAt": now,
		},
	}

	filter := bson.M{"_id": objectID}
	_, err = r.mongoCollection.UpdateOne(ctx, filter, update)
	return err
}

// PostgreSQL operations

func (r *achievementRepository) CreateReference(reference *model.AchievementReference) error {
	return r.db.Create(reference).Error
}

func (r *achievementRepository) UpdateReference(reference *model.AchievementReference) error {
	return r.db.Save(reference).Error
}

func (r *achievementRepository) FindReferenceByID(id uuid.UUID) (*model.AchievementReference, error) {
	var reference model.AchievementReference
	err := r.db.Preload("Student").Preload("Student.User").Preload("Verifier").
		Where("id = ?", id).First(&reference).Error
	if err != nil {
		return nil, err
	}
	return &reference, nil
}

func (r *achievementRepository) FindReferenceByMongoID(mongoID string) (*model.AchievementReference, error) {
	var reference model.AchievementReference
	err := r.db.Preload("Student").Preload("Student.User").Preload("Verifier").
		Where("mongo_achievement_id = ?", mongoID).First(&reference).Error
	if err != nil {
		return nil, err
	}
	return &reference, nil
}

func (r *achievementRepository) FindReferencesByStudentIDs(studentIDs []uuid.UUID) ([]model.AchievementReference, error) {
	var references []model.AchievementReference
	err := r.db.Preload("Student").Preload("Student.User").Preload("Verifier").
		Where("student_id IN ?", studentIDs).
		Order("created_at DESC").
		Find(&references).Error
	return references, err
}

func (r *achievementRepository) DeleteReference(id uuid.UUID) error {
	return r.db.Delete(&model.AchievementReference{}, id).Error
}

// Helper function untuk pagination
func (r *achievementRepository) FindReferencesWithPagination(studentIDs []uuid.UUID, page, limit int) ([]model.AchievementReference, int64, error) {
	var references []model.AchievementReference
	var total int64

	query := r.db.Model(&model.AchievementReference{}).Where("student_id IN ?", studentIDs)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Preload("Student").Preload("Student.User").Preload("Verifier").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&references).Error

	return references, total, err
}
