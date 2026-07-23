package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type PenggalangDanaService struct {
	db       *gorm.DB
	resource string
}

func NewPenggalangDanaService(db *gorm.DB) *PenggalangDanaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &PenggalangDanaService{db: db, resource: "penggalang-dana"}
}

func (s *PenggalangDanaService) List() []domain.PenggalangDana {
	var items []domain.PenggalangDana
	if err := s.db.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *PenggalangDanaService) Create(payload domain.PenggalangDana) (domain.PenggalangDana, error) {
	if payload.No == "" || payload.Nama == "" {
		return domain.PenggalangDana{}, fmt.Errorf("no and nama are required")
	}
	payload.Status = true
	payload.CreatedBy = 1
	payload.CreatedDate = time.Now()
	payload.UpdatedBy = 1
	payload.UpdatedDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.PenggalangDana{}, err
	}
	return payload, nil
}

func (s *PenggalangDanaService) Get(id string) (domain.PenggalangDana, error) {
	var item domain.PenggalangDana
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("penggalang dana %s not found", id)
	}
	return item, nil
}

func (s *PenggalangDanaService) Update(id string, payload domain.PenggalangDana) (domain.PenggalangDana, error) {
	var item domain.PenggalangDana
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("penggalang dana %s not found", id)
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
		return domain.PenggalangDana{}, err
	}
	return item, nil
}

func (s *PenggalangDanaService) Delete(id string) error {
	return s.db.Delete(&domain.PenggalangDana{}, "id = ?", id).Error
}
