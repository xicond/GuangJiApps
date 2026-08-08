package service

import (
	"errors"
	"fmt"
	"strconv"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KelasPengeluaranService struct {
	db       *gorm.DB
	resource string
}

func NewKelasPengeluaranService(db *gorm.DB) *KelasPengeluaranService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasPengeluaranService{db: db, resource: "kelas_pengeluaran"}
}

func (s *KelasPengeluaranService) List(trxID string, page int, limit int) ([]domain.KelasPengeluaran, int64, error) {
	var items []domain.KelasPengeluaran
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasPengeluaran{}).Where("status = ?", true)

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

func (s *KelasPengeluaranService) Create(payload domain.KelasPengeluaran, c *gin.Context) (domain.KelasPengeluaran, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("Validation failed: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_PENGELUARAN").Select("ISNULL(MAX(detailid), 0)").Row().Scan(&maxID)
		payload.DetailId = maxID + 1
	}

	userID := getUserID(c)
	statusTrue := true
	modActI := "I"
	now := domain.NowDateTime()

	payload.Status = &statusTrue
	payload.ModAct = &modActI
	payload.ModBy = &userID
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasPengeluaranService) Get(id string) (domain.KelasPengeluaran, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasPengeluaran
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPengeluaran{}, fmt.Errorf("kelas pengeluaran %s not found", id)
		}
		return domain.KelasPengeluaran{}, err
	}
	return item, nil
}

func (s *KelasPengeluaranService) Update(id string, payload domain.KelasPengeluaran, c *gin.Context) (domain.KelasPengeluaran, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasPengeluaran
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPengeluaran{}, fmt.Errorf("kelas pengeluaran %s not found", id)
		}
		return domain.KelasPengeluaran{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.TimKerja != nil {
		item.TimKerja = payload.TimKerja
	}
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}
	if payload.Biaya != nil {
		item.Biaya = payload.Biaya
	}

	userID := getUserID(c)
	modActU := "U"
	now := domain.NowDateTime()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasPengeluaranService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasPengeluaran
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas pengeluaran %s not found", id)
		}
		return err
	}

	userID := getUserID(c)
	statusFalse := false
	modActD := "D"
	now := domain.NowDateTime()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
