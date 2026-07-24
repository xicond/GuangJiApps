package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestPenggalangDanaService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewPenggalangDanaService(db)
	c := setupTestContext()

	// 1. Create
	p := domain.PenggalangDana{
		No:         "PG001",
		Nama:       "Dewi",
		Mandarin:   "德偉",
		Keterangan: "Penggalang SXY",
	}

	created, err := svc.Create(p, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}

	// 2. Get
	fetched, err := svc.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Nama != "Dewi" {
		t.Errorf("expected Dewi, got %s", fetched.Nama)
	}

	// 3. List
	filters := map[string]string{"nama": "Dewi"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.Nama = "Dewi Lestari"
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Nama != "Dewi Lestari" {
		t.Errorf("expected Dewi Lestari, got %s", updated.Nama)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
