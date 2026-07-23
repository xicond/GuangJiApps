package service

import (
	"fmt"

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

func (s *AdminService) List() []domain.Admin {
	var items []domain.Admin
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *AdminService) Create(payload domain.Admin) (domain.Admin, error) {
	if payload.Username == "" {
		return domain.Admin{}, fmt.Errorf("username is required")
	}
	if payload.GroupId == 0 {
		return domain.Admin{}, fmt.Errorf("group_id is required")
	}
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Admin{}, err
	}
	return payload, nil
}

func (s *AdminService) Get(id string) (domain.Admin, error) {
	var item domain.Admin
	if err := s.db.First(&item, "LoginId = ?", id).Error; err != nil {
		return domain.Admin{}, fmt.Errorf("admin %s not found", id)
	}
	return item, nil
}

func (s *AdminService) Update(id string, payload domain.Admin) (domain.Admin, error) {
	var item domain.Admin
	if err := s.db.First(&item, "LoginId = ?", id).Error; err != nil {
		return domain.Admin{}, fmt.Errorf("admin %s not found", id)
	}

	item.Username = payload.Username
	item.Email = payload.Email
	item.GroupId = payload.GroupId
	item.PhoneNumber = payload.PhoneNumber
	item.ImgUrl = payload.ImgUrl
	item.FlagUse = payload.FlagUse
	item.DateStart = payload.DateStart
	item.DateEnd = payload.DateEnd
	item.LoginDesc = payload.LoginDesc
	item.LastLogin = payload.LastLogin
	item.DepartmentId = payload.DepartmentId
	item.IsWarehouse = payload.IsWarehouse
	if payload.Password != "" {
		item.Password = payload.Password
	}
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Admin{}, err
	}
	return item, nil
}

func (s *AdminService) Delete(id string) error {
	return s.db.Delete(&domain.Admin{}, "LoginId = ?", id).Error
}
