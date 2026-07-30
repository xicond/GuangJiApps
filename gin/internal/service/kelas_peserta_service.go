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

type KelasPesertaService struct {
	db       *gorm.DB
	resource string
}

func NewKelasPesertaService(db *gorm.DB) *KelasPesertaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasPesertaService{db: db, resource: "kelas_peserta"}
}

func (s *KelasPesertaService) List(id string, c *gin.Context, page int, limit int) ([]domain.KelasPesertaResponse, int64, error) {
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

func (s *KelasPesertaService) Create(payload domain.KelasPeserta, c *gin.Context) (domain.KelasPeserta, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("validasi gagal: %w", err)
	}

	if payload.DetailId == 0 {
		var maxID int32
		s.db.Table("T_TRX_KELAS_PESERTA").Select("ISNULL(MAX(detailid), 0)").Row().Scan(&maxID)
		payload.DetailId = maxID + 1
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
		return domain.KelasPeserta{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasPesertaService) Get(id string) (domain.KelasPeserta, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasPeserta
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPeserta{}, fmt.Errorf("kelas peserta %s not found", id)
		}
		return domain.KelasPeserta{}, err
	}
	return item, nil
}

func (s *KelasPesertaService) Update(id string, payload domain.KelasPeserta, c *gin.Context) (domain.KelasPeserta, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasPeserta
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPeserta{}, fmt.Errorf("kelas peserta %s not found", id)
		}
		return domain.KelasPeserta{}, err
	}

	if payload.TrxId != 0 {
		item.TrxId = payload.TrxId
	}
	if payload.IdPeserta != nil {
		item.IdPeserta = payload.IdPeserta
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
	if payload.Keterangan != nil {
		item.Keterangan = payload.Keterangan
	}
	if payload.Lulus != nil {
		item.Lulus = payload.Lulus
	}
	if payload.KeteranganLulus != nil {
		item.KeteranganLulus = payload.KeteranganLulus
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
	now := time.Now()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasPesertaService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.KelasPeserta
	if err := s.db.Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas peserta %s not found", id)
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
