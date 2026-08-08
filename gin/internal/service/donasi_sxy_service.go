package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type nullStringScanner struct {
	target *string
}

func (s *nullStringScanner) Scan(value interface{}) error {
	if value == nil {
		*s.target = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*s.target = v
		return nil
	case []byte:
		*s.target = string(v)
		return nil
	default:
		*s.target = fmt.Sprintf("%v", v)
		return nil
	}
}

type DonasiSxyService struct {
	db                   *gorm.DB
	resource             string
	reportServerURL      string
	reportServerUsername string
	reportServerPassword string
}

func NewDonasiSxyService(db *gorm.DB) *DonasiSxyService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &DonasiSxyService{db: db, resource: "donasi-sxy", reportServerURL: getReportServerURL(), reportServerUsername: getReportServerUsername(), reportServerPassword: getReportServerPassword()}
}

func isValidDate(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

func (s *DonasiSxyService) List(page int, filters map[string]string, limit int) ([]domain.DonasiSxyResponse, int64, error) {
	items := make([]domain.DonasiSxyResponse, 0)
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	allowedFilters := map[string]bool{
		"no_kwitansi": true,
		"start_date":  true,
		"end_date":    true,
		"donatur":     true,
		// "nama":         true,
		// "donatur_nama": true,
	}

	var no_kwitansi string = ""
	var start_date *string = nil
	var end_date *string = nil
	var donatur string = ""

	for field, value := range filters {
		if value == "" {
			continue
		}
		if allowedFilters[field] {
			val := value
			switch field {
			case "no_kwitansi":
				no_kwitansi = val
			case "start_date":
				if isValidDate(val) {
					start_date = &val
				} else if donatur == "" {
					donatur = val
				}
			case "end_date":
				if isValidDate(val) {
					end_date = &val
				} else if donatur == "" {
					donatur = val
				}
			case "donatur":
				donatur = val
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	rows, err := s.db.Raw("EXEC SP_SXY_TRX_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu ASCENDING)
		no_kwitansi,   // @NoKwitansi (nvarchar)
		start_date,    // @StartDate (nvarchar)
		end_date,      // @EndDate (nvarchar)
		donatur,       // @Donatur (nvarchar)
	).Rows()

	if err != nil {
		return items, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("donasi sxy tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return items, total, findErr
	}

	for rows.Next() {
		var item domain.DonasiSxyResponse
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))
			cleanColNoUnderscore := strings.ReplaceAll(cleanCol, "_", "")

			switch cleanColNoUnderscore {
			case "totalrow", "totalcount", "rowcount":
				valuePtrs[i] = &nullFieldScanner{target: &totalRowScan}
			default:
				matched := false
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := strings.ToLower(field.Tag.Get("gorm"))
					gormTagNoUnderscore := strings.ReplaceAll(gormTag, "_", "")
					fieldNameNoUnderscore := strings.ReplaceAll(strings.ToLower(field.Name), "_", "")

					if strings.Contains(gormTagNoUnderscore, "column:"+cleanColNoUnderscore) ||
						fieldNameNoUnderscore == cleanColNoUnderscore {
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
			fmt.Printf("donasi_sxy rows.Scan error: %v\n", err)
			findErr = fmt.Errorf("scan error on row: %w", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.DonasiSxyResponse{}, 0, findErr
	}

	return items, total, nil
}

func (s *DonasiSxyService) Create(payload domain.DonasiSxy, c *gin.Context) (domain.DonasiSxy, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("Validation failed: %w", err)
	}

	var maxID int32
	s.db.Table("T_SXY_TRANSAKSI").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
	payload.ID = maxID + 1

	userID := int32(0)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	payload.Status = true
	payload.CreatedBy = userID
	payload.CreatedDate = domain.NowDateTime()
	payload.UpdatedBy = userID
	payload.UpdatedDate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *DonasiSxyService) Get(id string) (domain.DonasiSxy, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
		}
		return domain.DonasiSxy{}, err
	}
	return item, nil
}

func (s *DonasiSxyService) Update(id string, payload domain.DonasiSxy, c *gin.Context) (domain.DonasiSxy, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
		}
		return domain.DonasiSxy{}, err
	}

	userID := int32(0)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.NoKwitansi = payload.NoKwitansi
	item.Tanggal = payload.Tanggal
	item.Donatur = payload.Donatur
	item.Penggalang = payload.Penggalang
	item.Jumlah = payload.Jumlah
	item.TipeSumbangan = payload.TipeSumbangan
	item.NoKupon = payload.NoKupon
	item.Keterangan = payload.Keterangan
	item.TanggalTransfer = payload.TanggalTransfer
	item.AtasNama = payload.AtasNama
	item.TtkSent = payload.TtkSent
	item.UpdatedBy = userID
	item.UpdatedDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *DonasiSxyService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("donasi sxy %s not found", id)
		}
		return err
	}

	userID := int32(0)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.Status = false
	item.UpdatedBy = userID
	item.UpdatedDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

