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

type PenggalangDanaService struct {
	db       *gorm.DB
	resource string
}

func NewPenggalangDanaService(db *gorm.DB) *PenggalangDanaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &PenggalangDanaService{db: db, resource: "penggalang-dana"}
}

func (s *PenggalangDanaService) List(page int, filters map[string]string, limit int) ([]domain.PenggalangDana, int64, error) {
	var items []domain.PenggalangDana
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_SXY_MST_PENGGALANG").Where("STATUS = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"no":       {Column: "no", IsLike: true},
		"nama":     {Column: "nama", IsLike: true},
		"mandarin": {Column: "mandarin", IsLike: true},
		"email":    {Column: "email", IsLike: true},
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
		Order("id ASC").
		Find(&items).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.PenggalangDana{}, 0, errors.New("penggalang dana tidak ditemukan")
		}
		return []domain.PenggalangDana{}, 0, fmt.Errorf("database error: %w", err)
	}

	return items, total, nil
}

func (s *PenggalangDanaService) Create(payload domain.PenggalangDana, c *gin.Context) (domain.PenggalangDana, error) {
	if payload.No == "" || payload.Nama == "" {
		return domain.PenggalangDana{}, fmt.Errorf("no and nama are required")
	}

	var maxID int32
	s.db.Table("T_SXY_MST_PENGGALANG").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
	payload.ID = maxID + 1

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	payload.Status = true
	payload.CreatedBy = userID
	payload.CreatedDate = time.Now()
	payload.UpdatedBy = userID
	payload.UpdatedDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *PenggalangDanaService) Get(id string) (domain.PenggalangDana, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.PenggalangDana
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PenggalangDana{}, fmt.Errorf("penggalang dana %s not found", id)
		}
		return domain.PenggalangDana{}, err
	}
	return item, nil
}

func (s *PenggalangDanaService) Update(id string, payload domain.PenggalangDana, c *gin.Context) (domain.PenggalangDana, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.PenggalangDana
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PenggalangDana{}, fmt.Errorf("penggalang dana %s not found", id)
		}
		return domain.PenggalangDana{}, err
	}

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.No = payload.No
	item.Nama = payload.Nama
	item.Mandarin = payload.Mandarin
	item.Keterangan = payload.Keterangan
	item.LookupFothang = payload.LookupFothang
	item.Alamat = payload.Alamat
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile
	item.Email = payload.Email
	item.UpdatedBy = userID
	item.UpdatedDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *PenggalangDanaService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.PenggalangDana
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("penggalang dana %s not found", id)
		}
		return err
	}

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.Status = false
	item.UpdatedBy = userID
	item.UpdatedDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
