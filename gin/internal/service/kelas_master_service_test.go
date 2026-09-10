package service

import (
	"errors"
	"fmt"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestKelasMasterService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	svc := NewKelasMasterService(db)
	c := setupTestContext()

	// Clean up any test records
	db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = 'B_KELASKHUSUS' AND (LookupId LIKE 'B_KELASKHUSUSTEST%' OR LookupDescription LIKE 'Test Kelas%')")
	defer func() {
		db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = 'B_KELASKHUSUS' AND (LookupId LIKE 'B_KELASKHUSUSTEST%' OR LookupDescription LIKE 'Test Kelas%')")
	}()

	// 1. Create with automated LookupValue & LookupId
	desc1 := "Test Kelas Malaikat Baru"
	item1 := domain.AppLookup{
		LookupDescription: &desc1,
	}

	created1, err := svc.Create(item1, c)
	if err != nil {
		t.Fatalf("Create automated failed: %v", err)
	}
	if created1.LookupValue == nil || *created1.LookupValue == "" {
		t.Errorf("expected automated LookupValue, got nil or empty")
	}
	if created1.LookupId == "" {
		t.Errorf("expected automated LookupId, got empty")
	}
	if created1.LookupId != fmt.Sprintf("B_KELASKHUSUS%s", *created1.LookupValue) {
		t.Errorf("expected LookupId B_KELASKHUSUS%s, got %s", *created1.LookupValue, created1.LookupId)
	}

	// 2. Create duplicate name (should fail)
	descDup := "Test Kelas Malaikat Baru"
	itemDup := domain.AppLookup{
		LookupDescription: &descDup,
	}
	_, err = svc.Create(itemDup, c)
	if err == nil {
		t.Fatalf("expected error when creating duplicate class name, got nil")
	}
	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if len(vErr.Details["lookup_description"]) == 0 {
		t.Errorf("expected lookup_description error in details, got: %v", vErr.Details)
	}

	// 3. Create second record with manual LookupValue
	desc2 := "Test Kelas Pengabdi Lanjutan"
	val2 := "TEST01"
	id2 := "B_KELASKHUSUSTEST01"
	item2 := domain.AppLookup{
		LookupId:          id2,
		LookupValue:       &val2,
		LookupDescription: &desc2,
	}
	created2, err := svc.Create(item2, c)
	if err != nil {
		t.Fatalf("Create manual failed: %v", err)
	}
	if created2.LookupId != id2 {
		t.Errorf("expected %s, got %s", id2, created2.LookupId)
	}

	// 4. Get by ID and by LookupValue
	fetchedByID, err := svc.Get(created1.LookupId)
	if err != nil {
		t.Fatalf("Get by LookupId failed: %v", err)
	}
	if *fetchedByID.LookupDescription != desc1 {
		t.Errorf("expected '%s', got '%s'", desc1, *fetchedByID.LookupDescription)
	}

	fetchedByVal, err := svc.Get(*created1.LookupValue)
	if err != nil {
		t.Fatalf("Get by LookupValue failed: %v", err)
	}
	if fetchedByVal.LookupId != created1.LookupId {
		t.Errorf("expected LookupId '%s', got '%s'", created1.LookupId, fetchedByVal.LookupId)
	}

	// 5. List with filters
	filters := map[string]string{"lookup_description": "Malaikat Baru"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total < 1 || len(items) < 1 {
		t.Errorf("expected at least 1 item, got %d", total)
	}

	// 6. Update with same name (should succeed)
	updatePayload := created1
	updatedSame, err := svc.Update(created1.LookupId, updatePayload, c)
	if err != nil {
		t.Fatalf("Update with same description should succeed, got error: %v", err)
	}
	if *updatedSame.LookupDescription != desc1 {
		t.Errorf("expected '%s', got '%s'", desc1, *updatedSame.LookupDescription)
	}

	// 7. Update with duplicate name from item2 (should fail)
	updatePayload.LookupDescription = &desc2
	_, err = svc.Update(created1.LookupId, updatePayload, c)
	if err == nil {
		t.Fatalf("expected error updating to existing name '%s', got nil", desc2)
	}

	// 8. Delete
	err = svc.Delete(created1.LookupId, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify status is false
	deletedItem, err := svc.Get(created1.LookupId)
	if err != nil {
		t.Fatalf("Get deleted item failed: %v", err)
	}
	if deletedItem.Status != nil && *deletedItem.Status != false {
		t.Errorf("expected status false after delete, got %v", *deletedItem.Status)
	}
}
