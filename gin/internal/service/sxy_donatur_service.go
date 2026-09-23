package service

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SxyDonaturService manages SXY donor profiles and lookup operations.
type SxyDonaturService struct {
	db       *gorm.DB
	resource string
}

// NewSxyDonaturService initializes a new instance of SxyDonaturService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *SxyDonaturService: An initialized instance of SxyDonaturService.
func NewSxyDonaturService(db *gorm.DB) *SxyDonaturService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &SxyDonaturService{db: db, resource: "sxy-donatur"}
}

// List executes SP_SXY_DONATUR_SEARCH_DATA to search and paginate SXY donor records.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter parameters (e.g. "nama", "mandarin", "fotang").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.SxyDonaturResponse: Slice of donor records matching the query.
//   - int64: Total count of matching records.
//   - error: Error if stored procedure query fails.
func (s *SxyDonaturService) List(page int, filters map[string]string, limit int) ([]domain.SxyDonaturResponse, int64, error) {
	var items []domain.SxyDonaturResponse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	// offset := (page - 1) * limit

	allowedFilters := map[string]string{
		"lookup_description": "nama",
		"nama":               "nama",
		"mandarin":           "mandarin",
		"fotang":             "fotang",
	}

	var nama string
	var mandarin string
	var fotang int

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "lookup_description" || field == "nama" {
				nama = value
			}

			if field == "mandarin" {
				mandarin = value
			}

			if field == "fotang" {
				fotang = toInt(value)
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	rows, err := s.db.Raw("EXEC SP_SXY_DONATUR_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu null)
		nama,          // @nama (string / nvarchar)
		mandarin,      // @mandarin (nvarchar)
		fotang,        // @Date (int)
	).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("sxy donatur tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, findErr
	}

	for rows.Next() {
		var item domain.SxyDonaturResponse
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
			default:
				matched := false
				// Cari field di struct berdasarkan tag gorm "column" atau nama field
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					// Cocokkan dengan tag kolom GORM atau nama struct (case-insensitive)
					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						valuePtrs[i] = v.Field(j).Addr().Interface()
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
			continue
		}

		// Ambil total row jika ada
		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.SxyDonaturResponse{}, 0, findErr
	}

	return items, total, nil
}

// MapRowToStruct populates target struct fields from a map of column names to values via reflection.
//
// Parameters:
//   - rowMap: Map of column names to raw database values.
//   - target: Pointer to the destination struct to populate.
func MapRowToStruct(rowMap map[string]interface{}, target interface{}) {
	v := reflect.ValueOf(target).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		fieldType := t.Field(i)
		fieldName := strings.ToLower(fieldType.Name)

		// Cari key di rowMap secara case-insensitive
		var rawVal interface{}
		var found bool
		for k, v := range rowMap {
			if strings.ToLower(k) == fieldName {
				rawVal = v
				found = true
				break
			}
		}

		if !found || rawVal == nil {
			continue
		}

		// Assign nilai berdasarkan tipe data field pada struct
		switch fieldVal.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			fieldVal.SetInt(toInt64(rawVal))
		case reflect.String:
			fieldVal.SetString(toString(rawVal))
		case reflect.Bool:
			if b, ok := rawVal.(bool); ok {
				fieldVal.SetBool(b)
			}
			// Tambahkan jenis tipe lain jika diperlukan (misal: float, time.Time, dll)
		}
	}
}

// toInt64 converts arbitrary numeric, byte slice, or string values to int64.
//
// Parameters:
//   - val: The raw interface value to convert.
//
// Returns:
//   - int64: Converted integer value or 0 if conversion fails.
func toInt64(val interface{}) int64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	// case int:
	//     return int64(v)
	// case int32:
	//     return int64(v)
	// case float64:
	//     return int64(v)
	case []byte:
		// Konversi byte slice dari database driver ke string, lalu parse ke int64
		n, _ := strconv.ParseInt(string(v), 10, 64)
		return n
	case string:
		// Parse string langsung ke int64
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	default:
		// Fallback terakhir: ubah ke string dulu lalu parse
		n, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		return n
	}
}

// Create inserts a new SXY donor record after auto-generating an ID using SP_APP_GenerateId.
//
// Parameters:
//   - payload: Donor record to create (domain.SxyDonatur).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.SxyDonatur: The newly created donor record.
//   - error: Error if ID generation fails, validation fails, or database insert fails.
func (s *SxyDonaturService) Create(payload domain.SxyDonatur, c *gin.Context) (domain.SxyDonatur, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("Validation failed: %w", err)
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "SXYDONATURID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.SxyDonatur{}, fmt.Errorf("failed to generate ID: %w", errId)
	}

	payload.ID = genResult.GeneratedId

	userID := getUserID(c)
	payload.Status = true
	payload.CreatedBy = userID
	payload.CreatedDate = domain.NowDateTime()
	payload.UpdatedBy = userID
	payload.UpdatedDate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single SXY donor by primary key ID.
//
// Parameters:
//   - id: The primary key of the donor as a string.
//
// Returns:
//   - domain.SxyDonatur: The retrieved donor record.
//   - error: Error if ID format is invalid or record is not found.
func (s *SxyDonaturService) Get(id string) (domain.SxyDonatur, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.SxyDonatur
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.SxyDonatur{}, fmt.Errorf("sxy donatur %s not found", id)
		}
		return domain.SxyDonatur{}, err
	}
	return item, nil
}

// Update updates an existing SXY donor record's details.
//
// Parameters:
//   - id: The primary key of the donor to update as a string.
//   - payload: Updated donor fields (domain.SxyDonatur).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.SxyDonatur: The updated donor record.
//   - error: Error if the record is not found or database update fails.
func (s *SxyDonaturService) Update(id string, payload domain.SxyDonatur, c *gin.Context) (domain.SxyDonatur, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.SxyDonatur
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.SxyDonatur{}, fmt.Errorf("sxy donatur %s not found", id)
		}
		return domain.SxyDonatur{}, err
	}

	item.No = payload.No
	item.Nama = payload.Nama
	item.Mandarin = payload.Mandarin
	item.Keterangan = payload.Keterangan
	item.LookupFothang = payload.LookupFothang
	item.Alamat = payload.Alamat
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile
	item.Email = payload.Email
	item.UpdatedBy = getUserID(c)
	item.UpdatedDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete deactivates an SXY donor record (soft delete via Status = false).
//
// Parameters:
//   - id: The primary key of the donor to deactivate.
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - error: Error if the record is not found or database update fails.
func (s *SxyDonaturService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.SxyDonatur
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sxy donatur %s not found", id)
		}
		return err
	}

	item.Status = false
	item.UpdatedBy = getUserID(c)
	item.UpdatedDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
