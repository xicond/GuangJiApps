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

type SxyDonaturService struct {
	db       *gorm.DB
	resource string
}

func NewSxyDonaturService(db *gorm.DB) *SxyDonaturService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &SxyDonaturService{db: db, resource: "sxy-donatur"}
}

func (s *SxyDonaturService) List(page int, filters map[string]string, limit int) ([]domain.SxyDonaturResponse, int64, error) {
	var items []domain.SxyDonaturResponse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	if s.db.Dialector.Name() == "sqlite" {
		var count int64
		s.db.Model(&domain.SxyDonatur{}).Count(&count)
		var list []domain.SxyDonatur
		s.db.Limit(limit).Offset((page - 1) * limit).Find(&list)
		for _, d := range list {
			items = append(items, domain.SxyDonaturResponse{
				ID:   d.ID,
				No:   d.No,
				Nama: d.Nama,
			})
		}
		return items, count, nil
	}
	// offset := (page - 1) * limit

	allowedFilters := map[string]string{
		"nama":     "nama",
		"mandarin": "mandarin",
		"fotang":   "fotang",
	}

	var nama string = ""
	var mandarin string = ""
	var fotang int = 0

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "nama" {
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
	rows, err := s.db.Raw("EXEC SP_SXY_DONATUR_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?",
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
			findErr = errors.New("sxy donatur tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, findErr
	}

	for rows.Next() {
		var item domain.SxyDonaturResponse
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
		return []domain.SxyDonaturResponse{}, 0, findErr
	}

	return items, total, nil
}

func MapRowToStruct(rowMap map[string]interface{}, target interface{}) {
	v := reflect.ValueOf(target).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		fieldType := t.Field(i)
		fieldName := strings.ToLower(fieldType.Name)

		// Cari key di rowMap secara case-insensitive
		var rawVal interface{}
		var found bool
		for k, v := range rowMap {
			if strings.ToLower(k) == fieldName {
				rawVal = v
				found = true
				break
			}
		}

		if !found || rawVal == nil {
			continue
		}

		// Assign nilai berdasarkan tipe data field pada struct
		switch fieldVal.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			fieldVal.SetInt(toInt64(rawVal))
		case reflect.String:
			fieldVal.SetString(toString(rawVal))
		case reflect.Bool:
			if b, ok := rawVal.(bool); ok {
				fieldVal.SetBool(b)
			}
			// Tambahkan jenis tipe lain jika diperlukan (misal: float, time.Time, dll)
		}
	}
}

func toInt64(val interface{}) int64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	// case int:
	//     return int64(v)
	// case int32:
	//     return int64(v)
	// case float64:
	//     return int64(v)
	case []byte:
		// Konversi byte slice dari database driver ke string, lalu parse ke int64
		n, _ := strconv.ParseInt(string(v), 10, 64)
		return n
	case string:
		// Parse string langsung ke int64
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	default:
		// Fallback terakhir: ubah ke string dulu lalu parse
		n, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		return n
	}
}

func (s *SxyDonaturService) Create(payload domain.SxyDonatur, c *gin.Context) (domain.SxyDonatur, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("validasi gagal: %w", err)
	}

	var maxID int32
	s.db.Table("T_SXY_MST_DONATUR").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
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
		return domain.SxyDonatur{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *SxyDonaturService) Get(id string) (domain.SxyDonatur, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.SxyDonatur
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.SxyDonatur{}, fmt.Errorf("sxy donatur %s not found", id)
		}
		return domain.SxyDonatur{}, err
	}
	return item, nil
}

func (s *SxyDonaturService) Update(id string, payload domain.SxyDonatur, c *gin.Context) (domain.SxyDonatur, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.SxyDonatur
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.SxyDonatur{}, fmt.Errorf("sxy donatur %s not found", id)
		}
		return domain.SxyDonatur{}, err
	}

	userID := int32(1)
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			if uid, ok := val.(int); ok {
				userID = int32(uid)
			}
		}
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
	item.UpdatedBy = userID
	item.UpdatedDate = time.Now()

	if err := s.db.Save(&item).Error; err != nil {
		return domain.SxyDonatur{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *SxyDonaturService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.SxyDonatur
	if err := s.db.Where("id = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sxy donatur %s not found", id)
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
