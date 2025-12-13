package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const postgresSchemaSQL = `DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;

DROP TABLE IF EXISTS achievement_references CASCADE;
DROP TABLE IF EXISTS students CASCADE;
DROP TABLE IF EXISTS lecturers CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

DROP TYPE IF EXISTS achievement_status CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE achievement_status AS ENUM ('draft', 'submitted', 'verified', 'rejected');

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE lecturers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    lecturer_id VARCHAR(20) UNIQUE NOT NULL,
    department VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    student_id VARCHAR(20) UNIQUE NOT NULL,
    program_study VARCHAR(100),
    academic_year VARCHAR(10),
    advisor_id UUID REFERENCES lecturers(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE achievement_references (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    mongo_achievement_id VARCHAR(24) NOT NULL,
    status achievement_status NOT NULL DEFAULT 'draft',
    submitted_at TIMESTAMP,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    rejection_note TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_students_user_id ON students(user_id);
CREATE INDEX idx_students_advisor_id ON students(advisor_id);
CREATE INDEX idx_lecturers_user_id ON lecturers(user_id);
CREATE INDEX idx_achievement_references_student_id ON achievement_references(student_id);
CREATE INDEX idx_achievement_references_status ON achievement_references(status);
CREATE INDEX idx_achievement_references_verified_by ON achievement_references(verified_by);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_achievement_references_updated_at BEFORE UPDATE ON achievement_references
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`

const postgresSeedDataSQL = `DELETE FROM achievement_references;
DELETE FROM students;
DELETE FROM lecturers;
DELETE FROM role_permissions;
DELETE FROM users;
DELETE FROM permissions;
DELETE FROM roles;

INSERT INTO roles (name, description) VALUES
('Admin', 'Pengelola sistem dengan akses penuh'),
('Mahasiswa', 'Pelapor prestasi'),
('Dosen Wali', 'Verifikator prestasi mahasiswa bimbingannya');

INSERT INTO permissions (name, resource, action, description) VALUES
('achievements:create', 'achievements', 'create', 'Membuat prestasi baru'),
('achievements:read', 'achievements', 'read', 'Membaca data prestasi'),
('achievements:update', 'achievements', 'update', 'Mengupdate data prestasi'),
('achievements:delete', 'achievements', 'delete', 'Menghapus data prestasi'),
('achievements:verify', 'achievements', 'verify', 'Memverifikasi prestasi'),
('user:create', 'user', 'create', 'Membuat pengguna baru'),
('user:read', 'user', 'read', 'Membaca data pengguna'),
('user:update', 'user', 'update', 'Mengupdate data pengguna'),
('user:delete', 'user', 'delete', 'Menghapus pengguna'),
('user:manage', 'user', 'manage', 'Mengelola pengguna'),
('student:read', 'student', 'read', 'Membaca data mahasiswa'),
('student:update', 'student', 'update', 'Mengupdate data mahasiswa'),
('lecturer:read', 'lecturer', 'read', 'Membaca data dosen');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE (r.name = 'Admin' AND p.name IN (
    'achievements:create', 'achievements:read', 'achievements:update', 
    'achievements:delete', 'achievements:verify', 'user:create', 
    'user:read', 'user:update', 'user:delete', 'user:manage',
    'student:read', 'student:update', 'lecturer:read'
))
OR (r.name = 'Mahasiswa' AND p.name IN (
    'achievements:create', 'achievements:read', 'achievements:update', 'achievements:delete'
))
OR (r.name = 'Dosen Wali' AND p.name IN (
    'achievements:read', 'achievements:verify'
));

INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
SELECT 
    'admin',
    'admin@gmail.com',
    '$2y$10$4OE9fraSi08LQNvr9XTnxe8/EfXbsmU8j6TSBAMXg7ZT3V/JzFc9G',
    'Administrator',
    r.id,
    true
FROM roles r
WHERE r.name = 'Admin'
LIMIT 1;`

func main() {
	log.Println("Starting database migration...")

	config.LoadEnv()

	postgresDB, err := connectPostgreSQL()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	mongoDB, err := connectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoDB.Client().Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	if err := RunMigrations(postgresDB, mongoDB); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migration completed successfully")
}

func connectPostgreSQL() (*sql.DB, error) {
	var dsn string

	dbDSN := config.GetEnv("DB_DSN", "")
	if dbDSN != "" {
		dsn = dbDSN
		log.Println("Using DB_DSN for database connection")
	} else {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			config.DBHost,
			config.DBUser,
			config.DBPassword,
			config.DBName,
			config.DBPort,
		)
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("PostgreSQL connected successfully")
	return sqlDB, nil
}

func connectMongoDB() (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("MongoDB connected successfully")
	return client.Database(config.MongoDBName), nil
}

func RunMigrations(postgresDB *sql.DB, mongoDB *mongo.Database) error {
	log.Println("Running database migrations...")

	if err := runPostgresMigrations(postgresDB); err != nil {
		return fmt.Errorf("postgres migrations failed: %w", err)
	}

	if err := runMongoMigrations(mongoDB); err != nil {
		return fmt.Errorf("mongo migrations failed: %w", err)
	}

	log.Println("All migrations completed successfully")
	return nil
}

func runPostgresMigrations(db *sql.DB) error {
	log.Println("Running PostgreSQL schema and seed migrations...")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	log.Println("Executing PostgreSQL schema...")
	if _, err := tx.Exec(postgresSchemaSQL); err != nil {
		return fmt.Errorf("executing schema SQL: %w", err)
	}

	log.Println("Executing PostgreSQL seed data...")
	if _, err := tx.Exec(postgresSeedDataSQL); err != nil {
		return fmt.Errorf("executing seed data SQL: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.Println("PostgreSQL migrations completed")
	return nil
}

func runMongoMigrations(db *mongo.Database) error {
	log.Println("Running MongoDB migrations...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := dropCollectionIfExists(ctx, db, "achievements"); err != nil {
		return err
	}

	if err := createAchievementIndexes(ctx, db); err != nil {
		return err
	}

	log.Println("MongoDB migrations completed")
	return nil
}

func dropCollectionIfExists(ctx context.Context, db *mongo.Database, collectionName string) error {
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if err != nil {
		return fmt.Errorf("list collections for %s: %w", collectionName, err)
	}

	if len(collections) == 0 {
		log.Printf("Collection %s does not exist, skipping drop", collectionName)
		return nil
	}

	if err := db.Collection(collectionName).Drop(ctx); err != nil {
		return fmt.Errorf("drop collection %s: %w", collectionName, err)
	}

	log.Printf("Dropped collection: %s", collectionName)
	return nil
}

func createAchievementIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("achievements")

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "studentId", Value: 1}},
			Options: options.Index().SetName("idx_student_id"),
		},
		{
			Keys:    bson.D{{Key: "achievementType", Value: 1}},
			Options: options.Index().SetName("idx_achievement_type"),
		},
		{
			Keys:    bson.D{{Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "description", Value: "text"},
			},
			Options: options.Index().SetName("idx_text_search"),
		},
		{
			Keys:    bson.D{{Key: "deletedAt", Value: 1}},
			Options: options.Index().SetName("idx_deleted_at"),
		},
	}

	if _, err := collection.Indexes().CreateMany(ctx, indexModels); err != nil {
		return fmt.Errorf("create achievement indexes: %w", err)
	}

	log.Println("Created indexes for achievements collection")
	return nil
}
