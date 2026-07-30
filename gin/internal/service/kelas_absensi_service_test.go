package service

import (
	"strconv"
	"testing"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasAbsensiService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewKelasAbsensiService(db)
	c := setupTestContext()

	// 1. Validation error test - missing required fields
	_, err = svc.Create(domain.KelasAbsensi{}, c)
	if err == nil {
		t.Fatalf("expected error for missing required fields, got nil")
	}

	// 2. Successful Create
	trxId := int32(26160003)
	idPeserta := int32(301)
	status := true

	p := domain.KelasAbsensi{
		TrxId:     trxId,
		TrxDate:   time.Now(),
		IdPeserta: idPeserta,
		Status:    &status,
	}

	created, err := svc.Create(p, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Id == 0 {
		t.Errorf("expected non-zero Id")
	}

	// 3. Get
	idStr := strconv.Itoa(int(created.Id))
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.IdPeserta != 301 {
		t.Errorf("expected IdPeserta 301, got %v", fetched.IdPeserta)
	}

	// 4. Update
	created.IdPeserta = 302
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.IdPeserta != 302 {
		t.Errorf("expected IdPeserta 302, got %v", updated.IdPeserta)
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
