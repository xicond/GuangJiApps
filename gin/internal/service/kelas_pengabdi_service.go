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

type KelasPengabdiService struct {
	db       *gorm.DB
	resource string
}

func NewKelasPengabdiService(db *gorm.DB) *KelasPengabdiService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasPengabdiService{db: db, resource: "kelas_pengabdi"}
}

func (s *KelasPengabdiService) List(trxID string, page int, limit int) ([]domain.KelasPengabdi, int64, error) {
	var items []domain.KelasPengabdi
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasPengabdi{}).Where("status = ?", true)

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

func (s *KelasPengabdiService) Create(payload domain.KelasPengabdi, c *gin.Context) (domain.KelasPengabdi, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasPengabdi{}, fmt.Errorf("validasi gagal: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_PENGABDI").Select("ISNULL(MAX(detailid), 0)").Row().Scan(&maxID)
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
		return domain.KelasPengabdi{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasPengabdiService) Get(id string) (domain.KelasPengabdi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPengabdi{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasPengabdi
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPengabdi{}, fmt.Errorf("kelas pengabdi %s not found", id)
		}
		return domain.KelasPengabdi{}, err
	}
	return item, nil
}

func (s *KelasPengabdiService) Update(id string, payload domain.KelasPengabdi, c *gin.Context) (domain.KelasPengabdi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPengabdi{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasPengabdi
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPengabdi{}, fmt.Errorf("kelas pengabdi %s not found", id)
		}
		return domain.KelasPengabdi{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.IdPengabdi != nil {
		item.IdPengabdi = payload.IdPengabdi
	}
	if payload.Sumbangan != nil {
		item.Sumbangan = payload.Sumbangan
	}
	if payload.Barang != nil {
		item.Barang = payload.Barang
	}
	if payload.TimKerja != nil {
		item.TimKerja = payload.TimKerja
	}
	if payload.TimKerjaReport != nil {
		item.TimKerjaReport = payload.TimKerjaReport
	}
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}
	if payload.Hari != nil {
		item.Hari = payload.Hari
	}
	if payload.SubKerja != nil {
		item.SubKerja = payload.SubKerja
	}
	if payload.Anak != nil {
		item.Anak = payload.Anak
	}
	if payload.Suster != nil {
		item.Suster = payload.Suster
	}
	if payload.Menginap != nil {
		item.Menginap = payload.Menginap
	}
	if payload.MakananPagi != nil {
		item.MakananPagi = payload.MakananPagi
	}
	if payload.MakananSiang != nil {
		item.MakananSiang = payload.MakananSiang
	}
	if payload.MakananMalam != nil {
		item.MakananMalam = payload.MakananMalam
	}

	userID := getUserID(c)
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasPengabdi{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasPengabdiService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasPengabdi
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas pengabdi %s not found", id)
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
