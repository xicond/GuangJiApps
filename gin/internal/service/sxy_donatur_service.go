package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type SxyDonaturService struct {
	db       *gorm.DB
	resource string
}

func NewSxyDonaturService(db *gorm.DB) *SxyDonaturService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &SxyDonaturService{db: db, resource: "sxy-donatur"}
}

func (s *SxyDonaturService) List() []domain.SxyDonatur {
	var items []domain.SxyDonatur
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *SxyDonaturService) Create(payload domain.SxyDonatur) (domain.SxyDonatur, error) {
	if payload.No == "" || payload.Nama == "" {
		return domain.SxyDonatur{}, fmt.Errorf("no and nama are required")
	}
	payload.Status = true
	payload.CreatedBy = 1
	payload.CreatedDate = time.Now()
	payload.UpdatedBy = 1
	payload.UpdatedDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.SxyDonatur{}, err
	}
	return payload, nil
}

func (s *SxyDonaturService) Get(id string) (domain.SxyDonatur, error) {
	var item domain.SxyDonatur
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("sxy donatur %s not found", id)
	}
	return item, nil
}

func (s *SxyDonaturService) Update(id string, payload domain.SxyDonatur) (domain.SxyDonatur, error) {
	var item domain.SxyDonatur
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("sxy donatur %s not found", id)
	}

	item.No = payload.No
	item.Nama = payload.Nama
	item.Mandarin = payload.Mandarin
	item.Keterangan = payload.Keterangan
	item.LookupFothang = payload.LookupFothang
	item.Alamat = payload.Alamat
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile
	item.Email = payload.Email
	item.Status = payload.Status
	item.UpdatedBy = 1
	item.UpdatedDate = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return domain.SxyDonatur{}, err
	}
	return item, nil
}

func (s *SxyDonaturService) Delete(id string) error {
	return s.db.Delete(&domain.SxyDonatur{}, "id = ?", id).Error
}
