package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Fatalf("auto migrate failed: %v", err)
	// }

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

	// 4b. Update mobile and empty email (ensure no deadlock & email sanitized to nil)
	emptyMail := ""
	created.Email = &emptyMail
	newMobile := "08123456789"
	created.Mobile = &newMobile
	updated, err = svc.Update("1", created, nil, c)
	if err != nil {
		t.Fatalf("Update mobile & empty email failed: %v", err)
	}
	if updated.Mobile == nil || *updated.Mobile != "08123456789" {
		t.Errorf("expected mobile '08123456789', got %v", updated.Mobile)
	}
	if updated.Email != nil {
		t.Errorf("expected Email nil after setting empty string, got %v", *updated.Email)
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
		// Standard valid filenames
		{"normal.jpg", "normal.jpg"},
		{"photo_2026-08-15.png", "photo_2026-08-15.png"},
		{"my file name.jpg", "my_file_name.jpg"},

		// Directory / Path Traversal & Windows Drive Symbols (Linux & Windows)
		{"../path/to/file.png", "file.png"},
		{"../../../../etc/passwd", "passwd"},
		{"..\\..\\..\\boot.ini", "boot.ini"},
		{"C:\\Windows\\System32\\cmd.exe", "cmd.exe"},
		{"\\\\server\\share\\secret.png", "secret.png"},
		{"/var/www/uploads/shell.php", "shell.php"},
		{"segment1\\segment2\\..\\..\\..\\..\\..\\c:\\file.png", "file.png"},
		{"C:boot.ini", "C_boot.ini"},
		{"C:\\", "C_"},
		{"C:", "C_"},

		// Windows Reserved Device Names (Exact & Case-Insensitive)
		{"CON.jpg", "_CON.jpg"},
		{"con.PNG", "_con.PNG"},
		{"PRN.pdf", "_PRN.pdf"},
		{"AUX.png", "_AUX.png"},
		{"NUL.jpg", "_NUL.jpg"},
		{"COM1.png", "_COM1.png"},
		{"com1.jpg", "_com1.jpg"},
		{"COM9.png", "_COM9.png"},
		{"LPT1.txt", "_LPT1.txt"},
		{"lpt9.jpeg", "_lpt9.jpeg"},
		{"CLOCK$.txt", "_CLOCK$.txt"},
		{"COM1", "_COM1"},

		// Windows Device Names with Colons / Streams / Trailing Aliases
		{"COM1:", "_COM1_"},
		{"COM1:stream", "COM1_stream"},
		{"COM1. .jpg", "_COM1._.jpg"},
		{"file.png::$DATA", "file.png__$DATA"},
		{"file.png:stream", "file.png_stream"},

		// Invalid Characters & Control Characters (ASCII & DEL)
		{"invalid<>:?*name.png", "invalid_____name.png"},
		{"shell.php\x00.jpg", "shell.php_.jpg"},
		{"file\x07name\x1b.jpg", "file_name_.jpg"},
		{"file\x7fname.png", "file_name.png"},

		// Dangerous Unicode Characters (Zero-width space, RLO override, BOM)
		{"test\u202Egnp.exe", "test_gnp.exe"},   // Right-to-Left Override (RLO)
		{"file\u200Bname.jpg", "file_name.jpg"}, // Zero-Width Space
		{"file\u200Cname.png", "file_name.png"}, // Zero-Width Non-Joiner
		{"\uFEFFimage.jpg", "_image.jpg"},       // Byte Order Mark (BOM)

		// CJK (Chinese, Japanese, Korean) Allowed Filenames
		{"中文文件名.jpg", "中文文件名.jpg"},
		{"張三_楊慧鈴_佛堂.png", "張三_楊慧鈴_佛堂.png"},
		{"日本語画像.jpeg", "日本語画像.jpeg"},
		{"한국어_파일.png", "한국어_파일.png"},
		{"中文 文件 名.jpg", "中文_文件_名.jpg"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}

	// Test fallback default filename generation for empty / dot inputs
	emptyInputs := []string{"", "   ", ".", "..", "  .  ", "  ..  "}
	for _, input := range emptyInputs {
		res := sanitizeFilename(input)
		if !strings.HasPrefix(res, "image_") || !strings.HasSuffix(res, ".jpg") {
			t.Errorf("sanitizeFilename(%q) = %q, expected default 'image_*.jpg'", input, res)
		}
	}

	// Test long ASCII filename truncation (max 150 bytes, preserving extension)
	longInput := strings.Repeat("a", 200) + ".jpg"
	longResult := sanitizeFilename(longInput)
	if len(longResult) > 150 {
		t.Errorf("expected len <= 150 bytes, got %d for long filename", len(longResult))
	}
	if !strings.HasSuffix(longResult, ".jpg") {
		t.Errorf("expected long filename to preserve .jpg extension, got %q", longResult)
	}

	// Test CJK / multibyte long filename truncation (max runes = maxBytes / 2 = 75 runes, preserving extension & valid UTF-8)
	cjkLongInput := strings.Repeat("中文", 50) + ".jpg" // 100 CJK runes
	cjkLongResult := sanitizeFilename(cjkLongInput)
	if !utf8.ValidString(cjkLongResult) {
		t.Errorf("expected valid UTF-8 string for CJK long filename, got invalid bytes: %q", cjkLongResult)
	}
	cjkRunes := []rune(cjkLongResult)
	if len(cjkRunes) > 79 { // 75 base runes + 4 ext runes (.jpg)
		t.Errorf("expected max 79 runes for CJK long filename (75 base + 4 ext), got %d runes", len(cjkRunes))
	}
	if len(cjkLongResult) > 150 {
		t.Errorf("expected len <= 150 bytes, got %d bytes for CJK long filename", len(cjkLongResult))
	}
	if !strings.HasSuffix(cjkLongResult, ".jpg") {
		t.Errorf("expected CJK long filename to preserve .jpg extension, got %q", cjkLongResult)
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
	if parsed["tanggal_chiutao_int"] != "2026-08-09" {
		t.Errorf("expected tanggal_chiutao_int '2026-08-09', got '%v'", parsed["tanggal_chiutao_int"])
	}

	// Test Sample 2: 日期\n:\n22 Aug 26/丙午12/十日\nTANGGAL
	rawText2 := "日期\n:\n22 Aug 26/丙午12/十日\nTANGGAL"
	parsed2 := ParseUmatOcrText(rawText2)
	if parsed2["tanggal_chiutao_int"] != "2026-08-22" {
		t.Errorf("expected tanggal_chiutao_int '2026-08-22' for sample 2, got '%v'", parsed2["tanggal_chiutao_int"])
	}

	// Test Sample 3: PERANTARA Felx (no colon)
	rawText3 := "PERANTARA Felx"
	parsed3 := ParseUmatOcrText(rawText3)
	if parsed3["pengajak_manual"] != "Felx" {
		t.Errorf("expected pengajak_manual 'Felx' for sample 3, got '%v'", parsed3["pengajak_manual"])
	}

	// Test Sample 4: PENDIDIKAN: Si / S
	rawText4 := "PENDIDIKAN: Si"
	parsed4 := ParseUmatOcrText(rawText4)
	if parsed4["pendidikan"] != "S1" {
		t.Errorf("expected pendidikan 'S1' for 'PENDIDIKAN: Si', got '%v'", parsed4["pendidikan"])
	}

	rawText5 := "PENDIDIKAN: S"
	parsed5 := ParseUmatOcrText(rawText5)
	if parsed5["pendidikan"] != "S1" {
		t.Errorf("expected pendidikan 'S1' for 'PENDIDIKAN: S', got '%v'", parsed5["pendidikan"])
	}
}

func createMockFileHeader(filename string, content []byte) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("foto", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, err
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}
	return req.MultipartForm.File["foto"][0], nil
}

