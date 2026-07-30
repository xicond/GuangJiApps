package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestTimKerjaService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewTimKerjaService(db)
	c := setupTestContext()

	db.Exec("DELETE FROM T_APP_LOOKUP WHERE LookupId IN ('TK001', 'L_TK1', 'L_SK1', 'L_TR1')")
	db.Exec("DELETE FROM T_APP_LOOKUPCATEGORY WHERE CategoryId IN ('B_TIMKERJA', 'B_SUBKERJA', 'B_TIMKERJA_REPORT')")

	// 1. Create
	tim := domain.TimKerja{
		LookupId:          "TK001",
		LookupValue:       "Ketua Panitia",
		LookupDescription: "Head Coordinator",
	}

	created, err := svc.Create(tim, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.LookupId != "TK001" {
		t.Errorf("expected TK001, got %s", created.LookupId)
	}

	// 2. Get
	fetched, err := svc.Get("TK001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.LookupValue != "Ketua Panitia" {
		t.Errorf("expected Ketua Panitia, got %s", fetched.LookupValue)
	}

	// 3. List
	filters := map[string]string{"lookup_value": "Ketua"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.LookupValue = "Koordinator Utama"
	updated, err := svc.Update("TK001", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.LookupValue != "Koordinator Utama" {
		t.Errorf("expected Koordinator Utama, got %s", updated.LookupValue)
	}

	// 5. Delete
	err = svc.Delete("TK001", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 6. Test Lookup, LookupSub, LookupReport
	catTim := domain.AppLookupCategory{CategoryId: "B_TIMKERJA", Status: true}
	catSub := domain.AppLookupCategory{CategoryId: "B_SUBKERJA", Status: true}
	catRep := domain.AppLookupCategory{CategoryId: "B_TIMKERJA_REPORT", Status: true}
	db.Create(&catTim)
	db.Create(&catSub)
	db.Create(&catRep)

	v1, d1 := "V1", "D1"
	statusTrue := true
	db.Create(&domain.AppLookup{LookupId: "L_TK1", CategoryId: &catTim.CategoryId, LookupValue: &v1, LookupDescription: &d1, Status: &statusTrue})
	db.Create(&domain.AppLookup{LookupId: "L_SK1", CategoryId: &catSub.CategoryId, LookupValue: &v1, LookupDescription: &d1, Status: &statusTrue})
	db.Create(&domain.AppLookup{LookupId: "L_TR1", CategoryId: &catRep.CategoryId, LookupValue: &v1, LookupDescription: &d1, Status: &statusTrue})

	// Lookup (B_TIMKERJA)
	lTim, tTim, err := svc.Lookup(map[string]string{}, 1, 10)
	if err != nil || tTim != 1 || len(lTim) != 1 {
		t.Fatalf("Lookup B_TIMKERJA failed: %v, total=%d", err, tTim)
	}
	lTimAll, tTimAll, err := svc.Lookup(map[string]string{}, 0, 0)
	if err != nil || tTimAll != 1 || len(lTimAll) != 1 {
		t.Fatalf("Lookup B_TIMKERJA limit=0 failed: %v, total=%d", err, tTimAll)
	}

	// LookupSub (B_SUBKERJA)
	db.Exec("DELETE FROM T_BUS_WORK_MAPPING WHERE SubDivisi = 'V1'")
	wm := domain.WorkMapping{Divisi: "0", SubDivisi: "V1", Status: true}
	db.Create(&wm)

	lSub, tSub, err := svc.LookupSub(0, map[string]string{}, 1, 10)
	if err != nil || tSub != 1 || len(lSub) != 1 {
		t.Fatalf("LookupSub B_SUBKERJA failed: %v, total=%d", err, tSub)
	}

	// LookupReport (B_TIMKERJA_REPORT)
	lRep, tRep, err := svc.LookupReport(map[string]string{}, 1, 10)
	if err != nil || tRep != 1 || len(lRep) != 1 {
		t.Fatalf("LookupReport B_TIMKERJA_REPORT failed: %v, total=%d", err, tRep)
	}
}
