package service

import (
	"fmt"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type AdminGroupService struct {
	db       *gorm.DB
	resource string
}

func NewAdminGroupService(db *gorm.DB) *AdminGroupService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &AdminGroupService{db: db, resource: "admin-groups"}
}

func (s *AdminGroupService) List() []domain.AdminGroup {
	var items []domain.AdminGroup
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *AdminGroupService) Create(payload domain.AdminGroup) (domain.AdminGroup, error) {
	if payload.GroupName == "" {
		return domain.AdminGroup{}, fmt.Errorf("group_name is required")
	}
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AdminGroup{}, err
	}
	return payload, nil
}

func (s *AdminGroupService) Get(id string) (domain.AdminGroup, error) {
	var item domain.AdminGroup
	if err := s.db.First(&item, "GroupId = ?", id).Error; err != nil {
		return domain.AdminGroup{}, fmt.Errorf("admin group %s not found", id)
	}
	return item, nil
}

func (s *AdminGroupService) Update(id string, payload domain.AdminGroup) (domain.AdminGroup, error) {
	var item domain.AdminGroup
	if err := s.db.First(&item, "GroupId = ?", id).Error; err != nil {
		return domain.AdminGroup{}, fmt.Errorf("admin group %s not found", id)
	}

	item.GroupName = payload.GroupName
	item.RInsert = payload.RInsert
	item.REdit = payload.REdit
	item.RDelete = payload.RDelete
	item.RReporting = payload.RReporting
	item.RPositionId = payload.RPositionId
	item.GroupDesc = payload.GroupDesc
	if err := s.db.Save(&item).Error; err != nil {
		return domain.AdminGroup{}, err
	}
	return item, nil
}

func (s *AdminGroupService) Delete(id string) error {
	return s.db.Delete(&domain.AdminGroup{}, "GroupId = ?", id).Error
}
