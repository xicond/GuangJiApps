package service

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"golang.org/x/image/draw"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// UmatService manages operations on congregation member (umat) data including CRUD, photo processing, QR tokens, SSRS reports, and OCR.
type UmatService struct {
	db                   *gorm.DB
	cfg                  config.Config
	resource             string
	reportServerURL      string
	reportServerUsername string
	reportServerPassword string
}

// NewUmatService initializes a new instance of UmatService with the provided database and optional configuration.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//   - cfgs: Optional Config instance. If omitted, loaded from environment.
//
// Returns:
//   - *UmatService: An initialized instance of UmatService.
func NewUmatService(db *gorm.DB, cfgs ...config.Config) *UmatService {
	if db == nil {
		db = database.MustOpen("")
	}
	var cfg config.Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	} else {
		cfg = config.Load()
	}
	return &UmatService{
		db:                   db,
		cfg:                  cfg,
		resource:             "umats",
		reportServerURL:      getReportServerURL(),
		reportServerUsername: getReportServerUsername(),
		reportServerPassword: getReportServerPassword(),
	}
}

// derefString safely dereferences a string pointer, returning an empty string if nil.
//
// Parameters:
//   - s: Pointer to a string (*string).
//
// Returns:
//   - string: The dereferenced string value or "" if s is nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// validateUmatLookups validates lookup fields in an umat record against active master lookups in parallel.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB).
//   - payload: Pointer to the Umat record to validate (*domain.Umat).
//
// Returns:
//   - error: ValidationError containing invalid field details, or nil if all are valid.
func validateUmatLookups(db *gorm.DB, payload *domain.Umat) error {
	type lookupCheck struct {
		fieldName  string
		categoryID string
		val        string
	}

	checks := [...]lookupCheck{
		{fieldName: "status_umat", categoryID: "B_STATUS", val: derefString(payload.StatusUmat)},
		{fieldName: "tim_kerja", categoryID: "B_TIMKERJA", val: derefString(payload.TimKerja)},
		{fieldName: "kelas_khusus", categoryID: "B_KELAS", val: derefString(payload.KelasKhusus)},
		{fieldName: "kelas_umum", categoryID: "B_KELASUMUM", val: derefString(payload.KelasUmum)},
		{fieldName: "tempat_sd2", categoryID: "B_FOTHANG", val: derefString(payload.TempatSd2)},
		{fieldName: "tempat_sd3", categoryID: "B_FOTHANG", val: derefString(payload.TempatSd3)},
		{fieldName: "fotang_aktif", categoryID: "B_FOTHANG", val: payload.FotangAktif},
		{fieldName: "fotang_chiutao", categoryID: "B_FOTHANG", val: payload.FotangChiutao},
		{fieldName: "tcs", categoryID: "B_TCS", val: payload.Tcs},
		{fieldName: "waktu_chiutao_mandarin", categoryID: "B_WAKTUCIUTAO", val: payload.WaktuChiutaoMandarin},
		{fieldName: "pendidikan", categoryID: "B_PENDIDIKAN", val: derefString(payload.Pendidikan)},
		{fieldName: "pekerjaan", categoryID: "B_PEKERJAAN", val: derefString(payload.Pekerjaan)},
	}

	activeChecks := make([]lookupCheck, 0, len(checks))
	for _, c := range checks {
		if strings.TrimSpace(c.val) != "" {
			activeChecks = append(activeChecks, c)
		}
	}

	if len(activeChecks) == 0 {
		return nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		details = make(map[string][]string)
	)

	for _, check := range activeChecks {
		c := check
		wg.Go(func() {
			var count int64
			if err := db.Session(&gorm.Session{}).Model(&domain.AppLookup{}).
				Where("CategoryId = ? AND (LookupValue = ? OR LookupId = ?) AND Status = ?", c.categoryID, c.val, c.val, true).
				Count(&count).Error; err != nil {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], fmt.Sprintf("gagal memvalidasi %s: %v", c.fieldName, err))
				mu.Unlock()
				return
			}
			if count == 0 {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], fmt.Sprintf("field %s nilai '%s' tidak valid", c.categoryID, c.val))
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// List retrieves a paginated list of umat records matching the given search and filter parameters.
//
// Parameters:
//   - c: Gin context carrying HTTP query and header parameters.
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter conditions.
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.Umat: Slice of umat records matching the query.
//   - int64: Total count of matching records.
//   - error: Error if database query fails.
func (s *UmatService) List(c *gin.Context, page int, filters map[string]string, limit int) ([]domain.Umat, int64, error) { //[]domain.Umat {
	var items []domain.Umat
	var total int64

	// Validasi parameter pagination
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var subWhVal int64
	userID := getUserID(c)
	row := s.db.Model(&domain.AdminMatrix{}).
		Where("LOGINID = ?", userID).
		Select("SUBWHID").
		Row()
	if row != nil {
		_ = row.Scan(&subWhVal)
	}

	subQueryFotang := s.db.Table("T_APP_LOOKUP").
		Where("T_APP_LOOKUP.CategoryId = 'B_FOTHANG'").
		Where("T_APP_LOOKUP.LookupValue = T_BUS_UMAT.fotangaktif").
		Where("T_APP_LOOKUP.Status = ?", true)

	subQueryJK := s.db.Table("T_APP_LOOKUP").
		Where("T_APP_LOOKUP.CategoryId = 'B_JENISKELAMIN'").
		Where("T_APP_LOOKUP.LookupValue = T_BUS_UMAT.jeniskelamin").
		Where("T_APP_LOOKUP.Status = ?", true)

	// Base query
	query := s.db.Table("T_BUS_UMAT").
		Where("fotangaktif = ?", subWhVal).
		Where("EXISTS (?)", subQueryFotang).
		Where("EXISTS (?)", subQueryJK).
		Where("status = ?", true).
		Where("namaindonesia <> ''").
		Where("namaindonesia is not null ").
		Order("namaindonesia ASC")

	// Dynamic optional search on fields (dengan whitelist kolom aman dari SQL Injection)
	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"alias":                {Column: "alias", IsLike: true},
		"lookup_description":   {Column: "namaindonesia", IsLike: true},
		"namaindonesia":        {Column: "namaindonesia", IsLike: true},
		"namamandarin":         {Column: "namamandarin", IsLike: true},
		"tahunchiutaomandarin": {Column: "tahunchiutaomandarin", IsLike: false},
		"fotang_aktif":         {Column: "fotangaktif", IsLike: false},
		// "fotangaktif":          {Column: "fotangaktif", IsLike: false},
		"fotang_chiutao": {Column: "fotangciutao", IsLike: false},
		// "fotangciutao":         {Column: "fotangciutao", IsLike: false},
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if rule, exists := allowedFilters[field]; exists {
			if field == "tahunchiutaomandarin" {
				query = query.Where("YEAR([TanggalChiuTaoInt]) = ?", value)
			} else if rule.IsLike {
				// String / Varchar menggunakan LIKE
				query = query.Where(fmt.Sprintf("[%s] LIKE ?", rule.Column), "%"+value+"%")
			} else {
				// Tahun atau numerik menggunakan exact match (=)
				query = query.Where(fmt.Sprintf("[%s] = ?", rule.Column), value)
			}
		}
	}

	// Execute Count and List queries concurrently for optimal latency
	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	// Goroutine 1: Concurrent Count query
	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	})

	// Goroutine 2: Concurrent Find items query
	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).
			Preload("JenisKelaminInfo", "CategoryId = ? AND Status = ?", "B_JENISKELAMIN", true).
			Limit(limit).
			Offset(offset).
			Order("ID ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("umat tidak ditemukan")
			} else {
				findErr = fmt.Errorf("database error: %w", err)
			}
		}
	})

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if findErr != nil {
		return []domain.Umat{}, 0, findErr
	}

	// Dynamic post-processing for lowest latency
	now := time.Now()
	for i := range items {
		if !items[i].TanggalLahir.IsZero() && items[i].TanggalLahir.Year() > 1900 {
			u := int32(now.Year() - items[i].TanggalLahir.Year())
			items[i].Usia = &u
		}
		if items[i].JenisKelaminInfo != nil && items[i].JenisKelaminInfo.LookupDescription != nil && *items[i].JenisKelaminInfo.LookupDescription != "" {
			items[i].JenisKelamin = *items[i].JenisKelaminInfo.LookupDescription
		}
	}

	return items, total, nil
}

