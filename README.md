# Sistem Pelaporan Prestasi Mahasiswa

Sistem backend untuk pelaporan prestasi mahasiswa menggunakan Go dan Fiber framework.

## Struktur Project

```
Sistem-Pelaporan-Prestasi-Mahasiswa/
├── app/
│   ├── model/          # Struct data (User, Role, Student, dll)
│   ├── repository/     # Query database
│   └── service/        # Business logic layer
├── config/
│   ├── app.go         # Fiber instance, middleware, routes
│   ├── env.go         # Load environment variables
│   └── logger.go      # Log writer dan rotating files
├── database/
│   └── database.go    # Koneksi database PostgreSQL
├── helper/
│   └── helper.go      # Fungsi bantuan umum
├── logs/              # Hasil logger
├── middleware/
│   ├── jwt.go         # JWT authentication
│   └── error_handler.go # Error handling
├── route/
│   └── route.go       # Mendaftarkan routes
├── .env               # Environment variables
└── main.go            # Entry point aplikasi
```

## Setup

1. Install dependencies:

```bash
go mod download
```

2. Setup database PostgreSQL:

- Buat database dengan nama `backend`
- Jalankan query SQL yang sudah disediakan untuk membuat tabel

3. Konfigurasi environment:

- Copy `.env` dan sesuaikan dengan konfigurasi database Anda
- Update `DB_PASSWORD` dengan password PostgreSQL Anda

4. Jalankan aplikasi:

```bash
go run main.go
```

## API Endpoints

### Authentication

- `POST /api/auth/login` - Login user
- `POST /api/auth/register` - Register user baru (belum diimplementasi)
- `GET /api/auth/me` - Get user info (protected)

### Protected Routes

Semua route di `/api/*` memerlukan JWT token di header:

```
Authorization: Bearer <token>
```

## Testing dengan Postman

Lihat file `POSTMAN_GUIDE.md` untuk panduan lengkap testing API dengan Postman.

### Quick Start:

1. **Health Check**: `GET http://localhost:3000/`
2. **Login**: `POST http://localhost:3000/api/auth/login`
   ```json
   {
     "username": "your_username",
     "password": "your_password"
   }
   ```
3. **Get User Info**: `GET http://localhost:3000/api/auth/me`
   - Header: `Authorization: Bearer YOUR_TOKEN`

## Database

Database menggunakan PostgreSQL dengan struktur:

- `roles` - Role pengguna
- `permissions` - Permission sistem
- `users` - Data pengguna
- `students` - Data mahasiswa
- `lecturers` - Data dosen
- `achievement_references` - Referensi prestasi
