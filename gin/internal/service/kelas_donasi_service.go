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

// KelasDonasiService manages monetary donation records earmarked for a class event.
type KelasDonasiService struct {
	db       *gorm.DB
	resource string
}

// NewKelasDonasiService initializes a new KelasDonasiService instance.
//
// Parameters:
//   - db: *gorm.DB database connection pool (defaults to primary if nil)
//
// Returns:
//   - *KelasDonasiService: initialized service pointer
func NewKelasDonasiService(db *gorm.DB) *KelasDonasiService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasDonasiService{db: db, resource: "kelas_donasi"}
}

// List queries active monetary donations for a class session with pagination.
//
// Parameters:
//   - trxID: string representation of the class transaction ID
//   - page: page number (1-based)
//   - limit: page size limit
//
// Returns:
//   - []domain.KelasDonasi: list of class donation records
//   - int64: total matching count
//   - error: database error if query fails
func (s *KelasDonasiService) List(trxID string, page int, limit int) ([]domain.KelasDonasi, int64, error) {
	var items []domain.KelasDonasi
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	query := s.db.Model(&domain.KelasDonasi{}).Where("Status = ?", true)

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

// Create inserts a new monetary donation for a class session after validation and ID generation.
//
// Parameters:
//   - payload: domain.KelasDonasi entity payload to insert
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - domain.KelasDonasi: created donation record with generated DetailId
//   - error: validation, ID generation, or database error if failed
func (s *KelasDonasiService) Create(payload domain.KelasDonasi, c *gin.Context) (domain.KelasDonasi, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("Validation failed: %w", err)
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASDONASIID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.KelasDonasi{}, fmt.Errorf("failed to generate ID: %w", errId)
	}

	payload.DetailId = genResult.GeneratedId

	userIDStr := strconv.Itoa(int(getUserID(c)))
	statusTrue := true
	modActI := "I"
	now := domain.NowDateTime()

	payload.Status = &statusTrue
	payload.ModAct = &modActI
	payload.ModBy = &userIDStr
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get retrieves a single donation record by detail ID.
//
// Parameters:
//   - id: string representation of the DetailId primary key
//
// Returns:
//   - domain.KelasDonasi: retrieved donation entity
//   - error: not found or database error
func (s *KelasDonasiService) Get(id string) (domain.KelasDonasi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasDonasi
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasDonasi{}, fmt.Errorf("kelas donasi %s not found", id)
		}
		return domain.KelasDonasi{}, err
	}
	return item, nil
}

// Update modifies an existing class donation record.
//
// Parameters:
//   - id: string representation of the DetailId primary key
//   - payload: domain.KelasDonasi entity containing updated donation fields
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - domain.KelasDonasi: updated entity
//   - error: not found or database error if failed
func (s *KelasDonasiService) Update(id string, payload domain.KelasDonasi, c *gin.Context) (domain.KelasDonasi, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasDonasi
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasDonasi{}, fmt.Errorf("kelas donasi %s not found", id)
		}
		return domain.KelasDonasi{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.Donatur != "" {
		item.Donatur = payload.Donatur
	}
	if payload.Donasi != 0 {
		item.Donasi = payload.Donasi
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	modActU := "U"
	now := domain.NowDateTime()

	item.ModAct = &modActU
	item.ModBy = &userIDStr
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasDonasi{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

// Delete soft-deletes a class donation record (Status = false).
//
// Parameters:
//   - id: string representation of the DetailId primary key
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - error: not found or database error if failed
func (s *KelasDonasiService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasDonasi
	if err := s.db.Where("DetailId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas donasi %s not found", id)
		}
		return err
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	statusFalse := false
	modActD := "D"
	now := domain.NowDateTime()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userIDStr
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
