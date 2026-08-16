package service

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PenggalangDanaService struct {
	db       *gorm.DB
	resource string
}

func NewPenggalangDanaService(db *gorm.DB) *PenggalangDanaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &PenggalangDanaService{db: db, resource: "penggalang-dana"}
}

func (s *PenggalangDanaService) List(page int, filters map[string]string, limit int) ([]domain.PenggalangDanaResponse, int64, error) {
	var items []domain.PenggalangDanaResponse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	allowedFilters := map[string]string{
		"nama":               "nama",
		"lookup_description": "nama",
		"mandarin":           "mandarin",
		"fotang":             "fotang",
	}

	var nama string = ""
	var mandarin string = ""
	var fotang int = 0

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "nama" || field == "lookup_description" {
				nama = value
			}

			if field == "mandarin" {
				mandarin = value
			}

			if field == "fotang" {
				fotang = toInt(value)
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	rows, err := s.db.Raw("EXEC SP_SXY_PENGGALANG_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu null)
		nama,          // @nama (string / nvarchar)
		mandarin,      // @mandarin (nvarchar)
		fotang,        // @Date (int)
	).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("penggalang dana tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, findErr
	}

	for rows.Next() {
		var item domain.PenggalangDanaResponse
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		// Siapkan slice pointer untuk rows.Scan sepanjang jumlah kolom
		valuePtrs := make([]interface{}, len(cols))

		// Variabel penampung sementara untuk kolom khusus seperti TotalRow yang tidak ada di struct
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow":
				valuePtrs[i] = &totalRowScan
			default:
				matched := false
				// Cari field di struct berdasarkan tag gorm "column" atau nama field
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					// Cocokkan dengan tag kolom GORM atau nama struct (case-insensitive)
					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						valuePtrs[i] = v.Field(j).Addr().Interface()
						matched = true
						break
					}
				}
				// Jika kolom database tidak ada di struct, tampung ke dummy agar Scan tidak error
				if !matched {
					var dummy interface{}
					valuePtrs[i] = &dummy
				}
			}
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Ambil total row jika ada
		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.PenggalangDanaResponse{}, 0, findErr
	}

	return items, total, nil
}

func (s *PenggalangDanaService) Create(payload domain.PenggalangDana, c *gin.Context) (domain.PenggalangDana, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("Validation failed: %w", err)
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "SXYPENGGALANGID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.PenggalangDana{}, fmt.Errorf("failed to generate ID: %w", errId)
	}

	payload.ID = genResult.GeneratedId
	payload.No = strconv.Itoa(int(genResult.GeneratedId))

	userID := getUserID(c)

	payload.Status = true
	payload.CreatedBy = userID
	payload.CreatedDate = domain.NowDateTime()
	payload.UpdatedBy = userID
	payload.UpdatedDate = domain.NowDateTime()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *PenggalangDanaService) Get(id string) (domain.PenggalangDana, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.PenggalangDana
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PenggalangDana{}, fmt.Errorf("penggalang dana %s not found", id)
		}
		return domain.PenggalangDana{}, err
	}
	return item, nil
}

func (s *PenggalangDanaService) Update(id string, payload domain.PenggalangDana, c *gin.Context) (domain.PenggalangDana, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.PenggalangDana
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PenggalangDana{}, fmt.Errorf("penggalang dana %s not found", id)
		}
		return domain.PenggalangDana{}, err
	}

	item.No = payload.No
	item.Nama = payload.Nama
	item.Mandarin = payload.Mandarin
	item.Keterangan = payload.Keterangan
	item.LookupFothang = payload.LookupFothang
	item.Alamat = payload.Alamat
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile
	item.Email = payload.Email
	item.UpdatedBy = getUserID(c)
	item.UpdatedDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.PenggalangDana{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *PenggalangDanaService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.PenggalangDana
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("penggalang dana %s not found", id)
		}
		return err
	}

	item.Status = false
	item.UpdatedBy = getUserID(c)
	item.UpdatedDate = domain.NowDateTime()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
