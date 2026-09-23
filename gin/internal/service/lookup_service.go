package service

import (
	"errors"
	"fmt"
	"sync"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

// LookupService provides common master lookup retrieval methods for all lookup categories in the system.
type LookupService struct {
	db       *gorm.DB
	resource string
}

// NewLookupService initializes a new instance of LookupService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *LookupService: An initialized instance of LookupService.
func NewLookupService(db *gorm.DB) *LookupService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &LookupService{db: db, resource: "lookup"}
}

// Lookup fetches records from T_APP_LOOKUP for a specific category with pagination and options.
//
// Parameters:
//   - categoryID: Category identifier (e.g. "B_GENDER", "B_STATUS").
//   - page: Target page number (1-based).
//   - limit: Records per page.
//   - opts: Optional filter map (map[string]string) or custom subquery (*gorm.DB).
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count of matching records.
//   - error: Error if database query fails.
func (s *LookupService) Lookup(categoryID string, page int, limit int, opts ...interface{}) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, categoryID, page, limit, opts...)
}

// LookupWaktuCiuTao retrieves active initiation time period lookups (B_WAKTUCIUTAO).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupWaktuCiuTao(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_WAKTUCIUTAO", page, limit, filters)
}

// LookupGender retrieves active gender lookups (B_GENDER).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupGender(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_GENDER", page, limit, filters)
}

// LookupTcs retrieves active TCS (initiator) lookups (B_TCS).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupTcs(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_TCS", page, limit, filters)
}

// LookupFotang retrieves active temple lookups (B_FOTHANG).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupFotang(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_FOTHANG", page, limit, filters)
}

// LookupKelas retrieves active special class lookups (B_KELASKHUSUS).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupKelas(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELASKHUSUS", page, limit, filters)
}

// LookupPendidikan retrieves active education level lookups (B_PENDIDIKAN).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupPendidikan(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_PENDIDIKAN", page, limit, filters)
}

// LookupKelasUmum retrieves active general class lookups (B_KELASUMUM).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupKelasUmum(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELASUMUM", page, limit, filters)
}

// LookupPekerjaan retrieves active occupation lookups (B_PEKERJAAN).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupPekerjaan(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_PEKERJAAN", page, limit, filters)
}

// LookupKeluarga retrieves active family relationship lookups (B_KELUARGA).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupKeluarga(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KELUARGA", page, limit, filters)
}

// LookupKelasLevel retrieves active class level lookups (B_KLS_LEVEL).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupKelasLevel(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KLS_LEVEL", page, limit, filters)
}

// LookupStatus retrieves active status lookups (B_STATUS).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupStatus(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_STATUS", page, limit, filters)
}

// LookupKategoriTopic retrieves active topic category lookups (B_KATEGORI_TOPIK).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupKategoriTopic(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KATEGORI_TOPIK", page, limit, filters)
}

// LookupKategoriEvent retrieves active event category lookups (B_KATEGORI_EVENT).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupKategoriEvent(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_KATEGORI_EVENT", page, limit, filters)
}

// LookupTipeSumbangan retrieves active donation type lookups (SXY_TIPESUMBANGAN).
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count.
//   - error: Query error if any.
func (s *LookupService) LookupTipeSumbangan(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "SXY_TIPESUMBANGAN", page, limit, filters)
}

// Lookup fetches records from T_APP_LOOKUP for a given CategoryId with optional filters, pagination, and subQuery customizations.
// If limit <= 0, all records are returned without pagination (page is ignored).
// opts can optionally include a map[string]string (filters) and/or a *gorm.DB or func(*gorm.DB) *gorm.DB to customize the subQuery.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB).
//   - categoryID: Category identifier (e.g. "B_GENDER", "B_STATUS").
//   - page: Target page number (1-based index).
//   - limit: Maximum number of records per page (<= 0 for all).
//   - opts: Variadic options (filters map and/or subquery modifier).
//
// Returns:
//   - []domain.AppLookup: Slice of lookup records.
//   - int64: Total count of matching records.
//   - error: Error if database query fails.
func Lookup(db *gorm.DB, categoryID string, page int, limit int, opts ...interface{}) ([]domain.AppLookup, int64, error) {
	if db == nil {
		return nil, 0, errors.New("database connection is nil")
	}

	const defaultLimit = 10
	const maxLimit = 1000
	if limit <= 0 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
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
		Where("T_APP_LOOKUP.Status = ?", true).
		Order("LookupId")

	if filters != nil {
		var searchVal string
		for _, key := range []string{"search", "query", "q"} {
			if val, ok := filters[key]; ok && val != "" {
				searchVal = val
				break
			}
		}

		if searchVal != "" {
			query = query.Where("(T_APP_LOOKUP.LookupDescription LIKE ?)", "%"+searchVal+"%")
		} else {
			if val, ok := filters["lookup_description"]; ok && val != "" {
				query = query.Where("(T_APP_LOOKUP.LookupDescription LIKE ?)", "%"+val+"%")
			}

			if val, ok := filters["lookup_value"]; ok && val != "" {
				query = query.Where("(T_APP_LOOKUP.LookupValue = ?)", val)
			}

			if val, ok := filters["lookup_id"]; ok && val != "" {
				query = query.Where("(T_APP_LOOKUP.LookupId = ?)", val)
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
	})

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
