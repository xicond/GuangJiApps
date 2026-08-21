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

type TopicService struct {
	db       *gorm.DB
	resource string
}

func NewTopicService(db *gorm.DB) *TopicService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &TopicService{db: db, resource: "topics"}
}

func (s *TopicService) List(page int, filters map[string]string, limit int) ([]domain.Topic, int64, error) {
	var items []domain.Topic
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_BUS_TOPIC").Where("Status = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"lookup_description": {Column: "TopicName", IsLike: true},
		"topic_name":         {Column: "TopicName", IsLike: true},
		"topic_code":         {Column: "TopicCode", IsLike: false},
		"topic_category":     {Column: "TopicCategory", IsLike: true},
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
			Preload("TopicCategoryInfo", func(db *gorm.DB) *gorm.DB {
				// Tambahkan filter kategori secara spesifik di sini
				return db.Where("CategoryId = ?", "B_KATEGORI_TOPIK")
			}).
			Limit(limit).
			Offset(offset).
			Order("TopicCode ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("topic tidak ditemukan")
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
		return []domain.Topic{}, 0, findErr
	}

	for i := range items {
		if items[i].TopicCategoryInfo != nil && items[i].TopicCategoryInfo.LookupDescription != nil && *items[i].TopicCategoryInfo.LookupDescription != "" {
			items[i].TopicCategory = *items[i].TopicCategoryInfo.LookupDescription
		}
	}

	return items, total, nil
}

func (s *TopicService) Create(payload domain.Topic, c *gin.Context) (domain.Topic, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Topic{}, fmt.Errorf("Validation failed: %w", err)
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
		return domain.Topic{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *TopicService) Get(id string) (domain.Topic, error) {
	var item domain.Topic
	if err := s.db.Where("TopicCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Topic{}, fmt.Errorf("topic %s not found", id)
		}
		return domain.Topic{}, err
	}
	return item, nil
}

func (s *TopicService) Update(id string, payload domain.Topic, c *gin.Context) (domain.Topic, error) {
	var item domain.Topic
	if err := s.db.Where("TopicCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Topic{}, fmt.Errorf("topic %s not found", id)
		}
		return domain.Topic{}, err
	}

	userIDStr := "1"
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userIDStr = strconv.Itoa(uid)
			}
		}
	}

	item.TopicName = payload.TopicName
	item.TopicCategory = payload.TopicCategory
	item.Description = payload.Description
	item.ModAct = "U"
	item.ModBy = userIDStr
	item.ModDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Topic{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *TopicService) Delete(id string, c *gin.Context) error {
	var item domain.Topic
	if err := s.db.Where("TopicCode = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("topic %s not found", id)
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