func createValidTestImage(width, height int, format string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	buf := &bytes.Buffer{}
	if format == "png" {
		_ = png.Encode(buf, img)
	} else {
		_ = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	}
	return buf.Bytes()
}

func TestValidateImageHeader_Security(t *testing.T) {
	// 1. Attack Scenario 1: Polyglot GIF with PHP payload named shell.php
	polyglotContent := []byte("GIF89a<?php eval($_POST['c']); ?>")
	fh, err := createMockFileHeader("shell.php", polyglotContent)
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	err = validateImageHeader(fh)
	if err == nil {
		t.Fatalf("expected validateImageHeader to reject shell.php, got nil")
	}

	// 2. Attack Scenario 2: Polyglot GIF with PHP payload named shell.jpg
	fhGIFJpg, err := createMockFileHeader("shell.jpg", polyglotContent)
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	err = validateImageHeader(fhGIFJpg)
	if err == nil {
		t.Fatalf("expected validateImageHeader to reject GIF MIME disguised as .jpg, got nil")
	}

	// 3. Attack Scenario 3: Double extension shell.php.jpg
	validJpg := createValidTestImage(300, 400, "jpeg")
	fhDoubleExt, err := createMockFileHeader("shell.php.jpg", validJpg)
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	err = validateImageHeader(fhDoubleExt)
	if err == nil {
		t.Fatalf("expected validateImageHeader to reject dangerous double extension shell.php.jpg, got nil")
	}

	// 4. Attack Scenario 4: Other script extensions in name
	fhPhtml, err := createMockFileHeader("image.phtml.png", createValidTestImage(300, 400, "png"))
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	err = validateImageHeader(fhPhtml)
	if err == nil {
		t.Fatalf("expected validateImageHeader to reject image.phtml.png, got nil")
	}

	// 5. Valid PNG (300x400, ratio 0.75)
	fhValidPNG, err := createMockFileHeader("pasfoto.png", createValidTestImage(300, 400, "png"))
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	if err := validateImageHeader(fhValidPNG); err != nil {
		t.Fatalf("expected valid PNG to pass, got error: %v", err)
	}

	// 6. Valid JPEG (300x400, ratio 0.75)
	fhValidJPG, err := createMockFileHeader("pasfoto.jpg", createValidTestImage(300, 400, "jpeg"))
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	if err := validateImageHeader(fhValidJPG); err != nil {
		t.Fatalf("expected valid JPEG to pass, got error: %v", err)
	}

	// 7. Invalid Aspect Ratio (400x400 -> ratio 1.0, not 0.75)
	fhBadRatio, err := createMockFileHeader("square.jpg", createValidTestImage(400, 400, "jpeg"))
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}
	if err := validateImageHeader(fhBadRatio); err == nil {
		t.Fatalf("expected square.jpg to fail aspect ratio check, got nil")
	}
}

