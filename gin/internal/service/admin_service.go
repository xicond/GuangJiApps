package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type AdminService struct {
	db       *gorm.DB
	resource string
}

func NewAdminService(db *gorm.DB) *AdminService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &AdminService{db: db, resource: "admins"}
}

func (s *AdminService) List() []domain.Resource {
	var items []domain.Resource
	if err := s.db.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *AdminService) Create(payload domain.Resource) (domain.Resource, error) {
	if payload.Code == "" || payload.Name == "" {
		return domain.Resource{}, fmt.Errorf("code and name are required")
	}
	payload.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	payload.Category = s.resource
	payload.Active = true
	payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	payload.UpdatedAt = payload.CreatedAt
	payload.CreatedBy = "system"
	payload.UpdatedBy = "system"

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Resource{}, err
	}
	return payload, nil
}

func (s *AdminService) Get(id string) (domain.Resource, error) {
	var item domain.Resource
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.Resource{}, fmt.Errorf("admin %s not found", id)
	}
	return item, nil
}

func (s *AdminService) Update(id string, payload domain.Resource) (domain.Resource, error) {
	var item domain.Resource
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.Resource{}, fmt.Errorf("admin %s not found", id)
	}

	item.Code = payload.Code
	item.Name = payload.Name
	item.Active = payload.Active
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	item.UpdatedBy = "system"
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Resource{}, err
	}
	return item, nil
}

func (s *AdminService) Delete(id string) error {
	return s.db.Delete(&domain.Resource{}, "id = ?", id).Error
}
