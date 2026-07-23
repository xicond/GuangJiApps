package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type AdminSubWarehouseService struct {
	db       *gorm.DB
	resource string
}

func NewAdminSubWarehouseService(db *gorm.DB) *AdminSubWarehouseService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &AdminSubWarehouseService{db: db, resource: "admin-sub-warehouses"}
}

func (s *AdminSubWarehouseService) List() []domain.AdminSubWarehouse {
	var items []domain.AdminSubWarehouse
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *AdminSubWarehouseService) Create(payload domain.AdminSubWarehouse) (domain.AdminSubWarehouse, error) {
	if payload.FullName == "" {
		return domain.AdminSubWarehouse{}, fmt.Errorf("full_name is required")
	}
	if payload.SubWhType == "" {
		return domain.AdminSubWarehouse{}, fmt.Errorf("sub_wh_type is required")
	}
	payload.LstUpdate = time.Now()
	if payload.CruId == 0 {
		payload.CruId = 1
	}
	if payload.UpdateUId == 0 {
		payload.UpdateUId = 1
	}

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AdminSubWarehouse{}, err
	}
	return payload, nil
}

func (s *AdminSubWarehouseService) Get(id string) (domain.AdminSubWarehouse, error) {
	var item domain.AdminSubWarehouse
	if err := s.db.First(&item, "SubWhId = ?", id).Error; err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("admin sub warehouse %s not found", id)
	}
	return item, nil
}

func (s *AdminSubWarehouseService) Update(id string, payload domain.AdminSubWarehouse) (domain.AdminSubWarehouse, error) {
	var item domain.AdminSubWarehouse
	if err := s.db.First(&item, "SubWhId = ?", id).Error; err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("admin sub warehouse %s not found", id)
	}

	item.WhId = payload.WhId
	item.FullName = payload.FullName
	item.Pic = payload.Pic
	item.DocCode = payload.DocCode
	item.CruId = payload.CruId
	item.UpdateUId = payload.UpdateUId
	item.LstUpdate = time.Now()
	item.FlagProductions = payload.FlagProductions
	item.SubWhType = payload.SubWhType
	if err := s.db.Save(&item).Error; err != nil {
		return domain.AdminSubWarehouse{}, err
	}
	return item, nil
}

func (s *AdminSubWarehouseService) Delete(id string) error {
	return s.db.Delete(&domain.AdminSubWarehouse{}, "SubWhId = ?", id).Error
}
