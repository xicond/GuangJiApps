package service

import (
	"testing"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestDonasiSxyService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewDonasiSxyService(db)
	c := setupTestContext()

	// 1. Create
	donasi := domain.DonasiSxy{
		NoKwitansi: "KW001",
		Tanggal:    time.Now(),
		Donatur:    1,
		Penggalang: 1,
		Jumlah:     500000,
		NoKupon:    "KP001",
	}

	created, err := svc.Create(donasi, c)
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
	if fetched.NoKwitansi != "KW001" {
		t.Errorf("expected KW001, got %s", fetched.NoKwitansi)
	}

	// 3. List
	filters := map[string]string{"no_kwitansi": "KW001"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.Jumlah = 750000
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Jumlah != 750000 {
		t.Errorf("expected 750000, got %f", updated.Jumlah)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
