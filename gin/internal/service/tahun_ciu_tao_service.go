package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

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

func (s *TahunCiuTaoService) List() []domain.TahunCiuTao {
	var items []domain.TahunCiuTao
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *TahunCiuTaoService) Create(payload domain.TahunCiuTao) (domain.TahunCiuTao, error) {
	if payload.TahunMandarin == "" {
		return domain.TahunCiuTao{}, fmt.Errorf("tahun_mandarin is required")
	}
	payload.ModAct = "I"
	payload.ModBy = "system"
	payload.ModDate = time.Now()
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.TahunCiuTao{}, err
	}
	return payload, nil
}

func (s *TahunCiuTaoService) Get(id string) (domain.TahunCiuTao, error) {
	var item domain.TahunCiuTao
	if err := s.db.First(&item, "TahunMandarin = ?", id).Error; err != nil {
		return domain.TahunCiuTao{}, fmt.Errorf("tahun ciu tao %s not found", id)
	}
	return item, nil
}

func (s *TahunCiuTaoService) Update(id string, payload domain.TahunCiuTao) (domain.TahunCiuTao, error) {
	var item domain.TahunCiuTao
	if err := s.db.First(&item, "TahunMandarin = ?", id).Error; err != nil {
		return domain.TahunCiuTao{}, fmt.Errorf("tahun ciu tao %s not found", id)
	}

	item.TahunMandarin = payload.TahunMandarin
	item.StartDate = payload.StartDate
	item.EndDate = payload.EndDate
	item.Description = payload.Description
	item.Status = payload.Status
	item.ModAct = "U"
	item.ModBy = "system"
	item.ModDate = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return domain.TahunCiuTao{}, err
	}
	return item, nil
}

func (s *TahunCiuTaoService) Delete(id string) error {
	return s.db.Delete(&domain.TahunCiuTao{}, "TahunMandarin = ?", id).Error
}
