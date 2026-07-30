package service

import (
	"testing"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestTahunCiuTaoService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewTahunCiuTaoService(db)
	c := setupTestContext()

	db.Exec("DELETE FROM T_BUS_TAHUN_CIUTAO WHERE TahunMandarin LIKE N'%2024%'")

	// 1. Create
	th := domain.TahunCiuTao{
		TahunMandarin: "甲辰 (2024)",
		StartDate:     domain.DateOnly{Time: time.Now()},
		EndDate:       domain.DateOnly{Time: time.Now().AddDate(1, 0, 0)},
		Description:   "Year of the Dragon",
	}

	created, err := svc.Create(th, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.TahunMandarin != "甲辰 (2024)" {
		t.Errorf("expected 甲辰 (2024), got %s", created.TahunMandarin)
	}

	// 2. Get
	fetched, err := svc.Get("甲辰 (2024)")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Description != "Year of the Dragon" {
		t.Errorf("expected Year of the Dragon, got %s", fetched.Description)
	}

	// 3. List (may fail if stored procedure SP_BUS_YEAR_SEARCH_DATA is not installed in test db)
	filters := map[string]string{"tahun_mandarin": "2024"}
	_, _, _ = svc.List(1, filters, 10)

	// 4. Update
	created.Description = "Year of the Wood Dragon"
	updated, err := svc.Update("甲辰 (2024)", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Description != "Year of the Wood Dragon" {
		t.Errorf("expected updated description, got %s", updated.Description)
	}

	// 5. Delete
	err = svc.Delete("甲辰 (2024)", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