// PopUp executes an optimized search query intended for umat search modal/dialog popups with minimal fields.
//
// Parameters:
//   - c: Gin context carrying HTTP query and header parameters.
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter conditions.
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.UmatPopUpResponse: Lightweight umat records tailored for popup selection.
//   - int64: Total count of matching records.
//   - error: Error if database query fails.
func (s *UmatService) PopUp(c *gin.Context, page int, filters map[string]string, limit int) ([]domain.UmatPopUpResponse, int64, error) {
	const defaultLimit = 10
	const maxLimit = 1000

	if limit <= 0 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}
	if page <= 0 {
		page = 1
	}

	namaIndo := getFilterOrDefault(filters, []string{"nama_indonesia"}, "")
	namaMandarin := getFilterOrDefault(filters, []string{"nama_mandarin"}, "")
	alias := getFilterOrDefault(filters, []string{"alias"}, "")
	fotangAktif := toInt(getFilterOrDefault(filters, []string{"fotang_aktif"}, "0"))
	fotangChiuTao := toInt(getFilterOrDefault(filters, []string{"fotang_ciu_tao"}, "0"))

	/* sortDirection := getFilterOrDefault(filters, []string{"sort_direction", "sortDirection", "order", "direction"}, "ASCENDING")
	if strings.ToUpper(sortDirection) == "DESC" || strings.ToUpper(sortDirection) == "DESCENDING" {
		sortDirection = "DESCENDING"
	} else {
		sortDirection = "ASCENDING"
	} */
	sortDirection := "ASCENDING"
	// sortExpression := getFilterOrDefault(filters, []string{"sort_expression", "sortExpression", "sort", "sortBy"}, "namaindonesia")

	rows, err := s.db.Raw("EXEC dbo.SP_BUS_UMAT_SEARCH_POPUP ?, ?, ?, ?, ?, ?, ?, ?, ?",
		limit,
		page,
		nil,
		sortDirection,
		namaIndo,
		namaMandarin,
		alias,
		fotangAktif,
		fotangChiuTao,
	).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, 0, fmt.Errorf("database columns error: %w", err)
	}

	items := make([]domain.UmatPopUpResponse, 0, limit)
	var total int64

	for rows.Next() {
		var item domain.UmatPopUpResponse
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow", "total_row":
				valuePtrs[i] = &nullFieldScanner{target: &totalRowScan}
			case "rowno", "row_no":
				var dummy int64
				valuePtrs[i] = &nullFieldScanner{target: &dummy}
			default:
				matched := false
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						fieldVal := v.Field(j)
						valuePtrs[i] = &nullFieldScanner{target: fieldVal.Addr().Interface()}
						matched = true
						break
					}
				}
				if !matched {
					var dummy interface{}
					valuePtrs[i] = &dummy
				}
			}
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("[ERROR] Umat PopUp rows.Scan error: %v", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	return items, total, nil
}

