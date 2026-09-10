package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestAdminGroupService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

	svc := NewAdminGroupService(db)
	c := setupTestContext()

	// 1. Create
	group := domain.AdminGroup{
		GroupName: "SuperAdmin",
		GroupDesc: "Full System Control",
		RInsert:   true,
		REdit:     true,
	}

	created, err := svc.Create(group, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.GroupId != 1 {
		t.Errorf("expected GroupId 1, got %d", created.GroupId)
	}

	// 2. Get
	fetched, err := svc.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.GroupName != "SuperAdmin" {
		t.Errorf("expected SuperAdmin, got %s", fetched.GroupName)
	}

	// 3. List
	filters := map[string]string{"group_name": "Super"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.GroupName = "MasterAdmin"
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.GroupName != "MasterAdmin" {
		t.Errorf("expected MasterAdmin, got %s", updated.GroupName)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