func (s *DonasiSxyService) Report(page int, filters map[string]string, limit int) ([]domain.SxyDonasiReport, float64, int64, error) {
	items := make([]domain.SxyDonasiReport, 0)
	var total int64
	var totalJumlah float64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	allowedFilters := map[string]bool{
		"donatur":    true,
		"penggalang": true,
		"start_date": true,
		"end_date":   true,
		"fotang":     true,
	}

	var donatur int = 0
	var penggalang int = 0
	var startDate interface{} = nil
	var endDate interface{} = nil
	var fotang int = 0

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "donatur" {
				donatur = toInt(value)
			}
			if field == "penggalang" {
				penggalang = toInt(value)
			}
			if field == "start_date" {
				startDate = value
			}
			if field == "end_date" {
				endDate = value
			}
			if field == "fotang" {
				fotang = toInt(value)
			}
		}
	}

	var findErr error

	rows, err := s.db.Raw("EXEC SP_SXY_RPT_TRANSAKSI ?, ?, ?, ?, ?, ?, ?",
		donatur,
		penggalang,
		startDate,
		endDate,
		fotang,
		limit, // @PageSize
		page,  // @CurrentPage
	).Rows()

	// Log query and values
	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	// log.Printf("[SQL] %s | Query:\n%s\n", timestamp, s.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
	// 	return tx.Raw("EXEC SP_SXY_RPT_TRANSAKSI ?, ?, ?, ?, ?, ?, ?",
	// 		donatur,
	// 		penggalang,
	// 		startDate,
	// 		endDate,
	// 		fotang,
	// 		limit, // @PageSize
	// 		page,  // @CurrentPage
	// 	)
	// }))

	if err != nil {
		return nil, 0, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("sxy report tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, 0, findErr
	}

	for rows.Next() {
		var item domain.SxyDonasiReport
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		// Siapkan slice pointer untuk rows.Scan sepanjang jumlah kolom
		valuePtrs := make([]interface{}, len(cols))

		// Variabel penampung sementara untuk kolom khusus seperti TotalRow yang tidak ada di struct
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow":
				valuePtrs[i] = &totalRowScan
			case "totaljumlah":
				valuePtrs[i] = &totalJumlah
			default:
				matched := false
				// Cari field di struct berdasarkan tag gorm "column" atau nama field
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					// Cocokkan dengan tag kolom GORM atau nama struct (case-insensitive)
					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						fieldVal := v.Field(j)
						if fieldVal.Kind() == reflect.String {
							valuePtrs[i] = &nullStringScanner{target: fieldVal.Addr().Interface().(*string)}
						} else {
							valuePtrs[i] = fieldVal.Addr().Interface()
						}
						matched = true
						break
					}
				}
				// Jika kolom database tidak ada di struct, tampung ke dummy agar Scan tidak error
				if !matched {
					var dummy interface{}
					valuePtrs[i] = &dummy
				}
			}
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("[ERROR] Report rows.Scan error: %v", err)
			continue
		}

		// Ambil total row jika ada
		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	} /*

		if findErr != nil {
			return []domain.SxyDonasiReport{}, 0, 0, findErr
		} */

	return items, totalJumlah, total, nil
}

func (s *DonasiSxyService) ReportExcel(filters map[string]string, c *gin.Context) error {

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
				baseURL = fmt.Sprintf("%s://%s:%s@%s", schemeParts[0], url.QueryEscape(username), url.QueryEscape(password), schemeParts[1])
			}
		}
	}

	startDate := filters["start_date"]
	if startDate == "" {
		startDate = filters["startDate"]
	}
	endDate := filters["end_date"]
	if endDate == "" {
		endDate = filters["endDate"]
	}
	penggalang := filters["penggalang"]
	_, err := strconv.ParseUint(penggalang, 10, 64)
	if err != nil {
		penggalang = "0"
	}
	donatur := filters["donatur"]
	_, err = strconv.ParseUint(donatur, 10, 64)
	if err != nil {
		donatur = "0"
	}
	fotang := filters["fotang"]
	_, err = strconv.ParseUint(fotang, 10, 64)
	if err != nil {
		fotang = "0"
	}

	// &SubWhId=%s
	reportURL := fmt.Sprintf("%s/ReportServer?%%2fGuangJiReport%%2frpt_sxy_transaksi&StartDate=%s&EndDate=%s&Penggalang=%s&Donatur=%s&Fotang=%s&rs:Command=Render&rs:Format=EXCELOPENXML",
		baseURL,
		url.QueryEscape(startDate),
		url.QueryEscape(endDate),
		url.QueryEscape(penggalang),
		url.QueryEscape(donatur),
		url.QueryEscape(fotang),
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
		return fmt.Errorf("report server mengembalikan HTTP %d (fetch %v)", resp.StatusCode, fetchDuration)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition == "" {
		hash := sha256.Sum256([]byte(startDate + endDate + penggalang + donatur + fotang))
		filename := fmt.Sprintf("rpt_sxy_transaksi_%x.xlsx", hash[:4])
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
