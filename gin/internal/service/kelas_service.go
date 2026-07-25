package service

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KelasService struct {
	db       *gorm.DB
	resource string
}

func NewKelasService(db *gorm.DB) *KelasService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasService{db: db, resource: "kelas"}
}

func (s *KelasService) List(page int, filters map[string]string, limit int) ([]domain.Kelas, int64, error) {
	var items []domain.Kelas
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_APP_LOOKUP").Where("CategoryId = ? AND Status = ?", "B_KELASKHUSUS", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"lookup_id":          {Column: "LookupId", IsLike: true},
		"lookup_value":       {Column: "LookupValue", IsLike: true},
		"lookup_description": {Column: "LookupDescription", IsLike: true},
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

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).
			Limit(limit).
			Offset(offset).
			Order("LookupId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("kelas tidak ditemukan")
			} else {
				findErr = fmt.Errorf("database error: %w", err)
			}
		}
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if findErr != nil {
		return []domain.Kelas{}, 0, findErr
	}

	return items, total, nil
}

func (s *KelasService) Create(payload domain.Kelas, c *gin.Context) (domain.Kelas, error) {
	if payload.LookupValue == "" || payload.LookupDescription == "" {
		return domain.Kelas{}, fmt.Errorf("lookup_value and lookup_description are required")
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	payload.CategoryId = "B_KELASKHUSUS"
	payload.Status = true
	payload.ModAct = "I"
	payload.ModBy = userIDStr
	payload.ModDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasService) Get(id string) (domain.Kelas, error) {
	var item domain.Kelas
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_KELASKHUSUS").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
		}
		return domain.Kelas{}, err
	}
	return item, nil
}

func (s *KelasService) Update(id string, payload domain.Kelas, c *gin.Context) (domain.Kelas, error) {
	var item domain.Kelas
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_KELASKHUSUS").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
		}
		return domain.Kelas{}, err
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
	item.ModDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasService) Delete(id string, c *gin.Context) error {
	var item domain.Kelas
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_KELASKHUSUS").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas %s not found", id)
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
