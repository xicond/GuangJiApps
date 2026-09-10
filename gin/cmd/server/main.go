package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"guangjiapps/gin/internal/api"
	"guangjiapps/gin/internal/api/middleware"
	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/service"
)

func main() {
	// 1. Tangkap port dari argument flag CLI (misal: --port=111)
	//    Cek apakah port dikirim lewat argumen, jika tidak baru cek os.Getenv, jika tidak baru pakai default
	port := os.Getenv("PORT")

	cfg := config.Load()
	if _, _, err := cfg.GetJWTSigningKey(); err != nil {
		log.Printf("[WARNING] JWT signing key load check: %v", err)
	} else {
		log.Printf("[INFO] JWT signing key loaded successfully")
	}

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

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       50 * time.Second,
		WriteTimeout:      100 * time.Second,
		// IdleTimeout:       300 * time.Second,
	}

	// Initializing the server in a goroutine so that it doesn't block shutdown handling
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a 10-second timeout
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-quit
	log.Printf("[INFO] Shutdown signal received (%s), starting graceful shutdown...", sig)

	// Stop background tickers / fail2ban loops
	middleware.GetFail2Ban().Stop()

	// Context for HTTP server shutdown (allow up to 10 seconds for active requests to finish)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	} else {
		log.Println("[INFO] Server HTTP listener stopped cleanly")
	}

	// Close database connection pool
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("[ERROR] Failed to close database connections: %v", err)
			} else {
				log.Println("[INFO] Database connection pool closed")
			}
		}
	}

	// Close redis connections
	database.CloseRedisClient()

	log.Println("[INFO] Server exited gracefully")
}
