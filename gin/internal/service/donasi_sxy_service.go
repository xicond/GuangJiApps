package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type DonasiSxyService struct {
	db       *gorm.DB
	resource string
}

func NewDonasiSxyService(db *gorm.DB) *DonasiSxyService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &DonasiSxyService{db: db, resource: "donasi-sxy"}
}

func (s *DonasiSxyService) List() []domain.DonasiSxy {
	var items []domain.DonasiSxy
	if err := s.db.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *DonasiSxyService) Create(payload domain.DonasiSxy) (domain.DonasiSxy, error) {
	if payload.NoKwitansi == "" {
		return domain.DonasiSxy{}, fmt.Errorf("no_kwitansi is required")
	}
	payload.Status = true
	payload.CreatedBy = 1
	payload.CreatedDate = time.Now()
	payload.UpdatedBy = 1
	payload.UpdatedDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.DonasiSxy{}, err
	}
	return payload, nil
}

func (s *DonasiSxyService) Get(id string) (domain.DonasiSxy, error) {
	var item domain.DonasiSxy
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
	}
	return item, nil
}

func (s *DonasiSxyService) Update(id string, payload domain.DonasiSxy) (domain.DonasiSxy, error) {
	var item domain.DonasiSxy
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
	}

	item.NoKwitansi = payload.NoKwitansi
	item.Tanggal = payload.Tanggal
	item.Donatur = payload.Donatur
	item.Penggalang = payload.Penggalang
	item.Jumlah = payload.Jumlah
	item.TipeSumbangan = payload.TipeSumbangan
	item.NoKupon = payload.NoKupon
	item.Keterangan = payload.Keterangan
	item.Status = payload.Status
	item.TanggalTransfer = payload.TanggalTransfer
	item.AtasNama = payload.AtasNama
	item.TtkSent = payload.TtkSent
	item.UpdatedBy = 1
	item.UpdatedDate = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return domain.DonasiSxy{}, err
	}
	return item, nil
}

func (s *DonasiSxyService) Delete(id string) error {
	return s.db.Delete(&domain.DonasiSxy{}, "id = ?", id).Error
}
