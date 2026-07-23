package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

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

func (s *UmatService) List( /*limit int, offset int*/ ) ([]domain.Umat, error) { //[]domain.Umat {
	var items []domain.Umat

	var limit = 10
	var offset = 0
	// Default fallback values if pagination parameters are missing
	if limit <= 0 {
		limit = 10
	}

	// Prepare the query pool
	query := s.db.Table("T_BUS_UMAT").Where("status = ?", true)

	// Execute retrieval order
	err := query.
		Limit(limit).
		Offset(offset).
		Order("ID ASC"). // Explicitly sort by your real primary key column
		Find(&items).Error

	if err != nil {
		// 1. Jika error murni karena username tidak terdaftar di DB
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.Umat{}, errors.New("umat tidak ditemukan")
		}

		// 2. Jika error karena masalah MSSQL (misal: "invalid column name", "connection timeout")
		// Mengembalikan pesan error asli dari sistem SQL Server secara dinamis
		return []domain.Umat{}, fmt.Errorf("database error: %w", err)
		// return nil
	}
	return items, nil
}

func (s *UmatService) Create(payload domain.Umat) (domain.Umat, error) {
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
	payload.Status = true        // Active status mapping
	payload.ModAct = "I"         // 'I' standard legacy flag for Insert
	payload.ModBy = 1            // Default system user ID matching INT type
	payload.ModDate = time.Now() // Local server time object

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

func (s *UmatService) Update(id string, payload domain.Umat) (domain.Umat, error) {
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
	item.Status = payload.Status // Maps to legacy [STATUS] BIT flag
	item.ModAct = "U"            // 'U' standard legacy flag for Update
	item.ModBy = 1               // System user ID (int32)
	item.ModDate = time.Now()    // Actual time.Time object expected by DATETIME column

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *UmatService) Delete(id string) error {
	return s.db.Delete(&domain.Umat{}, "id = ?", id).Error
}
