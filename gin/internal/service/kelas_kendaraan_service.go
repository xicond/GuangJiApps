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

type KelasKendaraanService struct {
	db       *gorm.DB
	resource string
}

func NewKelasKendaraanService(db *gorm.DB) *KelasKendaraanService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasKendaraanService{db: db, resource: "kelas_kendaraan"}
}

func (s *KelasKendaraanService) List(trxID string, page int, limit int) ([]domain.KelasKendaraan, int64, error) {
	var items []domain.KelasKendaraan
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasKendaraan{}).Where("status = ?", true)

	if trxID != "" {
		if parsedID, err := strconv.Atoi(trxID); err == nil {
			query = query.Where("trxid = ?", parsedID)
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

func (s *KelasKendaraanService) Create(payload domain.KelasKendaraan, c *gin.Context) (domain.KelasKendaraan, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasKendaraan{}, fmt.Errorf("Validation failed: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_KENDARAAN").Select("ISNULL(MAX(detailid), 0)").Row().Scan(&maxID)
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
		return domain.KelasKendaraan{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasKendaraanService) Get(id string) (domain.KelasKendaraan, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasKendaraan{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasKendaraan
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasKendaraan{}, fmt.Errorf("kelas kendaraan %s not found", id)
		}
		return domain.KelasKendaraan{}, err
	}
	return item, nil
}

func (s *KelasKendaraanService) Update(id string, payload domain.KelasKendaraan, c *gin.Context) (domain.KelasKendaraan, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasKendaraan{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasKendaraan
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasKendaraan{}, fmt.Errorf("kelas kendaraan %s not found", id)
		}
		return domain.KelasKendaraan{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.NoPolisi != nil {
		item.NoPolisi = payload.NoPolisi
	}
	if payload.Pengendara != nil {
		item.Pengendara = payload.Pengendara
	}
	if payload.TipeKendaraan != nil {
		item.TipeKendaraan = payload.TipeKendaraan
	}
	if payload.Fotang != nil {
		item.Fotang = payload.Fotang
	}
	if payload.Hari != nil {
		item.Hari = payload.Hari
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
		return domain.KelasKendaraan{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasKendaraanService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasKendaraan
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas kendaraan %s not found", id)
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
