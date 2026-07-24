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

type UmatService struct {
	db       *gorm.DB
	resource string
}

func NewUmatService(db *gorm.DB) *UmatService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &UmatService{db: db, resource: "umats"}
}

func (s *UmatService) List(page int, filters map[string]string, limit int) ([]domain.Umat, int64, error) { //[]domain.Umat {
	var items []domain.Umat
	var total int64

	// Validasi parameter pagination
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Base query
	query := s.db.Table("T_BUS_UMAT").Where("status = ?", true)
	// Dynamic optional search on fields (dengan whitelist kolom aman dari SQL Injection)
	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"alias":                {Column: "alias", IsLike: true},
		"namaindonesia":        {Column: "namaindonesia", IsLike: true},
		"namamandarin":         {Column: "namamandarin", IsLike: true},          // Tahun menggunakan exact match (=)
		"tahunchiutaomandarin": {Column: "tahunchiutaomandarin", IsLike: false}, // Tahun menggunakan exact match (=)
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if rule, exists := allowedFilters[field]; exists {
			if rule.IsLike {
				// String / Varchar menggunakan LIKE
				query = query.Where(fmt.Sprintf("[%s] LIKE ?", rule.Column), "%"+value+"%")
			} else {
				// Tahun atau numerik menggunakan exact match (=)
				query = query.Where(fmt.Sprintf("[%s] = ?", rule.Column), value)
			}
		}
	}

	// 1. Hitung total data keseluruhan (untuk metadata pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("database count error: %w", err)
	}

	// Execute retrieval order
	err := query.
		Limit(limit).
		Offset(offset).
		Order("ID ASC"). // Explicitly sort by your real primary key column
		Find(&items).Error

	if err != nil {
		// 1. Jika error murni karena username tidak terdaftar di DB
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.Umat{}, 0, errors.New("umat tidak ditemukan")
		}

		// 2. Jika error karena masalah MSSQL (misal: "invalid column name", "connection timeout")
		// Mengembalikan pesan error asli dari sistem SQL Server secara dinamis
		return []domain.Umat{}, 0, fmt.Errorf("database error: %w", err)
		// return nil
	}
	return items, total, nil
}

func (s *UmatService) Create(payload domain.Umat, c *gin.Context) (domain.Umat, error) {
	// 1. Validate based on your actual struct parameters
	if payload.Kode == "" || payload.NamaIndonesia == "" {
		return domain.Umat{}, fmt.Errorf("kode and nama_indonesia are required")
	}

	// 2. IMPORTANT: Do NOT generate a random UnixNano string for ID!
	// Your SQL Server schema defines [id] INT NOT NULL.
	// If it is NOT an IDENTITY column, we calculate the next integer sequence manually.
	var maxID int32
	s.db.Table("T_BUS_UMAT").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
	payload.ID = maxID + 1

	// 3. Populate matching schema structural constraints
	payload.Status = true                            // Active status mapping
	payload.ModAct = "I"                             // 'I' standard legacy flag for Insert
	payload.ModBy = int32(c.MustGet("userID").(int)) // Default system user ID matching INT type
	payload.ModDate = time.Now()                     // Local server time object

	// 4. Persist the new entity to the database pool
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *UmatService) Get(id string) (domain.Umat, error) {
	var item domain.Umat
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("umat %s not found", id)
	}
	return item, nil
}

func (s *UmatService) Update(id string, payload domain.Umat, c *gin.Context) (domain.Umat, error) {
	// 1. Cast string ID parameter safely to int32 to prevent MSSQL query crashes
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Umat{}, fmt.Errorf("invalid ID format: %w", err)
	}
	userIDInt32 := int32(parsedInt)

	var item domain.Umat
	// 2. Fetch the existing item using .Take() to avoid default sorting bugs
	if err := s.db.Where("id = ?", userIDInt32).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Umat{}, fmt.Errorf("umat %s not found", id)
		}
		return domain.Umat{}, err
	}

	// 3. Map values onto the actual field variables present in your legacy schema
	item.Kode = payload.Kode
	item.NamaIndonesia = payload.NamaIndonesia
	item.NamaMandarin = payload.NamaMandarin
	item.Alamat = payload.Alamat
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile

	// Legacy metadata mappings
	// item.Status = payload.Status // Maps to legacy [STATUS] BIT flag
	item.ModAct = "U"                             // 'U' standard legacy flag for Update
	item.ModBy = int32(c.MustGet("userID").(int)) // System user ID (int32)
	item.ModDate = time.Now()                     // Actual time.Time object expected by DATETIME column

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *UmatService) Delete(id string, c *gin.Context) error {
	// 1. Cast string ID parameter safely to int32 to prevent MSSQL query crashes
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}
	userIDInt32 := int32(parsedInt)

	var item domain.Umat
	// 2. Fetch the existing item using .Take() to avoid default sorting bugs
	if err := s.db.Where("id = ?", userIDInt32).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("umat %s not found", id)
		}
		return err
	}

	// Legacy metadata mappings
	item.Status = false
	item.ModAct = "D"                             // 'D' standard legacy flag for Delete
	item.ModBy = int32(c.MustGet("userID").(int)) // Default system user ID matching INT type
	item.ModDate = time.Now()                     // Local server time object

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}
