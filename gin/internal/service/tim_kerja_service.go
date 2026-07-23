package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

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

func (s *TimKerjaService) List() []domain.TimKerja {
	var items []domain.TimKerja
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *TimKerjaService) Create(payload domain.TimKerja) (domain.TimKerja, error) {
	if payload.LookupValue == "" || payload.LookupDescription == "" {
		return domain.TimKerja{}, fmt.Errorf("lookup_value and lookup_description are required")
	}
	payload.ModAct = "I"
	payload.ModBy = "system"
	payload.ModDate = time.Now()
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.TimKerja{}, err
	}
	return payload, nil
}

func (s *TimKerjaService) Get(id string) (domain.TimKerja, error) {
	var item domain.TimKerja
	if err := s.db.First(&item, "LookupId = ?", id).Error; err != nil {
		return domain.TimKerja{}, fmt.Errorf("tim kerja %s not found", id)
	}
	return item, nil
}

func (s *TimKerjaService) Update(id string, payload domain.TimKerja) (domain.TimKerja, error) {
	var item domain.TimKerja
	if err := s.db.First(&item, "LookupId = ?", id).Error; err != nil {
		return domain.TimKerja{}, fmt.Errorf("tim kerja %s not found", id)
	}

	item.LookupValue = payload.LookupValue
	item.LookupDescription = payload.LookupDescription
	item.Status = payload.Status
	item.ModAct = "U"
	item.ModBy = "system"
	item.ModDate = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return domain.TimKerja{}, err
	}
	return item, nil
}

func (s *TimKerjaService) Delete(id string) error {
	return s.db.Delete(&domain.TimKerja{}, "LookupId = ?", id).Error
}
