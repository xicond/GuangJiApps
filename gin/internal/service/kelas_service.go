package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

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

func (s *KelasService) List() []domain.Kelas {
	var items []domain.Kelas
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *KelasService) Create(payload domain.Kelas) (domain.Kelas, error) {
	if payload.LookupValue == "" || payload.LookupDescription == "" {
		return domain.Kelas{}, fmt.Errorf("lookup_value and lookup_description are required")
	}
	payload.ModAct = "I"
	payload.ModBy = "system"
	payload.ModDate = time.Now()
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Kelas{}, err
	}
	return payload, nil
}

func (s *KelasService) Get(id string) (domain.Kelas, error) {
	var item domain.Kelas
	if err := s.db.First(&item, "LookupId = ?", id).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
	}
	return item, nil
}

func (s *KelasService) Update(id string, payload domain.Kelas) (domain.Kelas, error) {
	var item domain.Kelas
	if err := s.db.First(&item, "LookupId = ?", id).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
	}

	item.LookupValue = payload.LookupValue
	item.LookupDescription = payload.LookupDescription
	item.Status = payload.Status
	item.ModAct = "U"
	item.ModBy = "system"
	item.ModDate = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Kelas{}, err
	}
	return item, nil
}

func (s *KelasService) Delete(id string) error {
	return s.db.Delete(&domain.Kelas{}, "LookupId = ?", id).Error
}
