package service

import (
	"strconv"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasKendaraanService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewKelasKendaraanService(db)
	c := setupTestContext()

	// 1. Validation error test - missing TrxId
	_, err = svc.Create(domain.KelasKendaraan{}, c)
	if err == nil {
		t.Fatalf("expected error for missing required fields, got nil")
	}

	// 2. Successful Create
	trxId := int32(26160003)
	nopol := "B 1234 CD"
	pengendara := "Ahmad"
	tipe := "Mobil"

	p := domain.KelasKendaraan{
		TrxId:         trxId,
		NoPolisi:      &nopol,
		Pengendara:    &pengendara,
		TipeKendaraan: &tipe,
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
	if fetched.NoPolisi == nil || *fetched.NoPolisi != "B 1234 CD" {
		t.Errorf("expected NoPolisi 'B 1234 CD', got %v", fetched.NoPolisi)
	}

	// 4. Update
	newNopol := "B 5678 EF"
	created.NoPolisi = &newNopol
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.NoPolisi == nil || *updated.NoPolisi != "B 5678 EF" {
		t.Errorf("expected NoPolisi 'B 5678 EF', got %v", updated.NoPolisi)
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