// sanitizeFilename normalizes and cleans a file name to prevent directory traversal and remove unsafe characters.
//
// Parameters:
//   - name: The raw uploaded filename as a string.
//
// Returns:
//   - string: A sanitized and safe filename.
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "image_" + time.Now().Format("20060102150405") + ".jpg"
	}

	// Filter dangerous characters:
	// - Windows invalid chars: <>:"/\|?*
	// - ASCII control chars: \x00-\x1f, \x7f
	// - Unicode control & direction override chars: \x{200b}-\x{200d}, \x{202a}-\x{202e}, \x{feff}
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f\x7f\x{200b}-\x{200d}\x{202a}-\x{202e}\x{feff}]`)
	clean := re.ReplaceAllString(name, "_")
	clean = strings.ReplaceAll(clean, " ", "_")

	ext := filepath.Ext(clean)
	base := strings.TrimSuffix(clean, ext)
	upperBase := strings.ToUpper(base)
	upperBaseClean := strings.TrimRight(upperBase, " ._")

	reservedNames := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
		"COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
		"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true, "CLOCK$": true,
	}
	if reservedNames[upperBase] || reservedNames[upperBaseClean] {
		clean = "_" + clean
	}

	// Ensure length does not exceed 150 bytes and is rune-safe (no partial UTF-8 sequences)
	maxBytes := 150
	if len(clean) > maxBytes {
		targetLen := maxBytes - len(ext)
		if targetLen < 1 {
			targetLen = maxBytes
			ext = ""
		}

		baseStr := strings.TrimSuffix(clean, ext)
		baseRunes := []rune(baseStr)

		// If string contains non-ASCII / multibyte (e.g. CJK), cap max runes to maxBytes / 2 (75 runes)
		hasMultiByte := false
		for _, r := range baseRunes {
			if r > 127 {
				hasMultiByte = true
				break
			}
		}

		maxRunes := targetLen
		if hasMultiByte && maxRunes > maxBytes/2 {
			maxRunes = maxBytes / 2
		}

		if len(baseRunes) > maxRunes {
			baseRunes = baseRunes[:maxRunes]
		}

		truncatedBase := string(baseRunes)
		for len(truncatedBase)+len(ext) > maxBytes && len(baseRunes) > 0 {
			baseRunes = baseRunes[:len(baseRunes)-1]
			truncatedBase = string(baseRunes)
		}
		clean = truncatedBase + ext
	}
	return clean
}

// validateImageHeader checks the MIME type, extension, size, and decodability of an uploaded image file header.
//
// Parameters:
//   - fileHeader: The uploaded multipart file header (*multipart.FileHeader).
//
// Returns:
//   - error: Error if the file fails size, MIME, or decoding validations; nil if valid or nil header.
func validateImageHeader(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return nil
	}

	var errorsList []string

	// 1. File size check (< 8MB, > 0)
	maxSize := int64(8 * 1024 * 1024)
	if fileHeader.Size <= 0 {
		errorsList = append(errorsList, "File gambar kosong")
	} else if fileHeader.Size > maxSize {
		errorsList = append(errorsList, "Ukuran file tidak boleh melebihi 8MB")
	}

	// 2. Strict file extension whitelisting & dangerous script extension detection
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		errorsList = append(errorsList, "Ekstensi file harus berupa .jpg, .jpeg, atau .png")
	}

	lowerFilename := strings.ToLower(fileHeader.Filename)
	dangerousExts := []string{
		".php", ".phtml", ".php3", ".php4", ".php5", ".phps", ".pht", ".phar",
		".asp", ".aspx", ".ashx", ".asmx", ".cer", ".asa",
		".jsp", ".jspx", ".cgi", ".pl", ".py", ".sh", ".bash",
		".exe", ".bat", ".cmd", ".com", ".vbs", ".vbe", ".js", ".jse", ".wsf", ".wsh",
		".htaccess", ".config", ".svg", ".html", ".htm", ".shtml",
	}
	for _, dExt := range dangerousExts {
		if strings.Contains(lowerFilename, dExt) {
			errorsList = append(errorsList, "Nama file terdeteksi mengandung ekstensi berbahaya atau tidak diizinkan")
			break
		}
	}

	// 3. Open file & verify MIME type + full image decode (anti-polyglot check)
	src, err := fileHeader.Open()
	if err != nil {
		errorsList = append(errorsList, "Gagal membuka file gambar")
	} else {
		defer src.Close()

		// Read first 512 bytes for MIME type check (strictly image/jpeg or image/png)
		buffer := make([]byte, 512)
		n, _ := src.Read(buffer)
		contentType := http.DetectContentType(buffer[:n])
		if contentType != "image/jpeg" && contentType != "image/png" {
			errorsList = append(errorsList, "Tipe MIME file harus berupa image/jpeg atau image/png")
			return NewValidationError(map[string][]string{
				"foto": errorsList,
			})
		}

		// Reset file pointer to beginning for full image decoding
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			errorsList = append(errorsList, "Gagal memproses file gambar")
		} else {
			// Fully decode image into memory to guarantee genuine pixel structure and reject polyglots
			img, format, decodeErr := image.Decode(src)
			if decodeErr != nil {
				errorsList = append(errorsList, "Format file gambar tidak dapat dibaca atau rusak (harus berupa JPEG atau PNG yang valid)")
			} else {
				if format != "jpeg" && format != "png" {
					errorsList = append(errorsList, "Format file gambar tidak didukung (harus berupa JPEG atau PNG)")
				} else {
					bounds := img.Bounds()
					width := bounds.Dx()
					height := bounds.Dy()

					// Resolution bounds: width 150..3000, height 200..4000
					if width < 150 || width > 3000 {
						errorsList = append(errorsList, fmt.Sprintf("Lebar gambar (%dpx) harus di antara 150 hingga 3000 piksel", width))
					}
					if height < 200 || height > 4000 {
						errorsList = append(errorsList, fmt.Sprintf("Tinggi gambar (%dpx) harus di antara 200 hingga 4000 piksel", height))
					}

					// Aspect ratio 3:4 (ratio = 0.75)
					ratio := float64(width) / float64(height)
					if math.Abs(ratio-0.75) > 0.03 {
						errorsList = append(errorsList, fmt.Sprintf("Rasio gambar (saat ini %.2f) harus 3:4 (lebar : tinggi)", ratio))
					}
				}
			}
		}
	}

	if len(errorsList) > 0 {
		return NewValidationError(map[string][]string{
			"foto": errorsList,
		})
	}
	return nil
}

// processAndSaveUmatFoto saves an uploaded photo to disk (with optional thumbnailing and watermarking) and persists metadata to the database.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB).
//   - umatID: ID of the umat whose photo is being saved.
//   - namaIndo: Indonesian name of the umat, used in generating file names.
//   - fileHeader: The uploaded multipart file header.
//   - c: Gin context carrying request metadata.
//   - isUpdate: Flag indicating if this is an update to an existing photo.
//
// Returns:
//   - error: Error if file writing or database persistence fails.
func processAndSaveUmatFoto(db *gorm.DB, umatID int32, namaIndo string, fileHeader *multipart.FileHeader, c *gin.Context, isUpdate bool) error {
	if fileHeader == nil {
		return nil
	}

	uploadBaseDir := os.Getenv("UPLOAD_DIR")
	if uploadBaseDir == "" {
		uploadBaseDir = "uploads"
	}

	docPathStr := fmt.Sprintf("%d", umatID)
	targetDir := filepath.Join(uploadBaseDir, docPathStr)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("gagal membuka file upload: %w", err)
	}
	defer src.Close()

	// 1. Decode image into memory to strip polyglots, script tags, EXIF payloads, or malformed chunks
	img, format, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("gagal membaca data gambar: %w", err)
	}

	// 2. Select canonical extension and encoder based strictly on verified decoded image format
	var (
		ext        string
		encodeFunc func(io.Writer) error
	)
	switch format {
	case "png":
		ext = ".png"
		encodeFunc = func(w io.Writer) error {
			return png.Encode(w, img)
		}
	case "jpeg":
		ext = ".jpg"
		encodeFunc = func(w io.Writer) error {
			return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
		}
	default:
		return fmt.Errorf("format gambar tidak didukung: %s", format)
	}

	// 3. Sanitize filename base name and strictly enforce canonical extension (never trust raw extension)
	rawBase := filepath.Base(fileHeader.Filename)
	if idx := strings.LastIndex(rawBase, "."); idx != -1 {
		rawBase = rawBase[:idx]
	}
	cleanBase := sanitizeFilename(rawBase)
	cleanBase = strings.Trim(cleanBase, " ._")
	cleanBase = strings.ReplaceAll(cleanBase, ".", "_")
	if cleanBase == "" {
		cleanBase = fmt.Sprintf("foto_%d_%s", umatID, time.Now().Format("20060102150405"))
	}
	safeFileName := cleanBase + ext
	dstPath := filepath.Join(targetDir, safeFileName)

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file di disk: %w", err)
	}
	defer out.Close()

	// 4. Re-encode image from raw pixel memory (NEVER use raw io.Copy) to neutralize any polyglot payload
	if err := encodeFunc(out); err != nil {
		_ = os.Remove(dstPath)
		return fmt.Errorf("gagal meng-encode dan menyimpan gambar: %w", err)
	}

	docFileStr := docPathStr + "/" + safeFileName
	userID := getUserID(c)
	statusTrue := true
	nowTime := time.Now()

	if isUpdate {
		var existing domain.UmatFoto
		if err := db.Where("Id = ?", umatID).Take(&existing).Error; err == nil {
			modAct := "U"
			existing.DocPath = &docPathStr
			existing.DocFile = &docFileStr
			existing.DocFileName = &safeFileName
			existing.DocDesc = &namaIndo
			existing.Status = &statusTrue
			existing.ModBy = &userID
			existing.ModAct = &modAct
			existing.ModDate = &nowTime
			return db.Save(&existing).Error
		}
	}

	// Create new record in UmatFoto
	var genResult struct {
		GeneratedId int32
	}
	nowStr := nowTime.Format("2006-01-02 15:04:05")
	if errId := db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "FILE", nowStr, 1).Scan(&genResult).Error; errId != nil {
		return fmt.Errorf("failed to generate foto FileID: %w", errId)
	}

	modAct := "I"
	foto := domain.UmatFoto{
		FileID:      genResult.GeneratedId,
		Id:          &umatID,
		DocPath:     &docPathStr,
		DocFile:     &docFileStr,
		DocFileName: &safeFileName,
		DocDesc:     &namaIndo,
		Status:      &statusTrue,
		ModBy:       &userID,
		ModAct:      &modAct,
		ModDate:     &nowTime,
	}

	return db.Create(&foto).Error
}

// Create validates and inserts a new umat member record, generating a unique ID and processing the optional profile photo.
//
// Parameters:
//   - payload: Umat record to create (domain.Umat).
//   - fileHeader: Optional uploaded photo header (*multipart.FileHeader).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.Umat: The newly created umat record.
//   - error: Error if validation fails or database insert fails.
func (s *UmatService) Create(payload domain.Umat, fileHeader *multipart.FileHeader, c *gin.Context) (domain.Umat, error) {
	if fileHeader != nil {
		if err := validateImageHeader(fileHeader); err != nil {
			return domain.Umat{}, err
		}
	}

	if payload.Email != nil && strings.TrimSpace(*payload.Email) == "" {
		payload.Email = nil
	}

	if err := ValidateStruct(payload); err != nil {
		return domain.Umat{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validateUmatLookups(s.db, &payload); err != nil {
		return domain.Umat{}, err
	}

	type SubWhInfo struct {
		SubWhId  string `gorm:"column:SUBWHID"`
		FullName string `gorm:"column:FULL_NAME"`
	}

	var (
		genResult struct {
			GeneratedId int32
		}
		subWhResult   SubWhInfo
		errId         error
		errSubWh      error
		errPenanggung error
		errPengajak   error
		wg            sync.WaitGroup
	)

	userID := getUserID(c)

	needPenanggung := (payload.Penanggung != nil && string(*payload.Penanggung) != "" && string(*payload.Penanggung) != "0") && (payload.PenanggungManual == nil || *payload.PenanggungManual == "")
	needPengajak := (payload.Pengajak != nil && string(*payload.Pengajak) != "" && string(*payload.Pengajak) != "0") && (payload.PengajakManual == nil || *payload.PengajakManual == "")

	// Goroutine 1: Execute SP_APP_GenerateId concurrently
	wg.Go(func() {
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		errId = s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "UMAT", nowStr, 1).Scan(&genResult).Error
	})

	// Goroutine 2: Execute SubWhInfo query concurrently
	wg.Go(func() {
		subQuery := s.db.Table("T_WH_USER_MATRIX_MST AS A").
			Select("1").
			Where("A.LOGINID = ?", userID).
			Where("A.SUBWHID = T_WH_SUBWH_MST.SUBWHID")

		errSubWh = s.db.Table("T_WH_SUBWH_MST").
			Select("SUBWHID, FULL_NAME").
			Where("EXISTS (?)", subQuery).
			Limit(1).
			Scan(&subWhResult).Error
	})

	// Goroutine 3: Lookup Penanggung concurrently if needed
	if needPenanggung {
		wg.Go(func() {
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Penanggung).Limit(1).Find(&temp).Select("namaindonesia, namamandarin, alias").Error; err != nil {
				errPenanggung = fmt.Errorf("Penanggung %s not found: %w", *payload.Penanggung, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PenanggungManual = &pn
			}
		})
	}

	// Goroutine 4: Lookup Pengajak concurrently if needed
	if needPengajak {
		wg.Go(func() {
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Pengajak).Limit(1).Find(&temp).Select("namaindonesia, namamandarin, alias").Error; err != nil {
				errPengajak = fmt.Errorf("Pengajak %s not found: %w", *payload.Pengajak, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PengajakManual = &pn
			}
		})
	}

	wg.Wait()

	if errId != nil {
		return domain.Umat{}, fmt.Errorf("failed to generate id: %w", errId)
	}
	if errSubWh != nil {
		log.Printf("[UmatService.Create] warning querying sub warehouse info for user %d: %v", userID, errSubWh)
	}
	if errPenanggung != nil {
		return domain.Umat{}, errPenanggung
	}
	if errPengajak != nil {
		return domain.Umat{}, errPengajak
	}

	// payload.FotangAktif = subWhResult.SubWhId
	payload.ID = genResult.GeneratedId
	kode := fmt.Sprintf("%s-%d", subWhResult.FullName, genResult.GeneratedId)
	payload.Kode = &kode

	// 3. Populate matching schema structural constraints
	payload.Status = true                  // Active status mapping
	payload.ModAct = "I"                   // 'I' standard legacy flag for Insert
	payload.ModBy = getUserID(c)           // Default system user ID matching INT type
	payload.ModDate = domain.NowDateTime() // Local server time object

	// 4. Persist the new entity to the database pool
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to create record: %w", err)
	}

	if fileHeader != nil {
		if err := processAndSaveUmatFoto(s.db, payload.ID, payload.NamaIndonesia, fileHeader, c, false); err != nil {
			log.Printf("[UmatService.Create] Save UmatFoto error: %v", err)
		}
	}

	return payload, nil
}

// Get fetches a single umat member by ID, calculates age dynamically, and generates a signed QR token.
//
// Parameters:
//   - id: The primary key of the umat member as a string.
//
// Returns:
//   - domain.Umat: The retrieved umat member record.
//   - error: Error if the member is not found or database query fails.
func (s *UmatService) Get(id string) (domain.Umat, error) {
	var item domain.Umat
	if err := s.db.Preload("JenisKelaminInfo", "CategoryId = ? AND Status = ?", "B_JENISKELAMIN", true).First(&item, "id = ?", id).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("umat %s not found", id)
	}
	if !item.TanggalLahir.IsZero() && item.TanggalLahir.Year() > 1900 {
		u := int32(time.Now().Year() - item.TanggalLahir.Year())
		item.Usia = &u
	}

	qrToken, err := s.GenerateQRToken(item)
	if err == nil {
		item.QRToken = &qrToken
	} else {
		log.Printf("[UmatService.Get] GenerateQRToken error: %v", err)
	}

	// Dont activate this, get from JenisKelaminInfo
	/* if item.JenisKelaminInfo != nil && item.JenisKelaminInfo.LookupDescription != nil && *item.JenisKelaminInfo.LookupDescription != "" {
		item.JenisKelamin = *item.JenisKelaminInfo.LookupDescription
	} */
	return item, nil
}

// GenerateQRToken produces a cryptographically signed JWT token prefixed with "umat." for QR attendance.
//
// Parameters:
//   - item: The Umat record for which to issue the token (domain.Umat).
//
// Returns:
//   - string: The formatted QR token string.
//   - error: Error if key retrieval or signing fails.
func (s *UmatService) GenerateQRToken(item domain.Umat) (string, error) {
	key, method, err := s.cfg.GetJWTSigningKey()
	if err != nil {
		return "", fmt.Errorf("failed to get JWT signing key: %w", err)
	}

	claims := jwt.MapClaims{
		"sub": item.ID,
	}

	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return "umat." + signed, nil
}

// VerifyQR parses and cryptographically validates a QR token, returning basic member details.
//
// Parameters:
//   - tokenString: The raw QR token string to verify.
//
// Returns:
//   - *domain.VerifyQRResponse: Decoded umat identification response.
//   - error: Error if the token is empty, invalid, or expired.
func (s *UmatService) VerifyQR(tokenString string) (*domain.VerifyQRResponse, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("qr_token is required")
	}

	tokenString = strings.TrimPrefix(tokenString, "umat.")
	tokenString = strings.TrimSpace(tokenString)

	vKey, method, err := s.cfg.GetJWTVerificationKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT verification key: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != method.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return vKey, nil
	}, jwt.WithValidMethods([]string{
		jwt.SigningMethodRS256.Alg(),
		jwt.SigningMethodES256.Alg(),
	}))

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired qr_token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	var umatIDInt int32
	parseID := func(val interface{}) int32 {
		if val == nil {
			return 0
		}
		switch v := val.(type) {
		case float64:
			return int32(v)
		case float32:
			return int32(v)
		case int:
			return int32(v)
		case int32:
			return v
		case int64:
			return int32(v)
		case json.Number:
			if n, err := v.Int64(); err == nil {
				return int32(n)
			}
		case string:
			if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32); err == nil {
				return int32(n)
			}
		}
		return 0
	}

	if sub, ok := claims["sub"]; ok {
		umatIDInt = parseID(sub)
	}
	if umatIDInt == 0 {
		if idVal, ok := claims["id"]; ok {
			umatIDInt = parseID(idVal)
		}
	}

	if umatIDInt == 0 {
		return nil, errors.New("qr_token does not contain valid umat id")
	}

	var umat domain.Umat
	if err := s.db.First(&umat, "id = ?", umatIDInt).Error; err != nil {
		return nil, fmt.Errorf("umat not found: %w", err)
	}

	fotangCiuTaoName := umat.FotangChiutao
	fotangAktifName := umat.FotangAktif

	fotangCodes := make([]string, 0, 2)
	if strings.TrimSpace(umat.FotangChiutao) != "" {
		fotangCodes = append(fotangCodes, strings.TrimSpace(umat.FotangChiutao))
	}
	if strings.TrimSpace(umat.FotangAktif) != "" && strings.TrimSpace(umat.FotangAktif) != strings.TrimSpace(umat.FotangChiutao) {
		fotangCodes = append(fotangCodes, strings.TrimSpace(umat.FotangAktif))
	}

	if len(fotangCodes) > 0 {
		var lookups []domain.AppLookup
		if err := s.db.Where("CategoryId = ? AND (LookupValue IN (?) OR LookupId IN (?))", "B_FOTHANG", fotangCodes, fotangCodes).
			Order("Status DESC").
			Find(&lookups).Error; err == nil {
			fotangMap := make(map[string]string)
			for _, l := range lookups {
				name := ""
				if l.LookupDescription != nil && strings.TrimSpace(*l.LookupDescription) != "" {
					name = strings.TrimSpace(*l.LookupDescription)
				} else if l.LookupValue != nil && strings.TrimSpace(*l.LookupValue) != "" {
					name = strings.TrimSpace(*l.LookupValue)
				}
				if name != "" {
					if l.LookupValue != nil && *l.LookupValue != "" {
						if _, exists := fotangMap[strings.TrimSpace(*l.LookupValue)]; !exists {
							fotangMap[strings.TrimSpace(*l.LookupValue)] = name
						}
					}
					if l.LookupId != "" {
						if _, exists := fotangMap[strings.TrimSpace(l.LookupId)]; !exists {
							fotangMap[strings.TrimSpace(l.LookupId)] = name
						}
					}
				}
			}
			if name, ok := fotangMap[strings.TrimSpace(umat.FotangChiutao)]; ok && name != "" {
				fotangCiuTaoName = name
			}
			if name, ok := fotangMap[strings.TrimSpace(umat.FotangAktif)]; ok && name != "" {
				fotangAktifName = name
			}
		}
	}

	res := &domain.VerifyQRResponse{
		Claims:        claims,
		NamaIndonesia: umat.NamaIndonesia,
		NamaMandarin:  umat.NamaMandarin,
		Alias:         umat.Alias,
		FotangCiuTao:  fotangCiuTaoName,
		FotangAktif:   fotangAktifName,
	}

	return res, nil
}

// Update validates and modifies an existing umat member record and handles photo replacement if a new image is provided.
//
// Parameters:
//   - id: The primary key of the umat member to update as a string.
//   - payload: Updated umat member data (domain.Umat).
//   - fileHeader: Optional new photo header (*multipart.FileHeader).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.Umat: The updated umat record.
//   - error: Error if validation fails, record is not found, or database update fails.
func (s *UmatService) Update(id string, payload domain.Umat, fileHeader *multipart.FileHeader, c *gin.Context) (domain.Umat, error) {
	if fileHeader != nil {
		if err := validateImageHeader(fileHeader); err != nil {
			return domain.Umat{}, err
		}
	}

	if payload.Email != nil && strings.TrimSpace(*payload.Email) == "" {
		payload.Email = nil
	}

	if err := ValidateStruct(payload); err != nil {
		return domain.Umat{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validateUmatLookups(s.db, &payload); err != nil {
		return domain.Umat{}, err
	}

	// 1. Parse string ID directly as 32-bit signed integer to avoid narrowing overflow
	parsedInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return domain.Umat{}, fmt.Errorf("invalid ID format: %w", err)
	}
	userIDInt32 := int32(parsedInt)

	var item domain.Umat
	// Fetch existing item using .Take() to avoid default sorting bugs
	if err := s.db.Where("id = ?", userIDInt32).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Umat{}, fmt.Errorf("umat %s not found", id)
		}
		return domain.Umat{}, err
	}

	// 3. Map values onto the actual field variables present in your legacy schema
	// item.Kode = payload.Kode
	item.Alias = payload.Alias
	item.NamaIndonesia = payload.NamaIndonesia
	item.Marga = payload.Marga
	item.NamaMandarin = payload.NamaMandarin
	item.Alamat = payload.Alamat
	item.Alamat2 = payload.Alamat2
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile
	item.TempatLahir = payload.TempatLahir
	item.TanggalLahir = payload.TanggalLahir
	item.Usia = payload.Usia
	item.Wilayah = payload.Wilayah
	item.JenisKelamin = payload.JenisKelamin
	item.Pekerjaan = payload.Pekerjaan
	item.Pendidikan = payload.Pendidikan
	item.TanggalChiutaoInt = payload.TanggalChiutaoInt
	item.TanggalChiutaoMan = payload.TanggalChiutaoMan
	item.TahunChiutaoMandarin = payload.TahunChiutaoMandarin
	item.WaktuChiutaoMandarin = payload.WaktuChiutaoMandarin
	item.Pengajak = payload.Pengajak
	item.Penanggung = payload.Penanggung
	item.Tcs = payload.Tcs
	item.UangPahala = payload.UangPahala
	item.FotangChiutao = payload.FotangChiutao
	item.FotangAktif = payload.FotangAktif
	item.Sd2 = payload.Sd2
	item.TempatSd2 = payload.TempatSd2
	item.TanggalSd2 = payload.TanggalSd2
	item.Sd3 = payload.Sd3
	item.TempatSd3 = payload.TempatSd3
	item.TanggalSd3 = payload.TanggalSd3
	item.KelasUmum = payload.KelasUmum
	item.KelasKhusus = payload.KelasKhusus
	item.ChingKhou = payload.ChingKhou
	item.TanggalChingKhou = payload.TanggalChingKhou
	item.TanggalAncuo = payload.TanggalAncuo
	item.NamaCetyaRumah = payload.NamaCetyaRumah
	item.Meninggal = payload.Meninggal
	item.TanggalMeninggal = payload.TanggalMeninggal
	item.TimKerja = payload.TimKerja
	item.Posisi = payload.Posisi
	item.StatusUmat = payload.StatusUmat
	item.Keterangan = payload.Keterangan
	item.Email = payload.Email
	item.ImagePath = payload.ImagePath

	var (
		errPenanggung error
		errPengajak   error
		wg            sync.WaitGroup
	)

	needPenanggung := (payload.Penanggung != nil && string(*payload.Penanggung) != "") && (payload.PenanggungManual == nil || *payload.PenanggungManual == "")
	needPengajak := (payload.Pengajak != nil && string(*payload.Pengajak) != "") && (payload.PengajakManual == nil || *payload.PengajakManual == "")

	// Goroutine: Lookup Penanggung concurrently if needed
	if needPenanggung {
		wg.Go(func() {
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Penanggung).Limit(1).Select("namaindonesia, namamandarin, alias").Find(&temp).Error; err != nil {
				errPenanggung = fmt.Errorf("Penanggung %s not found: %w", *payload.Penanggung, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PenanggungManual = &pn
			}
		})
	}

	// Goroutine: Lookup Pengajak concurrently if needed
	if needPengajak {
		wg.Go(func() {
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Pengajak).Limit(1).Select("namaindonesia, namamandarin, alias").Find(&temp).Error; err != nil {
				errPengajak = fmt.Errorf("Pengajak %s not found: %w", *payload.Pengajak, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PengajakManual = &pn
			}
		})
	}

	wg.Wait()

	if errPenanggung != nil {
		return domain.Umat{}, errPenanggung
	}
	if errPengajak != nil {
		return domain.Umat{}, errPengajak
	}

	item.PenanggungManual = payload.PenanggungManual
	item.PengajakManual = payload.PengajakManual

	// Legacy metadata mappings
	item.ModAct = "U"                   // 'U' standard legacy flag for Update
	item.ModBy = getUserID(c)           // System user ID (int32)
	item.ModDate = domain.NowDateTime() // Actual DateTime object expected by DATETIME column

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to update record: %w", err)
	}

	if fileHeader != nil {
		if err := processAndSaveUmatFoto(s.db, item.ID, item.NamaIndonesia, fileHeader, c, true); err != nil {
			log.Printf("[UmatService.Update] Save UmatFoto error: %v", err)
		}
	}

	return item, nil
}

// Delete deactivates an umat member record (soft delete via Status = false and ModAct = 'D').
//
// Parameters:
//   - id: The primary key of the umat member to delete as a string.
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - error: Error if the record is not found or database update fails.
func (s *UmatService) Delete(id string, c *gin.Context) error {
	// 1. Parse ID directly as 32-bit signed integer to avoid narrowing overflow
	parsedInt64, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}
	userIDInt32 := int32(parsedInt64)

	var item domain.Umat
	// 2. Fetch the existing item using .Take() to avoid default sorting bugs
	if err := s.db.Where("id = ?", userIDInt32).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("umat %s not found", id)
		}
		return err
	}

	// Legacy metadata mappings
	item.Status = false
	item.ModAct = "D"                   // 'D' standard legacy flag for Delete
	item.ModBy = getUserID(c)           // Default system user ID matching INT type
	item.ModDate = domain.NowDateTime() // Local server time object

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}

// getFilterOrDefault extracts the first non-empty value for the specified candidate keys from filters map, or returns defaultVal.
//
// Parameters:
//   - filters: Key-value map of filter parameters.
//   - keys: Slice of key names to check in order of priority.
//   - defaultVal: Fallback value if none of the keys exist or are empty.
//
// Returns:
//   - string: The extracted filter value or defaultVal.
func getFilterOrDefault(filters map[string]string, keys []string, defaultVal string) string {
	for _, key := range keys {
		if val, exists := filters[key]; exists && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return defaultVal
}

// nullFieldScanner wraps a target pointer to safely scan database values that may be null or type-mismatched.
type nullFieldScanner struct {
	target interface{}
}

// Scan converts and assigns database raw values into the underlying target pointer.
//
// Parameters:
//   - value: Raw database value returned by the driver.
//
// Returns:
//   - error: Error if assignment to the target pointer fails.
func (s *nullFieldScanner) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	targetVal := reflect.ValueOf(s.target)
	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	elem := targetVal.Elem()

	for elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			if !elem.CanSet() {
				return nil
			}
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		elem = elem.Elem()
	}

	if elem.CanAddr() {
		if scanner, ok := elem.Addr().Interface().(sql.Scanner); ok {
			return scanner.Scan(value)
		}
	}
	if scanner, ok := elem.Interface().(sql.Scanner); ok {
		return scanner.Scan(value)
	}

	switch v := value.(type) {
	case string:
		vTrim := strings.TrimSpace(v)
		if elem.Kind() == reflect.String {
			elem.SetString(v)
			return nil
		}
		if elem.Kind() == reflect.Bool {
			vLower := strings.ToLower(vTrim)
			elem.SetBool(vLower == "1" || vLower == "true" || vLower == "y" || vLower == "ya")
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			if parsed, err := strconv.ParseInt(vTrim, 10, 64); err == nil {
				elem.SetInt(parsed)
			}
			return nil
		}
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			if parsed, err := strconv.ParseFloat(vTrim, 64); err == nil {
				elem.SetFloat(parsed)
			}
			return nil
		}
	case []byte:
		vStr := string(v)
		vTrim := strings.TrimSpace(vStr)
		if elem.Kind() == reflect.String {
			elem.SetString(vStr)
			return nil
		}
		if elem.Kind() == reflect.Bool {
			vLower := strings.ToLower(vTrim)
			elem.SetBool(vLower == "1" || vLower == "true" || vLower == "y" || vLower == "ya")
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			if parsed, err := strconv.ParseInt(vTrim, 10, 64); err == nil {
				elem.SetInt(parsed)
			}
			return nil
		}
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			if parsed, err := strconv.ParseFloat(vTrim, 64); err == nil {
				elem.SetFloat(parsed)
			}
			return nil
		}
	case int64:
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			elem.SetInt(v)
			return nil
		}
		if elem.Kind() == reflect.Bool {
			elem.SetBool(v != 0)
			return nil
		}
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			elem.SetFloat(float64(v))
			return nil
		}
		if elem.Kind() == reflect.String {
			elem.SetString(strconv.FormatInt(v, 10))
			return nil
		}
	case int32:
		return s.Scan(int64(v))
	case int:
		return s.Scan(int64(v))
	case bool:
		if elem.Kind() == reflect.Bool {
			elem.SetBool(v)
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			if v {
				elem.SetInt(1)
			} else {
				elem.SetInt(0)
			}
			return nil
		}
		if elem.Kind() == reflect.String {
			if v {
				elem.SetString("true")
			} else {
				elem.SetString("false")
			}
			return nil
		}
	case float64:
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			elem.SetFloat(v)
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			elem.SetInt(int64(v))
			return nil
		}
		if elem.Kind() == reflect.String {
			elem.SetString(fmt.Sprintf("%v", v))
			return nil
		}
	case time.Time:
		if elem.Kind() == reflect.String {
			elem.SetString(v.Format("2006-01-02"))
			return nil
		}
	}

	return nil
}

// Report executes the stored procedure SP_BUS_RPT_UMAT with filter criteria and returns paginated report items.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of report filter parameters.
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.UmatReport: Slice of report rows.
//   - int64: Total count of records matching the report filters.
//   - error: Error if stored procedure execution fails.
func (s *UmatService) Report(page int, filters map[string]string, limit int) ([]domain.UmatReport, int64, error) {
	if page <= 0 {
		page = 1
	}

	fotangAktif := getFilterOrDefault(filters, []string{"fotang_aktif", "FotangAktif", "fotangAktif"}, "0")
	fotangChiuTao := getFilterOrDefault(filters, []string{"fotang_chiutao", "fotang_chiu_tao", "FotangChiuTao", "fotangChiuTao"}, "0")
	namaMandarin := getFilterOrDefault(filters, []string{"nama_mandarin", "NamaMandarin", "namaMandarin"}, "")
	startDate := getFilterOrDefault(filters, []string{"start_date", "startDate", "StartDate"}, "1970-01-01")
	endDate := getFilterOrDefault(filters, []string{"end_date", "endDate", "EndDate"}, "1970-01-01")
	namaIndo := getFilterOrDefault(filters, []string{"nama_indo", "nama_indonesia", "NamaIndo", "namaIndo"}, "")
	pengajak := getFilterOrDefault(filters, []string{"pengajak", "Pengajak"}, "")
	alias := getFilterOrDefault(filters, []string{"alias", "Alias"}, "")
	usiaDari := toInt(getFilterOrDefault(filters, []string{"usia_dari", "UsiaDari", "usiaDari"}, "0"))
	usiaSampai := toInt(getFilterOrDefault(filters, []string{"usia_sampai", "UsiaSampai", "usiaSampai"}, "0"))
	isLulusSd := getFilterOrDefault(filters, []string{"is_lulus_sd", "IsLulusSd", "isLulusSd"}, "0")
	isVege := getFilterOrDefault(filters, []string{"is_vege", "IsVege", "isVege"}, "0")
	statusUmat := getFilterOrDefault(filters, []string{"status_umat", "StatusUmat", "statusUmat"}, "0")

	var findErr error

	rows, err := s.db.Raw("EXEC SP_BUS_RPT_UMAT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?",
		fotangAktif,
		fotangChiuTao,
		namaMandarin,
		startDate,
		endDate,
		namaIndo,
		pengajak,
		alias,
		usiaDari,
		usiaSampai,
		isLulusSd,
		isVege,
		statusUmat,
		limit, // @PageSize
		page,  // @CurrentPage
	).Rows()

	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	// log.Printf("[SQL] %s | Query:\n%s\n", timestamp, s.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
	// 	return tx.Raw("EXEC SP_BUS_RPT_UMAT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?",
	// 		fotangAktif,
	// 		fotangChiuTao,
	// 		namaMandarin,
	// 		startDate,
	// 		endDate,
	// 		namaIndo,
	// 		pengajak,
	// 		alias,
	// 		usiaDari,
	// 		usiaSampai,
	// 		isLulusSd,
	// 		isVege,
	// 		statusUmat,
	// 		limit, // @PageSize
	// 		page,  // @CurrentPage
	// 	)
	// }))

	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("umat report tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, findErr
	}

	var items []domain.UmatReport
	var total int64

	for rows.Next() {
		var item domain.UmatReport
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow":
				valuePtrs[i] = &nullFieldScanner{target: &totalRowScan}
			default:
				matched := false
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						fieldVal := v.Field(j)
						valuePtrs[i] = &nullFieldScanner{target: fieldVal.Addr().Interface()}
						matched = true
						break
					}
				}
				if !matched {
					var dummy interface{}
					valuePtrs[i] = &dummy
				}
			}
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("[ERROR] Umat Report rows.Scan error: %v", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	return items, total, nil
}

// ReportExcel generates and proxies or exports an Excel spreadsheet from the SSRS report server based on filter parameters.
//
// Parameters:
//   - filters: Key-value map of report filter criteria.
//   - c: Gin context used to stream file responses.
//
// Returns:
//   - error: Error if report generation or streaming fails.
func (s *UmatService) ReportExcel(filters map[string]string, c *gin.Context) error {
	baseURL := s.reportServerURL
	if baseURL == "" {
		baseURL = getReportServerURL()
	}
	baseURL = strings.TrimRight(baseURL, "/")

	username := s.reportServerUsername
	if username == "" {
		username = getReportServerUsername()
	}
	password := s.reportServerPassword
	if password == "" {
		password = getReportServerPassword()
	}

	if username != "" {
		parsedURL, err := url.Parse(baseURL)
		if err == nil {
			parsedURL.User = url.UserPassword(username, password)
			baseURL = parsedURL.String()
		} else {
			schemeParts := strings.SplitN(baseURL, "://", 2)
			if len(schemeParts) == 2 {
				baseURL = fmt.Sprintf("%s://%s", schemeParts[0], schemeParts[1])
			}
		}
	}

	fotangAktif := getFilterOrDefault(filters, []string{"fotang_aktif", "FotangAktif", "fotangAktif"}, "0")
	fotangChiuTao := getFilterOrDefault(filters, []string{"fotang_chiutao", "fotang_chiu_tao", "FotangChiuTao", "fotangChiuTao"}, "0")
	namaMandarin := getFilterOrDefault(filters, []string{"nama_mandarin", "NamaMandarin", "namaMandarin"}, "")
	startDate := getFilterOrDefault(filters, []string{"start_date", "startDate", "StartDate"}, "1970-01-01")
	endDate := getFilterOrDefault(filters, []string{"end_date", "endDate", "EndDate"}, "1970-01-01")
	namaIndo := getFilterOrDefault(filters, []string{"nama_indo", "nama_indonesia", "NamaIndo", "namaIndo"}, "")
	pengajak := getFilterOrDefault(filters, []string{"pengajak", "Pengajak"}, "")
	alias := getFilterOrDefault(filters, []string{"alias", "Alias"}, "")
	usiaDari := getFilterOrDefault(filters, []string{"usia_dari", "UsiaDari", "usiaDari"}, "0")
	usiaSampai := getFilterOrDefault(filters, []string{"usia_sampai", "UsiaSampai", "usiaSampai"}, "0")
	isLulusSd := getFilterOrDefault(filters, []string{"is_lulus_sd", "IsLulusSd", "isLulusSd"}, "0")
	isVege := getFilterOrDefault(filters, []string{"is_vege", "IsVege", "isVege"}, "0")
	statusUmat := getFilterOrDefault(filters, []string{"status_umat", "StatusUmat", "statusUmat"}, "0")

	reportURL := fmt.Sprintf("%s/ReportServer?%%2fGuangJiReport%%2frpt_bus_umat_list&FotangAktif=%s&FotangChiuTao=%s&NamaMandarin=%s&StartDate=%s&EndDate=%s&NamaIndo=%s&Pengajak=%s&Alias=%s&UsiaDari=%s&UsiaSampai=%s&IsLulusSd=%s&IsVege=%s&StatusUmat=%s&rs:Command=Render&rs:Format=EXCELOPENXML",
		baseURL,
		url.QueryEscape(fotangAktif),
		url.QueryEscape(fotangChiuTao),
		url.QueryEscape(namaMandarin),
		url.QueryEscape(startDate),
		url.QueryEscape(endDate),
		url.QueryEscape(namaIndo),
		url.QueryEscape(pengajak),
		url.QueryEscape(alias),
		url.QueryEscape(usiaDari),
		url.QueryEscape(usiaSampai),
		url.QueryEscape(isLulusSd),
		url.QueryEscape(isVege),
		url.QueryEscape(statusUmat),
	)

	var reqCtx = c.Request.Context()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, reportURL, nil)
	if err != nil {
		return fmt.Errorf("gagal membuat request report: %w", err)
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := defaultReportClient

	startTime := time.Now()
	resp, err := client.Do(req)
	fetchDuration := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("gagal mengambil report dari SSRS (%v): %w", fetchDuration, err)
	}

	if resp.StatusCode == http.StatusUnauthorized && username != "" {
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.HasPrefix(authHeader, "Digest ") {
			resp.Body.Close()

			req2, err := http.NewRequestWithContext(reqCtx, http.MethodGet, reportURL, nil)
			if err != nil {
				return fmt.Errorf("gagal membuat digest request: %w", err)
			}

			digestVal := formatDigestAuth(authHeader, username, password, http.MethodGet, req2.URL.RequestURI())
			req2.Header.Set("Authorization", digestVal)

			resp, err = client.Do(req2)
			if err != nil {
				return fmt.Errorf("gagal mengambil report dengan Digest auth: %w", err)
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("report server mengembalikan HTTP %d (fetch %v): %s", resp.StatusCode, fetchDuration, reportURL)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition == "" {
		hash := sha256.Sum256([]byte(fotangAktif + fotangChiuTao + namaMandarin + startDate + endDate + namaIndo + statusUmat))
		filename := fmt.Sprintf("rpt_bus_umat_list_%x.xlsx", hash[:4])
		contentDisposition = fmt.Sprintf("attachment; filename=\"%s\"", filename)
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", contentDisposition)
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	c.Status(http.StatusOK)

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	bufPtr := copyBufferPool.Get().(*[]byte)
	defer copyBufferPool.Put(bufPtr)

	// streamStart := time.Now()
	_, err = io.CopyBuffer(c.Writer, resp.Body, *bufPtr)
	// streamDuration := time.Since(streamStart)
	// totalDuration := time.Since(startTime)

	// log.Printf("[REPORT PERF] SSRS Fetch: %v, Client Stream (%d bytes): %v, Total: %v",
	// 	fetchDuration, nBytes, streamDuration, totalDuration)

	if err != nil {
		return fmt.Errorf("gagal stream report ke client: %w", err)
	}

	return nil
}

var ocrBufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 512*1024))
	},
}

// OCRSpaceResult models the structured JSON response returned from the OCR.Space API.
type OCRSpaceResult struct {
	ParsedResults []struct {
		TextOverlay struct {
			Lines []struct {
				Words []struct {
					WordText string  `json:"WordText"`
					Left     float64 `json:"Left"`
					Top      float64 `json:"Top"`
					Width    float64 `json:"Width"`
					Height   float64 `json:"Height"`
				} `json:"Words"`
				MaxHeight float64 `json:"MaxHeight"`
				MinTop    float64 `json:"MinTop"`
			} `json:"Lines"`
			HasOverlay bool   `json:"HasOverlay"`
			Message    string `json:"Message"`
		} `json:"TextOverlay"`
		TextOrientation   string `json:"TextOrientation"`
		FileParseExitCode int    `json:"FileParseExitCode"`
		ParsedText        string `json:"ParsedText"`
		ErrorMessage      string `json:"ErrorMessage"`
		ErrorDetails      string `json:"ErrorDetails"`
	} `json:"ParsedResults"`
	OCRExitCode                  int      `json:"OCRExitCode"`
	IsErroredOnProcessing        bool     `json:"IsErroredOnProcessing"`
	ProcessingTimeInMilliseconds string   `json:"ProcessingTimeInMilliseconds"`
	SearchablePDFURL             string   `json:"SearchablePDFURL"`
	ErrorMessage                 []string `json:"ErrorMessage"`
	ErrorDetails                 []string `json:"ErrorDetails"`
}

// Ocr uploads an image of a registration form to the OCR service, extracts text, and parses umat fields.
//
// Parameters:
//   - file: The opened multipart file reader (multipart.File).
//   - header: The uploaded file metadata header (*multipart.FileHeader).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - interface{}: OCRResponsePayload containing parsed fields and raw OCR output.
//   - error: Error if image decoding, network communication, or OCR parsing fails.
func (s *UmatService) Ocr(file multipart.File, header *multipart.FileHeader, c *gin.Context) (interface{}, error) {
	if file == nil || header == nil {
		return nil, errors.New("file upload tidak valid")
	}

	srcImg, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca format gambar: %w", err)
	}

	bounds := srcImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var finalImg image.Image = srcImg
	maxDim := 2800 // must 2800
	if width > maxDim || height > maxDim {
		scale := math.Min(float64(maxDim)/float64(width), float64(maxDim)/float64(height))
		targetW := int(float64(width) * scale)
		targetH := int(float64(height) * scale)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}

		dstRGBA := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		draw.BiLinear.Scale(dstRGBA, dstRGBA.Bounds(), srcImg, bounds, draw.Over, nil)
		finalImg = dstRGBA
	}

	buf := ocrBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer ocrBufferPool.Put(buf)

	quality := 100
	if err := jpeg.Encode(buf, finalImg, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("gagal kompresi gambar: %w", err)
	}

	for buf.Len() > 1000000 && quality > 30 {
		quality -= 15
		buf.Reset()
		if err := jpeg.Encode(buf, finalImg, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("gagal re-encode kompresi gambar: %w", err)
		}
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	apiKey := os.Getenv("OCR_SPACE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("API Key OCR.space tidak ditemukan")
	}

	formData := url.Values{}
	formData.Set("apikey", apiKey)
	formData.Set("base64Image", "data:image/jpeg;base64,"+base64Str)
	formData.Set("language", "cht")
	formData.Set("OCREngine", "3")

	client := &http.Client{
		Timeout: 45 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://api.ocr.space/parse/image", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request OCR.space: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim request ke OCR.space API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBuf bytes.Buffer
		_, _ = io.Copy(&errBuf, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("OCR.space API error HTTP %d: %s", resp.StatusCode, errBuf.String())
	}

	var result OCRSpaceResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gagal membaca respon JSON dari OCR.space: %w", err)
	}

	var parsedUmat map[string]interface{}
	if len(result.ParsedResults) > 0 && result.ParsedResults[0].ParsedText != "" {
		parsedUmat = ParseUmatOcrText(result.ParsedResults[0].ParsedText)
	} else {
		parsedUmat = make(map[string]interface{})
	}

	// Fetch subWhId from user matrix to populate fotang_aktif & fotang_chiutao
	if c != nil {
		userID := getUserID(c)
		if userID > 0 {
			type SubWhInfo struct {
				SubWhId  string `gorm:"column:SUBWHID"`
				FullName string `gorm:"column:FULL_NAME"`
			}
			var subWhResult SubWhInfo
			subQuery := s.db.Table("T_WH_USER_MATRIX_MST AS A").
				Select("1").
				Where("A.LOGINID = ?", userID).
				Where("A.SUBWHID = T_WH_SUBWH_MST.SUBWHID")

			_ = s.db.Table("T_WH_SUBWH_MST").
				Select("SUBWHID, FULL_NAME").
				Where("EXISTS (?)", subQuery).
				Limit(1).
				Scan(&subWhResult).Error

			if subWhResult.SubWhId != "" {
				parsedUmat["fotang_aktif"] = subWhResult.SubWhId
				parsedUmat["fotang_chiutao"] = subWhResult.SubWhId
			}
		}
	}

	return OCRResponsePayload{
		ParsedUmat: parsedUmat,
		OCRRaw:     result,
	}, nil
}

// OCRResponsePayload encapsulates the parsed umat attributes and raw OCR data returned to the caller.
type OCRResponsePayload struct {
	ParsedUmat map[string]interface{} `json:"parsed_umat"`
	OCRRaw     interface{}            `json:"ocr_raw"`
}

// cleanSymbols strips punctuation and non-alphanumeric characters, keeping Chinese characters, letters, digits, and spaces.
//
// Parameters:
//   - s: The raw string to clean.
//
// Returns:
//   - string: Sanitized string.
func cleanSymbols(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			builder.WriteRune(r)
		}
	}
	return strings.TrimSpace(builder.String())
}

// deduplicateName removes duplicated name suffixes caused by OCR reading both English and Chinese names side by side.
//
// Parameters:
//   - rawName: The scanned name string.
//
// Returns:
//   - string: Deduplicated name string.
func deduplicateName(rawName string) string {
	rawName = strings.TrimSpace(rawName)
	if rawName == "" {
		return ""
	}
	parts := strings.Fields(rawName)
	if len(parts) > 1 {
		firstPartRunes := []rune(parts[0])
		lastPartRunes := []rune(parts[len(parts)-1])

		if len(firstPartRunes) > 0 && len(lastPartRunes) > 0 {
			lastCharOfFirst := string(firstPartRunes[len(firstPartRunes)-1])
			if parts[len(parts)-1] == lastCharOfFirst {
				return parts[0]
			}
		}
	}
	return rawName
}

// convertToPinyin converts Hanzi characters to title-cased Pinyin representation.
//
// Parameters:
//   - chineseName: Hanzi characters string.
//
// Returns:
//   - string: Title-cased Pinyin transcription separated by spaces.
func convertToPinyin(chineseName string) string {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	pyList := pinyin.Pinyin(chineseName, a)
	var words []string
	for _, p := range pyList {
		if len(p) > 0 && p[0] != "" {
			word := strings.Title(strings.ToLower(p[0]))
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}

// parseGregorianDateFromText extracts and formats Gregorian dates (YYYY-MM-DD) from scanned text lines.
//
// Parameters:
//   - text: Raw text line containing date indications.
//
// Returns:
//   - string: Standardized date string ("YYYY-MM-DD") or empty string if not found.
func parseGregorianDateFromText(text string) string {
	clean := strings.ReplaceAll(text, "′", "/")
	clean = strings.ReplaceAll(clean, "’", "/")
	clean = strings.ReplaceAll(clean, "'", "/")
	clean = strings.ReplaceAll(clean, "`", "/")

	monthMap := map[string]int{
		"jan": 1, "januari": 1, "january": 1,
		"feb": 2, "februari": 2, "february": 2,
		"mar": 3, "maret": 3, "march": 3,
		"apr": 4, "april": 4,
		"may": 5, "mei": 5,
		"jun": 6, "juni": 6, "june": 6,
		"jul": 7, "juli": 7, "july": 7,
		"aug": 8, "agu": 8, "agustus": 8, "august": 8,
		"sep": 9, "sept": 9, "september": 9,
		"oct": 10, "okt": 10, "oktober": 10, "october": 10,
		"nov": 11, "november": 11,
		"dec": 12, "des": 12, "desember": 12, "december": 12,
	}

	reMMM := regexp.MustCompile(`(?i)\b(\d{1,2})[\s/\-\.]*([a-z]{3,9})[\s/\-\.]*(\d{2,4})\b`)
	if matches := reMMM.FindStringSubmatch(clean); len(matches) == 4 {
		day, _ := strconv.Atoi(matches[1])
		monthStr := strings.ToLower(matches[2])
		year, _ := strconv.Atoi(matches[3])
		if year < 100 {
			year += 2000
		}
		if monthNum, ok := monthMap[monthStr]; ok && day >= 1 && day <= 31 {
			return fmt.Sprintf("%04d-%02d-%02d", year, monthNum, day)
		}
	}

	reYYYY := regexp.MustCompile(`\b(20\d{2})[/\-\.](\d{1,2})[/\-\.](\d{1,2})\b`)
	if matches := reYYYY.FindStringSubmatch(clean); len(matches) == 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
			return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		}
	}

	reDDMM := regexp.MustCompile(`\b(\d{1,2})[/\-\.](\d{1,2})[/\-\.](\d{2,4})\b`)
	if matches := reDDMM.FindStringSubmatch(clean); len(matches) == 4 {
		day, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		year, _ := strconv.Atoi(matches[3])
		if year < 100 {
			year += 2000
		}
		if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
			return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		}
	}

	return ""
}

