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

type TahunCiuTaoService struct {
	db       *gorm.DB
	resource string
}

func NewTahunCiuTaoService(db *gorm.DB) *TahunCiuTaoService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &TahunCiuTaoService{db: db, resource: "tahun-ciu-tao"}
}

func (s *TahunCiuTaoService) List(page int, filters map[string]string, limit int) ([]domain.TahunCiuTao, int64, error) {
	var items []domain.TahunCiuTao
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_BUS_TAHUN_CIUTAO").Where("status = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"tahun_mandarin": {Column: "TahunMandarin", IsLike: true},
		"description":    {Column: "description", IsLike: true},
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
		Order("TahunMandarin ASC").
		Find(&items).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.TahunCiuTao{}, 0, errors.New("tahun ciu tao tidak ditemukan")
		}
		return []domain.TahunCiuTao{}, 0, fmt.Errorf("database error: %w", err)
	}

	return items, total, nil
}

func (s *TahunCiuTaoService) Create(payload domain.TahunCiuTao, c *gin.Context) (domain.TahunCiuTao, error) {
	if payload.TahunMandarin == "" {
		return domain.TahunCiuTao{}, fmt.Errorf("tahun_mandarin is required")
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
		return domain.TahunCiuTao{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

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
	item.ModDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.TahunCiuTao{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *TahunCiuTaoService) Delete(id string, c *gin.Context) error {
	var item domain.TahunCiuTao
	if err := s.db.Where("TahunMandarin = ?", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tahun ciu tao %s not found", id)
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
