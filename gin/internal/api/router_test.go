package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/service"
)

func TestPing(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	router := NewRouter(service.NewAuthService(cfg, db), db, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/ping", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	router := NewRouter(service.NewAuthService(cfg, db), db, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/login", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request from empty body, got %d", w.Code)
	}
}

func TestResourceEndpointsAreRegistered(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	router := NewRouter(service.NewAuthService(cfg, db), db, cfg)

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{name: "umat list", path: "/api/v1/umats", method: http.MethodGet},
		{name: "topic list", path: "/api/v1/topics", method: http.MethodGet},
		{name: "activity list", path: "/api/v1/activities", method: http.MethodGet},
		{name: "tim kerja list", path: "/api/v1/tim-kerja", method: http.MethodGet},
		{name: "tahun ciu tao list", path: "/api/v1/tahun-ciu-tao", method: http.MethodGet},
		{name: "penggalang dana list", path: "/api/v1/penggalang-dana", method: http.MethodGet},
		{name: "sxy donatur list", path: "/api/v1/sxy-donatur", method: http.MethodGet},
		{name: "kelas list", path: "/api/v1/kelas", method: http.MethodGet},
		{name: "fotang lookup", path: "/api/v1/fotang/lookup", method: http.MethodGet},
		{name: "kelas lookup", path: "/api/v1/kelas/lookup", method: http.MethodGet},
		{name: "donasi sxy list", path: "/api/v1/donasi-sxy", method: http.MethodGet},
		{name: "change password v1", path: "/api/v1/change-password", method: http.MethodPost},
		{name: "change password root", path: "/api/change-password", method: http.MethodPost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer testtoken")
			router.ServeHTTP(w, req)

			// Should return 401 (unauthorized due to fake token) or 200, but NOT 404 (not found)
			if w.Code == http.StatusNotFound {
				t.Fatalf("expected route to be registered for %s, got 404", tt.path)
			}
		})
	}
}
