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

type KelasDonasiBarangService struct {
	db       *gorm.DB
	resource string
}

func NewKelasDonasiBarangService(db *gorm.DB) *KelasDonasiBarangService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasDonasiBarangService{db: db, resource: "kelas_donasi_barang"}
}

func (s *KelasDonasiBarangService) List(trxID string, page int, limit int) ([]domain.KelasDonasiBarang, int64, error) {
	var items []domain.KelasDonasiBarang
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasDonasiBarang{}).Where("Status = ?", true)

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

func (s *KelasDonasiBarangService) Create(payload domain.KelasDonasiBarang, c *gin.Context) (domain.KelasDonasiBarang, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasDonasiBarang{}, fmt.Errorf("Validation failed: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_DONASI_BARANG").Select("ISNULL(MAX(DetailId), 0)").Row().Scan(&maxID)
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
		return domain.KelasDonasiBarang{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasDonasiBarangService) Get(id string) (domain.KelasDonasiBarang, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasDonasiBarang{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasDonasiBarang
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasDonasiBarang{}, fmt.Errorf("kelas donasi barang %s not found", id)
		}
		return domain.KelasDonasiBarang{}, err
	}
	return item, nil
}

func (s *KelasDonasiBarangService) Update(id string, payload domain.KelasDonasiBarang, c *gin.Context) (domain.KelasDonasiBarang, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasDonasiBarang{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasDonasiBarang
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasDonasiBarang{}, fmt.Errorf("kelas donasi barang %s not found", id)
		}
		return domain.KelasDonasiBarang{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.Donatur != "" {
		item.Donatur = payload.Donatur
	}
	if payload.Barang != "" {
		item.Barang = payload.Barang
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userIDStr
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasDonasiBarang{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasDonasiBarangService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasDonasiBarang
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas donasi barang %s not found", id)
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
