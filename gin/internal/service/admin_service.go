package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
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

func (s *AdminService) List(page int, filters map[string]string, limit int) ([]domain.Admin, int64, error) {
	var items []domain.Admin
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_Login_Mst").Where("FlagUse = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"username":      {Column: "Username", IsLike: true},
		"email":         {Column: "email", IsLike: true},
		"phone_number":  {Column: "PhoneNumber", IsLike: true},
		"department_id": {Column: "DepartmentId", IsLike: false},
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("database count error: %w", err)
	}

	err := query.
		Limit(limit).
		Offset(offset).
		Order("LoginId ASC").
		Find(&items).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.Admin{}, 0, errors.New("admin tidak ditemukan")
		}
		return []domain.Admin{}, 0, fmt.Errorf("database error: %w", err)
	}

	return items, total, nil
}

func (s *AdminService) Create(payload domain.Admin, c *gin.Context) (domain.Admin, error) {
	if payload.Username == "" {
		return domain.Admin{}, fmt.Errorf("username is required")
	}
	if payload.GroupId == 0 {
		return domain.Admin{}, fmt.Errorf("group_id is required")
	}

	var maxID int32
	s.db.Table("T_Login_Mst").Select("ISNULL(MAX(LoginId), 0)").Row().Scan(&maxID)
	payload.ID = maxID + 1

	payload.FlagUse = true
	if payload.DateStart.IsZero() {
		payload.DateStart = time.Now()
	}
	if payload.DateEnd.IsZero() {
		payload.DateEnd = time.Now().AddDate(10, 0, 0)
	}

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Admin{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *AdminService) Get(id string) (domain.Admin, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Admin{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.Admin
	if err := s.db.Where("LoginId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Admin{}, fmt.Errorf("admin %s not found", id)
		}
		return domain.Admin{}, err
	}
	return item, nil
}

func (s *AdminService) Update(id string, payload domain.Admin, c *gin.Context) (domain.Admin, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Admin{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Admin
	if err := s.db.Where("LoginId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Admin{}, fmt.Errorf("admin %s not found", id)
		}
		return domain.Admin{}, err
	}

	item.Username = payload.Username
	item.Email = payload.Email
	item.GroupId = payload.GroupId
	item.PhoneNumber = payload.PhoneNumber
	item.ImgUrl = payload.ImgUrl
	item.FlagUse = payload.FlagUse
	item.LoginDesc = payload.LoginDesc
	item.DepartmentId = payload.DepartmentId
	item.IsWarehouse = payload.IsWarehouse

	if !payload.DateStart.IsZero() {
		item.DateStart = payload.DateStart
	}
	if !payload.DateEnd.IsZero() {
		item.DateEnd = payload.DateEnd
	}
	if payload.Password != "" {
		item.Password = payload.Password
	}

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Admin{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *AdminService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Admin
	if err := s.db.Where("LoginId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin %s not found", id)
		}
		return err
	}

	item.FlagUse = false
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
