package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// KelasMusikService manages music session schedules and song assignments for a class session.
type KelasMusikService struct {
	db       *gorm.DB
	resource string
}

// NewKelasMusikService initializes a new KelasMusikService instance.
//
// Parameters:
//   - db: *gorm.DB database connection pool (defaults to primary if nil)
//
// Returns:
//   - *KelasMusikService: initialized service pointer
func NewKelasMusikService(db *gorm.DB) *KelasMusikService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasMusikService{db: db, resource: "kelas_musik"}
}

// List queries active music session items for a class session with pagination.
//
// Parameters:
//   - c: *gin.Context carrying HTTP request metadata
//   - trxID: string representation of the class transaction ID
//   - page: page number (1-based)
//   - limit: page size limit
//
// Returns:
//   - []domain.KelasMusik: list of class music schedule entries
//   - int64: total matching count
//   - error: database error if query fails
func (s *KelasMusikService) List(c *gin.Context, trxID string, page int, limit int) ([]domain.KelasMusik, int64, error) {
	var items []domain.KelasMusik
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	var reqCtx = context.Background()
	if c != nil && c.Request != nil {
		reqCtx = c.Request.Context()
	}

	query := s.db.WithContext(reqCtx).Model(&domain.KelasMusik{}).Where("status = ?", true)

	if trxID != "" {
		if parsedID, err := strconv.Atoi(trxID); err == nil {
			query = query.Where("TrxId = ?", parsedID)
		}
	}

	offset := (page - 1) * limit

	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			if errors.Is(err, context.Canceled) {
				countErr = fmt.Errorf("Query dibatalkan karena klien memutus koneksi.")
			} else {
				countErr = fmt.Errorf("failed to count record: %w", err)
			}
		}
	})

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).Offset(offset).Limit(limit).Find(&items).Error; err != nil {
			if errors.Is(err, context.Canceled) {
				findErr = errors.New("Query dibatalkan karena klien memutus koneksi.")
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("musik tidak ditemukan")
			} else {
				findErr = fmt.Errorf("failed to list records: %w", err)
			}
		}
	})

	wg.Wait()

	if countErr != nil {
		return items, 0, countErr
	}
	if findErr != nil {
		return items, 0, findErr
	}

	return items, total, nil
}

// Create inserts a new music schedule entry for a class session after validation and ID generation.
//
// Parameters:
//   - payload: domain.KelasMusik entity data to insert
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - domain.KelasMusik: created music schedule entity with generated DetailId
//   - error: validation, ID generation, or database error if failed
func (s *KelasMusikService) Create(payload domain.KelasMusik, c *gin.Context) (domain.KelasMusik, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasMusik{}, fmt.Errorf("Validation failed: %w", err)
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASMUSICID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.KelasMusik{}, fmt.Errorf("failed to generate ID: %w", errId)
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
		return domain.KelasMusik{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get retrieves a single music schedule item by detail ID.
//
// Parameters:
//   - id: string representation of the DetailId primary key
//
// Returns:
//   - domain.KelasMusik: retrieved music schedule entity
//   - error: not found or database error
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

// Update modifies an existing music schedule item.
//
// Parameters:
//   - id: string representation of the DetailId primary key
//   - payload: domain.KelasMusik entity containing updated fields
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - domain.KelasMusik: updated entity
//   - error: not found or database error if failed
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
	now := domain.NowDateTime()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasMusik{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete soft-deletes a music schedule item (Status = false).
//
// Parameters:
//   - id: string representation of the DetailId primary key
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - error: not found or database error if failed
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
