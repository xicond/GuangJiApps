package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestKelasPesertaService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

	svc := NewKelasPesertaService(db)
	c := setupTestContext()

	// Seed target FK records (Kelas & Umat) for lookup validation
	kodeK := "K01"
	db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", 26160003)
	db.Exec("DELETE FROM T_BUS_UMAT WHERE id = ?", 101)
	db.Create(&domain.Kelas{TrxId: 26160003, KodeKelas: &kodeK})
	kode101 := "UM101"
	db.Create(&domain.Umat{ID: 101, Kode: &kode101, NamaIndonesia: "Peserta Test", JenisKelamin: "001"})
	defer func() {
		db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", 26160003)
		db.Exec("DELETE FROM T_BUS_UMAT WHERE id = ?", 101)
	}()

	// 1. Validation error tests - missing TrxId or IdPeserta
	_, err = svc.Create(domain.KelasPeserta{}, c)
	if err == nil {
		t.Fatalf("expected error for missing required fields, got nil")
	}

	// Validation error for non-existent TrxId or IdPeserta
	invalidTrx := int32(999999)
	invalidUmat := int32(888888)
	_, err = svc.Create(domain.KelasPeserta{TrxId: invalidTrx, IdPeserta: &invalidUmat}, c)
	if err == nil {
		t.Fatalf("expected lookup validation error for invalid FKs, got nil")
	}

	// 2. Successful Create
	trxId := int32(26160003)
	idPeserta := int32(101)
	sumbangan := 500000.0
	barang := "Buku"
	ket := "Peserta baru"

	p := domain.KelasPeserta{
		TrxId:      trxId,
		IdPeserta:  &idPeserta,
		Sumbangan:  &sumbangan,
		Barang:     &barang,
		Keterangan: &ket,
	}

	created, err := svc.Create(p, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.DetailId == 0 {
		t.Errorf("expected non-zero DetailId")
	}
	if created.TrxId != trxId {
		t.Errorf("expected TrxId %d, got %d", trxId, created.TrxId)
	}

	// 3. Get
	idStr := strconv.Itoa(int(created.DetailId))
	fetched, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Keterangan == nil || *fetched.Keterangan != "Peserta baru" {
		t.Errorf("expected Keterangan 'Peserta baru', got %v", fetched.Keterangan)
	}

	// 4. Update
	newKet := "Peserta update"
	created.Keterangan = &newKet
	updated, err := svc.Update(idStr, created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Keterangan == nil || *updated.Keterangan != "Peserta update" {
		t.Errorf("expected Keterangan 'Peserta update', got %v", updated.Keterangan)
	}

	// 5. Delete
	err = svc.Delete(idStr, c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify status after soft delete
	deletedItem, err := svc.Get(idStr)
	if err != nil {
		t.Fatalf("Get after soft delete failed: %v", err)
	}
	if deletedItem.Status != nil && *deletedItem.Status != false {
		t.Errorf("expected status false after delete, got %v", *deletedItem.Status)
	}

	// 6. Test nested Umat ikrar saving for KodeKelas 004
	kode004 := "004"
	trx004 := int32(26160004)
	db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", trx004)
	db.Create(&domain.Kelas{TrxId: trx004, KodeKelas: &kode004})
	defer db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", trx004)

	p004 := domain.KelasPeserta{
		TrxId:     trx004,
		IdPeserta: &idPeserta,
		Umat: &domain.Umat{
			Ikrar1: true,
			Ikrar2: false,
			Ikrar3: true,
		},
	}
	created004, err := svc.Create(p004, c)
	if err != nil {
		t.Fatalf("Create for 004 failed: %v", err)
	}
	id004Str := strconv.Itoa(int(created004.DetailId))
	defer svc.Delete(id004Str, c)

	var umatCheck domain.Umat
	if err := db.Where("id = ?", idPeserta).Take(&umatCheck).Error; err != nil {
		t.Fatalf("Failed to fetch umat: %v", err)
	}
	if !umatCheck.Ikrar1 || !umatCheck.Ikrar3 || umatCheck.Ikrar2 {
		t.Errorf("Expected Ikrar1=true, Ikrar2=false, Ikrar3=true in Umat, got %+v", umatCheck)
	}

	// Test nested update
	p004.Umat.Ikrar2 = true
	lulusTrue := true
	p004.Lulus = &lulusTrue
	_, err = svc.Update(id004Str, p004, c)
	if err != nil {
		t.Fatalf("Update for 004 failed: %v", err)
	}
	if err := db.Where("id = ?", idPeserta).Take(&umatCheck).Error; err != nil {
		t.Fatalf("Failed to fetch umat after update: %v", err)
	}
	if !umatCheck.Ikrar2 {
		t.Errorf("Expected Ikrar2=true in Umat after update")
	}
	if !umatCheck.Sd3 {
		t.Errorf("Expected Sd3=true in Umat after Lulus=true update")
	}
	if umatCheck.StatusUmat == nil || *umatCheck.StatusUmat != "001" {
		t.Errorf("Expected StatusUmat='001' after Lulus=true update, got %v", umatCheck.StatusUmat)
	}

	// Test Lulus=false revert
	lulusFalse := false
	p004.Lulus = &lulusFalse
	_, err = svc.Update(id004Str, p004, c)
	if err != nil {
		t.Fatalf("Update for 004 Lulus=false failed: %v", err)
	}
	if err := db.Where("id = ?", idPeserta).Take(&umatCheck).Error; err != nil {
		t.Fatalf("Failed to fetch umat after Lulus=false update: %v", err)
	}
	if umatCheck.Sd3 {
		t.Errorf("Expected Sd3=false in Umat after Lulus=false update")
	}
	if umatCheck.StatusUmat == nil || *umatCheck.StatusUmat != "006" {
		t.Errorf("Expected StatusUmat='006' after Lulus=false update, got %v", umatCheck.StatusUmat)
	}
}

func TestKelasPesertaService_LoadPrevious(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	svc := NewKelasPesertaService(db)
	c := setupTestContext()

	// 1. Invalid TrxId format
	_, _, err = svc.LoadPrevious("invalid", c, 1, 10)
	if err == nil {
		t.Errorf("expected error for invalid ID format, got nil")
	}

	// 2. Non-existent Kelas
	_, _, err = svc.LoadPrevious("999999999", c, 1, 10)
	if err == nil {
		t.Errorf("expected error for non-existent kelas, got nil")
	}

	// 3. Valid Kelas (e.g. 26340002)
	items, total, err := svc.LoadPrevious("26340002", c, 1, 10)
	if err != nil {
		t.Fatalf("LoadPrevious failed: %v", err)
	}
	if total < 0 {
		t.Errorf("expected total >= 0, got %d", total)
	}

	// 4. Test pagination limit
	if len(items) > 1 {
		pagedItems, pagedTotal, err := svc.LoadPrevious("26340002", c, 1, 1)
		if err != nil {
			t.Fatalf("LoadPrevious pagination failed: %v", err)
		}
		if len(pagedItems) > 1 {
			t.Errorf("expected at most 1 item for limit 1, got %d", len(pagedItems))
		}
		if pagedTotal != total {
			t.Errorf("expected pagedTotal (%d) == total (%d)", pagedTotal, total)
		}
	}
}

func TestKelasPesertaService_CreateBulk(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	svc := NewKelasPesertaService(db)
	c := setupTestContext()

	// 1. Missing required fields
	_, err = svc.CreateBulk(domain.KelasPesertaBulkRequest{}, c)
	if err == nil {
		t.Errorf("expected error for empty bulk payload, got nil")
	}

	// 2. Invalid FKs (non-existent trx_id)
	_, err = svc.CreateBulk(domain.KelasPesertaBulkRequest{
		TrxId:     9999999,
		IdPeserta: []int32{888888},
	}, c)
	if err == nil {
		t.Errorf("expected lookup error for invalid FKs, got nil")
	}

	// 3. Seed test data
	testTrxId := int32(26160004)
	testUmatId1 := int32(1011)
	testUmatId2 := int32(1012)
	kodeK := "K01"
	kode1 := "UM1011"
	kode2 := "UM1012"

	db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", testTrxId)
	db.Exec("DELETE FROM T_BUS_UMAT WHERE id IN (?, ?)", testUmatId1, testUmatId2)
	db.Create(&domain.Kelas{TrxId: testTrxId, KodeKelas: &kodeK})
	db.Create(&domain.Umat{ID: testUmatId1, Kode: &kode1, NamaIndonesia: "Bulk Test 1", JenisKelamin: "001"})
	db.Create(&domain.Umat{ID: testUmatId2, Kode: &kode2, NamaIndonesia: "Bulk Test 2", JenisKelamin: "001"})

	defer func() {
		db.Exec("DELETE FROM T_TRX_KELAS_PESERTA WHERE trxid = ?", testTrxId)
		db.Exec("DELETE FROM T_TRX_KELAS WHERE trxid = ?", testTrxId)
		db.Exec("DELETE FROM T_BUS_UMAT WHERE id IN (?, ?)", testUmatId1, testUmatId2)
	}()

	sumbangan := 500000.0
	barang := "Buku Dharma"
	ket := "Bulk Created"

	bulkReq := domain.KelasPesertaBulkRequest{
		TrxId:      testTrxId,
		IdPeserta:  []int32{testUmatId1, testUmatId2},
		Sumbangan:  &sumbangan,
		Barang:     &barang,
		Keterangan: &ket,
	}

	created, err := svc.CreateBulk(bulkReq, c)
	if err != nil {
		t.Fatalf("CreateBulk failed: %v", err)
	}

	if len(created) != 2 {
		t.Fatalf("expected 2 created items, got %d", len(created))
	}

	for _, item := range created {
		if item.DetailId == 0 {
			t.Errorf("expected non-zero DetailId")
		}
		if item.TrxId != testTrxId {
			t.Errorf("expected TrxId %d, got %d", testTrxId, item.TrxId)
		}
		if item.Keterangan == nil || *item.Keterangan != ket {
			t.Errorf("expected Keterangan '%s', got %v", ket, item.Keterangan)
		}
	}
}

func TestKelasPeserta_PayloadTypeValidation(t *testing.T) {
	// 1. Array JSON into single KelasPeserta struct fails to unmarshal with custom error
	arrayJSON := `{"trx_id": 26160003, "id_peserta": [101, 102]}`
	var singlePayload domain.KelasPeserta
	err := json.Unmarshal([]byte(arrayJSON), &singlePayload)
	if err == nil {
		t.Errorf("expected error when unmarshaling array id_peserta into single KelasPeserta, got nil")
	} else if !strings.Contains(err.Error(), "idpeserta should be single value of umat") {
		t.Errorf("expected 'idpeserta should be single value of umat', got: %v", err)
	}

	// 2. Single number JSON into KelasPesertaBulkRequest struct fails to unmarshal with custom error
	singleJSON := `{"trx_id": 26160003, "id_peserta": 101}`
	var bulkPayload domain.KelasPesertaBulkRequest
	err = json.Unmarshal([]byte(singleJSON), &bulkPayload)
	if err == nil {
		t.Errorf("expected error when unmarshaling single number into KelasPesertaBulkRequest.id_peserta, got nil")
	} else if !strings.Contains(err.Error(), "idpeserta must array") {
		t.Errorf("expected 'idpeserta must array', got: %v", err)
	}

	// 3. Testing with Gin test context ShouldBindJSON
	c1, _ := gin.CreateTestContext(httptest.NewRecorder())
	c1.Request, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader(arrayJSON))
	c1.Request.Header.Set("Content-Type", "application/json")
	var bindSingle domain.KelasPeserta
	if err := c1.ShouldBindJSON(&bindSingle); err == nil {
		t.Errorf("expected ShouldBindJSON error for array id_peserta on single payload, got nil")
	} else if !strings.Contains(err.Error(), "idpeserta should be single value of umat") {
		t.Errorf("expected 'idpeserta should be single value of umat', got: %v", err)
	}

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader(singleJSON))
	c2.Request.Header.Set("Content-Type", "application/json")
	var bindBulk domain.KelasPesertaBulkRequest
	if err := c2.ShouldBindJSON(&bindBulk); err == nil {
		t.Errorf("expected ShouldBindJSON error for single id_peserta on bulk payload, got nil")
	} else if !strings.Contains(err.Error(), "idpeserta must array") {
		t.Errorf("expected 'idpeserta must array', got: %v", err)
	}
}
