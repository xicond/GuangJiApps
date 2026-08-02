package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestFotangServiceLookup(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = 'B_FOTHANG'")
	db.Exec("DELETE FROM T_APP_LOOKUPCATEGORY WHERE CategoryId = 'B_FOTHANG'")

	cat := domain.AppLookupCategory{
		CategoryId:          "B_FOTHANG",
		CategoryType:        "S",
		CategoryDescription: "Fotang",
		Status:              true,
	}
	db.Create(&cat)

	val1 := "FOTHANG_JKT"
	desc1 := "Fotang Jakarta Pusat"
	statusTrue := true
	f1 := domain.AppLookup{
		LookupId:          "FT001",
		CategoryId:        &cat.CategoryId,
		LookupValue:       &val1,
		LookupDescription: &desc1,
		Status:            &statusTrue,
	}
	db.Create(&f1)

	svc := NewFotangService(db)

	// 1. Fetch lookup list
	items, total, err := svc.Lookup(map[string]string{}, 1, 10)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 2. Search filter lookup_description
	filtered, fTotal, err := svc.Lookup(map[string]string{"lookup_description": "Jakarta"}, 1, 10)
	if err != nil {
		t.Fatalf("Lookup with filter failed: %v", err)
	}
	if fTotal != 1 || len(filtered) != 1 {
		t.Errorf("expected 1 filtered item, got %d", fTotal)
	}

	// 3. Fetch all with limit = 0
	all, allTotal, err := svc.Lookup(map[string]string{}, 0, 0)
	if err != nil {
		t.Fatalf("Lookup with limit=0 failed: %v", err)
	}
	if allTotal != 1 || len(all) != 1 {
		t.Errorf("expected 1 item when limit=0, got %d", len(all))
	}
}
