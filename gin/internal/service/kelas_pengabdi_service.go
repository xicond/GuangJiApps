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

func validatePengabdiLookups(db *gorm.DB, payload *domain.KelasPengabdi) error {
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

	// 2. IdPengabdi from Umat
	if payload.IdPengabdi != nil && *payload.IdPengabdi != 0 {
		checks = append(checks, lookupCheck{
			fieldName: "id_pengabdi",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.Umat{}).Where("id = ?", *payload.IdPengabdi).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("id_pengabdi %d tidak ditemukan di Umat", *payload.IdPengabdi)
				}
				return nil
			},
		})
	}

	// 3. TimKerja B_TIMKERJA
	if payload.TimKerja != nil && strings.TrimSpace(*payload.TimKerja) != "" {
		val := strings.TrimSpace(*payload.TimKerja)
		checks = append(checks, lookupCheck{
			fieldName: "tim_kerja",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.AppLookup{}).
					Where("CategoryId = ? AND (LookupValue = ? OR LookupId = ?)", "B_TIMKERJA", val, val).
					Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("field tim_kerja nilai '%s' tidak valid", val)
				}
				return nil
			},
		})
	}

	// 4. SubKerja if notempty B_SUBKERJA
	if payload.SubKerja != nil && strings.TrimSpace(*payload.SubKerja) != "" {
		val := strings.TrimSpace(*payload.SubKerja)
		checks = append(checks, lookupCheck{
			fieldName: "sub_kerja",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.AppLookup{}).
					Where("CategoryId = ? AND (LookupValue = ? OR LookupId = ?)", "B_SUBKERJA", val, val).
					Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("field sub_kerja nilai '%s' tidak valid", val)
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

	wg.Add(len(checks))
	for _, check := range checks {
		go func(c lookupCheck) {
			defer wg.Done()
			sess := db.Session(&gorm.Session{})
			if err := c.queryFn(sess); err != nil {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], err.Error())
				mu.Unlock()
			}
		}(check)
	}

	wg.Wait()

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

func (s *KelasPengabdiService) List(c *gin.Context, trxID string, page int, limit int) ([]domain.KelasPengabdi, int64, error) {
	var items []domain.KelasPengabdi
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	var subWhVal string
	userID := getUserID(c)
	row := s.db.Model(&domain.AdminMatrix{}).
		Where("LOGINID = ?", userID).
		Select("SUBWHID").
		Row()
	if row != nil {
		_ = row.Scan(&subWhVal)
	}

	query := s.db.Table("T_TRX_KELAS_PENGABDI").
		Select(`T_TRX_KELAS_PENGABDI.*, 
			u.namaindonesia AS nama_indonesia, 
			u.namamandarin AS nama_mandarin, 
			lf.LookupDescription AS fotang_aktif_desc, 
			lt.LookupDescription AS tim_kerja_desc, 
			ls.LookupDescription AS sub_kerja_desc`).
		Joins("LEFT JOIN T_BUS_UMAT u ON T_TRX_KELAS_PENGABDI.idpengabdi = u.id").
		Joins("LEFT JOIN T_APP_LOOKUP lf ON (u.fotangaktif = lf.LookupValue OR u.fotangaktif = lf.LookupId) AND lf.CategoryId = 'B_FOTHANG'").
		Joins("LEFT JOIN T_APP_LOOKUP lt ON (T_TRX_KELAS_PENGABDI.timkerja = lt.LookupValue OR T_TRX_KELAS_PENGABDI.timkerja = lt.LookupId) AND lt.CategoryId = 'B_TIMKERJA'").
		Joins("LEFT JOIN T_APP_LOOKUP ls ON (T_TRX_KELAS_PENGABDI.SubKerja = ls.LookupValue OR T_TRX_KELAS_PENGABDI.SubKerja = ls.LookupId) AND ls.CategoryId = 'B_SUBKERJA'").
		Where("u.fotangaktif = ?", subWhVal).
		Where("T_TRX_KELAS_PENGABDI.status = ?", true)

	if trxID != "" {
		if parsedID, err := strconv.Atoi(trxID); err == nil {
			query = query.Where("T_TRX_KELAS_PENGABDI.trxid = ?", parsedID)
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
		return domain.KelasPengabdi{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validatePengabdiLookups(s.db, &payload); err != nil {
		return domain.KelasPengabdi{}, err
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASPENGABDIID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.KelasPengabdi{}, fmt.Errorf("failed to generate ID: %w", errId)
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

	targetVal := item
	if payload.TrxId != 0 {
		targetVal.TrxId = payload.TrxId
	}
	if payload.IdPengabdi != nil {
		targetVal.IdPengabdi = payload.IdPengabdi
	}
	if payload.TimKerja != nil {
		targetVal.TimKerja = payload.TimKerja
	}
	if payload.SubKerja != nil {
		targetVal.SubKerja = payload.SubKerja
	}

	if err := validatePengabdiLookups(s.db, &targetVal); err != nil {
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
	now := domain.NowDateTime()

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
