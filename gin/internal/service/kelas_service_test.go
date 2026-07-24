package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewKelasService(db)
	c := setupTestContext()

	// 1. Create
	k := domain.Kelas{
		LookupId:          "KL001",
		LookupValue:       "Kelas Tingkat Dasar",
		LookupDescription: "Basic Class",
	}

	created, err := svc.Create(k, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.LookupId != "KL001" {
		t.Errorf("expected KL001, got %s", created.LookupId)
	}

	// 2. Get
	fetched, err := svc.Get("KL001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.LookupValue != "Kelas Tingkat Dasar" {
		t.Errorf("expected Kelas Tingkat Dasar, got %s", fetched.LookupValue)
	}

	// 3. List
	filters := map[string]string{"lookup_value": "Tingkat"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.LookupValue = "Kelas Dasar Utama"
	updated, err := svc.Update("KL001", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.LookupValue != "Kelas Dasar Utama" {
		t.Errorf("expected Kelas Dasar Utama, got %s", updated.LookupValue)
	}

	// 5. Delete
	err = svc.Delete("KL001", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
