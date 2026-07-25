package service

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DonasiSxyService struct {
	db       *gorm.DB
	resource string
}

func NewDonasiSxyService(db *gorm.DB) *DonasiSxyService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &DonasiSxyService{db: db, resource: "donasi-sxy"}
}

func (s *DonasiSxyService) List(page int, filters map[string]string, limit int) ([]domain.DonasiSxy, int64, error) {
	var items []domain.DonasiSxy
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_SXY_TRANSAKSI").Where("STATUS = ?", true)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"no_kwitansi":   {Column: "nokwitansi", IsLike: true},
		"no_kupon":      {Column: "nokupon", IsLike: true},
		"donatur_id":    {Column: "donatur", IsLike: false},
		"penggalang_id": {Column: "penggalang", IsLike: false},
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
			Order("id ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("donasi sxy tidak ditemukan")
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
		return []domain.DonasiSxy{}, 0, findErr
	}

	return items, total, nil
}

func (s *DonasiSxyService) Create(payload domain.DonasiSxy, c *gin.Context) (domain.DonasiSxy, error) {
	if payload.NoKwitansi == "" {
		return domain.DonasiSxy{}, fmt.Errorf("no_kwitansi is required")
	}

	var maxID int32
	s.db.Table("T_SXY_TRANSAKSI").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
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
		return domain.DonasiSxy{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *DonasiSxyService) Get(id string) (domain.DonasiSxy, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
		}
		return domain.DonasiSxy{}, err
	}
	return item, nil
}

func (s *DonasiSxyService) Update(id string, payload domain.DonasiSxy, c *gin.Context) (domain.DonasiSxy, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
		}
		return domain.DonasiSxy{}, err
	}

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.NoKwitansi = payload.NoKwitansi
	item.Tanggal = payload.Tanggal
	item.Donatur = payload.Donatur
	item.Penggalang = payload.Penggalang
	item.Jumlah = payload.Jumlah
	item.TipeSumbangan = payload.TipeSumbangan
	item.NoKupon = payload.NoKupon
	item.Keterangan = payload.Keterangan
	item.TanggalTransfer = payload.TanggalTransfer
	item.AtasNama = payload.AtasNama
	item.TtkSent = payload.TtkSent
	item.UpdatedBy = userID
	item.UpdatedDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *DonasiSxyService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("donasi sxy %s not found", id)
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
