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
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatal("Gagal menghubungkan ke database:", err)
	}

	log.Println("Database berhasil terhubung")

	// Fix schema issues (column types) - harus dipanggil sebelum AutoMigrate
	FixSchema()

	// Auto migrate tables
	AutoMigrate()
}

// AutoMigrate melakukan migrasi tabel secara otomatis
func AutoMigrate() {
	// Cek apakah student_id sudah bertipe VARCHAR
	var studentColumnType string
	var studentTableExists bool
	var needToFixAfterMigration bool

	DB.Raw(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'students'
	)`).Scan(&studentTableExists)

	if studentTableExists {
		DB.Raw(`
			SELECT data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = 'students' 
			AND column_name = 'student_id'
		`).Scan(&studentColumnType)

		// Jika kolom sudah VARCHAR, tandai untuk diperbaiki setelah migrasi
		if studentColumnType == "character varying" || studentColumnType == "varchar" {
			needToFixAfterMigration = true
		}
	}

	// Lakukan migrasi normal
	err := DB.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.User{},
		&model.Student{},
		&model.Lecturer{},
		&model.AchievementReference{},
	)

	// Jika terjadi error karena perubahan tipe kolom student_id, perbaiki
	if err != nil {
		if needToFixAfterMigration {
			log.Println("Memperbaiki tipe kolom student_id setelah error migrasi...")
			// Pastikan student_id tetap VARCHAR
			fixErr := DB.Exec(`ALTER TABLE students ALTER COLUMN student_id TYPE VARCHAR(20)`).Error
			if fixErr != nil {
				log.Printf("Warning: Gagal memperbaiki tipe kolom student_id: %v", fixErr)
			} else {
				log.Println("Tipe kolom student_id berhasil diperbaiki menjadi VARCHAR(20)")
				// Coba migrasi lagi
				err = DB.AutoMigrate(
					&model.Role{},
					&model.Permission{},
					&model.User{},
					&model.Lecturer{},
					&model.AchievementReference{},
				)
				if err != nil {
					log.Fatal("Gagal melakukan migrasi database:", err)
				}
			}
		} else {
			log.Fatal("Gagal melakukan migrasi database:", err)
		}
	}

	// Setelah migrasi berhasil, pastikan student_id tetap VARCHAR jika sebelumnya sudah VARCHAR
	if needToFixAfterMigration && err == nil {
		var currentType string
		DB.Raw(`
			SELECT data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = 'students' 
			AND column_name = 'student_id'
		`).Scan(&currentType)

		if currentType != "character varying" && currentType != "varchar" {
			log.Println("Memperbaiki tipe kolom student_id menjadi VARCHAR(20)...")
			DB.Exec(`ALTER TABLE students ALTER COLUMN student_id TYPE VARCHAR(20)`)
		}
	}

	log.Println("Migrasi database berhasil")
}

// FixSchema memperbaiki tipe kolom yang tidak sesuai dengan model
func FixSchema() {
	log.Println("Memeriksa dan memperbaiki schema database...")

	// Fix student_id column type dari UUID ke VARCHAR jika diperlukan
	var studentColumnType string
	var studentTableExists bool

	// Cek apakah tabel students ada
	DB.Raw(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'students'
	)`).Scan(&studentTableExists)

	if studentTableExists {
		err := DB.Raw(`
			SELECT data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = 'students' 
			AND column_name = 'student_id'
		`).Scan(&studentColumnType).Error

		if err == nil {
			log.Printf("Tipe kolom student_id saat ini: %s", studentColumnType)
			switch studentColumnType {
			case "uuid":
				log.Println("Memperbaiki tipe kolom student_id dari UUID ke VARCHAR...")

				// Hapus foreign key constraint yang mungkin menghalangi
				// Foreign key dari achievement_references harus merujuk ke students.id (UUID), bukan students.student_id (VARCHAR)
				// Tapi kita hapus dulu untuk memastikan tidak ada konflik
				DB.Exec(`ALTER TABLE achievement_references DROP CONSTRAINT IF EXISTS fk_achievement_student`)
				DB.Exec(`ALTER TABLE achievement_references DROP CONSTRAINT IF EXISTS fk_achievement_references_student`)

				// Hapus constraint unique jika ada
				DB.Exec(`ALTER TABLE students DROP CONSTRAINT IF EXISTS students_student_id_key`)
				// Hapus constraint not null sementara
				DB.Exec(`ALTER TABLE students ALTER COLUMN student_id DROP NOT NULL`)

				// Hapus data jika ada (karena UUID tidak bisa dikonversi langsung ke VARCHAR dengan data)
				log.Println("Menghapus data lama dari tabel students untuk memungkinkan konversi tipe...")
				DB.Exec(`DELETE FROM students`)

				// Ubah tipe kolom
				err := DB.Exec(`
					ALTER TABLE students 
					ALTER COLUMN student_id TYPE VARCHAR(20)
				`).Error
				if err != nil {
					log.Printf("Error: Gagal mengubah tipe kolom student_id: %v", err)
				} else {
					log.Println("Kolom student_id berhasil diperbaiki menjadi VARCHAR(20)")
					// Tambahkan kembali constraint unique dan not null
					DB.Exec(`ALTER TABLE students ALTER COLUMN student_id SET NOT NULL`)
					DB.Exec(`ALTER TABLE students ADD CONSTRAINT students_student_id_key UNIQUE (student_id)`)
				}
			case "character varying", "varchar":
				log.Println("Kolom student_id sudah bertipe VARCHAR, tidak perlu diperbaiki")
			}
		} else {
			log.Printf("Warning: Gagal memeriksa tipe kolom student_id: %v", err)
		}
	}

	// Fix lecturer_id column type dari UUID ke VARCHAR jika diperlukan
	var lecturerColumnType string
	var lecturerTableExists bool

	// Cek apakah tabel lecturers ada
	DB.Raw(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'lecturers'
	)`).Scan(&lecturerTableExists)

	if lecturerTableExists {
		err := DB.Raw(`
			SELECT data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			AND table_name = 'lecturers' 
			AND column_name = 'lecturer_id'
		`).Scan(&lecturerColumnType).Error

		if err == nil {
			log.Printf("Tipe kolom lecturer_id saat ini: %s", lecturerColumnType)
			switch lecturerColumnType {
			case "uuid":
				log.Println("Memperbaiki tipe kolom lecturer_id dari UUID ke VARCHAR...")
				// Hapus constraint unique jika ada
				DB.Exec(`ALTER TABLE lecturers DROP CONSTRAINT IF EXISTS lecturers_lecturer_id_key`)
				// Hapus constraint not null sementara
				DB.Exec(`ALTER TABLE lecturers ALTER COLUMN lecturer_id DROP NOT NULL`)
				// Ubah tipe kolom
				err := DB.Exec(`
					ALTER TABLE lecturers 
					ALTER COLUMN lecturer_id TYPE VARCHAR(20) USING COALESCE(lecturer_id::text, '')
				`).Error
				if err != nil {
					log.Printf("Error: Gagal mengubah tipe kolom lecturer_id: %v", err)
				} else {
					log.Println("Kolom lecturer_id berhasil diperbaiki menjadi VARCHAR(20)")
					// Tambahkan kembali constraint unique dan not null
					DB.Exec(`ALTER TABLE lecturers ALTER COLUMN lecturer_id SET NOT NULL`)
					DB.Exec(`ALTER TABLE lecturers ADD CONSTRAINT lecturers_lecturer_id_key UNIQUE (lecturer_id)`)
				}
			case "character varying", "varchar":
				log.Println("Kolom lecturer_id sudah bertipe VARCHAR, tidak perlu diperbaiki")
			}
		} else {
			log.Printf("Warning: Gagal memeriksa tipe kolom lecturer_id: %v", err)
		}
	}

	// Pastikan enum achievement_status sudah dibuat
	var enumExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_type 
			WHERE typname = 'achievement_status'
		)
	`).Scan(&enumExists)

	if !enumExists {
		log.Println("Membuat enum achievement_status...")
		err := DB.Exec(`
			CREATE TYPE achievement_status AS ENUM ('draft', 'submitted', 'verified', 'rejected')
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat enum achievement_status: %v", err)
		} else {
			log.Println("Enum achievement_status berhasil dibuat")
		}
	}

	// Pastikan foreign key constraint yang benar ada setelah perbaikan
	// Foreign key dari achievement_references ke students.id (bukan student_id)
	var fkExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.table_constraints 
			WHERE constraint_schema = 'public' 
			AND constraint_name = 'fk_achievement_references_student'
		)
	`).Scan(&fkExists)

	if !fkExists {
		log.Println("Membuat foreign key constraint fk_achievement_references_student...")
		err := DB.Exec(`
			ALTER TABLE achievement_references 
			ADD CONSTRAINT fk_achievement_references_student 
			FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat foreign key fk_achievement_references_student: %v", err)
		}
	}

	// Pastikan foreign key dari achievement_references ke users (verified_by)
	var fkVerifiedByExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.table_constraints 
			WHERE constraint_schema = 'public' 
			AND constraint_name = 'fk_achievement_references_verifier'
		)
	`).Scan(&fkVerifiedByExists)

	if !fkVerifiedByExists {
		log.Println("Membuat foreign key constraint fk_achievement_references_verifier...")
		err := DB.Exec(`
			ALTER TABLE achievement_references 
			ADD CONSTRAINT fk_achievement_references_verifier 
			FOREIGN KEY (verified_by) REFERENCES users(id) ON DELETE SET NULL
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat foreign key fk_achievement_references_verifier: %v", err)
		}
	}

	// Pastikan foreign key dari students ke users
	var fkStudentUserExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.table_constraints 
			WHERE constraint_schema = 'public' 
			AND constraint_name = 'fk_students_user'
		)
	`).Scan(&fkStudentUserExists)

	if !fkStudentUserExists {
		log.Println("Membuat foreign key constraint fk_students_user...")
		err := DB.Exec(`
			ALTER TABLE students 
			ADD CONSTRAINT fk_students_user 
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat foreign key fk_students_user: %v", err)
		}
	}

	// Pastikan foreign key dari students ke lecturers (advisor_id)
	var fkStudentAdvisorExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.table_constraints 
			WHERE constraint_schema = 'public' 
			AND constraint_name = 'fk_students_advisor'
		)
	`).Scan(&fkStudentAdvisorExists)

	if !fkStudentAdvisorExists {
		log.Println("Membuat foreign key constraint fk_students_advisor...")
		err := DB.Exec(`
			ALTER TABLE students 
			ADD CONSTRAINT fk_students_advisor 
			FOREIGN KEY (advisor_id) REFERENCES lecturers(id) ON DELETE SET NULL
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat foreign key fk_students_advisor: %v", err)
		}
	}

	// Pastikan foreign key dari lecturers ke users
	var fkLecturerUserExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.table_constraints 
			WHERE constraint_schema = 'public' 
			AND constraint_name = 'fk_lecturers_user'
		)
	`).Scan(&fkLecturerUserExists)

	if !fkLecturerUserExists {
		log.Println("Membuat foreign key constraint fk_lecturers_user...")
		err := DB.Exec(`
			ALTER TABLE lecturers 
			ADD CONSTRAINT fk_lecturers_user 
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat foreign key fk_lecturers_user: %v", err)
		}
	}

	// Pastikan foreign key dari users ke roles
	var fkUserRoleExists bool
	DB.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.table_constraints 
			WHERE constraint_schema = 'public' 
			AND constraint_name = 'fk_users_role'
		)
	`).Scan(&fkUserRoleExists)

	if !fkUserRoleExists {
		log.Println("Membuat foreign key constraint fk_users_role...")
		err := DB.Exec(`
			ALTER TABLE users 
			ADD CONSTRAINT fk_users_role 
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL
		`).Error
		if err != nil {
			log.Printf("Warning: Gagal membuat foreign key fk_users_role: %v", err)
		}
	}

	log.Println("Pemeriksaan schema selesai")
}

// getEnv helper untuk membaca environment variable
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
