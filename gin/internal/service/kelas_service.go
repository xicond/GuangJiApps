package service

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KelasService struct {
	db       *gorm.DB
	resource string
}

func NewKelasService(db *gorm.DB) *KelasService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasService{db: db, resource: "kelas"}
}

func (s *KelasService) List(page int, filters map[string]string, c *gin.Context, limit int) ([]domain.KelasResponse, int64, error) {
	var items []domain.KelasResponse
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	allowedFilters := map[string]bool{
		"kelas": true,
		// "kwitansi":     true,
		"start_date": true,
		// "startdate":    true,
		"end_date": true,
		// "enddate":      true,
		// "tanggal":      true,
		// "date":         true,
		"fotang": true,
		// "nama":         true,
		// "donatur_nama": true,
	}

	var kelas string = "0"
	var start_date string = "2018-07-01"
	var end_date string = "2026-07-25"
	var fotang string = "0"

	for field, value := range filters {
		if value == "" {
			continue
		}
		if allowedFilters[field] {
			val := value
			switch field {
			case "kelas":
				kelas = val
			case "start_date":
				if isValidDate(val) {
					start_date = val
				}
			case "end_date":
				if isValidDate(val) {
					end_date = val
				}
			case "fotang":
				fotang = val
			}
		}
	}

	var findErr error

	sortDirection := "ASCENDING"
	var userID int32
	if userIDVal, exists := c.Get("userID"); exists {
		userID = toInt32(userIDVal)
	}

	rows, err := s.db.Raw("EXEC SP_TRX_KELAS_SEARCH_DATA ?, ?, ?, ?, ?, ?, ?, ?",
		limit,         // @PageSize
		page,          // @CurrentPage
		nil,           // @SortExpression (selalu null)
		sortDirection, // @SortDirection (selalu ASCENDING)
		kelas,         // @NoKwitansi (nvarchar)
		start_date,    // @StartDate (Date)
		end_date,      // @EndDate (Date)
		fotang,        // @fotang (int)
		userID,        // @Donatur (int32)
	).Rows()

	if err != nil {
		return items, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("kelas tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return items, total, findErr
	}

	for rows.Next() {
		var item domain.KelasResponse
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
			fmt.Printf("kelas rows.Scan error: %v\n", err)
			findErr = fmt.Errorf("scan error on row: %w", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.KelasResponse{}, 0, findErr
	}

	return items, total, nil
}

func (s *KelasService) Lookup(filters map[string]string, page int, limit int) ([]domain.AppLookup, int64, error) {
	var items []domain.AppLookup
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	subQuery := s.db.Table("T_APP_LOOKUPCATEGORY").
		Where("T_APP_LOOKUPCATEGORY.CategoryId = T_APP_LOOKUP.CategoryId")

	query := s.db.Model(&domain.AppLookup{}).
		Where("CategoryId = ?", "B_KELASKHUSUS").
		Where("EXISTS (?)", subQuery)

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]bool{
		"lookup_description": true,
		// "lookup_value":       true,
		// "lookup_id":          true,
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if _, exists := allowedFilters[field]; exists {
			if field == "lookup_description" {
				query = query.Where("LookupDescription LIKE ?", "%"+value+"%")
			} /*  else if field == "lookup_value" {
				query = query.Where("LookupValue LIKE ?", "%"+value+"%")
			} else if field == "lookup_id" {
				query = query.Where("LookupId = ?", value)
			} */
		}
	}

	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).
			Limit(limit).
			Offset(offset).
			Order("LookupValue ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("activity tidak ditemukan")
			} else {
				findErr = fmt.Errorf("database error: %w", err)
			}
		}
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if findErr != nil {
		return []domain.AppLookup{}, 0, findErr
	}

	return items, total, nil
}

func toInt32(val interface{}) int32 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int32:
		return v
	case []byte:
		n, _ := strconv.ParseInt(string(v), 10, 32)
		return int32(n)
	case string:
		n, _ := strconv.ParseInt(v, 10, 32)
		return int32(n)
	default:
		n, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 32)
		return int32(n)
	}
}

func getUserID(c *gin.Context) int32 {
	if c != nil {
		if val, exists := c.Get("userID"); exists {
			switch v := val.(type) {
			case int:
				return int32(v)
			case int32:
				return v
			case float64:
				return int32(v)
			case string:
				if n, err := strconv.Atoi(v); err == nil {
					return int32(n)
				}
			}
		}
	}
	return 1
}

func validateKelasLookups(db *gorm.DB, kodeKelas *string, kodeFotang *string) error {
	if kodeKelas == nil || strings.TrimSpace(*kodeKelas) == "" {
		return errors.New("kode_kelas is required")
	}

	var countKelas int64
	if err := db.Model(&domain.AppLookup{}).
		Where("CategoryId = ? AND LookupValue = ?", "B_KELASKHUSUS", *kodeKelas).
		Count(&countKelas).Error; err != nil {
		return fmt.Errorf("failed to validate kode_kelas: %w", err)
	}
	if countKelas == 0 {
		return fmt.Errorf("kode_kelas %s not found in lookup", *kodeKelas)
	}

	if kodeFotang != nil && strings.TrimSpace(*kodeFotang) != "" {
		var countFotang int64
		if err := db.Model(&domain.AppLookup{}).
			Where("CategoryId = ? AND LookupValue = ?", "B_FOTHANG", *kodeFotang).
			Count(&countFotang).Error; err != nil {
			return fmt.Errorf("failed to validate kode_fotang: %w", err)
		}
		if countFotang == 0 {
			return fmt.Errorf("kode_fotang %s not found in lookup", *kodeFotang)
		}
	}
	return nil
}

