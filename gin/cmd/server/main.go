package main

import (
	"log"
	"os"
	"strings" // WAJIB: Tambahkan import strings
	"time"

	"guangjiapps/gin/internal/api"
	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/service"
)

func main() {
	// 1. Tangkap port dari argument flag CLI (misal: --port=111)
	//    Cek apakah port dikirim lewat argumen, jika tidak baru cek os.Getenv, jika tidak baru pakai default
	port := os.Getenv("PORT")

	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("database open failed: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	log.Fatalf("database migration failed: %v", err)
	// }

	jakartaLoc, err := time.LoadLocation("Asia/Jakarta")
	if err == nil {
		time.Local = jakartaLoc
	}

	// Atur waktu lokal default sistem (opsional, mempengaruhi time.Now() tanpa .UTC())

	authService := service.NewAuthService(cfg, db)
	router := api.NewRouter(authService, db, cfg)

	if port == "" {
		port = cfg.Port // Port default jika dijalankan lokal tanpa IIS
	}
	if port == "" {
		port = "8080"
	}

	port = strings.TrimSpace(port)
	log.Printf("starting gin server on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
