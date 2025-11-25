package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/app/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect menghubungkan ke database PostgreSQL
func Connect() {
	// Load .env file jika ada
	godotenv.Load()

	var dsn string

	// Cek apakah menggunakan DB_DSN (connection string) atau variabel terpisah
	dbDSN := getEnv("DB_DSN", "")
	if dbDSN != "" {
		// Gunakan connection string langsung
		dsn = dbDSN
		log.Println("Menggunakan DB_DSN untuk koneksi database")
	} else {
		// Baca env terpisah
		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "5432")
		dbUser := getEnv("DB_USER", "postgres")
		dbPassword := getEnv("DB_PASSWORD", "")
		dbName := getEnv("DB_NAME", "backend")

		// Validasi konfigurasi penting
		if dbPassword == "" {
			log.Println("Warning: DB_PASSWORD tidak di-set. Pastikan password PostgreSQL sudah dikonfigurasi di .env")
		}
		if dbName == "" {
			log.Fatal("DB_NAME tidak boleh kosong")
		}

		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			dbHost,
			dbUser,
			dbPassword,
			dbName,
			dbPort,
		)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Gagal menghubungkan ke database:", err)
	}

	log.Println("Database berhasil terhubung")

	// Auto migrate tables
	AutoMigrate()
}

// AutoMigrate melakukan migrasi tabel secara otomatis
func AutoMigrate() {
	err := DB.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.User{},
		&model.Student{},
		&model.Lecturer{},
		&model.AchievementReference{},
	)

	if err != nil {
		log.Fatal("Gagal melakukan migrasi database:", err)
	}

	log.Println("Migrasi database berhasil")
}

// getEnv helper untuk membaca environment variable
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

