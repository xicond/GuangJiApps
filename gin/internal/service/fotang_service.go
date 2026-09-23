package service

import (
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

// FotangService provides temple (fotang) lookup helpers across B_FOTHANG and SXY_FOTHANG categories.
type FotangService struct {
	db       *gorm.DB
	resource string
}

// NewFotangService initializes a new instance of FotangService.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *FotangService: An initialized instance of FotangService.
func NewFotangService(db *gorm.DB) *FotangService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &FotangService{db: db, resource: "kelas"}
}

// Lookup retrieves active temple options under CategoryId = B_FOTHANG.
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of temple lookup records.
//   - int64: Total count of matching records.
//   - error: Error if query fails.
func (s *FotangService) Lookup(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "B_FOTHANG", page, limit, filters)
}

// LookupSxy retrieves active SXY temple options under CategoryId = SXY_FOTHANG.
//
// Parameters:
//   - filters: Filter criteria map.
//   - page: Page number (1-based).
//   - limit: Records per page.
//
// Returns:
//   - []domain.AppLookup: Slice of SXY temple lookup records.
//   - int64: Total count of matching records.
//   - error: Error if query fails.
func (s *FotangService) LookupSxy(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	return Lookup(s.db, "SXY_FOTHANG", page, limit, filters)
}
