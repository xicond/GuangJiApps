package config

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
	RedisAddr            string
	RedisPassword        string
	RedisDB              string
}

func Load() Config {
	_ = godotenv.Load()
	// _ = godotenv.Load("../.env")
	// _ = godotenv.Load("../../.env")
	_ = godotenv.Overload(".env.local")
	// _ = godotenv.Overload("../.env.local")
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
		RedisAddr:            getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getenv("REDIS_PASSWORD", ""),
		RedisDB:              getenv("REDIS_DB", "0"),
	}

	if cfg.GinMode == "release" {
		// log.Println("running in release mode")
	} else {
		log.Println("running in debug mode")
	}

	return cfg
}

func (c Config) GetJWTSigningKey() (interface{}, jwt.SigningMethod, error) {
	return LoadJWTSigningKey(c.JWTSecret)
}

func (c Config) GetJWTVerificationKey() (interface{}, jwt.SigningMethod, error) {
	return LoadJWTVerificationKey(c.JWTSecret)
}

func (c Config) GetJWTRefreshSigningKey() (interface{}, jwt.SigningMethod, error) {
	return LoadJWTSigningKey(c.RefreshSecret)
}

func (c Config) GetJWTRefreshVerificationKey() (interface{}, jwt.SigningMethod, error) {
	return LoadJWTVerificationKey(c.RefreshSecret)
}

func LoadJWTSigningKey(path string) (interface{}, jwt.SigningMethod, error) {
	bytesData, err := getSecretBytes(path)
	if err != nil {
		return nil, nil, err
	}

	// 1. Try RSA Private Key
	if rsaKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytesData); err == nil {
		return rsaKey, jwt.SigningMethodRS256, nil
	}

	// 2. Try EC Private Key
	if ecKey, err := jwt.ParseECPrivateKeyFromPEM(bytesData); err == nil {
		return ecKey, jwt.SigningMethodES256, nil
	}

	return nil, nil, errors.New("JWT key must be a valid RSA or ECDSA private key PEM file")
}

func LoadJWTVerificationKey(path string) (interface{}, jwt.SigningMethod, error) {
	bytesData, err := getSecretBytes(path)
	if err != nil {
		return nil, nil, err
	}

	// 1. Try RSA Private Key (extract public key)
	if rsaPrivKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytesData); err == nil {
		return &rsaPrivKey.PublicKey, jwt.SigningMethodRS256, nil
	}

	// 2. Try RSA Public Key
	if rsaPubKey, err := jwt.ParseRSAPublicKeyFromPEM(bytesData); err == nil {
		return rsaPubKey, jwt.SigningMethodRS256, nil
	}

	// 3. Try EC Private Key (extract public key)
	if ecPrivKey, err := jwt.ParseECPrivateKeyFromPEM(bytesData); err == nil {
		return &ecPrivKey.PublicKey, jwt.SigningMethodES256, nil
	}

	// 4. Try EC Public Key
	if ecPubKey, err := jwt.ParseECPublicKeyFromPEM(bytesData); err == nil {
		return ecPubKey, jwt.SigningMethodES256, nil
	}

	return nil, nil, errors.New("JWT key must be a valid RSA or ECDSA public/private key PEM file")
}

func getSecretBytes(path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("jwt secret not configured")
	}

	// 1. Check if path is raw PEM content string
	if strings.HasPrefix(path, "-----BEGIN") {
		return []byte(path), nil
	}

	// Clean path and normalize slashes for cross-platform (Windows & Linux)
	cleanPath := filepath.Clean(path)
	slashPath := strings.ReplaceAll(path, "\\", "/")
	slashPath = strings.TrimPrefix(slashPath, "./")

	execDir := ""
	if execPath, err := os.Executable(); err == nil {
		execDir = filepath.Dir(execPath)
	}

	// 2. Candidate paths to attempt reading
	candidates := []string{
		path,
		cleanPath,
		slashPath,
	}

	if strings.HasPrefix(cleanPath, "/app/") {
		relPath := strings.TrimPrefix(cleanPath, "/app/")
		candidates = append(candidates, relPath, filepath.Join(".", relPath), filepath.Join("gin", relPath), filepath.Join("..", relPath), filepath.Join("../..", relPath))
	}

	if !filepath.IsAbs(cleanPath) {
		candidates = append(candidates, filepath.Join(".", cleanPath))
		if execDir != "" {
			candidates = append(candidates, filepath.Join(execDir, cleanPath))
			candidates = append(candidates, filepath.Join(execDir, "certs", filepath.Base(cleanPath)))
		}
		candidates = append(candidates, filepath.Join("gin", cleanPath))
		candidates = append(candidates, filepath.Join("..", cleanPath))
		candidates = append(candidates, filepath.Join("../..", cleanPath))
	}

	searched := make([]string, 0, len(candidates))
	for _, cand := range candidates {
		if cand == "" {
			continue
		}
		searched = append(searched, cand)
		if data, err := os.ReadFile(cand); err == nil {
			return bytes.TrimSpace(data), nil
		}
	}

	return nil, fmt.Errorf("JWT RSA/ECDSA key PEM file not found at '%s'. Tried paths: %s", path, strings.Join(searched, ", "))
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
