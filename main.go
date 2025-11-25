package main

import (
	"log"

	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/config"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/database"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Setup logger
	config.SetupLogger()

	// Connect to database
	database.Connect()

	// Setup Fiber app
	app := config.SetupApp()

	// Start server
	port := config.Port
	log.Printf("Server berjalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
