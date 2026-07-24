package service

import (
	"net/http/httptest"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
)

func setupTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", 1)
	return c
}

func TestUmatService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewUmatService(db)
	c := setupTestContext()

	// 1. Create
	umat := domain.Umat{
		Kode:                 "UM001",
		NamaIndonesia:        "Budi Santoso",
		NamaMandarin:         "武帝",
		Alias:                "Budi",
		TahunChiutaoMandarin: "2024",
	}

	created, err := svc.Create(umat, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}
	if !created.Status {
		t.Errorf("expected Status to be true")
	}

	// 2. Get
	fetched, err := svc.Get("1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.NamaIndonesia != "Budi Santoso" {
		t.Errorf("expected 'Budi Santoso', got '%s'", fetched.NamaIndonesia)
	}

	// 3. List & Filter
	filters := map[string]string{"namaindonesia": "Budi", "tahunchiutaomandarin": "2024"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got total %d, items len %d", total, len(items))
	}

	// 4. Update
	created.NamaIndonesia = "Budi Updated"
	updated, err := svc.Update("1", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.NamaIndonesia != "Budi Updated" {
		t.Errorf("expected 'Budi Updated', got '%s'", updated.NamaIndonesia)
	}

	// 5. Delete
	err = svc.Delete("1", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify soft delete filter
	itemsAfterDelete, totalAfter, _ := svc.List(1, nil, 10)
	if totalAfter != 0 || len(itemsAfterDelete) != 0 {
		t.Errorf("expected 0 active items after delete, got total %d", totalAfter)
	}
}
