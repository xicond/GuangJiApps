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

type KelasTopikService struct {
	db       *gorm.DB
	resource string
}

func NewKelasTopikService(db *gorm.DB) *KelasTopikService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasTopikService{db: db, resource: "kelas_topik"}
}

func (s *KelasTopikService) List(trxID string, page int, limit int) ([]domain.KelasTopik, int64, error) {
	var items []domain.KelasTopik
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Table("T_TRX_KELAS_TOPIK").
		Select("T_TRX_KELAS_TOPIK.*, T_BUS_TOPIC.TopicName AS nama_topik, T_BUS_TOPIC.TopicCategory AS topic_category, T_BUS_TOPIC.Description AS topic_desc").
		Joins("LEFT JOIN T_BUS_TOPIC ON T_TRX_KELAS_TOPIK.kodetopik = T_BUS_TOPIC.TopicCode").
		Where("T_TRX_KELAS_TOPIK.status = ?", true)

	if trxID != "" {
		if parsedID, err := strconv.Atoi(trxID); err == nil {
			query = query.Where("T_TRX_KELAS_TOPIK.trxid = ?", parsedID)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return items, 0, fmt.Errorf("failed to count record: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("T_TRX_KELAS_TOPIK.topikdate asc, T_TRX_KELAS_TOPIK.urutan asc").Find(&items).Error; err != nil {
		return items, 0, fmt.Errorf("failed to list records: %w", err)
	}

	return items, total, nil
}

func (s *KelasTopikService) Create(payload domain.KelasTopik, c *gin.Context) (domain.KelasTopik, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasTopik{}, fmt.Errorf("validasi gagal: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_TOPIK").Select("ISNULL(MAX(detailid), 0)").Row().Scan(&maxID)
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
		return domain.KelasTopik{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasTopikService) Get(id string) (domain.KelasTopik, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasTopik{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasTopik
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasTopik{}, fmt.Errorf("kelas topik %s not found", id)
		}
		return domain.KelasTopik{}, err
	}
	return item, nil
}

func (s *KelasTopikService) Update(id string, payload domain.KelasTopik, c *gin.Context) (domain.KelasTopik, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasTopik{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasTopik
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasTopik{}, fmt.Errorf("kelas topik %s not found", id)
		}
		return domain.KelasTopik{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.KodeTopik != nil {
		item.KodeTopik = payload.KodeTopik
	}
	if payload.Urutan != nil {
		item.Urutan = payload.Urutan
	}
	if payload.TopikDate != nil {
		item.TopikDate = payload.TopikDate
	}
	if payload.Penceramah != nil {
		item.Penceramah = payload.Penceramah
	}
	if payload.PenceramahExt != nil {
		item.PenceramahExt = payload.PenceramahExt
	}
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}
	if payload.Durasi != nil {
		item.Durasi = payload.Durasi
	}
	if payload.Penterjemah != nil {
		item.Penterjemah = payload.Penterjemah
	}

	userID := getUserID(c)
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasTopik{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasTopikService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasTopik
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas topik %s not found", id)
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
