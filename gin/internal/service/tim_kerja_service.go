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

type TimKerjaService struct {
	db       *gorm.DB
	resource string
}

func NewTimKerjaService(db *gorm.DB) *TimKerjaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &TimKerjaService{db: db, resource: "tim-kerja"}
}

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
				findErr = errors.New("tim kerja tidak ditemukan")
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
		return []domain.TimKerja{}, 0, findErr
	}

	return items, total, nil
}

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
	payload.Status = true
	payload.ModAct = "I"
	payload.ModBy = userIDStr
	payload.ModDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.TimKerja{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

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
	item.ModDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.TimKerja{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *TimKerjaService) Delete(id string, c *gin.Context) error {
	var item domain.TimKerja
	if err := s.db.Where("LookupId = ? AND CategoryId = ?", id, "B_POSISI").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tim kerja %s not found", id)
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

func (s *TimKerjaService) Lookup(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_TIMKERJA", page, limit, filters)
}

func (s *TimKerjaService) LookupSub(timkerjaValue string, filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	subQuery := s.db.Table("T_BUS_WORK_MAPPING").
		Where("T_BUS_WORK_MAPPING.SubDivisi = T_APP_LOOKUP.LookupValue").
		Where("divisi = ?", timkerjaValue).
		Where("T_BUS_WORK_MAPPING.Status = 1")
	return Lookup(s.db, "B_SUBKERJA", page, limit, filters, subQuery)
}

func (s *TimKerjaService) LookupReport(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_TIMKERJA_REPORT", page, limit, filters)
}
