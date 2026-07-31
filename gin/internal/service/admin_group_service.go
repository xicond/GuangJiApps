package service

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
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

func (s *AdminGroupService) List(page int, filters map[string]string, limit int) ([]domain.AdminGroup, int64, error) {
	var items []domain.AdminGroup
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_Login_Group")

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"group_name": {Column: "GroupName", IsLike: true},
		// "group_desc": {Column: "GroupDesc", IsLike: true},
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if rule, exists := allowedFilters[field]; exists {
			if rule.IsLike {
				query = query.Where(fmt.Sprintf("[%s] LIKE ?", rule.Column), "%"+value+"%")
			} else {
				query = query.Where(fmt.Sprintf("[%s] = ?", rule.Column), value)
			}
		}
	}

	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).
			Limit(limit).
			Offset(offset).
			Order("GroupId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("admin group tidak ditemukan")
			} else {
				findErr = fmt.Errorf("database error: %w", err)
			}
		}
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if findErr != nil {
		return []domain.AdminGroup{}, 0, findErr
	}

	return items, total, nil
}

func (s *AdminGroupService) Create(payload domain.AdminGroup, c *gin.Context) (domain.AdminGroup, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.AdminGroup{}, fmt.Errorf("Validation failed: %w", err)
	}

	var maxID int32
	s.db.Table("T_Login_Group").Select("ISNULL(MAX(GroupId), 0)").Row().Scan(&maxID)
	payload.GroupId = maxID + 1

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AdminGroup{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *AdminGroupService) Get(id string) (domain.AdminGroup, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminGroup{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.AdminGroup
	if err := s.db.Where("GroupId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminGroup{}, fmt.Errorf("admin group %s not found", id)
		}
		return domain.AdminGroup{}, err
	}
	return item, nil
}

func (s *AdminGroupService) Update(id string, payload domain.AdminGroup, c *gin.Context) (domain.AdminGroup, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminGroup{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminGroup
	if err := s.db.Where("GroupId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminGroup{}, fmt.Errorf("admin group %s not found", id)
		}
		return domain.AdminGroup{}, err
	}

	item.GroupName = payload.GroupName
	item.RInsert = payload.RInsert
	item.REdit = payload.REdit
	item.RDelete = payload.RDelete
	item.RReporting = payload.RReporting
	item.RPositionId = payload.RPositionId
	item.GroupDesc = payload.GroupDesc

	if err := s.db.Save(&item).Error; err != nil {
		return domain.AdminGroup{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *AdminGroupService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminGroup
	if err := s.db.Where("GroupId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin group %s not found", id)
		}
		return err
	}

	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
