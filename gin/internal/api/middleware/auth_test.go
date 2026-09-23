package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"guangjiapps/gin/internal/config"

	"github.com/gin-gonic/gin"
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
