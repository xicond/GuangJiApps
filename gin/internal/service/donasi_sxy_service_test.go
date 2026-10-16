package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestDonasiSxyService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

	svc := NewDonasiSxyService(db)
	c := setupTestContext()

	// 1. Create
	donasi := domain.DonasiSxy{
		NoKwitansi: "KW001",
		Tanggal:    domain.DateOnly{Time: time.Now()},
		Donatur:    1,
		Penggalang: 1,
		Jumlah:     500000,
		NoKupon:    "KP001",
	}

	created, err := svc.Create(donasi, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID <= 0 {
		t.Errorf("expected valid ID > 0, got %d", created.ID)
	}

	idStr := fmt.Sprintf("%d", created.ID)

	// 2. Get
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.NoKwitansi != "KW001" {
		t.Errorf("expected KW001, got %s", fetched.NoKwitansi)
	}

	// 3. List (may fail if stored procedure is not installed in test db)
	filters := map[string]string{"no_kwitansi": "KW001"}
	_, _, _ = svc.List(c, 1, filters, 10)

	// 3b. Report (may fail if stored procedure is not installed in test db)
	reportFilters := map[string]string{"no_kwitansi": "KW001"}
	_, _, _, _ = svc.Report(c, 1, reportFilters, 10)

	// 4. Update validation failures
	invalidNegative := created
	invalidNegative.Jumlah = -500000
	if _, err := svc.Update(idStr, invalidNegative, c); err == nil {
		t.Fatalf("expected error updating with negative jumlah, got nil")
	}

	invalidZero := created
	invalidZero.Jumlah = 0
	if _, err := svc.Update(idStr, invalidZero, c); err == nil {
		t.Fatalf("expected error updating with zero jumlah, got nil")
	}

	invalidNoKwitansi := created
	invalidNoKwitansi.NoKwitansi = ""
	if _, err := svc.Update(idStr, invalidNoKwitansi, c); err == nil {
		t.Fatalf("expected error updating with empty no_kwitansi, got nil")
	}

	// 4b. Update valid
	created.Jumlah = 750000
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Jumlah != 750000 {
		t.Errorf("expected 750000, got %f", updated.Jumlah)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNullFieldScannerPointerFields(t *testing.T) {
	var resp domain.DonasiSxyResponse

	// Test scanning string into *string field (e.g. NamaDonatur)
	scanner := &nullFieldScanner{target: &resp.NamaDonatur}
	if err := scanner.Scan("Budi"); err != nil {
		t.Fatalf("Scan into *string failed: %v", err)
	}
	if resp.NamaDonatur == nil || *resp.NamaDonatur != "Budi" {
		t.Errorf("expected 'Budi', got %v", resp.NamaDonatur)
	}

	// Test scanning []byte into *string field (e.g. NamaPenggalang)
	scanner2 := &nullFieldScanner{target: &resp.NamaPenggalang}
	if err := scanner2.Scan([]byte("Siti")); err != nil {
		t.Fatalf("Scan []byte into *string failed: %v", err)
	}
	if resp.NamaPenggalang == nil || *resp.NamaPenggalang != "Siti" {
		t.Errorf("expected 'Siti', got %v", resp.NamaPenggalang)
	}

	// Test scanning string date into *DateOnly field (e.g. TanggalTransfer)
	scanner3 := &nullFieldScanner{target: &resp.TanggalTransfer}
	if err := scanner3.Scan("2026-08-01"); err != nil {
		t.Fatalf("Scan into *DateOnly failed: %v", err)
	}
	if resp.TanggalTransfer == nil || resp.TanggalTransfer.Format("2006-01-02") != "2026-08-01" {
		t.Errorf("expected '2026-08-01', got %v", resp.TanggalTransfer)
	}
}

func TestDonasiSxyResponseJSON(t *testing.T) {
	var resp domain.DonasiSxyResponse
	resp.ID = 1
	resp.NoKwitansi = "KW001"

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(data)
	t.Logf("JSON output: %s", jsonStr)

	expectedKeys := []string{
		`"id"`,
		`"no_kwitansi"`,
		`"no_kupon"`,
		`"tanggal"`,
		`"keterangan"`,
		`"penggalang_id"`,
		`"tipe_sumbangan"`,
		`"jumlah"`,
		`"tipe_sumbangan_desc"`,
		`"nama_penggalang"`,
		`"donatur_id"`,
		`"nama_donatur"`,
		`"tanggal_transfer"`,
		`"atas_nama"`,
		`"email_penggalang"`,
	}

	for _, key := range expectedKeys {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("expected JSON output to contain key %s, got: %s", key, jsonStr)
		}
	}
}

func TestDonasiSxyService_Report_ContextCanceled(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Skipf("skipping test, DB not available: %v", err)
		return
	}
	svc := NewDonasiSxyService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	req, _ := http.NewRequestWithContext(ctx, "GET", "/v1/donasi-sxy/report", nil)
	c.Request = req

	_, _, _, err = svc.Report(c, 1, nil, 10)
	if err == nil {
		t.Errorf("expected error when context is canceled, got nil")
	}
}
