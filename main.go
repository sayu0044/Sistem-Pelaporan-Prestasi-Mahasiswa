package main

import (
	"log"
	"time"

	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/config"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/database"
)

func main() {
	config.LoadEnv()
	config.SetupLogger()
	database.Connect()
	database.ConnectMongoDB(config.MongoURI, config.MongoDBName)

	jwtExpiry, _ := time.ParseDuration(config.JWTExpiry)
	app := config.SetupApp(database.DB, database.MongoDB, config.JWTSecret, jwtExpiry)

	port := config.Port
	log.Printf("Server berjalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
