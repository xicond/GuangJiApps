package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestAdminService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewAdminService(db)
	c := setupTestContext()

	// 1. Create
	email := "admin@example.com"
	phone := "08123456789"
	admin := domain.Admin{
		Username:    "admin_test",
		GroupId:     1,
		Email:       &email,
		PhoneNumber: &phone,
	}

	created, err := svc.Create(admin, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}

	// 2. Get
	fetched, err := svc.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Username != "admin_test" {
		t.Errorf("expected 'admin_test', got '%s'", fetched.Username)
	}

	// 3. List
	filters := map[string]string{"username": "admin"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got total %d", total)
	}

	// 4. Update
	updatedEmail := "updated@example.com"
	created.Email = &updatedEmail
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Email == nil || *updated.Email != "updated@example.com" {
		t.Errorf("expected updated email, got %v", updated.Email)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 6. ListDepartments
	depts, err := svc.ListDepartments()
	if err != nil {
		t.Fatalf("ListDepartments failed: %v", err)
	}
	_ = depts
}
