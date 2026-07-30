package service

import (
	"errors"
	"fmt"
	"sync"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type LookupService struct {
	db       *gorm.DB
	resource string
}

func NewLookupService(db *gorm.DB) *LookupService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &LookupService{db: db, resource: "lookup"}
}

func (s *LookupService) Lookup(categoryID string, page int, limit int, opts ...interface{}) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, categoryID, page, limit, opts...)
}

func (s *LookupService) LookupWaktuCiuTao(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_WAKTUCIUTAO", page, limit, filters)
}

func (s *LookupService) LookupGender(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_GENDER", page, limit, filters)
}

func (s *LookupService) LookupTcs(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_TCS", page, limit, filters)
}

func (s *LookupService) LookupFotang(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_FOTANG", page, limit, filters)
}

func (s *LookupService) LookupKelas(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELASKHUSUS", page, limit, filters)
}

func (s *LookupService) LookupPendidikan(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_PENDIDIKAN", page, limit, filters)
}

func (s *LookupService) LookupKelasUmum(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELASUMUM", page, limit, filters)
}

func (s *LookupService) LookupPekerjaan(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_PEKERJAAN", page, limit, filters)
}

func (s *LookupService) LookupKeluarga(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELUARGA", page, limit, filters)
}

func (s *LookupService) LookupKelasLevel(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KLS_LEVEL", page, limit, filters)
}

func (s *LookupService) LookupStatus(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_STATUS", page, limit, filters)
}

// Lookup fetches records from T_APP_LOOKUP for a given CategoryId with optional filters, pagination, and subQuery customizations.
// If limit <= 0, all records are returned without pagination (page is ignored).
// opts can optionally include a map[string]string (filters) and/or a *gorm.DB or func(*gorm.DB) *gorm.DB to customize the subQuery.
func Lookup(db *gorm.DB, categoryID string, page int, limit int, opts ...interface{}) ([]domain.AppLookup, int64, error) {
	if db == nil {
		return nil, 0, errors.New("database connection is nil")
	}

	var items []domain.AppLookup
	var total int64
	var filters map[string]string
	var customSubQuery interface{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		switch v := opt.(type) {
		case map[string]string:
			filters = v
		case *gorm.DB, func(*gorm.DB) *gorm.DB:
			customSubQuery = v
		}
	}

	subQuery := db.Table("T_APP_LOOKUPCATEGORY").
		Where("T_APP_LOOKUPCATEGORY.CategoryId = T_APP_LOOKUP.CategoryId").
		Where("T_APP_LOOKUPCATEGORY.Status = ?", true)

	if customSubQuery != nil {
		switch v := customSubQuery.(type) {
		case *gorm.DB:
			subQuery = v
		case func(*gorm.DB) *gorm.DB:
			subQuery = v(subQuery)
		}
	}

	query := db.Model(&domain.AppLookup{}).
		Where("T_APP_LOOKUP.CategoryId = ?", categoryID).
		Where("EXISTS (?)", subQuery).
		Where("T_APP_LOOKUP.Status = ?", true)

	if filters != nil {
		if val, ok := filters["lookup_description"]; ok && val != "" {
			query = query.Where("LookupDescription LIKE ?", "%"+val+"%")
		}
		if val, ok := filters["lookup_value"]; ok && val != "" {
			query = query.Where("LookupValue LIKE ?", "%"+val+"%")
		}
		if val, ok := filters["lookup_id"]; ok && val != "" {
			query = query.Where("LookupId = ?", val)
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
		q := query.Session(&gorm.Session{})
		if limit > 0 {
			if page <= 0 {
				page = 1
			}
			offset := (page - 1) * limit
			q = q.Limit(limit).Offset(offset)
			items = make([]domain.AppLookup, 0, limit)
		}
		if err := q.Order("LookupValue ASC").Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("lookup data tidak ditemukan")
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

	if items == nil {
		items = []domain.AppLookup{}
	}

	return items, total, nil
}
