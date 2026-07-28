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

type nullStringScanner struct {
	target *string
}

func (s *nullStringScanner) Scan(value interface{}) error {
	if value == nil {
		*s.target = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*s.target = v
		return nil
	case []byte:
		*s.target = string(v)
		return nil
	default:
		*s.target = fmt.Sprintf("%v", v)
		return nil
	}
}

type DonasiSxyService struct {
	db       *gorm.DB
	resource string
}

func NewDonasiSxyService(db *gorm.DB) *DonasiSxyService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &DonasiSxyService{db: db, resource: "donasi-sxy"}
}

func isValidDate(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

func (s *DonasiSxyService) List(page int, filters map[string]string, limit int) ([]domain.DonasiSxyResponse, int64, error) {
	items := make([]domain.DonasiSxyResponse, 0)
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	if s.db.Dialector.Name() == "sqlite" {
		var count int64
		s.db.Model(&domain.DonasiSxy{}).Count(&count)
		var list []domain.DonasiSxy
		s.db.Limit(limit).Offset((page - 1) * limit).Find(&list)
		for _, d := range list {
			items = append(items, domain.DonasiSxyResponse{
				ID:         d.ID,
				NoKwitansi: d.NoKwitansi,
				Jumlah:     d.Jumlah,
			})
		}
		return items, count, nil
	}

	allowedFilters := map[string]bool{
		"no_kwitansi": true,
		"nokwitansi":  true,
		// "kwitansi":     true,
		"start_date": true,
		// "startdate":    true,
		"end_date": true,
		// "enddate":      true,
		// "tanggal":      true,
		// "date":         true,
		"donatur": true,
		// "nama":         true,
		// "donatur_nama": true,
	}

	var no_kwitansi string = ""
	var start_date *string = nil
	var end_date *string = nil
	var donatur string = ""

	for field, value := range filters {
		if value == "" {
			continue
		}
		if allowedFilters[field] {
			val := value
			switch field {
			case "no_kwitansi", "nokwitansi", "kwitansi":
				no_kwitansi = val
			case "start_date", "startdate":
				if isValidDate(val) {
					start_date = &val
				} else if donatur == "" {
					donatur = val
				}
			case "end_date", "enddate", "date", "tanggal":
				if isValidDate(val) {
					end_date = &val
				} else if donatur == "" {
					donatur = val
				}
			case "donatur", "nama", "donatur_nama":
				donatur = val
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	rows, err := s.db.Raw("EXEC SP_SXY_TRX_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu ASCENDING)
		no_kwitansi,   // @NoKwitansi (nvarchar)
		start_date,    // @StartDate (nvarchar)
		end_date,      // @EndDate (nvarchar)
		donatur,       // @Donatur (nvarchar)
	).Rows()

	if err != nil {
		return items, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("donasi sxy tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return items, total, findErr
	}

	for rows.Next() {
		var item domain.DonasiSxyResponse
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow", "total_row", "totalcount", "total_count", "rowcount":
				valuePtrs[i] = &totalRowScan
			default:
				matched := false
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						fieldVal := v.Field(j)
						if fieldVal.Kind() == reflect.String {
							valuePtrs[i] = &nullStringScanner{target: fieldVal.Addr().Interface().(*string)}
						} else {
							valuePtrs[i] = fieldVal.Addr().Interface()
						}
						matched = true
						break
					}
				}
				if !matched {
					var dummy interface{}
					valuePtrs[i] = &dummy
				}
			}
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("donasi_sxy rows.Scan error: %v\n", err)
			findErr = fmt.Errorf("scan error on row: %w", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.DonasiSxyResponse{}, 0, findErr
	}

	return items, total, nil
}

func (s *DonasiSxyService) Create(payload domain.DonasiSxy, c *gin.Context) (domain.DonasiSxy, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("validasi gagal: %w", err)
	}

	var maxID int32
	s.db.Table("T_SXY_TRANSAKSI").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
	payload.ID = maxID + 1

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	payload.Status = true
	payload.CreatedBy = userID
	payload.CreatedDate = time.Now()
	payload.UpdatedBy = userID
	payload.UpdatedDate = time.Now()

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *DonasiSxyService) Get(id string) (domain.DonasiSxy, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
		}
		return domain.DonasiSxy{}, err
	}
	return item, nil
}

func (s *DonasiSxyService) Update(id string, payload domain.DonasiSxy, c *gin.Context) (domain.DonasiSxy, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DonasiSxy{}, fmt.Errorf("donasi sxy %s not found", id)
		}
		return domain.DonasiSxy{}, err
	}

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.NoKwitansi = payload.NoKwitansi
	item.Tanggal = payload.Tanggal
	item.Donatur = payload.Donatur
	item.Penggalang = payload.Penggalang
	item.Jumlah = payload.Jumlah
	item.TipeSumbangan = payload.TipeSumbangan
	item.NoKupon = payload.NoKupon
	item.Keterangan = payload.Keterangan
	item.TanggalTransfer = payload.TanggalTransfer
	item.AtasNama = payload.AtasNama
	item.TtkSent = payload.TtkSent
	item.UpdatedBy = userID
	item.UpdatedDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.DonasiSxy{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *DonasiSxyService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.DonasiSxy
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("donasi sxy %s not found", id)
		}
		return err
	}

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
	}

	item.Status = false
	item.UpdatedBy = userID
	item.UpdatedDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
