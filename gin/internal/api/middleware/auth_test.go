package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{
		JWTSecret: "test-secret-key-12345",
	}

	r := gin.New()
	r.Use(AuthMiddleware(cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 1. Missing Authorization header
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing header, got %d", w1.Code)
	}

	// 2. Invalid Authorization header format
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "InvalidFormatToken")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for invalid format, got %d", w2.Code)
	}

	// 3. Invalid token payload
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer invalid.jwt.token")
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for invalid token, got %d", w3.Code)
	}
}

func TestRefreshTokenAuthMiddleware(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate EC key: %v", err)
	}
	ecDER, _ := x509.MarshalECPrivateKey(ecKey)
	ecPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: ecDER}))
	cfg := config.Config{GinMode: "test", BaseURL: "/api", RefreshSecret: ecPEM,
		JWTSecret: ecPEM}
	gin.SetMode(gin.TestMode)

	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Logf("auto migrate warning: %v", err)
	}

	// Clean up any residual test data before and after
	db.Exec("DELETE FROM T_Auth_Passkey WHERE LoginId IN (10, 20, 30)")
	defer db.Exec("DELETE FROM T_Auth_Passkey WHERE LoginId IN (10, 20, 30)")

	cfg = config.Config{
		RefreshSecret: ecPEM,
		JWTSecret:     ecPEM,
	}

	r := gin.New()
	r.Use(RefreshTokenAuthMiddleware(cfg, db))
	r.POST("/login-renew", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"status": "ok", "user_id": userID})
	})

	// 1. Missing refresh token
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/login-renew", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing refresh token, got %d", w1.Code)
	}

	// 2. Invalid JWT token
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/login-renew", nil)
	req2.Header.Set("Authorization", "Bearer random.invalid.token")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid JWT, got %d", w2.Code)
	}

	// 3. Valid signed token, but not in DB
	validToken := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub":  int32(10),
		"exp":  time.Now().Add(time.Hour).Unix(),
		"type": "refresh",
	})
	tokenStr, _ := validToken.SignedString(ecKey)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/login-renew", nil)
	req3.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for token not found in DB, got %d", w3.Code)
	}

	// 4. Token present in DB, but Status = false (revoked)
	statusRevoked := false
	tokenHash := fmt.Sprintf("%x", md5.Sum([]byte(tokenStr)))
	if err := db.Create(&domain.AuthPasskey{
		LoginID:          10,
		CredentialID:     "cred-revoked",
		Credential:       "dummy-pubkey",
		RefreshToken:     &tokenStr,
		RefreshTokenHash: tokenHash,
		Status:           &statusRevoked,
	}).Error; err != nil {
		t.Fatalf("failed to insert revoked passkey: %v", err)
	}

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "/login-renew", nil)
	req4.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for revoked token in DB, got %d", w4.Code)
	}

	// 5. Active token in DB (Status = 1) -> Success!
	activeToken := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub":  int32(20),
		"exp":  time.Now().Add(time.Hour).Unix(),
		"type": "refresh",
	})
	activeTokenStr, _ := activeToken.SignedString(ecKey)
	statusActive := true
	expAt := time.Now().Add(time.Hour)
	activeHash := fmt.Sprintf("%x", md5.Sum([]byte(activeTokenStr)))
	if err := db.Create(&domain.AuthPasskey{
		LoginID:          20,
		CredentialID:     "cred-active",
		Credential:       "dummy-pubkey",
		RefreshToken:     &activeTokenStr,
		RefreshTokenHash: activeHash,
		ExpiredAt:        &expAt,
		Status:           &statusActive,
	}).Error; err != nil {
		t.Fatalf("failed to insert active passkey: %v", err)
	}

	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodPost, "/login-renew", nil)
	req5.Header.Set("Authorization", "Bearer "+activeTokenStr)
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Errorf("expected 200 OK for active token in DB, got %d: %s", w5.Code, w5.Body.String())
	}

	// 6. Token with NULL RefreshToken in DB cannot do LoginRenew
	nullToken := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub":  int32(30),
		"exp":  time.Now().Add(time.Hour).Unix(),
		"type": "refresh",
	})
	nullTokenStr, _ := nullToken.SignedString(ecKey)
	nullHash := fmt.Sprintf("%x", md5.Sum([]byte(nullTokenStr)))
	if err := db.Create(&domain.AuthPasskey{
		LoginID:          30,
		CredentialID:     "cred-null-rt",
		Credential:       "dummy-pubkey",
		RefreshToken:     nil, // null RefreshToken
		RefreshTokenHash: nullHash,
		ExpiredAt:        &expAt,
		Status:           &statusActive,
	}).Error; err != nil {
		t.Fatalf("failed to insert null rt passkey: %v", err)
	}

	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest(http.MethodPost, "/login-renew", nil)
	req6.Header.Set("Authorization", "Bearer "+nullTokenStr)
	r.ServeHTTP(w6, req6)
	if w6.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for null RefreshToken in DB, got %d", w6.Code)
	}
}
