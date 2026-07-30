package service

import (
	"strconv"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasPengabdiService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewKelasPengabdiService(db)
	c := setupTestContext()

	// 1. Validation error test - missing TrxId or IdPengabdi
	_, err = svc.Create(domain.KelasPengabdi{}, c)
	if err == nil {
		t.Fatalf("expected error for missing required fields, got nil")
	}

	// 2. Successful Create
	trxId := int32(26160003)
	idPengabdi := int32(201)
	sumbangan := 250000.0
	barang := "Buku Panduan"
	ket := "Pengabdi baru"

	p := domain.KelasPengabdi{
		TrxId:      trxId,
		IdPengabdi: &idPengabdi,
		Sumbangan:  &sumbangan,
		Barang:     &barang,
		Keterangan: &ket,
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
	if fetched.Keterangan == nil || *fetched.Keterangan != "Pengabdi baru" {
		t.Errorf("expected Keterangan 'Pengabdi baru', got %v", fetched.Keterangan)
	}

	// 4. Update
	newKet := "Pengabdi updated"
	created.Keterangan = &newKet
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Keterangan == nil || *updated.Keterangan != "Pengabdi updated" {
		t.Errorf("expected Keterangan 'Pengabdi updated', got %v", updated.Keterangan)
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
