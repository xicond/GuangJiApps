package service

import (
	"fmt"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

func TestLookupService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// if err := db.AutoMigrate(&domain.AppLookupCategory{}, &domain.AppLookup{}); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

	// 1. Database nil check
	_, _, err = Lookup(nil, "B_TEST", 1, 10)
	if err == nil {
		t.Errorf("expected error for nil db, got nil")
	}

	// Seed Category
	catID := "B_TEST"
	db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = ?", catID)
	db.Exec("DELETE FROM T_APP_LOOKUPCATEGORY WHERE CategoryId = ?", catID)

	cat := domain.AppLookupCategory{
		CategoryId:          catID,
		CategoryType:        "S",
		CategoryDescription: "Test Category",
		Status:              true,
	}
	db.Create(&cat)

	trueStatus := true
	// Seed Lookups
	for i := 1; i <= 15; i++ {
		val := fmt.Sprintf("VAL_%02d", i)
		desc := fmt.Sprintf("Description Test %d", i)
		id := fmt.Sprintf("ID_%02d", i)
		item := domain.AppLookup{
			LookupId:          id,
			CategoryId:        &cat.CategoryId,
			LookupValue:       &val,
			LookupDescription: &desc,
			Status:            &trueStatus,
		}
		db.Create(&item)
	}

	// 2. Limit = 0 (Return all, ignore page)
	itemsAll, totalAll, err := Lookup(db, catID, 5, 0) // page=5 ignored because limit=0
	if err != nil {
		t.Fatalf("Lookup with limit=0 failed: %v", err)
	}
	if totalAll != 15 {
		t.Errorf("expected total 15, got %d", totalAll)
	}
	if len(itemsAll) != 15 {
		t.Errorf("expected 15 items when limit=0, got %d", len(itemsAll))
	}

	// 3. Limit > 0 (Paginated)
	itemsP1, totalP1, err := Lookup(db, catID, 1, 10)
	if err != nil {
		t.Fatalf("Lookup paginated page 1 failed: %v", err)
	}
	if totalP1 != 15 {
		t.Errorf("expected total 15, got %d", totalP1)
	}
	if len(itemsP1) != 10 {
		t.Errorf("expected 10 items on page 1, got %d", len(itemsP1))
	}

	itemsP2, _, err := Lookup(db, catID, 2, 10)
	if err != nil {
		t.Fatalf("Lookup paginated page 2 failed: %v", err)
	}
	if len(itemsP2) != 5 {
		t.Errorf("expected 5 items on page 2, got %d", len(itemsP2))
	}

	// 4. Filters (lookup_description, lookup_value, lookup_id)
	filteredDesc, totalDesc, err := Lookup(db, catID, 1, 10, map[string]string{"lookup_description": "Test 15"})
	if err != nil {
		t.Fatalf("Lookup filter description failed: %v", err)
	}
	if totalDesc != 1 || len(filteredDesc) != 1 {
		t.Errorf("expected 1 item filtered by description, got total=%d len=%d", totalDesc, len(filteredDesc))
	}

	filteredVal, totalVal, err := Lookup(db, catID, 1, 10, map[string]string{"lookup_value": "VAL_05"})
	if err != nil {
		t.Fatalf("Lookup filter value failed: %v", err)
	}
	if totalVal != 1 || len(filteredVal) != 1 {
		t.Errorf("expected 1 item filtered by value, got total=%d len=%d", totalVal, len(filteredVal))
	}

	filteredID, totalID, err := Lookup(db, catID, 1, 10, map[string]string{"lookup_id": "ID_08"})
	if err != nil {
		t.Fatalf("Lookup filter id failed: %v", err)
	}
	if totalID != 1 || len(filteredID) != 1 {
		t.Errorf("expected 1 item filtered by id, got total=%d len=%d", totalID, len(filteredID))
	}

	// 4b. Test searching by lookup_value via lookup_description filter or search parameter
	filteredByValInDesc, totalByValInDesc, err := Lookup(db, catID, 1, 10, map[string]string{"lookup_description": "VAL_07"})
	if err != nil {
		t.Fatalf("Lookup filter search by value in description filter failed: %v", err)
	}
	if totalByValInDesc != 1 || len(filteredByValInDesc) != 1 {
		t.Errorf("expected 1 item searching by value in description filter, got total=%d len=%d", totalByValInDesc, len(filteredByValInDesc))
	}

	filteredSearch, totalSearch, err := Lookup(db, catID, 1, 10, map[string]string{"search": "VAL_09"})
	if err != nil {
		t.Fatalf("Lookup filter search parameter failed: %v", err)
	}
	if totalSearch != 1 || len(filteredSearch) != 1 {
		t.Errorf("expected 1 item using search parameter, got total=%d len=%d", totalSearch, len(filteredSearch))
	}

	// 5. Test LookupService struct helper methods
	svc := NewLookupService(db)

	categories := []string{
		"B_WAKTUCIUTAO", "B_GENDER", "B_TCS", "B_FOTHANG", "B_KELAS",
		"B_PENDIDIKAN", "B_KELASUMUM", "B_PEKERJAAN", "B_KELUARGA", "B_STATUS",
	}
	for _, cID := range categories {
		id := "ID_" + cID
		db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = ?", cID)
		db.Exec("DELETE FROM T_APP_LOOKUPCATEGORY WHERE CategoryId = ?", cID)

		cObj := domain.AppLookupCategory{CategoryId: cID, CategoryType: "S", CategoryDescription: cID, Status: true}
		db.Create(&cObj)
		val := "VAL_" + cID
		desc := "Desc " + cID
		lObj := domain.AppLookup{LookupId: id, CategoryId: &cObj.CategoryId, LookupValue: &val, LookupDescription: &desc, Status: &trueStatus}
		db.Create(&lObj)
	}

	if res, count, err := svc.LookupWaktuCiuTao(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupWaktuCiuTao failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupGender(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupGender failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupTcs(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupTcs failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupFotang(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupFotang failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupKelas(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupKelas failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupPendidikan(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupPendidikan failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupKelasUmum(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupKelasUmum failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupPekerjaan(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupPekerjaan failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupKeluarga(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupKeluarga failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.LookupStatus(nil, 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("LookupStatus failed: %v, count=%d", err, count)
	}
	if res, count, err := svc.Lookup("B_STATUS", 1, 10); err != nil || count < 1 || len(res) < 1 {
		t.Errorf("svc.Lookup failed: %v, count=%d", err, count)
	}

	// 6. Test custom subQuery parameter (modifier function and *gorm.DB)
	resMod, countMod, err := Lookup(db, "B_STATUS", 1, 10, nil, func(sub *gorm.DB) *gorm.DB {
		return sub.Where("CategoryType = ?", "S")
	})
	if err != nil || countMod != 1 || len(resMod) != 1 {
		t.Errorf("Lookup with subQuery modifier func failed: %v, count=%d", err, countMod)
	}

	customSub := db.Table("T_APP_LOOKUPCATEGORY").
		Where("T_APP_LOOKUPCATEGORY.CategoryId = T_APP_LOOKUP.CategoryId").
		Where("T_APP_LOOKUPCATEGORY.CategoryType = ?", "S")
	resDB, countDB, err := Lookup(db, "B_STATUS", 1, 10, nil, customSub)
	if err != nil || countDB != 1 || len(resDB) != 1 {
		t.Errorf("Lookup with custom *gorm.DB subQuery failed: %v, count=%d", err, countDB)
	}
}
