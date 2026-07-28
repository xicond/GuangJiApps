package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestActivityService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewActivityService(db)
	c := setupTestContext()

	// 1. Create
	act := domain.Activity{
		EventCode:     "EV001",
		EventName:     "Annual Gathering",
		EventCategory: "CER",
		Description:   "Yearly event",
	}

	created, err := svc.Create(act, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.EventCode != "EV001" {
		t.Errorf("expected EV001, got %s", created.EventCode)
	}

	// 2. Get
	fetched, err := svc.Get("EV001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.EventName != "Annual Gathering" {
		t.Errorf("expected Annual Gathering, got %s", fetched.EventName)
	}

	// 3. List
	filters := map[string]string{"event_name": "Gathering"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.EventName = "Grand Gathering"
	updated, err := svc.Update("EV001", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.EventName != "Grand Gathering" {
		t.Errorf("expected Grand Gathering, got %s", updated.EventName)
	}

	// 5. Delete
	err = svc.Delete("EV001", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
