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

	// Clean up existing test umat records to ensure test isolation
	db.Exec("DELETE FROM T_BUS_UMAT")

	// Seed lookup category and values for validation testing safely using FirstOrCreate
	catStatus := domain.AppLookupCategory{CategoryId: "B_STATUS", Status: true}
	db.FirstOrCreate(&catStatus, domain.AppLookupCategory{CategoryId: "B_STATUS"})
	valStatus := "001"
	statusTrue := true
	lookupItem := domain.AppLookup{LookupId: "L001", CategoryId: &catStatus.CategoryId, LookupValue: &valStatus, Status: &statusTrue}
	db.FirstOrCreate(&lookupItem, domain.AppLookup{LookupId: "L001"})

	invalidStatus := "INVALID_STATUS"
	// Test Invalid Lookup Rejection
	invalidUmat := domain.Umat{
		NamaIndonesia: "Test Invalid",
		JenisKelamin:  "L",
		StatusUmat:    &invalidStatus,
	}
	_, err = svc.Create(invalidUmat, nil, c)
	if err == nil {
		t.Fatalf("expected error when creating umat with invalid StatusUmat, got nil")
	}

	kode := "UM001"
	mandarin := "武帝"
	alias := "Budi"
	statusUmat := "001"

	// 1. Create with Valid Lookup
	umat := domain.Umat{
		Kode:                 &kode,
		NamaIndonesia:        "Budi Santoso",
		NamaMandarin:         &mandarin,
		Alias:                &alias,
		TahunChiutaoMandarin: "2024",
		JenisKelamin:         "L",
		StatusUmat:           &statusUmat,
	}

	created, err := svc.Create(umat, nil, c)
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
	items, total, err := svc.List(c, 1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got total %d, items len %d", total, len(items))
	}

	// 4. Update
	created.NamaIndonesia = "Budi Updated"
	updated, err := svc.Update("1", created, nil, c)
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
	itemsAfterDelete, totalAfter, _ := svc.List(c, 1, nil, 10)
	if totalAfter != 0 || len(itemsAfterDelete) != 0 {
		t.Errorf("expected 0 active items after delete, got total %d", totalAfter)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal.jpg", "normal.jpg"},
		{"../path/to/file.png", "file.png"},
		{"CON.jpg", "_CON.jpg"},
		{"invalid<>:?*name.png", "invalid_____name.png"},
		{"my file name.jpg", "my_file_name.jpg"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseUmatOcrText(t *testing.T) {
	rawText := "一掛號單\n編\n號:3986\n姓名\nNAMA:楊慧鈴 鈴\n年齡\nUMUR: 51\n教育\nPENDIDIKAN: SI\n地址\n性別 Wanita\nJENIS KELAMIN\nALAMAT: Il. Fatmawati No-30\n電話: 081933363639\n引師\nPERANTARA:伍輝生\n保師\nPENANGGUNG:徐·盛超\n點傳師:林榮茂\n日期\nTANGGAL:丙午年六月二十七日 9/8′26\n功德費: Rp.10.000 未"

	parsed := ParseUmatOcrText(rawText)

	if parsed["nama_indonesia"] != "楊慧鈴" {
		t.Errorf("expected nama_indonesia '楊慧鈴', got '%v'", parsed["nama_indonesia"])
	}
	if parsed["nama_mandarin"] != "楊慧鈴" {
		t.Errorf("expected nama_mandarin '楊慧鈴', got '%v'", parsed["nama_mandarin"])
	}
	if parsed["alias"] != "Yang Hui Ling" {
		t.Errorf("expected alias 'Yang Hui Ling', got '%v'", parsed["alias"])
	}
	if parsed["usia"] != 51 {
		t.Errorf("expected usia 51, got '%v'", parsed["usia"])
	}
	if parsed["pendidikan"] != "S1" {
		t.Errorf("expected pendidikan 'S1', got '%v'", parsed["pendidikan"])
	}
	if parsed["jenis_kelamin"] != "WANITA" {
		t.Errorf("expected jenis_kelamin 'WANITA', got '%v'", parsed["jenis_kelamin"])
	}
	if parsed["alamat"] != "JL. Fatmawati No-30" {
		t.Errorf("expected alamat 'JL. Fatmawati No-30', got '%v'", parsed["alamat"])
	}
	if parsed["mobile"] != "081933363639" {
		t.Errorf("expected mobile '081933363639', got '%v'", parsed["mobile"])
	}
	if parsed["pengajak_manual"] != "伍輝生" {
		t.Errorf("expected pengajak_manual '伍輝生', got '%v'", parsed["pengajak_manual"])
	}
	if parsed["penanggung_manual"] != "徐盛超" {
		t.Errorf("expected penanggung_manual '徐盛超', got '%v'", parsed["penanggung_manual"])
	}
	if parsed["tcs"] != "林榮茂" {
		t.Errorf("expected tcs '林榮茂', got '%v'", parsed["tcs"])
	}
	if parsed["uang_pahala"] != float64(10000) {
		t.Errorf("expected uang_pahala 10000, got '%v'", parsed["uang_pahala"])
	}
	if parsed["waktu_chiutao_mandarin"] != "未" {
		t.Errorf("expected waktu_chiutao_mandarin '未', got '%v'", parsed["waktu_chiutao_mandarin"])
	}
}

