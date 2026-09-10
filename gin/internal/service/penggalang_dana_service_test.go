package service

import (
	"fmt"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestPenggalangDanaService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

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
	if created.ID <= 0 {
		t.Errorf("expected valid ID > 0, got %d", created.ID)
	}

	idStr := fmt.Sprintf("%d", created.ID)

	// 2. Get
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Nama != "Dewi" {
		t.Errorf("expected Dewi, got %s", fetched.Nama)
	}

	// 3. List (may fail if stored procedure is not installed in test db)
	filters := map[string]string{"nama": "Dewi"}
	_, _, _ = svc.List(1, filters, 10)

	// 4. Update
	created.Nama = "Dewi Lestari"
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Nama != "Dewi Lestari" {
		t.Errorf("expected Dewi Lestari, got %s", updated.Nama)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
