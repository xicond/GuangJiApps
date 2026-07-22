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
	cfg := config.Config{GinMode: "test"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	router := NewRouter(service.NewAuthService(cfg), db, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	cfg := config.Config{GinMode: "test"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	router := NewRouter(service.NewAuthService(cfg), db, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request from empty body, got %d", w.Code)
	}
}

func TestResourceEndpointsAreRegistered(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "http://localhost"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	router := NewRouter(service.NewAuthService(cfg), db, cfg)

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{name: "umat list", path: "/api/umats", method: http.MethodGet},
		{name: "topic list", path: "/api/topics", method: http.MethodGet},
		{name: "activity list", path: "/api/activities", method: http.MethodGet},
		{name: "tim kerja list", path: "/api/tim-kerja", method: http.MethodGet},
		{name: "tahun ciu tao list", path: "/api/tahun-ciu-tao", method: http.MethodGet},
		{name: "penggalang dana list", path: "/api/penggalang-dana", method: http.MethodGet},
		{name: "sxy donatur list", path: "/api/sxy-donatur", method: http.MethodGet},
		{name: "kelas list", path: "/api/kelas", method: http.MethodGet},
		{name: "donasi sxy list", path: "/api/donasi-sxy", method: http.MethodGet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", tt.path, w.Code)
			}
		})
	}
}
