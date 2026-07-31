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

type KelasDonasiService struct {
	db       *gorm.DB
	resource string
}

func NewKelasDonasiService(db *gorm.DB) *KelasDonasiService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasDonasiService{db: db, resource: "kelas_donasi"}
}

func (s *KelasDonasiService) List(trxID string, page int, limit int) ([]domain.KelasDonasi, int64, error) {
	var items []domain.KelasDonasi
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasDonasi{}).Where("Status = ?", true)

	if trxID != "" {
		if parsedID, err := strconv.Atoi(trxID); err == nil {
			query = query.Where("TrxId = ?", parsedID)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return items, 0, fmt.Errorf("failed to count record: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return items, 0, fmt.Errorf("failed to list records: %w", err)
	}

	return items, total, nil
}

func (s *KelasDonasiService) Create(payload domain.KelasDonasi, c *gin.Context) (domain.KelasDonasi, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("Validation failed: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_DONASI").Select("ISNULL(MAX(DetailId), 0)").Row().Scan(&maxID)
		payload.DetailId = maxID + 1
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	statusTrue := true
	modActI := "I"
	now := time.Now()

	payload.Status = &statusTrue
	payload.ModAct = &modActI
	payload.ModBy = &userIDStr
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasDonasiService) Get(id string) (domain.KelasDonasi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasDonasi
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasDonasi{}, fmt.Errorf("kelas donasi %s not found", id)
		}
		return domain.KelasDonasi{}, err
	}
	return item, nil
}

func (s *KelasDonasiService) Update(id string, payload domain.KelasDonasi, c *gin.Context) (domain.KelasDonasi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasDonasi
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasDonasi{}, fmt.Errorf("kelas donasi %s not found", id)
		}
		return domain.KelasDonasi{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.Donatur != "" {
		item.Donatur = payload.Donatur
	}
	if payload.Donasi != 0 {
		item.Donasi = payload.Donasi
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userIDStr
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasDonasiService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasDonasi
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas donasi %s not found", id)
		}
		return err
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	statusFalse := false
	modActD := "D"
	now := time.Now()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userIDStr
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
