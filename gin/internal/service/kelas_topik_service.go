package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
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

func validateTopikLookups(db *gorm.DB, payload *domain.KelasTopik) error {
	type lookupCheck struct {
		fieldName string
		queryFn   func(db *gorm.DB) error
	}

	var checks []lookupCheck

	// 1. TrxId from Kelas
	if payload.TrxId != 0 {
		checks = append(checks, lookupCheck{
			fieldName: "trx_id",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.Kelas{}).Where("trxid = ?", payload.TrxId).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("trx_id %d tidak ditemukan di Kelas", payload.TrxId)
				}
				return nil
			},
		})
	}

	// 2. KodeTopik from Topic
	if payload.KodeTopik != nil && strings.TrimSpace(*payload.KodeTopik) != "" {
		val := strings.TrimSpace(*payload.KodeTopik)
		checks = append(checks, lookupCheck{
			fieldName: "kode_topik",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.Topic{}).Where("TopicCode = ?", val).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("kode_topik '%s' tidak ditemukan di Topic", val)
				}
				return nil
			},
		})
	}

	// 3. Penceramah if notempty from Umat
	if payload.Penceramah != nil && *payload.Penceramah != 0 {
		checks = append(checks, lookupCheck{
			fieldName: "penceramah",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.Umat{}).Where("id = ?", *payload.Penceramah).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("penceramah %d tidak ditemukan di Umat", *payload.Penceramah)
				}
				return nil
			},
		})
	}

	if len(checks) == 0 {
		return nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		details = make(map[string][]string)
	)

	for _, check := range checks {
		c := check
		wg.Go(func() {
			sess := db.Session(&gorm.Session{})
			if err := c.queryFn(sess); err != nil {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], err.Error())
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
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
		return domain.KelasTopik{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validateTopikLookups(s.db, &payload); err != nil {
		return domain.KelasTopik{}, err
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASTOPIKID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.KelasTopik{}, fmt.Errorf("failed to generate ID: %w", errId)
	}

	payload.DetailId = genResult.GeneratedId

	userID := getUserID(c)
	statusTrue := true
	modActI := "I"
	now := domain.NowDateTime()

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

	targetVal := item
	if payload.TrxId != 0 {
		targetVal.TrxId = payload.TrxId
	}
	if payload.KodeTopik != nil {
		targetVal.KodeTopik = payload.KodeTopik
	}
	if payload.Penceramah != nil {
		targetVal.Penceramah = payload.Penceramah
	}

	if err := validateTopikLookups(s.db, &targetVal); err != nil {
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
	now := domain.NowDateTime()

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
