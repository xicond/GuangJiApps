package service

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TimKerjaService handles work team positions and division mappings under the B_POSISI category.
type TimKerjaService struct {
	db       *gorm.DB
	resource string
}

// NewTimKerjaService initializes a new instance of TimKerjaService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *TimKerjaService: An initialized instance of TimKerjaService.
func NewTimKerjaService(db *gorm.DB) *TimKerjaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &TimKerjaService{db: db, resource: "tim-kerja"}
}

// List retrieves a paginated list of active work team position records matching the given filter criteria.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter parameters (e.g. "search", "lookup_description", "lookup_value").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.TimKerja: Slice of work team records.
//   - int64: Total count of matching records.
//   - error: Error if database query fails.
func (s *TimKerjaService) List(page int, filters map[string]string, limit int) ([]domain.TimKerja, int64, error) {
	var items []domain.TimKerja
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_APP_LOOKUP").Where("CategoryId = ? AND Status = ?", "B_POSISI", true)

	if filters != nil {
		var searchVal string
		for _, key := range []string{"search", "query", "q"} {
			if val, ok := filters[key]; ok && val != "" {
				searchVal = val
				break
			}
		}

		if searchVal != "" {
			query = query.Where("([LookupDescription] LIKE ? OR [LookupValue] LIKE ? OR [LookupId] LIKE ?)", "%"+searchVal+"%", "%"+searchVal+"%", "%"+searchVal+"%")
		} else if val, ok := filters["lookup_description"]; ok && val != "" {
			query = query.Where("([LookupDescription] LIKE ? OR [LookupValue] LIKE ?)", "%"+val+"%", "%"+val+"%")
		}

		if val, ok := filters["lookup_value"]; ok && val != "" && searchVal == "" {
			query = query.Where("[LookupValue] LIKE ?", "%"+val+"%")
		}

		if val, ok := filters["lookup_id"]; ok && val != "" {
			query = query.Where("[LookupId] = ?", val)
		}
	}

	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	})

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).
			Limit(limit).
			Offset(offset).
			Order("LookupId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("tim kerja tidak ditemukan")
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
		return []domain.TimKerja{}, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new work team position into T_APP_LOOKUP under CategoryId = B_POSISI.
//
// Parameters:
//   - payload: Work team record to create (domain.TimKerja).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.TimKerja: The newly created work team record.
//   - error: Error if validation fails or database insert fails.
func (s *TimKerjaService) Create(payload domain.TimKerja, c *gin.Context) (domain.TimKerja, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.TimKerja{}, fmt.Errorf("Validation failed: %w", err)
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	payload.CategoryId = "B_POSISI"
	if payload.LookupId == "" {
		payload.LookupId = payload.LookupValue
	}
	payload.Status = true
	payload.ModAct = "I"
	payload.ModBy = userIDStr
	payload.ModDate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.TimKerja{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single work team position by its LookupId.
//
// Parameters:
//   - id: The LookupId of the work team position.
//
// Returns:
//   - domain.TimKerja: The retrieved work team record.
//   - error: Error if the record is not found or database query fails.
func (s *TimKerjaService) Get(id string) (domain.TimKerja, error) {
	var item domain.TimKerja
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_POSISI").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.TimKerja{}, fmt.Errorf("tim kerja %s not found", id)
		}
		return domain.TimKerja{}, err
	}
	return item, nil
}

// Update updates an existing work team position's details.
//
// Parameters:
//   - id: The LookupId of the work team position to update.
//   - payload: Updated work team fields (domain.TimKerja).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.TimKerja: The updated work team record.
//   - error: Error if the record is not found or database update fails.
func (s *TimKerjaService) Update(id string, payload domain.TimKerja, c *gin.Context) (domain.TimKerja, error) {
	var item domain.TimKerja
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_POSISI").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.TimKerja{}, fmt.Errorf("tim kerja %s not found", id)
		}
		return domain.TimKerja{}, err
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	item.LookupValue = payload.LookupValue
	item.LookupDescription = payload.LookupDescription
	item.ModAct = "U"
	item.ModBy = userIDStr
	item.ModDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.TimKerja{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete deactivates a work team position (soft delete via Status = false and ModAct = 'D').
//
// Parameters:
//   - id: The LookupId of the work team position to deactivate.
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - error: Error if the record is not found or database update fails.
func (s *TimKerjaService) Delete(id string, c *gin.Context) error {
	var item domain.TimKerja
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_POSISI").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tim kerja %s not found", id)
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

// Lookup retrieves active work team options under CategoryId = B_TIMKERJA.
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count of matching records.
//   - error: Error if query fails.
func (s *TimKerjaService) Lookup(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_TIMKERJA", page, limit, filters)
}

// LookupSub retrieves active sub-division options under CategoryId = B_SUBKERJA mapped to the specified division value.
//
// Parameters:
//   - timkerjaValue: The parent division value to match against.
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of mapped sub-division lookup records.
//   - int64: Total count of matching records.
//   - error: Error if query fails.
func (s *TimKerjaService) LookupSub(timkerjaValue string, filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	subQuery := s.db.Table("T_BUS_WORK_MAPPING").
		Where("T_BUS_WORK_MAPPING.SubDivisi = T_APP_LOOKUP.LookupValue").
		Where("divisi = ?", timkerjaValue).
		Where("T_BUS_WORK_MAPPING.Status = 1")
	return Lookup(s.db, "B_SUBKERJA", page, limit, filters, subQuery)
}

// LookupReport retrieves active work team report options under CategoryId = B_TIMKERJA_REPORT.
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of report lookup records.
//   - int64: Total count of matching records.
//   - error: Error if query fails.
func (s *TimKerjaService) LookupReport(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_TIMKERJA_REPORT", page, limit, filters)
}
