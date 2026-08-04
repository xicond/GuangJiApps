package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/service"

	"github.com/gin-gonic/gin"
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
		{name: "kelas peserta list", path: "/api/v1/kelas/1/peserta", method: http.MethodGet},
		{name: "kelas pengabdi list", path: "/api/v1/kelas/1/pengabdi", method: http.MethodGet},
		{name: "kelas topik list", path: "/api/v1/kelas/1/topik", method: http.MethodGet},
		{name: "kelas kendaraan list", path: "/api/v1/kelas/1/kendaraan", method: http.MethodGet},
		{name: "kelas donasi list", path: "/api/v1/kelas/1/donasi", method: http.MethodGet},
		{name: "kelas donasi barang list", path: "/api/v1/kelas/1/donasi-barang", method: http.MethodGet},
		{name: "kelas pengeluaran list", path: "/api/v1/kelas/1/pengeluaran", method: http.MethodGet},
		{name: "kelas musik list", path: "/api/v1/kelas/1/musik", method: http.MethodGet},
		{name: "kelas absensi list", path: "/api/v1/kelas/1/absensi", method: http.MethodGet},
		{name: "change password v1", path: "/api/v1/change-password", method: http.MethodPost},
		{name: "report sxy", path: "/api/v1/donasi-sxy/report", method: http.MethodGet},
		{name: "report sxy excel", path: "/api/v1/donasi-sxy/report/excel", method: http.MethodGet},
		{name: "lookup waktu ciu tao", path: "/api/v1/lookup/waktu-ciu-tao", method: http.MethodGet},
		{name: "lookup gender", path: "/api/v1/lookup/gender", method: http.MethodGet},
		{name: "lookup tcs", path: "/api/v1/lookup/tcs", method: http.MethodGet},
		{name: "lookup fotang", path: "/api/v1/lookup/fotang", method: http.MethodGet},
		{name: "lookup kelas", path: "/api/v1/lookup/kelas", method: http.MethodGet},
		{name: "lookup pendidikan", path: "/api/v1/lookup/pendidikan", method: http.MethodGet},
		{name: "lookup kelas umum", path: "/api/v1/lookup/kelas-umum", method: http.MethodGet},
		{name: "lookup pekerjaan", path: "/api/v1/lookup/pekerjaan", method: http.MethodGet},
		{name: "lookup keluarga", path: "/api/v1/lookup/keluarga", method: http.MethodGet},
		{name: "lookup status", path: "/api/v1/lookup/status", method: http.MethodGet},
		{name: "lookup category param", path: "/api/v1/lookup/category/B_STATUS", method: http.MethodGet},
		{name: "tim kerja lookup", path: "/api/v1/tim-kerja/lookup", method: http.MethodGet},
		{name: "tim kerja lookup sub", path: "/api/v1/tim-kerja/lookup/1/sub", method: http.MethodGet},
		{name: "tim kerja lookup report", path: "/api/v1/tim-kerja/lookup-report", method: http.MethodGet},
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

func TestValidationErrorReturns400(t *testing.T) {
	vErr := service.NewValidationError(map[string][]string{
		"fotang_aktif":   {"field B_FOTHANG nilai '31' tidak valid"},
		"fotang_chiutao": {"field B_FOTHANG nilai '31' tidak valid"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondError(c, vErr)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "tidak valid") {
		t.Fatalf("expected body to contain error detail, got: %s", body)
	}
}
