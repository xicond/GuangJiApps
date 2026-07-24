package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ActivityService struct {
	db       *gorm.DB
	resource string
}

func NewActivityService(db *gorm.DB) *ActivityService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &ActivityService{db: db, resource: "activities"}
}

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
		"event_code":     {Column: "EventCode", IsLike: true},
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("database count error: %w", err)
	}

	err := query.
		Limit(limit).
		Offset(offset).
		Order("EventCode ASC").
		Find(&items).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.Activity{}, 0, errors.New("activity tidak ditemukan")
		}
		return []domain.Activity{}, 0, fmt.Errorf("database error: %w", err)
	}

	return items, total, nil
}

func (s *ActivityService) Create(payload domain.Activity, c *gin.Context) (domain.Activity, error) {
	if payload.EventCode == "" || payload.EventName == "" {
		return domain.Activity{}, fmt.Errorf("event_code and event_name are required")
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
	payload.ModDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Activity{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

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
	item.ModDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Activity{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *ActivityService) Delete(id string, c *gin.Context) error {
	var item domain.Activity
	if err := s.db.Where("EventCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("activity %s not found", id)
		}
		return err
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	item.Status = false
	item.ModAct = "D"
	item.ModBy = userIDStr
	item.ModDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
