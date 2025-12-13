package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementStatistics struct {
	TotalByType                  map[string]int64
	TotalByPeriod               []PeriodStat
	TopStudents                 []TopStudentStat
	CompetitionLevelDistribution map[string]int64
}

type PeriodStat struct {
	Period string
	Count  int64
}

type TopStudentStat struct {
	StudentID         string
	TotalPoints       float64
	TotalAchievements int64
}

type AchievementRepository interface {
	// MongoDB operations
	CreateAchievement(ctx context.Context, achievement *model.Achievement) (*model.Achievement, error)
	FindAchievementByID(ctx context.Context, id string) (*model.Achievement, error)
	FindAchievementsByStudentID(ctx context.Context, studentID string) ([]model.Achievement, error)
	UpdateAchievement(ctx context.Context, id string, achievement *model.Achievement) error
	SoftDeleteAchievement(ctx context.Context, id string) error

	// PostgreSQL operations
	CreateReference(ctx context.Context, reference *model.AchievementReference) error
	UpdateReference(ctx context.Context, reference *model.AchievementReference) error
	FindReferenceByID(ctx context.Context, id uuid.UUID) (*model.AchievementReference, error)
	FindReferenceByMongoID(ctx context.Context, mongoID string) (*model.AchievementReference, error)
	FindReferencesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID) ([]model.AchievementReference, error)
	FindReferencesWithPagination(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]model.AchievementReference, int64, error)
	DeleteReference(ctx context.Context, id uuid.UUID) error

	// Statistics operations
	GetAchievementStatistics(ctx context.Context, studentIDs []string) (*AchievementStatistics, error)
}

type achievementRepository struct {
	mongoCollection *mongo.Collection
	db              *gorm.DB
}

func NewAchievementRepository(db *gorm.DB, mongoDB *mongo.Database) AchievementRepository {
	return &achievementRepository{
		mongoCollection: mongoDB.Collection("achievements"),
		db:              db,
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

func (r *achievementRepository) CreateReference(ctx context.Context, reference *model.AchievementReference) error {
	return r.db.WithContext(ctx).Create(reference).Error
}

func (r *achievementRepository) UpdateReference(ctx context.Context, reference *model.AchievementReference) error {
	return r.db.WithContext(ctx).Save(reference).Error
}

func (r *achievementRepository) FindReferenceByID(ctx context.Context, id uuid.UUID) (*model.AchievementReference, error) {
	var reference model.AchievementReference
	err := r.db.WithContext(ctx).Preload("Student").Preload("Student.User").Preload("Verifier").
		Where("id = ?", id).First(&reference).Error
	if err != nil {
		return nil, err
	}
	return &reference, nil
}

func (r *achievementRepository) FindReferenceByMongoID(ctx context.Context, mongoID string) (*model.AchievementReference, error) {
	var reference model.AchievementReference
	err := r.db.WithContext(ctx).Preload("Student").Preload("Student.User").Preload("Verifier").
		Where("mongo_achievement_id = ?", mongoID).First(&reference).Error
	if err != nil {
		return nil, err
	}
	return &reference, nil
}

func (r *achievementRepository) FindReferencesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID) ([]model.AchievementReference, error) {
	var references []model.AchievementReference
	err := r.db.WithContext(ctx).Preload("Student").Preload("Student.User").Preload("Verifier").
		Where("student_id IN ?", studentIDs).
		Order("created_at DESC").
		Find(&references).Error
	return references, err
}

func (r *achievementRepository) DeleteReference(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.AchievementReference{}, id).Error
}

func (r *achievementRepository) FindReferencesWithPagination(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]model.AchievementReference, int64, error) {
	var references []model.AchievementReference
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AchievementReference{}).Where("student_id IN ?", studentIDs)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Student").Preload("Student.User").Preload("Verifier").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&references).Error

	return references, total, err
}

func (r *achievementRepository) GetAchievementStatistics(ctx context.Context, studentIDs []string) (*AchievementStatistics, error) {
	stats := &AchievementStatistics{
		TotalByType:                  make(map[string]int64),
		TotalByPeriod:               []PeriodStat{},
		TopStudents:                 []TopStudentStat{},
		CompetitionLevelDistribution: make(map[string]int64),
	}

	matchFilter := bson.M{
		"deletedAt": bson.M{"$exists": false},
	}

	if len(studentIDs) > 0 {
		matchFilter["studentId"] = bson.M{"$in": studentIDs}
	}

	totalByTypePipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id":   "$achievementType",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.mongoCollection.Aggregate(ctx, totalByTypePipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.TotalByType[result.ID] = result.Count
		}
	}

	totalByPeriodPipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$createdAt"},
				"month": bson.M{"$month": "$createdAt"},
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id.year": -1, "_id.month": -1}},
	}

	cursor, err = r.mongoCollection.Aggregate(ctx, totalByPeriodPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    struct {
				Year  int `bson:"year"`
				Month int `bson:"month"`
			} `bson:"_id"`
			Count int64 `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			period := fmt.Sprintf("%d-%02d", result.ID.Year, result.ID.Month)
			stats.TotalByPeriod = append(stats.TotalByPeriod, PeriodStat{
				Period: period,
				Count:  result.Count,
			})
		}
	}

	topStudentsPipeline := []bson.M{
		{"$match": matchFilter},
		{"$group": bson.M{
			"_id":              "$studentId",
			"totalPoints":      bson.M{"$sum": "$points"},
			"totalAchievements": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"totalPoints": -1}},
		{"$limit": 10},
	}

	cursor, err = r.mongoCollection.Aggregate(ctx, topStudentsPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID               string  `bson:"_id"`
			TotalPoints      float64 `bson:"totalPoints"`
			TotalAchievements int64  `bson:"totalAchievements"`
		}
		if err := cursor.Decode(&result); err == nil {
			stats.TopStudents = append(stats.TopStudents, TopStudentStat{
				StudentID:         result.ID,
				TotalPoints:       result.TotalPoints,
				TotalAchievements: result.TotalAchievements,
			})
		}
	}

	competitionMatchFilter := bson.M{
		"achievementType": "competition",
		"deletedAt":       bson.M{"$exists": false},
	}

	if len(studentIDs) > 0 {
		competitionMatchFilter["studentId"] = bson.M{"$in": studentIDs}
	}

	competitionLevelPipeline := []bson.M{
		{"$match": competitionMatchFilter},
		{"$group": bson.M{
			"_id":   "$details.competitionLevel",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err = r.mongoCollection.Aggregate(ctx, competitionLevelPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    *string `bson:"_id"`
			Count int64   `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			if result.ID != nil {
				stats.CompetitionLevelDistribution[*result.ID] = result.Count
			}
		}
	}

	return stats, nil
}
