package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/Azure/go-ntlmssp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KelasService struct {
	db                   *gorm.DB
	resource             string
	reportServerURL      string
	reportServerUsername string
	reportServerPassword string
}

func NewKelasService(db *gorm.DB) *KelasService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasService{
		db:                   db,
		resource:             "kelas",
		reportServerURL:      getReportServerURL(),
		reportServerUsername: getReportServerUsername(),
		reportServerPassword: getReportServerPassword(),
	}
}

func getReportServerURL() string {
	if url := os.Getenv("REPORT_BASE_URL"); url != "" {
		return strings.TrimRight(url, "/")
	}
	return "http://localhost"
}

func getReportServerUsername() string {
	return os.Getenv("REPORT_USERNAME")
}

func getReportServerPassword() string {
	return os.Getenv("REPORT_PASSWORD")
}

var defaultReportClient = &http.Client{
	Timeout: 120 * time.Second,
	Transport: ntlmssp.Negotiator{
		RoundTripper: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	},
}

var copyBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 32*1024)
		return &buf
	},
}

func (s *KelasService) List(page int, filters map[string]string, c *gin.Context, limit int) ([]domain.KelasResponse, int64, error) {
	var items []domain.KelasResponse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	allowedFilters := map[string]bool{
		"kelas":              true,
		"lookup_description": true,
		"start_date":         true,
		"startdate":          true,
		"end_date":           true,
		"enddate":            true,
		"fotang":             true,
	}

	var kelas string = "0"
	var start_date string = "2018-07-01"
	var end_date string = "2026-07-25"
	var fotang string = "0"

	for field, value := range filters {
		if value == "" {
			continue
		}
		if allowedFilters[field] {
			val := value
			switch field {
			case "kelas":
				kelas = val
			case "lookup_description":
				kelas = val
			case "start_date":
				if isValidDate(val) {
					start_date = val
				}
			case "end_date":
				if isValidDate(val) {
					end_date = val
				}
			case "fotang":
				fotang = val
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	var userID int32
	if userIDVal, exists := c.Get("userID"); exists {
		userID = ToInt32(userIDVal)
	}

	rows, err := s.db.Raw("EXEC SP_TRX_KELAS_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu ASCENDING)
		kelas,         // @Kelas (nvarchar)
		start_date,    // @StartDate (Date)
		end_date,      // @EndDate (Date)
		fotang,        // @Fotang (nvarchar)
		userID,        // @LoginId (int32)
	).Rows()

	if err != nil {
		return items, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("kelas tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return items, total, findErr
	}

	for rows.Next() {
		var item domain.KelasResponse
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow", "total_row", "totalcount", "total_count", "rowcount":
				valuePtrs[i] = &totalRowScan
			default:
				matched := false
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

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
				if !matched {
					var dummy interface{}
					valuePtrs[i] = &dummy
				}
			}
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("kelas rows.Scan error: %v\n", err)
			findErr = fmt.Errorf("scan error on row: %w", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.KelasResponse{}, 0, findErr
	}

	return items, total, nil
}

func (s *KelasService) Lookup(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELASKHUSUS", page, limit, filters)
}

func getUserID(c *gin.Context) int32 {
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			switch v := val.(type) {
			case int:
				return int32(v)
			case int32:
				return v
			case float64:
				return int32(v)
			case string:
				if n, err := strconv.ParseInt(v, 10, 32); err == nil {
					return int32(n)
				}
			}
		}
	}
	return 1
}

func validateKelasLookups(db *gorm.DB, kodeKelas *string, kodeFotang *string, level *string) error {
	type lookupCheck struct {
		fieldName  string
		categoryID string
		val        string
	}

	var activeChecks []lookupCheck
	if kodeKelas != nil && strings.TrimSpace(*kodeKelas) != "" {
		activeChecks = append(activeChecks, lookupCheck{
			fieldName:  "kode_kelas",
			categoryID: "B_KELASKHUSUS",
			val:        strings.TrimSpace(*kodeKelas),
		})
	}
	if kodeFotang != nil && strings.TrimSpace(*kodeFotang) != "" {
		activeChecks = append(activeChecks, lookupCheck{
			fieldName:  "kode_fotang",
			categoryID: "B_FOTHANG",
			val:        strings.TrimSpace(*kodeFotang),
		})
	}
	if level != nil && strings.TrimSpace(*level) != "" {
		activeChecks = append(activeChecks, lookupCheck{
			fieldName:  "level",
			categoryID: "B_KLS_LEVEL",
			val:        strings.TrimSpace(*level),
		})
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
				Where("CategoryId = ? AND (LookupValue = ? OR LookupId = ?)", c.categoryID, c.val, c.val).
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

func (s *KelasService) Report(trxId string /* , subWhId string */, c *gin.Context) error {
	if trxId == "" && c != nil {
		trxId = c.Param("id")
		if trxId == "" {
			trxId = c.Query("trx_id")
		}
		if trxId == "" {
			trxId = c.Query("TrxId")
		}
	}

	/*

		if subWhId == "" && c != nil {
			var subWhVal int64
			userID := getUserID(c)
			row := s.db.Model(&domain.AdminMatrix{}).
				Where("LOGINID = ?", userID).
				Select("SUBWHID").
				Row()
			if row != nil {
				_ = row.Scan(&subWhVal)
			}
			if subWhVal > 0 {
				subWhId = strconv.FormatInt(subWhVal, 10)
			}
		} */

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

	// &SubWhId=%s
	reportURL := fmt.Sprintf("%s/ReportServer?%%2fGuangJiReport%%2frpt_trx_kelas&TrxId=%s&rs:Command=Render&rs:Format=EXCELOPENXML",
		baseURL,
		url.QueryEscape(trxId),
		// url.QueryEscape(subWhId),
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
		filename := fmt.Sprintf("rpt_trx_kelas_%s.xlsx", trxId)
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

func formatDigestAuth(authHeader, username, password, method, uri string) string {
	parts := parseHeaderParts(authHeader)
	realm := parts["realm"]
	nonce := parts["nonce"]
	qop := parts["qop"]
	opaque := parts["opaque"]

	ha1 := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password))))
	ha2 := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri))))

	nc := "00000001"
	cnonce := fmt.Sprintf("%08x", time.Now().UnixNano())

	var response string
	if strings.Contains(qop, "auth") {
		qop = "auth"
		response = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, nonce, nc, cnonce, qop, ha2))))
	} else {
		response = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", ha1, nonce, ha2))))
	}

	digest := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		username, realm, nonce, uri, response)

	if qop != "" {
		digest += fmt.Sprintf(`, qop=%s, nc=%s, cnonce="%s"`, qop, nc, cnonce)
	}
	if opaque != "" {
		digest += fmt.Sprintf(`, opaque="%s"`, opaque)
	}
	if parts["algorithm"] != "" {
		digest += fmt.Sprintf(`, algorithm="%s"`, parts["algorithm"])
	}

	return digest
}

