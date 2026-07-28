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

	cat := domain.AppLookupCategory{
		CategoryId:          "B_FOTHANG",
		CategoryType:        "S",
		CategoryDescription: "Fotang",
		Status:              true,
	}
	db.Create(&cat)

	val1 := "FOTHANG_JKT"
	desc1 := "Fotang Jakarta Pusat"
	f1 := domain.AppLookup{
		LookupId:          "FT001",
		CategoryId:        &cat.CategoryId,
		LookupValue:       &val1,
		LookupDescription: &desc1,
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
}
