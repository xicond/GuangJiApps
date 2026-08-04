package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	GinMode              string
	BaseURL              string
	JWTSecret            string
	RefreshSecret        string
	DatabaseDSN          string
	ReportServerUsername string
	ReportServerPassword string
}

func Load() Config {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")
	_ = godotenv.Overload(".env.local")
	_ = godotenv.Overload("../.env.local")
	_ = godotenv.Overload("../../.env.local")

	cfg := Config{
		Port:                 getenv("PORT", "8080"),
		GinMode:              getenv("GIN_MODE", "debug"),
		BaseURL:              NormalizeBaseURL(getenv("BASE_URL", "http://localhost:8080")),
		JWTSecret:            getenv("JWT_SECRET", "change-me"),
		RefreshSecret:        getenv("REFRESH_SECRET", "change-me"),
		DatabaseDSN:          getenv("DATABASE_DSN", ""),
		ReportServerUsername: getenv("REPORT_USERNAME", ""),
		ReportServerPassword: getenv("REPORT_PASSWORD", ""),
	}

	if cfg.GinMode == "release" {
		// log.Println("running in release mode")
	} else {
		log.Println("running in debug mode")
	}

	return cfg
}

func NormalizeBaseURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/"
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return strings.TrimRight(value, "/")
	}

	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	return strings.TrimRight(value, "/")
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
