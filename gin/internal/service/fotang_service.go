package service

import (
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
	return Lookup(s.db, "B_FOTHANG", page, limit, filters)
}