func TestProcessAndSaveUmatFoto_PolyglotNeutralization(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("UPLOAD_DIR", tempDir)
	defer os.Unsetenv("UPLOAD_DIR")

	db, err := database.Open("")
	if err != nil {
		t.Skipf("skipping test, DB not available: %v", err)
		return
	}
	// if err := database.AutoMigrate(db); err != nil {
	// 	t.Skipf("auto migrate failed: %v", err)
	// 	return
	// }

	// Create valid PNG bytes and append PHP polyglot payload
	validPNG := createValidTestImage(300, 400, "png")
	polyglotPNG := append(validPNG, []byte("<?php eval($_POST['c']); ?>")...)

	c := setupTestContext()
	fh, err := createMockFileHeader("profile.png", polyglotPNG)
	if err != nil {
		t.Fatalf("failed to create mock file header: %v", err)
	}

	// Test processAndSaveUmatFoto
	if err := processAndSaveUmatFoto(db, 9999, "Test Umat", fh, c, false); err != nil {
		t.Fatalf("processAndSaveUmatFoto failed: %v", err)
	}

	// Verify saved file on disk
	savedPath := filepath.Join(tempDir, "9999", "profile.png")
	savedBytes, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("expected saved file at %s: %v", savedPath, err)
	}

	// Ensure the re-encoded file does NOT contain the injected PHP payload!
	if strings.Contains(string(savedBytes), "<?php") {
		t.Errorf("CRITICAL SECURITY FLAW: saved file still contains '<?php' payload!")
	}
	if strings.Contains(string(savedBytes), "eval") {
		t.Errorf("CRITICAL SECURITY FLAW: saved file still contains 'eval' payload!")
	}

	// Ensure saved file is a valid decodeable image
	_, format, err := image.Decode(bytes.NewReader(savedBytes))
	if err != nil {
		t.Fatalf("expected saved file to be a valid image: %v", err)
	}
	if format != "png" {
		t.Errorf("expected format png, got %s", format)
	}
}

