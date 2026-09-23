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

// AdminSubWarehouseService manages sub-warehouse locations and assignments.
type AdminSubWarehouseService struct {
	db       *gorm.DB
	resource string
}

// NewAdminSubWarehouseService initializes a new instance of AdminSubWarehouseService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *AdminSubWarehouseService: An initialized instance of AdminSubWarehouseService.
func NewAdminSubWarehouseService(db *gorm.DB) *AdminSubWarehouseService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &AdminSubWarehouseService{db: db, resource: "admin-sub-warehouses"}
}

// List retrieves a paginated list of sub-warehouse master records matching the given filter criteria.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter conditions (e.g. "full_name", "pic", "doc_code", "sub_wh_type").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.AdminSubWarehouse: Slice of sub-warehouse records.
//   - int64: Total count of records matching the filters.
//   - error: Error if the database query fails.
func (s *AdminSubWarehouseService) List(page int, filters map[string]string, limit int) ([]domain.AdminSubWarehouse, int64, error) {
	var items []domain.AdminSubWarehouse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_WH_SUBWH_MST")

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"full_name":   {Column: "FULL_NAME", IsLike: true},
		"pic":         {Column: "PIC", IsLike: true},
		"doc_code":    {Column: "DOCCODE", IsLike: true},
		"sub_wh_type": {Column: "SubWhType", IsLike: false},
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if rule, exists := allowedFilters[field]; exists {
			if rule.IsLike {
				query = query.Where(fmt.Sprintf("[%s] LIKE ?", rule.Column), "%"+value+"%")
			} else {
				query = query.Where(fmt.Sprintf("[%s] = ?", rule.Column), value)
			}
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
			Order("SUBWHID ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("admin sub warehouse tidak ditemukan")
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
		return []domain.AdminSubWarehouse{}, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new sub-warehouse master record into the database, computing the next SubWhId.
//
// Parameters:
//   - payload: The sub-warehouse record to create (domain.AdminSubWarehouse).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.AdminSubWarehouse: The newly created sub-warehouse record.
//   - error: Error if validation fails or database insert fails.
func (s *AdminSubWarehouseService) Create(payload domain.AdminSubWarehouse, c *gin.Context) (domain.AdminSubWarehouse, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("Validation failed: %w", err)
	}

	var maxID int32
	s.db.Table("T_WH_SUBWH_MST").Select("ISNULL(MAX(SUBWHID), 0)").Row().Scan(&maxID)
	payload.SubWhId = maxID + 1

	userID := int64(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int64(uid)
			}
		}
	}

	payload.CruId = userID
	payload.UpdateUId = userID
	payload.LstUpdate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single sub-warehouse record by its SubWhId.
//
// Parameters:
//   - id: The primary key (SubWhId) of the sub-warehouse as a string.
//
// Returns:
//   - domain.AdminSubWarehouse: The retrieved sub-warehouse record.
//   - error: Error if ID format is invalid or record is not found.
func (s *AdminSubWarehouseService) Get(id string) (domain.AdminSubWarehouse, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.AdminSubWarehouse
	if err := s.db.Where("SUBWHID = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminSubWarehouse{}, fmt.Errorf("admin sub warehouse %s not found", id)
		}
		return domain.AdminSubWarehouse{}, err
	}
	return item, nil
}

// Update updates an existing sub-warehouse master record and records the modifying user ID and timestamp.
//
// Parameters:
//   - id: The primary key (SubWhId) of the sub-warehouse to update as a string.
//   - payload: Updated sub-warehouse fields (domain.AdminSubWarehouse).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.AdminSubWarehouse: The updated sub-warehouse record.
//   - error: Error if the record is not found or database update fails.
func (s *AdminSubWarehouseService) Update(id string, payload domain.AdminSubWarehouse, c *gin.Context) (domain.AdminSubWarehouse, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminSubWarehouse
	if err := s.db.Where("SUBWHID = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminSubWarehouse{}, fmt.Errorf("admin sub warehouse %s not found", id)
		}
		return domain.AdminSubWarehouse{}, err
	}

	item.WhId = payload.WhId
	item.FullName = payload.FullName
	item.Pic = payload.Pic
	item.DocCode = payload.DocCode
	item.UpdateUId = toInt64(getUserID(c))
	item.LstUpdate = domain.NowDateTime()
	item.FlagProductions = payload.FlagProductions
	item.SubWhType = payload.SubWhType

	if err := s.db.Save(&item).Error; err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete removes a sub-warehouse record from the database.
//
// Parameters:
//   - id: The primary key (SubWhId) of the sub-warehouse to delete as a string.
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - error: Error if the record is not found or database delete fails.
func (s *AdminSubWarehouseService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminSubWarehouse
	if err := s.db.Where("SUBWHID = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin sub warehouse %s not found", id)
		}
		return err
	}

	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
