package service

import (
	"strconv"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasPengeluaranService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewKelasPengeluaranService(db)
	c := setupTestContext()

	// 1. Validation error test - missing required fields
	_, err = svc.Create(domain.KelasPengeluaran{}, c)
	if err == nil {
		t.Fatalf("expected error for missing required fields, got nil")
	}

	// 2. Successful Create
	trxId := int32(26160003)
	ket := "Biaya Konsumsi"
	biaya := 150000.0

	p := domain.KelasPengeluaran{
		TrxId:      trxId,
		Keterangan: &ket,
		Biaya:      &biaya,
	}

	created, err := svc.Create(p, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.DetailId == 0 {
		t.Errorf("expected non-zero DetailId")
	}

	// 3. Get
	idStr := strconv.Itoa(int(created.DetailId))
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Keterangan == nil || *fetched.Keterangan != "Biaya Konsumsi" {
		t.Errorf("expected Keterangan 'Biaya Konsumsi', got %v", fetched.Keterangan)
	}

	// 4. Update
	newKet := "Biaya Konsumsi Updated"
	created.Keterangan = &newKet
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Keterangan == nil || *updated.Keterangan != "Biaya Konsumsi Updated" {
		t.Errorf("expected Keterangan 'Biaya Konsumsi Updated', got %v", updated.Keterangan)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	deletedItem, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get after soft delete failed: %v", err)
	}
	if deletedItem.Status != nil && *deletedItem.Status != false {
		t.Errorf("expected status false after delete, got %v", *deletedItem.Status)
	}
}
