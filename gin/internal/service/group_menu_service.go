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

func (s *GroupMenuService) List(page int, filters map[string]string, limit int) ([]domain.GroupMenuMapping, int64, error) {
	var items []domain.GroupMenuMapping
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_Login_Menu").Where("FlagActive = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"menu_name": {Column: "MenuName", IsLike: true},
		"page_url":  {Column: "PageUrl", IsLike: true},
		"parent_id": {Column: "ParentId", IsLike: false},
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
			Order("MenuId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("group menu mapping tidak ditemukan")
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
		return []domain.GroupMenuMapping{}, 0, findErr
	}

	return items, total, nil
}

func (s *GroupMenuService) Create(payload domain.GroupMenuMapping, c *gin.Context) (domain.GroupMenuMapping, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("validasi gagal: %w", err)
	}

	var maxID int32
	s.db.Table("T_Login_Menu").Select("ISNULL(MAX(MenuId), 0)").Row().Scan(&maxID)
	payload.MenuId = maxID + 1

	if payload.Sequence == 0 {
		payload.Sequence = 1
	}
	if payload.ParentLevel1 == 0 {
		payload.ParentLevel1 = 1
	}
	payload.FlagActive = true

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *GroupMenuService) Get(id string) (domain.GroupMenuMapping, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.GroupMenuMapping
	if err := s.db.Where("MenuId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.GroupMenuMapping{}, fmt.Errorf("group menu mapping %s not found", id)
		}
		return domain.GroupMenuMapping{}, err
	}
	return item, nil
}

func (s *GroupMenuService) Update(id string, payload domain.GroupMenuMapping, c *gin.Context) (domain.GroupMenuMapping, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.GroupMenuMapping
	if err := s.db.Where("MenuId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.GroupMenuMapping{}, fmt.Errorf("group menu mapping %s not found", id)
		}
		return domain.GroupMenuMapping{}, err
	}

	item.ParentId = payload.ParentId
	item.MenuName = payload.MenuName
	item.PageUrl = payload.PageUrl
	item.Sequence = payload.Sequence
	item.MenuDesc = payload.MenuDesc
	item.ParentLevel1 = payload.ParentLevel1
	item.FlagActive = payload.FlagActive

	if err := s.db.Save(&item).Error; err != nil {
		return domain.GroupMenuMapping{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *GroupMenuService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.GroupMenuMapping
	if err := s.db.Where("MenuId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("group menu mapping %s not found", id)
		}
		return err
	}

	item.FlagActive = false
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