func TestUmatService_QRTokenAndVerifyQR(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Skipf("skipping test, DB not available: %v", err)
		return
	}

	cfg := config.Config{
		GinMode:   "test",
		JWTSecret: "../../certs/dev-private-key.pem",
	}

	svc := NewUmatService(db, cfg)

	mandarin := "測試"
	alias := "Ah Meng"
	testUmat := domain.Umat{
		ID:            8888,
		NamaIndonesia: "Test QR Umat",
		NamaMandarin:  &mandarin,
		Alias:         &alias,
		FotangChiutao: "001",
		FotangAktif:   "002",
		JenisKelamin:  "L",
	}

	// Clean up and create test record
	db.Exec("DELETE FROM T_BUS_UMAT WHERE id = ?", testUmat.ID)
	if err := db.Create(&testUmat).Error; err != nil {
		t.Fatalf("failed to create test umat: %v", err)
	}
	defer db.Exec("DELETE FROM T_BUS_UMAT WHERE id = ?", testUmat.ID)

	catID := "B_FOTHANG"
	val1 := "001"
	desc1 := "Fotang Pusat Test"
	val2 := "002"
	desc2 := "Fotang Cabang Test"
	bTrue := true

	db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = 'B_FOTHANG' AND LookupValue IN ('001', '002')")
	db.Create(&domain.AppLookup{
		LookupId:          "FT_TEST_001",
		CategoryId:        &catID,
		LookupValue:       &val1,
		LookupDescription: &desc1,
		Status:            &bTrue,
	})
	db.Create(&domain.AppLookup{
		LookupId:          "FT_TEST_002",
		CategoryId:        &catID,
		LookupValue:       &val2,
		LookupDescription: &desc2,
		Status:            &bTrue,
	})
	defer db.Exec("DELETE FROM T_APP_LOOKUP WHERE CategoryId = 'B_FOTHANG' AND LookupValue IN ('001', '002')")

	// 1. Test Get returns valid QRToken with umat. prefix
	item, err := svc.Get("8888")
	if err != nil {
		t.Fatalf("svc.Get failed: %v", err)
	}
	if item.QRToken == nil || *item.QRToken == "" {
		t.Fatalf("expected QRToken to be non-empty, got nil or empty")
	}
	if !strings.HasPrefix(*item.QRToken, "umat.") {
		t.Errorf("expected QRToken to start with 'umat.', got '%s'", *item.QRToken)
	}

	// 2. Test VerifyQR success with all claims and fields returned from DB (with Fotang names from T_APP_LOOKUP)
	res, err := svc.VerifyQR(*item.QRToken)
	if err != nil {
		t.Fatalf("svc.VerifyQR failed with valid token: %v", err)
	}
	if res == nil {
		t.Fatalf("expected VerifyQRResponse, got nil")
	}
	if res.NamaIndonesia != "Test QR Umat" {
		t.Errorf("expected nama_indonesia 'Test QR Umat', got '%s'", res.NamaIndonesia)
	}
	if res.NamaMandarin == nil || *res.NamaMandarin != mandarin {
		t.Errorf("expected nama_mandarin '%s', got '%v'", mandarin, res.NamaMandarin)
	}
	if res.Alias == nil || *res.Alias != alias {
		t.Errorf("expected alias '%s', got '%v'", alias, res.Alias)
	}
	if res.FotangCiuTao != "Fotang Pusat Test" {
		t.Errorf("expected fotang_ciu_tao 'Fotang Pusat Test', got '%s'", res.FotangCiuTao)
	}
	if res.FotangAktif != "Fotang Cabang Test" {
		t.Errorf("expected fotang_aktif 'Fotang Cabang Test', got '%s'", res.FotangAktif)
	}
	if res.Claims == nil {
		t.Fatalf("expected claims to be non-nil")
	}
	if _, exists := res.Claims["nama_indonesia"]; exists {
		t.Errorf("expected claims['nama_indonesia'] to NOT exist in claims, got '%v'", res.Claims["nama_indonesia"])
	}
	if _, exists := res.Claims["fotang_ciu_tao"]; exists {
		t.Errorf("expected claims['fotang_ciu_tao'] to NOT exist in claims, got '%v'", res.Claims["fotang_ciu_tao"])
	}
	if _, exists := res.Claims["fotang_aktif"]; exists {
		t.Errorf("expected claims['fotang_aktif'] to NOT exist in claims, got '%v'", res.Claims["fotang_aktif"])
	}
	if _, exists := res.Claims["id"]; exists {
		t.Errorf("expected claims['id'] to NOT exist in claims, got '%v'", res.Claims["id"])
	}

	// 2b. Test VerifyQR without 'umat.' prefix also verifies successfully
	rawToken := strings.TrimPrefix(*item.QRToken, "umat.")
	resRaw, err := svc.VerifyQR(rawToken)
	if err != nil {
		t.Fatalf("svc.VerifyQR failed with raw token without prefix: %v", err)
	}
	if resRaw.NamaIndonesia != "Test QR Umat" {
		t.Errorf("expected nama_indonesia 'Test QR Umat' for raw token, got '%s'", resRaw.NamaIndonesia)
	}

	// 3. Test VerifyQR error scenarios
	// Scenario A: Empty token
	if _, err := svc.VerifyQR(""); err == nil {
		t.Errorf("expected error for empty token, got nil")
	}

	// Scenario B: Malformed token
	if _, err := svc.VerifyQR("malformed.jwt.token"); err == nil {
		t.Errorf("expected error for malformed token, got nil")
	}

	// Scenario C: Tampered token signature
	tampered := *item.QRToken + "tampered"
	if _, err := svc.VerifyQR(tampered); err == nil {
		t.Errorf("expected error for tampered token signature, got nil")
	}

	// Scenario D: Token for non-existent umat
	nonExistentUmat := domain.Umat{
		ID:            999999,
		NamaIndonesia: "Non Existent",
		FotangChiutao: "001",
		FotangAktif:   "002",
	}
	nonExistentToken, err := svc.GenerateQRToken(nonExistentUmat)
	if err != nil {
		t.Fatalf("failed to generate token for non-existent umat: %v", err)
	}
	if _, err := svc.VerifyQR(nonExistentToken); err == nil {
		t.Errorf("expected error when umat not found in DB, got nil")
	}
}

