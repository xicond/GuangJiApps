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

// ActivityService handles events and activities master records.
type ActivityService struct {
	db       *gorm.DB
	resource string
}

// NewActivityService initializes a new instance of ActivityService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *ActivityService: An initialized instance of ActivityService.
func NewActivityService(db *gorm.DB) *ActivityService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &ActivityService{db: db, resource: "activities"}
}

// List retrieves a paginated list of active activities matching the given filter criteria.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter parameters (e.g. "event_name", "event_category").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.Activity: Slice of activity records.
//   - int64: Total count of matching records.
//   - error: Error if database query fails.
func (s *ActivityService) List(page int, filters map[string]string, limit int) ([]domain.Activity, int64, error) {
	var items []domain.Activity
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_BUS_EVENT").Where("Status = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		// "event_code":     {Column: "EventCode", IsLike: true},
		"event_name":     {Column: "EventName", IsLike: true},
		"event_category": {Column: "EventCategory", IsLike: true},
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
			Order("EventCode ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("activity tidak ditemukan")
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
		return []domain.Activity{}, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new activity record into the database.
//
// Parameters:
//   - payload: Activity record to create (domain.Activity).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.Activity: The newly created activity record.
//   - error: Error if validation fails or database insert fails.
func (s *ActivityService) Create(payload domain.Activity, c *gin.Context) (domain.Activity, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Activity{}, fmt.Errorf("Validation failed: %w", err)
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
		return domain.Activity{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get fetches a single activity by its EventCode.
//
// Parameters:
//   - id: The EventCode of the activity.
//
// Returns:
//   - domain.Activity: The retrieved activity record.
//   - error: Error if the activity is not found or database query fails.
func (s *ActivityService) Get(id string) (domain.Activity, error) {
	var item domain.Activity
	if err := s.db.Where("EventCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Activity{}, fmt.Errorf("activity %s not found", id)
		}
		return domain.Activity{}, err
	}
	return item, nil
}

// Update updates an existing activity record's details.
//
// Parameters:
//   - id: The EventCode of the activity to update.
//   - payload: Updated activity fields (domain.Activity).
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - domain.Activity: The updated activity record.
//   - error: Error if the activity is not found or database update fails.
func (s *ActivityService) Update(id string, payload domain.Activity, c *gin.Context) (domain.Activity, error) {
	var item domain.Activity
	if err := s.db.Where("EventCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Activity{}, fmt.Errorf("activity %s not found", id)
		}
		return domain.Activity{}, err
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	item.EventName = payload.EventName
	item.EventCategory = payload.EventCategory
	item.Description = payload.Description
	item.ModAct = "U"
	item.ModBy = userIDStr
	item.ModDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Activity{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete deactivates an activity record (soft delete via Status = false and ModAct = 'D').
//
// Parameters:
//   - id: The EventCode of the activity to deactivate.
//   - c: Gin context carrying HTTP request metadata for user tracking.
//
// Returns:
//   - error: Error if the activity is not found or database update fails.
func (s *ActivityService) Delete(id string, c *gin.Context) error {
	var item domain.Activity
	if err := s.db.Where("EventCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("activity %s not found", id)
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
