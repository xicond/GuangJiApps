package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsScannerURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"/.env", true},
		{"/.env.local", true},
		{"/wp-admin", true},
		{"/wp-login.php", true},
		{"/wordpress/index.php", true},
		{"/.git/config", true},
		{"/phpmyadmin/index.php", true},
		{"/actuator/health", true},
		{"/shell.php", true},
		{"/test.aspx", true},
		{"/v1/umats", false},
		{"/v1/lookup/fotang", false},
		{"/login", false},
	}

	for _, tt := range tests {
		result := IsScannerURL(tt.url)
		if result != tt.expected {
			t.Errorf("IsScannerURL(%s) = %v; want %v", tt.url, result, tt.expected)
		}
	}
}

func TestFail2Ban_General404Rule(t *testing.T) {
	fb := NewFail2Ban()
	defer fb.Stop()
	targetIP := "203.0.113.100"

	// 9 404s should NOT block
	for i := 0; i < 9; i++ {
		blocked := fb.Record404(targetIP, "/non-existent-page")
		if blocked {
			t.Fatalf("IP should not be blocked after 9 404s (iteration %d)", i+1)
		}
	}

	// 10th 404 within 20s SHOULD block
	blocked := fb.Record404(targetIP, "/non-existent-page")
	if !blocked {
		t.Fatalf("IP should be blocked on 10th 404 within 20s")
	}

	if !fb.IsBlocked(targetIP) {
		t.Fatalf("fb.IsBlocked(%s) should return true", targetIP)
	}
}

func TestFail2Ban_Scanner404Rule(t *testing.T) {
	fb := NewFail2Ban()
	defer fb.Stop()
	targetIP := "203.0.113.101"

	// 1st scanner 404 should NOT block
	blocked := fb.Record404(targetIP, "/.env")
	if blocked {
		t.Fatalf("IP should not be blocked after 1st scanner 404")
	}

	// 2nd scanner 404 within 1 min SHOULD block
	blocked = fb.Record404(targetIP, "/wp-login.php")
	if !blocked {
		t.Fatalf("IP should be blocked on 2nd scanner 404 within 1 minute")
	}

	if !fb.IsBlocked(targetIP) {
		t.Fatalf("fb.IsBlocked(%s) should return true", targetIP)
	}
}

func TestFail2Ban_LoopbackNotBlocked(t *testing.T) {
	fb := NewFail2Ban()
	defer fb.Stop()
	loopbackIPs := []string{"127.0.0.1", "::1", "192.168.1.1", "10.0.0.1"}

	for _, ip := range loopbackIPs {
		for i := 0; i < 15; i++ {
			fb.Record404(ip, "/.env")
		}

		if fb.IsBlocked(ip) {
			t.Fatalf("Whitelisted/Private IP %s should NEVER be blocked", ip)
		}
	}
}

func TestFail2Ban_MiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fb := NewFail2Ban()
	defer fb.Stop()

	r := gin.New()
	r.Use(fb.Middleware())
	r.GET("/valid-route", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// 1. Valid request -> 200 OK
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/valid-route", nil)
	req.RemoteAddr = "203.0.113.50:12345"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// 2. Trigger 2 scanner 404s for 203.0.113.50
	fb.Record404("203.0.113.50", "/.env")
	fb.Record404("203.0.113.50", "/wp-admin")

	// 3. Subsequent request -> 403 Forbidden (in TestMode/debug) or 404
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/valid-route", nil)
	req2.RemoteAddr = "203.0.113.50:12345"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden && w2.Code != http.StatusNotFound {
		t.Fatalf("Expected blocked status (403 or 404), got %d", w2.Code)
	}
}
