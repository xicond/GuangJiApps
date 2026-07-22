package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestAdminServiceCRUD(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test db failed: %v", err)
	}
	service := NewAdminService(db)

	created, err := service.Create(domain.Resource{Code: "ADM001", Name: "Admin One"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected created id")
	}

	items := service.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	got, err := service.Get(created.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "Admin One" {
		t.Fatalf("expected name Admin One, got %s", got.Name)
	}

	updated, err := service.Update(created.ID, domain.Resource{Code: "ADM002", Name: "Admin Two", Active: false})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Code != "ADM002" || updated.Active {
		t.Fatalf("unexpected updated payload: %+v", updated)
	}

	if err := service.Delete(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if len(service.List()) != 0 {
		t.Fatalf("expected empty list after delete")
	}
}
