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

// KelasPengeluaranService manages expense records incurred during a class session.
type KelasPengeluaranService struct {
	db       *gorm.DB
	resource string
}

// NewKelasPengeluaranService initializes a new KelasPengeluaranService instance.
//
// Parameters:
//   - db: *gorm.DB database connection pool (defaults to primary if nil)
//
// Returns:
//   - *KelasPengeluaranService: initialized service pointer
func NewKelasPengeluaranService(db *gorm.DB) *KelasPengeluaranService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasPengeluaranService{db: db, resource: "kelas_pengeluaran"}
}

// List queries active expense records for a class session with pagination.
//
// Parameters:
//   - trxID: string representation of the class transaction ID
//   - page: page number (1-based)
//   - limit: page size limit
//
// Returns:
//   - []domain.KelasPengeluaran: list of expense entries
//   - int64: total matching count
//   - error: database error if query fails
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

// Create inserts a new expense record for a class session after validation and ID generation.
//
// Parameters:
//   - payload: domain.KelasPengeluaran entity data to insert
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - domain.KelasPengeluaran: created expense entity with generated DetailId
//   - error: validation, ID generation, or database error if failed
func (s *KelasPengeluaranService) Create(payload domain.KelasPengeluaran, c *gin.Context) (domain.KelasPengeluaran, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("Validation failed: %w", err)
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASPENGELUARANID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.KelasPengeluaran{}, fmt.Errorf("failed to generate ID: %w", errId)
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
		return domain.KelasPengeluaran{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

// Get retrieves a single expense record by detail ID.
//
// Parameters:
//   - id: string representation of the detailid primary key
//
// Returns:
//   - domain.KelasPengeluaran: retrieved expense entity
//   - error: not found or database error
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

// Update modifies an existing expense record.
//
// Parameters:
//   - id: string representation of the detailid primary key
//   - payload: domain.KelasPengeluaran entity containing updated fields
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - domain.KelasPengeluaran: updated entity
//   - error: not found or database error if failed
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

// Delete soft-deletes an expense record (Status = false).
//
// Parameters:
//   - id: string representation of the detailid primary key
//   - c: *gin.Context containing user auth session for audit metadata
//
// Returns:
//   - error: not found or database error if failed
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
