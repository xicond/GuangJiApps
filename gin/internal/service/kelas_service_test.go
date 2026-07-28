package service

import (
	"strconv"
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

	// Seed lookup categories & lookups for validation
	catKelas := domain.AppLookupCategory{
		CategoryId:          "B_KELASKHUSUS",
		CategoryType:        "S",
		CategoryDescription: "Kelas Khusus",
		Status:              true,
	}
	db.Create(&catKelas)

	valKelas := "K01"
	descKelas := "Kelas Tingkat Dasar"
	lookupKelas := domain.AppLookup{
		LookupId:          "LK001",
		CategoryId:        &catKelas.CategoryId,
		LookupValue:       &valKelas,
		LookupDescription: &descKelas,
	}
	db.Create(&lookupKelas)

	catFotang := domain.AppLookupCategory{
		CategoryId:          "B_FOTHANG",
		CategoryType:        "S",
		CategoryDescription: "Fotang",
		Status:              true,
	}
	db.Create(&catFotang)

	valFotang := "F01"
	descFotang := "Fotang Utama"
	lookupFotang := domain.AppLookup{
		LookupId:          "LF001",
		CategoryId:        &catFotang.CategoryId,
		LookupValue:       &valFotang,
		LookupDescription: &descFotang,
	}
	db.Create(&lookupFotang)

	svc := NewKelasService(db)
	c := setupTestContext()

	// 1. Validation error tests
	// Missing KodeKelas
	_, err = svc.Create(domain.Kelas{}, c)
	if err == nil {
		t.Fatalf("expected error for missing KodeKelas, got nil")
	}

	// Invalid KodeKelas
	invalidK := "INVALID_K"
	_, err = svc.Create(domain.Kelas{KodeKelas: &invalidK}, c)
	if err == nil {
		t.Fatalf("expected error for invalid KodeKelas, got nil")
	}

	// Valid KodeKelas but invalid KodeFotang
	invalidF := "INVALID_F"
	_, err = svc.Create(domain.Kelas{KodeKelas: &valKelas, KodeFotang: &invalidF}, c)
	if err == nil {
		t.Fatalf("expected error for invalid KodeFotang, got nil")
	}

	// 2. Successful Create
	lokasi := "Jakarta"
	pic := "Budi"
	k := domain.Kelas{
		KodeKelas:  &valKelas,
		KodeFotang: &valFotang,
		Lokasi:     &lokasi,
		PIC:        &pic,
	}

	created, err := svc.Create(k, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.TrxId == 0 {
		t.Errorf("expected non-zero TrxId")
	}
	if created.KodeKelas == nil || *created.KodeKelas != "K01" {
		t.Errorf("expected K01, got %v", created.KodeKelas)
	}

	// 3. Get
	idStr := strconv.Itoa(int(created.TrxId))
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Lokasi == nil || *fetched.Lokasi != "Jakarta" {
		t.Errorf("expected Jakarta, got %v", fetched.Lokasi)
	}
	if fetched.KelasName == nil || fetched.KelasName.LookupDescription == nil || *fetched.KelasName.LookupDescription != "Kelas Tingkat Dasar" {
		t.Errorf("expected preloaded KelasName, got %v", fetched.KelasName)
	}

	// 4. Update
	newLokasi := "Surabaya"
	created.Lokasi = &newLokasi
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Lokasi == nil || *updated.Lokasi != "Surabaya" {
		t.Errorf("expected Surabaya, got %v", updated.Lokasi)
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

func TestKelasServiceLookup(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	db.Exec("DELETE FROM T_APP_LOOKUP")
	db.Exec("DELETE FROM T_APP_LOOKUPCATEGORY")

	cat := domain.AppLookupCategory{
		CategoryId:          "B_KELASKHUSUS",
		CategoryType:        "S",
		CategoryDescription: "Kelas Khusus",
		Status:              true,
	}
	db.Create(&cat)

	val1 := "KELAS_DASAR"
	desc1 := "Kelas Tingkat Dasar"
	k1 := domain.AppLookup{
		LookupId:          "KL001",
		CategoryId:        &cat.CategoryId,
		LookupValue:       &val1,
		LookupDescription: &desc1,
	}
	db.Create(&k1)

	svc := NewKelasService(db)

	items, total, err := svc.Lookup(map[string]string{}, 1, 10)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	filtered, fTotal, err := svc.Lookup(map[string]string{"lookup_description": "Tingkat"}, 1, 10)
	if err != nil {
		t.Fatalf("Lookup with filter failed: %v", err)
	}
	if fTotal != 1 || len(filtered) != 1 {
		t.Errorf("expected 1 filtered item, got %d", fTotal)
	}
}
