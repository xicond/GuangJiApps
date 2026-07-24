package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestAdminSubWarehouseService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewAdminSubWarehouseService(db)
	c := setupTestContext()

	// 1. Create
	subWh := domain.AdminSubWarehouse{
		FullName:  "Central Warehouse",
		SubWhType: "MAIN",
		Pic:       "John Doe",
		DocCode:   "WH001",
	}

	created, err := svc.Create(subWh, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.SubWhId != 1 {
		t.Errorf("expected SubWhId 1, got %d", created.SubWhId)
	}

	// 2. Get
	fetched, err := svc.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.FullName != "Central Warehouse" {
		t.Errorf("expected Central Warehouse, got %s", fetched.FullName)
	}

	// 3. List
	filters := map[string]string{"full_name": "Central"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.FullName = "Main Warehouse Hub"
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.FullName != "Main Warehouse Hub" {
		t.Errorf("expected Main Warehouse Hub, got %s", updated.FullName)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
