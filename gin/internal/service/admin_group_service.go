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

// AdminGroupService handles administrative role and permission group operations.
type AdminGroupService struct {
	db       *gorm.DB
	resource string
}

// NewAdminGroupService initializes a new instance of AdminGroupService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *AdminGroupService: An initialized instance of AdminGroupService.
func NewAdminGroupService(db *gorm.DB) *AdminGroupService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &AdminGroupService{db: db, resource: "admin-groups"}
}

// List retrieves a paginated list of admin groups matching the given filter criteria.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter conditions (e.g. "group_name").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.AdminGroup: Slice of admin group records.
//   - int64: Total count of records matching the filters.
//   - error: Error if the database query fails.
func (s *AdminGroupService) List(page int, filters map[string]string, limit int) ([]domain.AdminGroup, int64, error) {
	var items []domain.AdminGroup
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_Login_Group")

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"group_name": {Column: "GroupName", IsLike: true},
		// "group_desc": {Column: "GroupDesc", IsLike: true},
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
			Order("GroupId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("admin group tidak ditemukan")
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
		return []domain.AdminGroup{}, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new admin group into the database after validating fields.
//
// Parameters:
//   - payload: The admin group record to create (domain.AdminGroup).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - domain.AdminGroup: The newly created admin group record.
//   - error: Error if validation fails or the insert query fails.
func (s *AdminGroupService) Create(payload domain.AdminGroup, c *gin.Context) (domain.AdminGroup, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.AdminGroup{}, fmt.Errorf("Validation failed: %w", err)
	}


	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AdminGroup{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single admin group by its ID.
//
// Parameters:
//   - id: The primary key (GroupId) of the admin group as a string.
//
// Returns:
//   - domain.AdminGroup: The retrieved admin group record.
//   - error: Error if ID format is invalid or record is not found.
func (s *AdminGroupService) Get(id string) (domain.AdminGroup, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminGroup{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.AdminGroup
	if err := s.db.Where("GroupId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminGroup{}, fmt.Errorf("admin group %s not found", id)
		}
		return domain.AdminGroup{}, err
	}
	return item, nil
}

// Update updates an existing admin group's permissions and details.
//
// Parameters:
//   - id: The primary key (GroupId) of the admin group to update as a string.
//   - payload: Updated admin group fields (domain.AdminGroup).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - domain.AdminGroup: The updated admin group record.
//   - error: Error if the record is not found or database update fails.
func (s *AdminGroupService) Update(id string, payload domain.AdminGroup, c *gin.Context) (domain.AdminGroup, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminGroup{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminGroup
	if err := s.db.Where("GroupId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminGroup{}, fmt.Errorf("admin group %s not found", id)
		}
		return domain.AdminGroup{}, err
	}

	item.GroupName = payload.GroupName
	item.RInsert = payload.RInsert
	item.REdit = payload.REdit
	item.RDelete = payload.RDelete
	item.RReporting = payload.RReporting
	item.RPositionId = payload.RPositionId
	item.GroupDesc = payload.GroupDesc

	if err := s.db.Save(&item).Error; err != nil {
		return domain.AdminGroup{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete removes an admin group from the database.
//
// Parameters:
//   - id: The primary key (GroupId) of the admin group to delete as a string.
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - error: Error if the record is not found or database delete fails.
func (s *AdminGroupService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminGroup
	if err := s.db.Where("GroupId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin group %s not found", id)
		}
		return err
	}

	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
