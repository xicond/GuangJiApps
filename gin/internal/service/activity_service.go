package service

import (
	"fmt"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type ActivityService struct {
	db       *gorm.DB
	resource string
}

func NewActivityService(db *gorm.DB) *ActivityService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &ActivityService{db: db, resource: "activities"}
}

func (s *ActivityService) List() []domain.Activity {
	var items []domain.Activity
	if err := s.db.Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *ActivityService) Create(payload domain.Activity) (domain.Activity, error) {
	if payload.EventCode == "" || payload.EventName == "" {
		return domain.Activity{}, fmt.Errorf("event_code and event_name are required")
	}
	payload.ModAct = "I"
	payload.ModBy = "system"
	payload.ModDate = time.Now()
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Activity{}, err
	}
	return payload, nil
}

func (s *ActivityService) Get(id string) (domain.Activity, error) {
	var item domain.Activity
	if err := s.db.First(&item, "EventCode = ?", id).Error; err != nil {
		return domain.Activity{}, fmt.Errorf("activity %s not found", id)
	}
	return item, nil
}

func (s *ActivityService) Update(id string, payload domain.Activity) (domain.Activity, error) {
	var item domain.Activity
	if err := s.db.First(&item, "EventCode = ?", id).Error; err != nil {
		return domain.Activity{}, fmt.Errorf("activity %s not found", id)
	}

	item.EventCode = payload.EventCode
	item.EventName = payload.EventName
	item.EventCategory = payload.EventCategory
	item.Description = payload.Description
	item.Status = payload.Status
	item.ModAct = "U"
	item.ModBy = "system"
	item.ModDate = time.Now()
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Activity{}, err
	}
	return item, nil
}

func (s *ActivityService) Delete(id string) error {
	return s.db.Delete(&domain.Activity{}, "EventCode = ?", id).Error
}
