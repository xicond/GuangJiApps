package service

import (
	"testing"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
)

func TestAuthService(t *testing.T) {
	cfg := config.Config{
		JWTSecret: "test-secret-key-12345",
	}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	authSvc := NewAuthService(cfg, db)

	// 1. Password encryption test (legacy MD5 Base64)
	encrypted := EncryptPassword("123456")
	if encrypted == "" {
		t.Errorf("expected non-empty encrypted password")
	}

	// 2. JWT Generation test
	token, err := authSvc.GenerateToken(1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Errorf("expected non-empty token")
	}

	// 3. Login with empty parameters
	_, _, _, err = authSvc.Login("", "")
	if err == nil {
		t.Errorf("expected error for empty credentials, got nil")
	}
}
