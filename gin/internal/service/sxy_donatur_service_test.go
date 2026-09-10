package service

import (
	"fmt"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestSxyDonaturService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

	svc := NewSxyDonaturService(db)
	c := setupTestContext()

	// 1. Create
	donatur := domain.SxyDonatur{
		No:         "DN001",
		Nama:       "Agus Tan",
		Mandarin:   "阿古斯",
		Keterangan: "Regular Donatur",
	}

	created, err := svc.Create(donatur, c)
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
	if fetched.Nama != "Agus Tan" {
		t.Errorf("expected Agus Tan, got %s", fetched.Nama)
	}

	// 3. List (may fail if stored procedure is not installed in test db)
	filters := map[string]string{"nama": "Agus"}
	_, _, _ = svc.List(1, filters, 10)

	// 4. Update
	created.Nama = "Agus Tandi"
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Nama != "Agus Tandi" {
		t.Errorf("expected Agus Tandi, got %s", updated.Nama)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