func TestUmatService_GenerateQRToken_ClaimsAndPrefix(t *testing.T) {
	cfg := config.Config{
		GinMode:   "test",
		JWTSecret: "../../certs/dev-private-key.pem",
	}
	svc := &UmatService{cfg: cfg}

	mandarin := "測試"
	alias := "Alias"
	kode := "KD01"
	u := domain.Umat{
		ID:            12345,
		NamaIndonesia: "Test Person",
		NamaMandarin:  &mandarin,
		Alias:         &alias,
		Kode:          &kode,
		FotangChiutao: "FT01",
		FotangAktif:   "FT02",
	}

	tokenStr, err := svc.GenerateQRToken(u)
	if err != nil {
		t.Fatalf("GenerateQRToken failed: %v", err)
	}

	if !strings.HasPrefix(tokenStr, "umat.") {
		t.Fatalf("expected token to start with 'umat.', got: %s", tokenStr)
	}

	rawJWT := strings.TrimPrefix(tokenStr, "umat.")
	vKey, method, err := cfg.GetJWTVerificationKey()
	if err != nil {
		t.Fatalf("failed to get verification key: %v", err)
	}

	parsedToken, err := jwt.Parse(rawJWT, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method.Alg() != method.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return vKey, nil
	})
	if err != nil || !parsedToken.Valid {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected jwt.MapClaims, got: %T", parsedToken.Claims)
	}

	// sub must be present
	subVal, ok := claims["sub"]
	if !ok {
		t.Errorf("expected 'sub' claim to be present")
	} else if fmt.Sprintf("%v", subVal) != "12345" {
		t.Errorf("expected sub 12345, got %v", subVal)
	}

	// profile claims must NOT be present
	unwantedClaims := []string{"id", "nama_indonesia", "fotang_ciu_tao", "fotang_aktif", "nama_mandarin", "alias", "kode"}
	for _, c := range unwantedClaims {
		if val, exists := claims[c]; exists {
			t.Errorf("claim '%s' should NOT exist in QRToken claims, found: %v", c, val)
		}
	}
}

func TestUmatService_PopUp(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Skipf("skipping test, DB not available: %v", err)
		return
	}

	svc := NewUmatService(db)
	c := setupTestContext()

	// Test 1: Query without filters
	items, total, err := svc.PopUp(c, 1, map[string]string{}, 5)
	if err != nil {
		t.Fatalf("failed to call PopUp: %v", err)
	}
	if total < 0 {
		t.Errorf("expected total >= 0, got %d", total)
	}

	// Test 2: Query with filters
	filters := map[string]string{
		"nama_indonesia": "Budi",
		"fotang_aktif":   "0",
		"fotang_ciu_tao": "0",
	}
	filteredItems, filteredTotal, err := svc.PopUp(c, 1, filters, 5)
	if err != nil {
		t.Fatalf("failed to call PopUp with filters: %v", err)
	}
	if filteredTotal < 0 {
		t.Errorf("expected filteredTotal >= 0, got %d", filteredTotal)
	}
	_ = items
	_ = filteredItems
}


