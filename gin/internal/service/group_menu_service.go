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

// GroupMenuService manages navigation menu items and group menu access mappings.
type GroupMenuService struct {
	db       *gorm.DB
	resource string
}

// NewGroupMenuService initializes a new instance of GroupMenuService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *GroupMenuService: An initialized instance of GroupMenuService.
func NewGroupMenuService(db *gorm.DB) *GroupMenuService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &GroupMenuService{db: db, resource: "group-menu-mappings"}
}

// List retrieves a paginated list of active group menu mappings matching the given filter criteria.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter conditions (e.g. "menu_name", "page_url", "parent_id").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.GroupMenuMapping: Slice of menu mapping records.
//   - int64: Total count of records matching the filters.
//   - error: Error if the database query fails.
func (s *GroupMenuService) List(page int, filters map[string]string, limit int) ([]domain.GroupMenuMapping, int64, error) {
	var items []domain.GroupMenuMapping
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_Login_Menu").Where("FlagActive = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"menu_name": {Column: "MenuName", IsLike: true},
		"page_url":  {Column: "PageUrl", IsLike: true},
		"parent_id": {Column: "ParentId", IsLike: false},
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
			Order("MenuId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("group menu mapping tidak ditemukan")
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
		return []domain.GroupMenuMapping{}, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new menu item mapping into the database with default hierarchy sequence values.
//
// Parameters:
//   - payload: The menu mapping record to create (domain.GroupMenuMapping).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - domain.GroupMenuMapping: The newly created menu mapping record.
//   - error: Error if validation fails or database insert fails.
func (s *GroupMenuService) Create(payload domain.GroupMenuMapping, c *gin.Context) (domain.GroupMenuMapping, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("Validation failed: %w", err)
	}


	if payload.Sequence == 0 {
		payload.Sequence = 1
	}
	if payload.ParentLevel1 == 0 {
		payload.ParentLevel1 = 1
	}
	payload.FlagActive = true

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single group menu mapping by its MenuId.
//
// Parameters:
//   - id: The primary key (MenuId) of the menu mapping as a string.
//
// Returns:
//   - domain.GroupMenuMapping: The retrieved menu mapping record.
//   - error: Error if ID format is invalid or record is not found.
func (s *GroupMenuService) Get(id string) (domain.GroupMenuMapping, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.GroupMenuMapping
	if err := s.db.Where("MenuId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.GroupMenuMapping{}, fmt.Errorf("group menu mapping %s not found", id)
		}
		return domain.GroupMenuMapping{}, err
	}
	return item, nil
}

// Update updates an existing menu mapping's attributes.
//
// Parameters:
//   - id: The primary key (MenuId) of the menu mapping to update as a string.
//   - payload: Updated menu mapping fields (domain.GroupMenuMapping).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - domain.GroupMenuMapping: The updated menu mapping record.
//   - error: Error if the record is not found or database update fails.
func (s *GroupMenuService) Update(id string, payload domain.GroupMenuMapping, c *gin.Context) (domain.GroupMenuMapping, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.GroupMenuMapping
	if err := s.db.Where("MenuId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.GroupMenuMapping{}, fmt.Errorf("group menu mapping %s not found", id)
		}
		return domain.GroupMenuMapping{}, err
	}

	item.ParentId = payload.ParentId
	item.MenuName = payload.MenuName
	item.PageUrl = payload.PageUrl
	item.Sequence = payload.Sequence
	item.MenuDesc = payload.MenuDesc
	item.ParentLevel1 = payload.ParentLevel1
	item.FlagActive = payload.FlagActive

	if err := s.db.Save(&item).Error; err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete deactivates a menu mapping (soft delete via FlagActive = false).
//
// Parameters:
//   - id: The primary key (MenuId) of the menu mapping to deactivate.
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - error: Error if the record is not found or database update fails.
func (s *GroupMenuService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.GroupMenuMapping
	if err := s.db.Where("MenuId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("group menu mapping %s not found", id)
		}
		return err
	}

	item.FlagActive = false
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
