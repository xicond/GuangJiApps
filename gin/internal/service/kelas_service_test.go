package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestKelasService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	db.Exec("DELETE FROM T_APP_LOOKUP WHERE LookupId = 'LK001'")
	db.Exec("DELETE FROM T_APP_LOOKUPCATEGORY WHERE CategoryId = 'B_KELASKHUSUS'")

	// Seed lookup categories & lookups for validation
	catKelas := domain.AppLookupCategory{
		CategoryId:          "B_KELASKHUSUS",
		CategoryType:        "S",
		CategoryDescription: "Kelas Khusus",
		Status:              true,
	}
	db.Create(&catKelas)

	valKelas := "K01"
	descKelas := "Kelas Tingkat Dasar"
	statusTrue := true
	lookupKelas := domain.AppLookup{
		LookupId:          "LK001",
		CategoryId:        &catKelas.CategoryId,
		LookupValue:       &valKelas,
		LookupDescription: &descKelas,
		Status:            &statusTrue,
	}
	db.Create(&lookupKelas)

	catFotang := domain.AppLookupCategory{
		CategoryId:          "B_FOTHANG",
		CategoryType:        "S",
		CategoryDescription: "Fotang",
		Status:              true,
	}
	db.Create(&catFotang)

	valFotang := "F01"
	descFotang := "Fotang Utama"
	lookupFotang := domain.AppLookup{
		LookupId:          "LF001",
		CategoryId:        &catFotang.CategoryId,
		LookupValue:       &valFotang,
		LookupDescription: &descFotang,
	}
	db.Create(&lookupFotang)

	svc := NewKelasService(db)
	c := setupTestContext()

	// 1. Validation error tests
	// Missing KodeKelas
	_, err = svc.Create(domain.Kelas{}, c)
	if err == nil {
		t.Fatalf("expected error for missing KodeKelas, got nil")
	}

	// Invalid KodeKelas returns ValidationError
	invalidK := "INVALID_K"
	_, err = svc.Create(domain.Kelas{KodeKelas: &invalidK}, c)
	if err == nil {
		t.Fatalf("expected error for invalid KodeKelas, got nil")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected *ValidationError for invalid KodeKelas, got %T (%v)", err, err)
	} else if len(valErr.Details["kode_kelas"]) == 0 {
		t.Errorf("expected details for 'kode_kelas', got: %v", valErr.Details)
	}

	// Valid KodeKelas but invalid KodeFotang
	invalidF := "INVALID_F"
	_, err = svc.Create(domain.Kelas{KodeKelas: &valKelas, KodeFotang: &invalidF}, c)
	if err == nil {
		t.Fatalf("expected error for invalid KodeFotang, got nil")
	}
	if !errors.As(err, &valErr) {
		t.Errorf("expected *ValidationError for invalid KodeFotang, got %T (%v)", err, err)
	} else if len(valErr.Details["kode_fotang"]) == 0 {
		t.Errorf("expected details for 'kode_fotang', got: %v", valErr.Details)
	}

	// Date range validation: EndDate < StartDate
	tMulai := domain.DateOnly{Time: time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)}
	tSelesai := domain.DateOnly{Time: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)}
	_, err = svc.Create(domain.Kelas{KodeKelas: &valKelas, StartDate: &tMulai, EndDate: &tSelesai}, c)
	if err == nil {
		t.Fatalf("expected error for StartDate > EndDate, got nil")
	}

	// 2. Successful Create
	lokasi := "Jakarta"
	pic := "Budi"
	k := domain.Kelas{
		KodeKelas:  &valKelas,
		KodeFotang: &valFotang,
		Lokasi:     &lokasi,
		PIC:        &pic,
	}

	created, err := svc.Create(k, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.TrxId == 0 {
		t.Errorf("expected non-zero TrxId")
	}
	if created.KodeKelas == nil || *created.KodeKelas != "K01" {
		t.Errorf("expected K01, got %v", created.KodeKelas)
	}

	// 3. Get
	idStr := strconv.Itoa(int(created.TrxId))
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Lokasi == nil || *fetched.Lokasi != "Jakarta" {
		t.Errorf("expected Jakarta, got %v", fetched.Lokasi)
	}
	if fetched.KelasName == nil || fetched.KelasName.LookupDescription == nil || *fetched.KelasName.LookupDescription != "Kelas Tingkat Dasar" {
		t.Errorf("expected preloaded KelasName, got %v", fetched.KelasName)
	}

	// 4. Update
	newLokasi := "Surabaya"
	created.Lokasi = &newLokasi
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Lokasi == nil || *updated.Lokasi != "Surabaya" {
		t.Errorf("expected Surabaya, got %v", updated.Lokasi)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify status after soft delete
	deletedItem, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get after soft delete failed: %v", err)
	}
	if deletedItem.Status != nil && *deletedItem.Status != false {
		t.Errorf("expected status false after delete, got %v", *deletedItem.Status)
	}

	// 6. Lookup (paginated & limit=0)
	lookups, totalLookup, err := svc.Lookup(map[string]string{}, 1, 10)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if totalLookup != 1 || len(lookups) != 1 {
		t.Errorf("expected 1 lookup item, got %d", totalLookup)
	}

	lookupsAll, totalAll, err := svc.Lookup(map[string]string{}, 0, 0)
	if err != nil {
		t.Fatalf("Lookup with limit=0 failed: %v", err)
	}
	if totalAll != 1 || len(lookupsAll) != 1 {
		t.Errorf("expected 1 lookup item with limit=0, got %d", totalAll)
	}
}

func TestKelasServiceReport(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	db.Exec("DELETE FROM T_WH_USER_MATRIX_MST WHERE CRUID = 999999 OR LOGINID = 999999")
	if err := db.Exec("INSERT INTO T_WH_USER_MATRIX_MST (CRUID, LOGINID, SUBWHID) VALUES (?, ?, ?)", 999999, 999999, 160).Error; err != nil {
		t.Fatalf("failed to seed matrix: %v", err)
	}
	defer db.Exec("DELETE FROM T_WH_USER_MATRIX_MST WHERE CRUID = 999999")

	mockReportContent := []byte("PK\x03\x04mock_excel_openxml_data")

	digestTested := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Digest ") {
			w.Header().Set("WWW-Authenticate", `Digest realm="TestRealm", nonce="test_nonce_123", qop="auth"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if strings.HasPrefix(authHeader, "Digest ") {
			digestTested = true
			if !strings.Contains(authHeader, `username="admin"`) {
				t.Errorf("expected Digest auth username=admin, got %s", authHeader)
			}
		}

		if r.URL.Query().Get("TrxId") != "26040001" {
			t.Errorf("expected TrxId=26040001, got %s", r.URL.Query().Get("TrxId"))
		}
		// if subWh := r.URL.Query().Get("SubWhId"); subWh != "" && subWh != "160" {
		// 	t.Errorf("expected SubWhId=160, got %s", subWh)
		// }

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.WriteHeader(http.StatusOK)
		w.Write(mockReportContent)
	}))
	defer ts.Close()

	svc := NewKelasService(db)
	svc.reportServerURL = ts.URL
	svc.reportServerUsername = "admin"
	svc.reportServerPassword = "password123"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/kelas/26040001/report", nil)
	c.Set("userID", int32(999999))

	err = svc.Report("26040001", c)
	if err != nil {
		t.Fatalf("Report streaming failed: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}

	if w.Body.String() != string(mockReportContent) {
		t.Errorf("expected body %q, got %q", string(mockReportContent), w.Body.String())
	}

	if !digestTested {
		t.Errorf("expected Digest Auth challenge to be processed")
	}
}
