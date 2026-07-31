package service

import (
	"strconv"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasPesertaService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewKelasPesertaService(db)
	c := setupTestContext()

	// Seed target FK records (Kelas & Umat) for lookup validation
	kodeK := "K01"
	db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", 26160003)
	db.Exec("DELETE FROM T_BUS_UMAT WHERE id = ?", 101)
	db.Create(&domain.Kelas{TrxId: 26160003, KodeKelas: &kodeK})
	db.Create(&domain.Umat{ID: 101, Kode: "UM101", NamaIndonesia: "Peserta Test", JenisKelamin: "001"})
	defer func() {
		db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", 26160003)
		db.Exec("DELETE FROM T_BUS_UMAT WHERE id = ?", 101)
	}()

	// 1. Validation error tests - missing TrxId or IdPeserta
	_, err = svc.Create(domain.KelasPeserta{}, c)
	if err == nil {
		t.Fatalf("expected error for missing required fields, got nil")
	}

	// Validation error for non-existent TrxId or IdPeserta
	invalidTrx := int32(999999)
	invalidUmat := int32(888888)
	_, err = svc.Create(domain.KelasPeserta{TrxId: invalidTrx, IdPeserta: &invalidUmat}, c)
	if err == nil {
		t.Fatalf("expected lookup validation error for invalid FKs, got nil")
	}

	// 2. Successful Create
	trxId := int32(26160003)
	idPeserta := int32(101)
	sumbangan := 500000.0
	barang := "Buku"
	ket := "Peserta baru"

	p := domain.KelasPeserta{
		TrxId:      trxId,
		IdPeserta:  &idPeserta,
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
	if created.TrxId != trxId {
		t.Errorf("expected TrxId %d, got %d", trxId, created.TrxId)
	}

	// 3. Get
	idStr := strconv.Itoa(int(created.DetailId))
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Keterangan == nil || *fetched.Keterangan != "Peserta baru" {
		t.Errorf("expected Keterangan 'Peserta baru', got %v", fetched.Keterangan)
	}

	// 4. Update
	newKet := "Peserta update"
	created.Keterangan = &newKet
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Keterangan == nil || *updated.Keterangan != "Peserta update" {
		t.Errorf("expected Keterangan 'Peserta update', got %v", updated.Keterangan)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify status after soft delete
	deletedItem, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get after soft delete failed: %v", err)
	}
	if deletedItem.Status != nil && *deletedItem.Status != false {
		t.Errorf("expected status false after delete, got %v", *deletedItem.Status)
	}
}
