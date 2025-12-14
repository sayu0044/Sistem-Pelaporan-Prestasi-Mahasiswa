package repository_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAchievementTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "cgo") {
			t.Skip("Skipping test: CGO is disabled. Run tests with CGO_ENABLED=1")
		}
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Student{}, &model.AchievementReference{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewAchievementRepository(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)

		assert.NotNil(t, repo)
	})
}

func TestAchievementRepository_CreateAchievement(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		achievement := &model.Achievement{
			StudentID:       uuid.New().String(),
			AchievementType: model.AchievementTypeCompetition,
			Title:           "Test Achievement",
			Description:     "Test Description",
			Status:          model.StatusDraft,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		result, err := repo.CreateAchievement(ctx, achievement)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, primitive.NilObjectID, result.ID)
	})
}

func TestAchievementRepository_FindAchievementByID(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		objectID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    0,
			Message: "not found",
		}))

		result, err := repo.FindAchievementByID(ctx, objectID.Hex())

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAchievementRepository_CreateReference(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		reference := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: primitive.NewObjectID().Hex(),
			Status:             model.StatusDraft,
		}

		err := repo.CreateReference(ctx, reference)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, reference.ID)
	})
}

func TestAchievementRepository_FindReferenceByID(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		reference := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: primitive.NewObjectID().Hex(),
			Status:             model.StatusDraft,
		}
		repo.CreateReference(ctx, reference)

		found, err := repo.FindReferenceByID(ctx, reference.ID)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, reference.ID, found.ID)
	})
}

func TestAchievementRepository_FindReferenceByMongoID(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		mongoID := primitive.NewObjectID().Hex()
		reference := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: mongoID,
			Status:             model.StatusDraft,
		}
		repo.CreateReference(ctx, reference)

		found, err := repo.FindReferenceByMongoID(ctx, mongoID)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, mongoID, found.MongoAchievementID)
	})
}

func TestAchievementRepository_UpdateReference(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		reference := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: primitive.NewObjectID().Hex(),
			Status:             model.StatusDraft,
		}
		repo.CreateReference(ctx, reference)

		now := time.Now()
		reference.Status = model.StatusSubmitted
		reference.SubmittedAt = &now

		err := repo.UpdateReference(ctx, reference)

		assert.NoError(t, err)

		updated, _ := repo.FindReferenceByID(ctx, reference.ID)
		assert.Equal(t, model.StatusSubmitted, updated.Status)
	})
}

func TestAchievementRepository_DeleteReference(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		reference := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: primitive.NewObjectID().Hex(),
			Status:             model.StatusDraft,
		}
		repo.CreateReference(ctx, reference)

		err := repo.DeleteReference(ctx, reference.ID)
		assert.NoError(t, err)

		_, err = repo.FindReferenceByID(ctx, reference.ID)
		assert.Error(t, err)
	})
}

func TestAchievementRepository_FindReferencesByStudentIDs(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		ref1 := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: primitive.NewObjectID().Hex(),
			Status:             model.StatusDraft,
		}
		ref2 := &model.AchievementReference{
			StudentID:          student.ID,
			MongoAchievementID: primitive.NewObjectID().Hex(),
			Status:             model.StatusSubmitted,
		}
		repo.CreateReference(ctx, ref1)
		repo.CreateReference(ctx, ref2)

		references, err := repo.FindReferencesByStudentIDs(ctx, []uuid.UUID{student.ID})

		assert.NoError(t, err)
		assert.Len(t, references, 2)
	})
}

func TestAchievementRepository_FindReferencesWithPagination(t *testing.T) {
	db := setupAchievementTestDB(t)
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mongoDB := mt.DB
		repo := repository.NewAchievementRepository(db, mongoDB)
		ctx := context.Background()

		user := &model.User{
			Username:     "student1",
			Email:        "student1@example.com",
			PasswordHash: "hash",
			FullName:     "Student 1",
			IsActive:     true,
		}
		db.Create(user)

		student := &model.Student{
			UserID:       user.ID,
			StudentID:    "1234567890",
			ProgramStudy: "CS",
			AcademicYear: "2024",
		}
		db.Create(student)

		for i := 0; i < 5; i++ {
			ref := &model.AchievementReference{
				StudentID:          student.ID,
				MongoAchievementID: primitive.NewObjectID().Hex(),
				Status:             model.StatusDraft,
			}
			repo.CreateReference(ctx, ref)
		}

		references, total, err := repo.FindReferencesWithPagination(ctx, []uuid.UUID{student.ID}, 1, 2)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, references, 2)
	})
}
