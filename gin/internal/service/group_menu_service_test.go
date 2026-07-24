package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestGroupMenuService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewGroupMenuService(db)
	c := setupTestContext()

	// 1. Create
	menu := domain.GroupMenuMapping{
		MenuName: "Umat Management",
		PageUrl:  "/master-data/umat",
		MenuDesc: "Manage Umat Records",
	}

	created, err := svc.Create(menu, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.MenuId != 1 {
		t.Errorf("expected MenuId 1, got %d", created.MenuId)
	}

	// 2. Get
	fetched, err := svc.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.MenuName != "Umat Management" {
		t.Errorf("expected 'Umat Management', got '%s'", fetched.MenuName)
	}

	// 3. List
	filters := map[string]string{"menu_name": "Umat"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.MenuName = "Umat Master Data"
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.MenuName != "Umat Master Data" {
		t.Errorf("expected updated menu name, got %s", updated.MenuName)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
