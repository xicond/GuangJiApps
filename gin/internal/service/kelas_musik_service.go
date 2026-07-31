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

type KelasMusikService struct {
	db       *gorm.DB
	resource string
}

func NewKelasMusikService(db *gorm.DB) *KelasMusikService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasMusikService{db: db, resource: "kelas_musik"}
}

func (s *KelasMusikService) List(trxID string, page int, limit int) ([]domain.KelasMusik, int64, error) {
	var items []domain.KelasMusik
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasMusik{}).Where("status = ?", true)

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

func (s *KelasMusikService) Create(payload domain.KelasMusik, c *gin.Context) (domain.KelasMusik, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasMusik{}, fmt.Errorf("Validation failed: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_MUSIK").Select("ISNULL(MAX(DetailId), 0)").Row().Scan(&maxID)
		payload.DetailId = maxID + 1
	}

	userID := getUserID(c)
	statusTrue := true
	modActI := "I"
	now := time.Now()

	payload.Status = &statusTrue
	payload.ModAct = &modActI
	payload.ModBy = &userID
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.KelasMusik{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasMusikService) Get(id string) (domain.KelasMusik, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasMusik{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasMusik
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasMusik{}, fmt.Errorf("kelas musik %s not found", id)
		}
		return domain.KelasMusik{}, err
	}
	return item, nil
}

func (s *KelasMusikService) Update(id string, payload domain.KelasMusik, c *gin.Context) (domain.KelasMusik, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasMusik{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasMusik
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasMusik{}, fmt.Errorf("kelas musik %s not found", id)
		}
		return domain.KelasMusik{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.MusicId != 0 {
		item.MusicId = payload.MusicId
	}
	if payload.Urutan != nil {
		item.Urutan = payload.Urutan
	}
	if payload.MusicDate != nil {
		item.MusicDate = payload.MusicDate
	}
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}

	userID := getUserID(c)
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasMusik{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasMusikService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasMusik
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas musik %s not found", id)
		}
		return err
	}

	userID := getUserID(c)
	statusFalse := false
	modActD := "D"
	now := time.Now()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
