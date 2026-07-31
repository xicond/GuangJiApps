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

type KelasAbsensiService struct {
	db       *gorm.DB
	resource string
}

func NewKelasAbsensiService(db *gorm.DB) *KelasAbsensiService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasAbsensiService{db: db, resource: "kelas_absensi"}
}

func (s *KelasAbsensiService) List(trxID string, page int, limit int) ([]domain.KelasAbsensi, int64, error) {
	var items []domain.KelasAbsensi
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasAbsensi{}).Where("Status = ?", true)

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

func (s *KelasAbsensiService) Create(payload domain.KelasAbsensi, c *gin.Context) (domain.KelasAbsensi, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasAbsensi{}, fmt.Errorf("Validation failed: %w", err)
	}

	if payload.Id == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_ABSENSI").Select("ISNULL(MAX(Id), 0)").Row().Scan(&maxID)
		payload.Id = maxID + 1
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
		return domain.KelasAbsensi{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasAbsensiService) Get(id string) (domain.KelasAbsensi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasAbsensi{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasAbsensi
	if err := s.db.Where("Id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasAbsensi{}, fmt.Errorf("kelas absensi %s not found", id)
		}
		return domain.KelasAbsensi{}, err
	}
	return item, nil
}

func (s *KelasAbsensiService) Update(id string, payload domain.KelasAbsensi, c *gin.Context) (domain.KelasAbsensi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasAbsensi{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasAbsensi
	if err := s.db.Where("Id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasAbsensi{}, fmt.Errorf("kelas absensi %s not found", id)
		}
		return domain.KelasAbsensi{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if !payload.TrxDate.IsZero() {
		item.TrxDate = payload.TrxDate
	}
	if payload.IdPeserta != 0 {
		item.IdPeserta = payload.IdPeserta
	}
	if payload.Status != nil {
		item.Status = payload.Status
	}

	userID := getUserID(c)
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasAbsensi{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasAbsensiService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasAbsensi
	if err := s.db.Where("Id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas absensi %s not found", id)
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
