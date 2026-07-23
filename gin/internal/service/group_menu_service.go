package service

import (
	"fmt"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type GroupMenuService struct {
	db       *gorm.DB
	resource string
}

func NewGroupMenuService(db *gorm.DB) *GroupMenuService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &GroupMenuService{db: db, resource: "group-menu-mappings"}
}

func (s *GroupMenuService) List() []domain.GroupMenuMapping {
	var items []domain.GroupMenuMapping
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *GroupMenuService) Create(payload domain.GroupMenuMapping) (domain.GroupMenuMapping, error) {
	if payload.MenuName == "" || payload.PageUrl == "" {
		return domain.GroupMenuMapping{}, fmt.Errorf("menu_name and page_url are required")
	}
	if payload.Sequence == 0 {
		payload.Sequence = 1
	}
	if payload.ParentLevel1 == 0 {
		payload.ParentLevel1 = 1
	}
	payload.FlagActive = true
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.GroupMenuMapping{}, err
	}
	return payload, nil
}

func (s *GroupMenuService) Get(id string) (domain.GroupMenuMapping, error) {
	var item domain.GroupMenuMapping
	if err := s.db.First(&item, "MenuId = ?", id).Error; err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("group menu mapping %s not found", id)
	}
	return item, nil
}

func (s *GroupMenuService) Update(id string, payload domain.GroupMenuMapping) (domain.GroupMenuMapping, error) {
	var item domain.GroupMenuMapping
	if err := s.db.First(&item, "MenuId = ?", id).Error; err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("group menu mapping %s not found", id)
	}

	item.ParentId = payload.ParentId
	item.MenuName = payload.MenuName
	item.PageUrl = payload.PageUrl
	item.Sequence = payload.Sequence
	item.MenuDesc = payload.MenuDesc
	item.ParentLevel1 = payload.ParentLevel1
	item.FlagActive = payload.FlagActive
	if err := s.db.Save(&item).Error; err != nil {
		return domain.GroupMenuMapping{}, err
	}
	return item, nil
}

func (s *GroupMenuService) Delete(id string) error {
	return s.db.Delete(&domain.GroupMenuMapping{}, "MenuId = ?", id).Error
}
