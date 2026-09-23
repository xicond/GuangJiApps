package api

import (
	"encoding/json"
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
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate test db failed: %v", err)
	// }
	router := NewRouter(service.NewAuthService(cfg, db), db, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/ping", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}
}

func TestLogin(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate test db failed: %v", err)
	// }
	router := NewRouter(service.NewAuthService(cfg, db), db, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/login", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request from empty body, got %d", w.Code)
	}
}

func TestLoginRateLimiting(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate test db failed: %v", err)
	// }
	router := NewRouter(service.NewAuthService(cfg, db), db, cfg)

	invalidBody := `{"username":"wronguser","password":"wrongpassword"}`

	// 3 Failed Attempts (each returns 401 Unauthorized)
	for i := 1; i <= 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/login", strings.NewReader(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.1.100:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected status 401, got %d", i, w.Code)
		}
	}

	// 4th Attempt should be immediately blocked with 429 Too Many Requests
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/login", strings.NewReader(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("4th attempt: expected status 429, got %d", w.Code)
	}
	if w.Header().Get("X-Retry-After") == "" || w.Header().Get("Retry-After") == "" {
		t.Fatalf("4th attempt: expected X-Retry-After header, got empty")
	}
}

func TestResourceEndpointsAreRegistered(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate test db failed: %v", err)
	// }
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
		{name: "fotang lookup sxy", path: "/api/v1/fotang/lookup-sxy", method: http.MethodGet},
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
		{name: "lookup kategori topic", path: "/api/v1/lookup/kategori-topic", method: http.MethodGet},
		{name: "lookup kategori event", path: "/api/v1/lookup/kategori-event", method: http.MethodGet},
		{name: "lookup tipe sumbangan", path: "/api/v1/lookup/tipe-sumbangan", method: http.MethodGet},
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
		"topic_name": {"Nama topik 'Dharma' sudah ada"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondError(c, vErr)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Nama topik 'Dharma' sudah ada") {
		t.Fatalf("expected body to contain error clue, got: %s", body)
	}
}

func TestCombinedValidationErrors(t *testing.T) {
	vErr := service.NewValidationError(map[string][]string{
		"topic_category": {"Kategori Topik melebihi nilai maksimum 3 (max:3)"},
		"topic_name":     {"Nama topik 'Dharma' sudah ada"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondError(c, vErr)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "topic_category") || !strings.Contains(body, "topic_name") {
		t.Fatalf("expected body to contain both error fields, got: %s", body)
	}
	if !strings.Contains(body, "Nama topik 'Dharma' sudah ada") {
		t.Fatalf("expected body to contain duplicate name clue, got: %s", body)
	}
}

func TestKelasPeserta_PayloadTypeValidation(t *testing.T) {
	cfg := config.Config{GinMode: "test", BaseURL: "/api", JWTSecret: "../../certs/dev-private-key.pem"}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	authSvc := service.NewAuthService(cfg, db)
	router := NewRouter(authSvc, db, cfg)

	token, err := authSvc.GenerateToken(1)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	// 1. Array sent to single create endpoint POST /api/v1/kelas/26340002/peserta
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/kelas/26340002/peserta", strings.NewReader(`{"id_peserta": [101, 102]}`))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for array id_peserta on single create, got %d (body: %s)", w1.Code, w1.Body.String())
	}
	var resp1 struct {
		Error   string              `json:"error"`
		Details map[string][]string `json:"details"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("failed to parse response 1 json: %v", err)
	}
	if len(resp1.Details["idpeserta"]) == 0 || resp1.Details["idpeserta"][0] != "idpeserta should be single value of umat" {
		t.Fatalf("expected details.idpeserta to contain 'idpeserta should be single value of umat', got: %+v", resp1.Details)
	}

	// 2. Single number sent to bulk create endpoint POST /api/v1/kelas/26340002/peserta/bulk
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/kelas/26340002/peserta/bulk", strings.NewReader(`{"id_peserta": 101}`))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for single id_peserta on bulk create, got %d (body: %s)", w2.Code, w2.Body.String())
	}
	var resp2 struct {
		Error   string              `json:"error"`
		Details map[string][]string `json:"details"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("failed to parse response 2 json: %v", err)
	}
	if len(resp2.Details["idpeserta"]) == 0 || resp2.Details["idpeserta"][0] != "idpeserta must array" {
		t.Fatalf("expected details.idpeserta to contain 'idpeserta must array', got: %+v", resp2.Details)
	}

	// 3. Syntax error JSON sent to endpoint
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/api/v1/kelas/26340002/peserta", strings.NewReader(`{"id_peserta": `))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for syntax error, got %d", w3.Code)
	}
	var resp3 struct {
		Error   string              `json:"error"`
		Details map[string][]string `json:"details"`
	}
	_ = json.Unmarshal(w3.Body.Bytes(), &resp3)
	if resp3.Error != "Format payload JSON tidak valid atau rusak." {
		t.Errorf("expected error 'Format payload JSON tidak valid atau rusak.', got '%s'", resp3.Error)
	}
	if len(resp3.Details["general"]) == 0 || resp3.Details["general"][0] != "Format payload JSON tidak valid atau rusak." {
		t.Errorf("expected details.general to have format error, got: %+v", resp3.Details)
	}

	// 4. Empty body (EOF) sent to endpoint
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "/api/v1/kelas/26340002/peserta", strings.NewReader(``))
	req4.Header.Set("Authorization", "Bearer "+token)
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty body, got %d", w4.Code)
	}
	var resp4 struct {
		Error   string              `json:"error"`
		Details map[string][]string `json:"details"`
	}
	_ = json.Unmarshal(w4.Body.Bytes(), &resp4)
	if resp4.Error != "Payload request tidak boleh kosong." {
		t.Errorf("expected error 'Payload request tidak boleh kosong.', got '%s'", resp4.Error)
	}
	if len(resp4.Details["general"]) == 0 || resp4.Details["general"][0] != "Payload request tidak boleh kosong." {
		t.Errorf("expected details.general to have empty body error, got: %+v", resp4.Details)
	}
}

func TestParseJSONError(t *testing.T) {
	// 1. Syntax Error
	var dummy struct {
		Name string `json:"name"`
	}
	errSyntax := json.Unmarshal([]byte(`{"name": `), &dummy)
	res1 := ParseJSONError(errSyntax)
	if res1["general"] != "Format payload JSON tidak valid atau rusak." {
		t.Errorf("expected syntax error message, got: %v", res1)
	}

	// 2. Type Error
	var numStruct struct {
		Age int `json:"age"`
	}
	errType := json.Unmarshal([]byte(`{"age": "twenty"}`), &numStruct)
	res2 := ParseJSONError(errType)
	if res2["age"] != "Field 'age' memiliki tipe data yang tidak sesuai." {
		t.Errorf("expected field type error message, got: %v", res2)
	}

	// 3. Nil error
	res3 := ParseJSONError(nil)
	if len(res3) != 0 {
		t.Errorf("expected empty map for nil error, got: %v", res3)
	}
}
