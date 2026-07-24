package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestTimKerjaService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewTimKerjaService(db)
	c := setupTestContext()

	// 1. Create
	tim := domain.TimKerja{
		LookupId:          "TK001",
		LookupValue:       "Ketua Panitia",
		LookupDescription: "Head Coordinator",
	}

	created, err := svc.Create(tim, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.LookupId != "TK001" {
		t.Errorf("expected TK001, got %s", created.LookupId)
	}

	// 2. Get
	fetched, err := svc.Get("TK001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.LookupValue != "Ketua Panitia" {
		t.Errorf("expected Ketua Panitia, got %s", fetched.LookupValue)
	}

	// 3. List
	filters := map[string]string{"lookup_value": "Ketua"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.LookupValue = "Koordinator Utama"
	updated, err := svc.Update("TK001", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.LookupValue != "Koordinator Utama" {
		t.Errorf("expected Koordinator Utama, got %s", updated.LookupValue)
	}

	// 5. Delete
	err = svc.Delete("TK001", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
