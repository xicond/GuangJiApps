package service

import (
	"errors"
	"fmt"
	"sync"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type FotangService struct {
	db       *gorm.DB
	resource string
}

func NewFotangService(db *gorm.DB) *FotangService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &FotangService{db: db, resource: "kelas"}
}

func (s *FotangService) Lookup(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	var items []domain.AppLookup
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	subQuery := s.db.Table("T_APP_LOOKUPCATEGORY").
		Where("T_APP_LOOKUPCATEGORY.CategoryId = T_APP_LOOKUP.CategoryId")

	query := s.db.Model(&domain.AppLookup{}).
		Where("CategoryId = ?", "B_FOTHANG").
		Where("EXISTS (?)", subQuery)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]bool{
		"lookup_description": true,
		// "lookup_value":       true,
		// "lookup_id":          true,
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "lookup_description" {
				query = query.Where("LookupDescription LIKE ?", "%"+value+"%")
			} /*  else if field == "lookup_value" {
				query = query.Where("LookupValue LIKE ?", "%"+value+"%")
			} else if field == "lookup_id" {
				query = query.Where("LookupId = ?", value)
			} */
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
			Order("LookupValue ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("activity tidak ditemukan")
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
		return []domain.AppLookup{}, 0, findErr
	}

	return items, total, nil
}