// ParseUmatOcrText parses raw line-by-line OCR string output into structured umat registration fields.
//
// Parameters:
//   - rawText: Raw OCR text output.
//
// Returns:
//   - map[string]interface{}: Structured key-value mapping of recognized umat fields.
func ParseUmatOcrText(rawText string) map[string]interface{} {
	parsed := make(map[string]interface{})
	lines := strings.Split(rawText, "\n")

	cleanSymbols := func(str string) string {
		re := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
		cleaned := re.ReplaceAllString(str, "")
		return strings.TrimSpace(cleaned)
	}

	deduplicateName := func(str string) string {
		str = strings.TrimSpace(str)
		parts := strings.Fields(str)
		if len(parts) == 0 {
			return ""
		}
		if len(parts) == 2 && strings.HasSuffix(parts[0], parts[1]) {
			return parts[0]
		}
		seen := make(map[string]bool)
		var unique []string
		for _, p := range parts {
			if !seen[p] {
				seen[p] = true
				unique = append(unique, p)
			}
		}
		return strings.Join(unique, " ")
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 1. NAMA / 姓名
		if (strings.Contains(line, "NAMA") || strings.Contains(line, "姓名")) && (strings.Contains(line, ":") || strings.Contains(line, "：")) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				cleaned := deduplicateName(val)
				parsed["nama_indonesia"] = cleaned
				parsed["nama_mandarin"] = cleaned
				if cleaned != "" {
					py := convertToPinyin(cleaned)
					if py != "" {
						parsed["alias"] = py
					}
				}
			}
		}

		// 2. UMUR / 年齡
		if strings.Contains(line, "UMUR") || strings.Contains(line, "年齡") {
			reDigits := regexp.MustCompile(`\d+`)
			if numStr := reDigits.FindString(line); numStr != "" {
				if ageNum, err := strconv.Atoi(numStr); err == nil {
					parsed["usia"] = ageNum
					currentYear := time.Now().Year()
					birthYear := currentYear - ageNum
					parsed["tanggal_lahir"] = fmt.Sprintf("%d-01-01", birthYear)
				}
			}
		}

		// 3. PENDIDIKAN / 教育
		if strings.Contains(line, "PENDIDIKAN") || strings.Contains(line, "教育") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			var val string
			if len(parts) == 2 {
				val = strings.TrimSpace(parts[1])
			} else {
				val = strings.TrimSpace(strings.ReplaceAll(line, "PENDIDIKAN", ""))
				val = strings.TrimSpace(strings.ReplaceAll(val, "教育", ""))
			}
			valUpper := strings.ToUpper(strings.TrimSuffix(val, "."))
			valUpper = strings.ReplaceAll(valUpper, " ", "")
			if valUpper == "SI" || valUpper == "S" || valUpper == "S1" || valUpper == "S-1" || valUpper == "S/1" {
				val = "S1"
			} else if valUpper == "SII" || valUpper == "S2" || valUpper == "S-2" {
				val = "S2"
			} else if valUpper == "SIII" || valUpper == "S3" || valUpper == "S-3" {
				val = "S3"
			} else if val != "" {
				val = strings.ToUpper(val)
			}
			parsed["pendidikan"] = val
		}

		// 4. JENIS KELAMIN / 性別
		if strings.Contains(line, "性別") || strings.Contains(line, "JENIS KELAMIN") || strings.Contains(line, "KELAMIN") {
			lineUpper := strings.ToUpper(line)
			if strings.Contains(lineUpper, "童") || lineUpper == "ANAK PRIA" || lineUpper == "ANAK LAKI-LAKI" || lineUpper == "ANAK L" || lineUpper == "ANAK LAKI" {
				parsed["jenis_kelamin"] = "ANAK PRIA"
			} else if strings.Contains(lineUpper, "女") || lineUpper == "ANAK WANITA" || lineUpper == "ANAK W" {
				parsed["jenis_kelamin"] = "ANAK WANITA"
			} else if strings.Contains(lineUpper, "坤") || lineUpper == "P" || strings.Contains(lineUpper, "WANITA") || strings.Contains(lineUpper, "PEREMPUAN") || strings.Contains(lineUpper, " W ") || strings.HasSuffix(lineUpper, " W") || strings.HasSuffix(lineUpper, "WANITA") {
				parsed["jenis_kelamin"] = "WANITA"
			} else if strings.Contains(lineUpper, "乾") || lineUpper == "PRIA" || lineUpper == "LAKI-LAKI" || lineUpper == "L" {
				parsed["jenis_kelamin"] = "PRIA"
			}
		}

		// 5. ALAMAT / 地址
		if strings.Contains(line, "ALAMAT") || strings.Contains(line, "地址") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				addr := strings.TrimSpace(parts[1])
				reIl := regexp.MustCompile(`(?i)\bIl\.\s*`)
				addr = reIl.ReplaceAllString(addr, "JL. ")
				reJl := regexp.MustCompile(`(?i)\bJl\.\s*`)
				addr = reJl.ReplaceAllString(addr, "JL. ")
				parsed["alamat"] = addr
			}
		}

		// 6. TELEPON / 電話 / MOBILE
		if strings.Contains(line, "電話") || strings.Contains(line, "TELEPON") || strings.Contains(line, "TELP") || strings.Contains(line, "MOBILE") || strings.Contains(line, "HP") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				mob := strings.TrimSpace(parts[1])
				mob = strings.ReplaceAll(mob, "-", "")
				mob = strings.ReplaceAll(mob, " ", "")
				parsed["mobile"] = mob
			}
		}

		// 7. PERANTARA / 引師 (pengajak_manual)
		if strings.Contains(line, "PERANTARA") || strings.Contains(line, "引師") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				parsed["pengajak_manual"] = cleanSymbols(parts[1])
			} else {
				reKeyword := regexp.MustCompile(`(?i)(PERANTARA|引師)\s*`)
				val := reKeyword.ReplaceAllString(line, "")
				parsed["pengajak_manual"] = cleanSymbols(val)
			}
		}

		// 8. PENANGGUNG / 保師 (penanggung_manual)
		if strings.Contains(line, "PENANGGUNG") || strings.Contains(line, "保師") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				parsed["penanggung_manual"] = cleanSymbols(parts[1])
			} else {
				reKeyword := regexp.MustCompile(`(?i)(PENANGGUNG|保師)\s*`)
				val := reKeyword.ReplaceAllString(line, "")
				parsed["penanggung_manual"] = cleanSymbols(val)
			}
		}

		// 9. TCS / 點傳師
		if strings.Contains(line, "點傳師") || strings.Contains(line, "TCS") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				parsed["tcs"] = cleanSymbols(parts[1])
			} else {
				reKeyword := regexp.MustCompile(`(?i)(點傳師|TCS)\s*`)
				val := reKeyword.ReplaceAllString(line, "")
				parsed["tcs"] = cleanSymbols(val)
			}
		}

		// 10. 日期 / TANGGAL / tanggal_chiutao_int
		if strings.Contains(line, "日期") || strings.Contains(line, "TANGGAL") {
			if dt := parseGregorianDateFromText(line); dt != "" {
				parsed["tanggal_chiutao_int"] = dt
			}
		}

		// 11. 功德費 / uang_pahala & waktu_chiutao_mandarin
		if strings.Contains(line, "功德費") || strings.Contains(line, "UANG PAHALA") {
			reDigits := regexp.MustCompile(`\d[\d\.\,]*`)
			if numMatch := reDigits.FindString(line); numMatch != "" {
				numStr := strings.ReplaceAll(numMatch, ".", "")
				numStr = strings.ReplaceAll(numStr, ",", "")
				if amount, err := strconv.ParseFloat(numStr, 64); err == nil {
					parsed["uang_pahala"] = amount
				}
			}
			lunarHours := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
			for _, lh := range lunarHours {
				if strings.Contains(line, lh) {
					parsed["waktu_chiutao_mandarin"] = lh
					break
				}
			}
		}
	}

	// Fallback date check across whole rawText if tanggal_chiutao_int was not found in specific line
	if parsed["tanggal_chiutao_int"] == nil {
		if dt := parseGregorianDateFromText(rawText); dt != "" {
			parsed["tanggal_chiutao_int"] = dt
		}
	}

	return parsed
}
