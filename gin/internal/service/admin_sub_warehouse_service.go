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

func (s *AdminSubWarehouseService) List(page int, filters map[string]string, limit int) ([]domain.AdminSubWarehouse, int64, error) {
	var items []domain.AdminSubWarehouse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_WH_SUBWH_MST")

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"full_name":   {Column: "FULL_NAME", IsLike: true},
		"pic":         {Column: "PIC", IsLike: true},
		"doc_code":    {Column: "DOCCODE", IsLike: true},
		"sub_wh_type": {Column: "SubWhType", IsLike: false},
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

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	})

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).
			Limit(limit).
			Offset(offset).
			Order("SUBWHID ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("admin sub warehouse tidak ditemukan")
			} else {
				findErr = fmt.Errorf("database error: %w", err)
			}
		}
	})

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if findErr != nil {
		return []domain.AdminSubWarehouse{}, 0, findErr
	}

	return items, total, nil
}

func (s *AdminSubWarehouseService) Create(payload domain.AdminSubWarehouse, c *gin.Context) (domain.AdminSubWarehouse, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("Validation failed: %w", err)
	}

	var maxID int32
	s.db.Table("T_WH_SUBWH_MST").Select("ISNULL(MAX(SUBWHID), 0)").Row().Scan(&maxID)
	payload.SubWhId = maxID + 1

	userID := int64(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int64(uid)
			}
		}
	}

	payload.CruId = userID
	payload.UpdateUId = userID
	payload.LstUpdate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *AdminSubWarehouseService) Get(id string) (domain.AdminSubWarehouse, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.AdminSubWarehouse
	if err := s.db.Where("SUBWHID = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminSubWarehouse{}, fmt.Errorf("admin sub warehouse %s not found", id)
		}
		return domain.AdminSubWarehouse{}, err
	}
	return item, nil
}

func (s *AdminSubWarehouseService) Update(id string, payload domain.AdminSubWarehouse, c *gin.Context) (domain.AdminSubWarehouse, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminSubWarehouse
	if err := s.db.Where("SUBWHID = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AdminSubWarehouse{}, fmt.Errorf("admin sub warehouse %s not found", id)
		}
		return domain.AdminSubWarehouse{}, err
	}

	item.WhId = payload.WhId
	item.FullName = payload.FullName
	item.Pic = payload.Pic
	item.DocCode = payload.DocCode
	item.UpdateUId = toInt64(getUserID(c))
	item.LstUpdate = domain.NowDateTime()
	item.FlagProductions = payload.FlagProductions
	item.SubWhType = payload.SubWhType

	if err := s.db.Save(&item).Error; err != nil {
		return domain.AdminSubWarehouse{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *AdminSubWarehouseService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.AdminSubWarehouse
	if err := s.db.Where("SUBWHID = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin sub warehouse %s not found", id)
		}
		return err
	}

	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