func parseHeaderParts(header string) map[string]string {
	result := make(map[string]string)
	if !strings.HasPrefix(header, "Digest ") {
		return result
	}
	content := strings.TrimPrefix(header, "Digest ")
	parts := strings.Split(content, ",")
	for _, part := range parts {
		keyValue := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(keyValue) == 2 {
			key := strings.TrimSpace(keyValue[0])
			value := strings.Trim(strings.TrimSpace(keyValue[1]), `"`)
			result[key] = value
		}
	}
	return result
}

func (s *KelasService) Create(payload domain.Kelas, c *gin.Context) (domain.Kelas, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Kelas{}, fmt.Errorf("Validation failed: %w", err)
	}
	if err := validateKelasLookups(s.db, payload.KodeKelas, payload.KodeFotang, payload.Level); err != nil {
		return domain.Kelas{}, err
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASHEADERID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.Kelas{}, fmt.Errorf("failed to generate ID: %w", errId)
	}

	userID := getUserID(c)
	statusTrue := true
	modActI := "I"
	now := domain.NowDateTime()

	payload.TrxId = genResult.GeneratedId

	payload.Status = &statusTrue
	payload.ModAct = &modActI
	payload.ModBy = &userID
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasService) Get(id string) (domain.Kelas, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Kelas{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.Kelas
	if err := s.db.
		Preload("KelasName", "CategoryId = ?", "B_KELASKHUSUS").
		Preload("FotangName", "CategoryId = ?", "B_FOTHANG").
		Where("trxid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
		}
		return domain.Kelas{}, err
	}
	return item, nil
}

func (s *KelasService) Update(id string, payload domain.Kelas, c *gin.Context) (domain.Kelas, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Kelas{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Kelas
	if err := s.db.Where("trxid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
		}
		return domain.Kelas{}, err
	}

	if err := validateKelasLookups(s.db, payload.KodeKelas, payload.KodeFotang, payload.Level); err != nil {
		return domain.Kelas{}, err
	}

	if payload.KodeKelas != nil {
		item.KodeKelas = payload.KodeKelas
	}

	if payload.KodeFotang != nil {
		if strings.TrimSpace(*payload.KodeFotang) != "" {
			item.KodeFotang = payload.KodeFotang
		} else {
			item.KodeFotang = nil
		}
	}

	if payload.StartDate != nil {
		item.StartDate = payload.StartDate
	}
	if payload.EndDate != nil {
		item.EndDate = payload.EndDate
	}
	if payload.Lokasi != nil {
		item.Lokasi = payload.Lokasi
	}
	if payload.PIC != nil {
		item.PIC = payload.PIC
	}
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}
	if payload.Level != nil {
		item.Level = payload.Level
	}
	if payload.Mc1 != nil {
		item.Mc1 = payload.Mc1
	}
	if payload.Mc2 != nil {
		item.Mc2 = payload.Mc2
	}
	if payload.Mc3 != nil {
		item.Mc3 = payload.Mc3
	}
	if payload.Mc4 != nil {
		item.Mc4 = payload.Mc4
	}
	if payload.Mc5 != nil {
		item.Mc5 = payload.Mc5
	}
	if payload.Deadline != nil {
		item.Deadline = payload.Deadline
	}

	userID := getUserID(c)
	modActU := "U"
	now := domain.NowDateTime()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := ValidateStruct(item); err != nil {
		return domain.Kelas{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Kelas
	if err := s.db.Where("trxid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas %s not found", id)
		}
		return err
	}

	userID := getUserID(c)
	statusFalse := false
	modActD := "D"
	now := domain.NowDateTime()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

func ToInt32(val interface{}) int32 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int32:
		return v
	case []byte:
		n, _ := strconv.ParseInt(string(v), 10, 32)
		return int32(n)
	case string:
		n, _ := strconv.ParseInt(v, 10, 32)
		return int32(n)
	default:
		n, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 32)
		return int32(n)
	}
}