func (s *KelasService) Create(payload domain.Kelas, c *gin.Context) (domain.Kelas, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Kelas{}, fmt.Errorf("validasi gagal: %w", err)
	}
	if err := validateKelasLookups(s.db, payload.KodeKelas, payload.KodeFotang); err != nil {
		return domain.Kelas{}, err
	}

	if payload.TrxId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS").Select("ISNULL(MAX(trxid), 0)").Row().Scan(&maxID)
		payload.TrxId = maxID + 1
	}

	userID := getUserID(c)
	statusTrue := true
	modActI := "I"
	now := time.Now()

	payload.Status = &statusTrue
	payload.ModAct = &modActI
	payload.ModBy = &userID
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasService) Get(id string) (domain.Kelas, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Kelas{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.Kelas
	if err := s.db.
		Preload("KelasName", "CategoryId = ?", "B_KELASKHUSUS").
		Preload("FotangName", "CategoryId = ?", "B_FOTHANG").
		Where("trxid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
		}
		return domain.Kelas{}, err
	}
	return item, nil
}

func (s *KelasService) Peserta(id string, c *gin.Context, page int, limit int) ([]domain.KelasPesertaResponse, int64, error) {
	var items []domain.KelasPesertaResponse

	var subWhId int64

	err := s.db.Model(&domain.AdminMatrix{}).
		Where("LOGINID = ?", getUserID(c)).
		Limit(1).
		Pluck("SUBWHID", &subWhId).Error

	if err != nil {
		return items, 0, fmt.Errorf("database query error: %w", err)
	}

	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	var findErr error

	var trxID int64
	if id != "" {
		trxID, _ = strconv.ParseInt(id, 10, 64)
	}

	// before: SP_TRX_KELAS_GET_PESERTA, now: SP_TRX_KELAS_PESERTA_SEARCH_DATA
	sortDirection := "ASCENDING"
	rows, err := s.db.Raw("EXEC [dbo].[SP_TRX_KELAS_PESERTA_SEARCH_DATA] @PageSize = ?, @CurrentPage = ?, @SortDirection = ?, @TrxId = ?, @FotangId = ?",
		limit,
		page,
		sortDirection,
		trxID,
		subWhId,
	).Rows()

	if err != nil {
		return items, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("kelas tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return items, 0, findErr
	}

	for rows.Next() {
		var item domain.KelasPesertaResponse
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
			fmt.Printf("kelas rows.Scan error: %v\n", err)
			findErr = fmt.Errorf("scan error on row: %w", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	if findErr != nil {
		return []domain.KelasPesertaResponse{}, 0, findErr
	}

	return items, total, nil
}

func (s *KelasService) Update(id string, payload domain.Kelas, c *gin.Context) (domain.Kelas, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Kelas{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Kelas
	if err := s.db.Where("trxid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Kelas{}, fmt.Errorf("kelas %s not found", id)
		}
		return domain.Kelas{}, err
	}

	if payload.KodeKelas != nil {
		if strings.TrimSpace(*payload.KodeKelas) == "" {
			return domain.Kelas{}, errors.New("kode_kelas cannot be empty")
		}
		var countKelas int64
		if err := s.db.Model(&domain.AppLookup{}).
			Where("CategoryId = ? AND LookupValue = ?", "B_KELASKHUSUS", *payload.KodeKelas).
			Count(&countKelas).Error; err != nil {
			return domain.Kelas{}, fmt.Errorf("failed to validate kode_kelas: %w", err)
		}
		if countKelas == 0 {
			return domain.Kelas{}, fmt.Errorf("kode_kelas %s not found in lookup", *payload.KodeKelas)
		}
		item.KodeKelas = payload.KodeKelas
	}

	if payload.KodeFotang != nil {
		if strings.TrimSpace(*payload.KodeFotang) != "" {
			var countFotang int64
			if err := s.db.Model(&domain.AppLookup{}).
				Where("CategoryId = ? AND LookupValue = ?", "B_FOTHANG", *payload.KodeFotang).
				Count(&countFotang).Error; err != nil {
				return domain.Kelas{}, fmt.Errorf("failed to validate kode_fotang: %w", err)
			}
			if countFotang == 0 {
				return domain.Kelas{}, fmt.Errorf("kode_fotang %s not found in lookup", *payload.KodeFotang)
			}
			item.KodeFotang = payload.KodeFotang
		} else {
			item.KodeFotang = nil
		}
	}

	if payload.StartDate != nil {
		item.StartDate = payload.StartDate
	}
	if payload.EndDate != nil {
		item.EndDate = payload.EndDate
	}
	if payload.Lokasi != nil {
		item.Lokasi = payload.Lokasi
	}
	if payload.PIC != nil {
		item.PIC = payload.PIC
	}
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}
	if payload.Level != nil {
		item.Level = payload.Level
	}
	if payload.Mc1 != nil {
		item.Mc1 = payload.Mc1
	}
	if payload.Mc2 != nil {
		item.Mc2 = payload.Mc2
	}
	if payload.Mc3 != nil {
		item.Mc3 = payload.Mc3
	}
	if payload.Mc4 != nil {
		item.Mc4 = payload.Mc4
	}
	if payload.Mc5 != nil {
		item.Mc5 = payload.Mc5
	}
	if payload.Deadline != nil {
		item.Deadline = payload.Deadline
	}

	userID := getUserID(c)
	modActU := "U"
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.Kelas{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Kelas
	if err := s.db.Where("trxid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas %s not found", id)
		}
		return err
	}

	userID := getUserID(c)
	statusFalse := false
	modActD := "D"
	now := time.Now()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
