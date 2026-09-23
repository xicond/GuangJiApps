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

// TahunCiuTaoService handles Mandarin lunar calendar year ranges and metadata.
type TahunCiuTaoService struct {
	db       *gorm.DB
	resource string
}

// NewTahunCiuTaoService initializes a new instance of TahunCiuTaoService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *TahunCiuTaoService: An initialized instance of TahunCiuTaoService.
func NewTahunCiuTaoService(db *gorm.DB) *TahunCiuTaoService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &TahunCiuTaoService{db: db, resource: "tahun-ciu-tao"}
}

// List executes SP_BUS_YEAR_SEARCH_DATA to search Mandarin lunar calendar year records.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter parameters (e.g. "tahun_mandarin", "date").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.TahunCiuTao: Slice of lunar year records.
//   - int64: Total count of matching records.
//   - error: Error if stored procedure query fails.
func (s *TahunCiuTaoService) List(page int, filters map[string]string, limit int) ([]domain.TahunCiuTao, int64, error) {
	var items []domain.TahunCiuTao
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	allowedFilters := map[string]string{
		"tahun_mandarin": "TahunMandarin",
		"date":           "date",
	}

	var tahun string
	var dateVal *time.Time = nil

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "tahun_mandarin" {
				tahun = value
			}

			if field == "date" {
				t, err := time.Parse("2006-01-02", value)
				if err != nil {
					t, err = time.Parse(time.RFC3339, value)
					if err != nil {
						return nil, 0, fmt.Errorf("invalid date format: %w", err)
					}
				}
				dateVal = &t
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	rows, err := s.db.Raw("EXEC SP_BUS_YEAR_SEARCH_DATA ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu null)
		tahun,         // @Tahun (string / nvarchar)
		dateVal,       // @Date (time.Time atau nil)
	).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("tahun ciu tao tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, findErr
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.TahunCiuTao
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
		return []domain.TahunCiuTao{}, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new Mandarin year record into the database.
//
// Parameters:
//   - payload: Lunar year record to create (domain.TahunCiuTao).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.TahunCiuTao: The newly created year record.
//   - error: Error if validation fails or database insert fails.
func (s *TahunCiuTaoService) Create(payload domain.TahunCiuTao, c *gin.Context) (domain.TahunCiuTao, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.TahunCiuTao{}, fmt.Errorf("Validation failed: %w", err)
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	payload.Status = true
	payload.ModAct = "I"
	payload.ModBy = userIDStr
	payload.ModDate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.TahunCiuTao{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single Mandarin year record by TahunMandarin string.
//
// Parameters:
//   - id: The TahunMandarin primary key.
//
// Returns:
//   - domain.TahunCiuTao: The retrieved lunar year record.
//   - error: Error if the record is not found or database query fails.
func (s *TahunCiuTaoService) Get(id string) (domain.TahunCiuTao, error) {
	var item domain.TahunCiuTao
	if err := s.db.Where("TahunMandarin = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.TahunCiuTao{}, fmt.Errorf("tahun ciu tao %s not found", id)
		}
		return domain.TahunCiuTao{}, err
	}
	return item, nil
}

// Update updates an existing Mandarin lunar year record.
//
// Parameters:
//   - id: The TahunMandarin identifier of the record to update.
//   - payload: Updated year fields (domain.TahunCiuTao).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.TahunCiuTao: The updated year record.
//   - error: Error if the record is not found or database update fails.
func (s *TahunCiuTaoService) Update(id string, payload domain.TahunCiuTao, c *gin.Context) (domain.TahunCiuTao, error) {
	var item domain.TahunCiuTao
	if err := s.db.Where("TahunMandarin = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.TahunCiuTao{}, fmt.Errorf("tahun ciu tao %s not found", id)
		}
		return domain.TahunCiuTao{}, err
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	item.TahunMandarin = payload.TahunMandarin
	item.StartDate = payload.StartDate
	item.EndDate = payload.EndDate
	item.Description = payload.Description
	item.ModAct = "U"
	item.ModBy = userIDStr
	item.ModDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.TahunCiuTao{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete deactivates a Mandarin lunar year record (soft delete via Status = false and ModAct = 'D').
//
// Parameters:
//   - id: The TahunMandarin identifier of the record to deactivate.
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - error: Error if the record is not found or database update fails.
func (s *TahunCiuTaoService) Delete(id string, c *gin.Context) error {
	var item domain.TahunCiuTao
	if err := s.db.Where("TahunMandarin = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tahun ciu tao %s not found", id)
		}
		return err
	}

	item.Status = false
	item.ModAct = "D"
	item.ModBy = strconv.Itoa(int(getUserID(c)))
	item.ModDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
