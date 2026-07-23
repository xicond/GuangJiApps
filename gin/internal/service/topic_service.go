package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"
)

type TopicService struct {
	db       *gorm.DB
	resource string
}

func NewTopicService(db *gorm.DB) *TopicService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &TopicService{db: db, resource: "topics"}
}

func (s *TopicService) List() []domain.Topic {
	var items []domain.Topic
	if err := s.db.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil
	}
	return items
}

func (s *TopicService) Create(payload domain.Topic) (domain.Topic, error) {
	// 1. Validate based on your actual struct parameters
	if payload.TopicName == "" || payload.TopicCategory == "" {
		return domain.Topic{}, fmt.Errorf("Name and Category are required")
	}

	// 2. IMPORTANT: Do NOT generate a random UnixNano string for ID!
	// Your SQL Server schema defines [id] INT NOT NULL.
	// If it is NOT an IDENTITY column, we calculate the next integer sequence manually.
	var maxID int32
	s.db.Table("T_BUS_UMAT").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
	// payload.ID = maxID + 1

	// 3. Populate matching schema structural constraints
	payload.Status = true        // Active status mapping
	payload.ModAct = "I"         // 'I' standard legacy flag for Insert
	payload.ModBy = "1"          // Default system user ID matching INT type
	payload.ModDate = time.Now() // Local server time object

	// 4. Persist the new entity to the database pool
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Topic{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *TopicService) Get(id string) (domain.Topic, error) {
	var item domain.Topic
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.Topic{}, fmt.Errorf("topic %s not found", id)
	}
	return item, nil
}

func (s *TopicService) Update(id string, payload domain.Topic) (domain.Topic, error) {
	// 1. Cast string ID parameter safely to int32 to prevent MSSQL query crashes
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Topic{}, fmt.Errorf("invalid ID format: %w", err)
	}
	userIDInt32 := int32(parsedInt)

	var item domain.Topic
	// 2. Fetch the existing item using .Take() to avoid default sorting bugs
	if err := s.db.Where("id = ?", userIDInt32).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Topic{}, fmt.Errorf("umat %s not found", id)
		}
		return domain.Topic{}, err
	}

	// 3. Map values onto the actual field variables present in your legacy schema
	item.TopicName = payload.TopicName         // string
	item.TopicCategory = payload.TopicCategory // string
	item.Description = payload.Description     // string
	item.Status = payload.Status               // bool

	// Metadata
	item.ModDate = time.Now() // time.Time
	item.ModAct = "U"         // 'U' standard legacy flag for Update
	item.ModBy = "1"          // System user ID (int32)

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Topic{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *TopicService) Delete(id string) error {
	return s.db.Delete(&domain.Topic{}, "id = ?", id).Error
}
