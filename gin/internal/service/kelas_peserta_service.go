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

// KelasPesertaService manages participant (peserta) enrollments and attendance within a class session.
type KelasPesertaService struct {
	db       *gorm.DB
	resource string
}

// NewKelasPesertaService initializes a new KelasPesertaService instance.
//
// Parameters:
//   - db: *gorm.DB database connection pool (defaults to primary if nil)
//
// Returns:
//   - *KelasPesertaService: initialized service pointer
func NewKelasPesertaService(db *gorm.DB) *KelasPesertaService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasPesertaService{db: db, resource: "kelas_peserta"}
}

// validatePesertaLookups verifies foreign keys and lookup values for class participants.
//
// Parameters:
//   - db: *gorm.DB database connection
//   - payload: *domain.KelasPesertaBulkRequest containing participant IDs and lookup values
//
// Returns:
//   - error: *ValidationError with error details if validation fails, or nil
func validatePesertaLookups(db *gorm.DB, payload *domain.KelasPesertaBulkRequest) error {
	type lookupCheck struct {
		fieldName string
		queryFn   func(db *gorm.DB) error
	}

	var checks []lookupCheck

	// 1. TrxId from Kelas
	if payload.TrxId != 0 {
		checks = append(checks, lookupCheck{
			fieldName: "trx_id",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.Kelas{}).Where("trxid = ?", payload.TrxId).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("trx_id %d tidak ditemukan di Kelas", payload.TrxId)
				}
				return nil
			},
		})
	}

	// 2. IdPeserta from Umat
	if len(payload.IdPeserta) > 0 {
		checks = append(checks, lookupCheck{
			fieldName: "id_peserta",
			queryFn: func(db *gorm.DB) error {
				var foundIds []int32
				if err := db.Model(&domain.Umat{}).Where("id IN (?)", payload.IdPeserta).Pluck("id", &foundIds).Error; err != nil {
					return err
				}
				foundMap := make(map[int32]bool, len(foundIds))
				for _, fid := range foundIds {
					foundMap[fid] = true
				}
				var missing []string
				for _, id := range payload.IdPeserta {
					if !foundMap[id] {
						missing = append(missing, strconv.Itoa(int(id)))
					}
				}
				if len(missing) == 1 {
					return fmt.Errorf("id_peserta %s tidak ditemukan di Umat", missing[0])
				} else if len(missing) > 1 {
					return fmt.Errorf("id_peserta %s tidak ditemukan di Umat", strings.Join(missing, ", "))
				}
				return nil
			},
		})
	}

	// 3. TimKerja if notempty B_TIMKERJA
	if payload.TimKerja != nil && strings.TrimSpace(*payload.TimKerja) != "" {
		val := strings.TrimSpace(*payload.TimKerja)
		checks = append(checks, lookupCheck{
			fieldName: "tim_kerja",
			queryFn: func(db *gorm.DB) error {
				var count int64
				if err := db.Model(&domain.AppLookup{}).
					Where("CategoryId = ? AND (LookupValue = ? OR LookupId = ?)", "B_TIMKERJA", val, val).
					Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					return fmt.Errorf("field tim_kerja nilai '%s' tidak valid", val)
				}
				return nil
			},
		})
	}

	if len(checks) == 0 {
		return nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		details = make(map[string][]string)
	)

	for _, check := range checks {
		c := check
		wg.Go(func() {
			sess := db.Session(&gorm.Session{})
			if err := c.queryFn(sess); err != nil {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], err.Error())
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// List queries participants of a class session using stored procedure with pagination.
//
// Parameters:
//   - id: class TrxId string
//   - c: *gin.Context containing user auth session for branch isolation
//   - page: page number (1-based)
//   - limit: page size limit
//
// Returns:
//   - []domain.KelasPesertaResponse: list of participant records
//   - int64: total matching count
//   - error: query or scan error if failed
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

// updateUmatSDPemula updates umat status, graduation date, and ikrar attributes for SD Pemula (004) classes.
//
// Parameters:
//   - tx: *gorm.DB transaction instance
//   - idPeserta: umat participant ID
//   - trxId: kelas transaction ID
//   - detailId: current participant detail record ID
//   - lulus: graduation boolean pointer
//   - umat: optional domain.Umat payload containing ikrar updates
//   - userID: modifying admin user ID
//   - now: audit timestamp
//
// Returns:
//   - error: database error if update fails
func updateUmatSDPemula(tx *gorm.DB, idPeserta int32, trxId int32, detailId int32, lulus *bool, umat *domain.Umat, userID int32, now domain.DateTime) error {
	if idPeserta == 0 || trxId == 0 {
		return nil
	}

	var kelas domain.Kelas
	if err := tx.Where("trxid = ?", trxId).First(&kelas).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if kelas.KodeKelas == nil || strings.TrimSpace(*kelas.KodeKelas) != "004" {
		return nil
	}

	umatUpdates := make(map[string]interface{})

	// 1. Update Ikrar if provided in payload
	if umat != nil {
		umatUpdates["ikrar1"] = umat.Ikrar1
		umatUpdates["ikrar2"] = umat.Ikrar2
		umatUpdates["ikrar3"] = umat.Ikrar3
		umatUpdates["ikrar4"] = umat.Ikrar4
		umatUpdates["ikrar5"] = umat.Ikrar5
		umatUpdates["ikrar6"] = umat.Ikrar6
	}

	// 2. Check if participant is passed (Lulus) for this class or any other active 004 class
	isLulusCurrent := lulus != nil && *lulus

	var otherPassCount int64

	if !isLulusCurrent {
		query := tx.Table("T_TRX_KELAS_PESERTA kp").
			Joins("JOIN T_TRX_KELAS k ON k.trxid = kp.trxid").
			Where("kp.idpeserta = ? AND kp.status = 1 AND kp.lulus = 1 AND k.kodekelas = '004'", idPeserta)
		if detailId != 0 {
			query = query.Where("kp.detailid != ?", detailId)
		}
		if err := query.Count(&otherPassCount).Error; err != nil {
			return err
		}
	}

	hasPassedSD3 := isLulusCurrent || (otherPassCount > 0)

	var currentUmat domain.Umat
	if err := tx.Select("id, statusumat").Where("id = ?", idPeserta).First(&currentUmat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("umat %d not found", idPeserta)
		}
		return err
	}

	if hasPassedSD3 {
		umatUpdates["sd3"] = true
		if kelas.EndDate != nil {
			umatUpdates["tanggalsd3"] = kelas.EndDate
		}
		if kelas.KodeFotang != nil {
			umatUpdates["tempatsd3"] = kelas.KodeFotang
		}
		// Update statusumat to 001 (Pengabdi) if previous is 006 (Umat Baru) or nil/empty
		if currentUmat.StatusUmat == nil || strings.TrimSpace(*currentUmat.StatusUmat) == "" || strings.TrimSpace(*currentUmat.StatusUmat) == "006" {
			status001 := "001"
			umatUpdates["statusumat"] = status001
		}
	} else {
		umatUpdates["sd3"] = false
		umatUpdates["tanggalsd3"] = nil
		umatUpdates["tempatsd3"] = nil
		// Revert statusumat to 006 (Umat Baru) if previous status is 001 (Pengabdi)
		if currentUmat.StatusUmat != nil && strings.TrimSpace(*currentUmat.StatusUmat) == "001" {
			status006 := "006"
			umatUpdates["statusumat"] = status006
		}
	}

	if len(umatUpdates) > 0 {
		umatUpdates["modby"] = userID
		umatUpdates["moddate"] = now
		umatUpdates["modact"] = "U"

		if err := tx.Model(&domain.Umat{}).Where("id = ?", idPeserta).Updates(umatUpdates).Error; err != nil {
			return fmt.Errorf("failed to update umat for SD Pemula: %w", err)
		}
	}

	return nil
}

// Create registers a single participant in a class session.
//
// Parameters:
//   - payload: domain.KelasPeserta entity payload
//   - c: *gin.Context containing user auth session
//
// Returns:
//   - domain.KelasPeserta: created participant entity with generated DetailId
//   - error: validation, ID generation, or database error if failed
func (s *KelasPesertaService) Create(payload domain.KelasPeserta, c *gin.Context) (domain.KelasPeserta, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("Validation failed: %w", err)
	}

	var ids []int32
	if payload.IdPeserta != nil && *payload.IdPeserta != 0 {
		ids = []int32{*payload.IdPeserta}
	}
	bulkReq := domain.KelasPesertaBulkRequest{
		TrxId:     payload.TrxId,
		IdPeserta: ids,
		TimKerja:  payload.TimKerja,
	}
	if err := validatePesertaLookups(s.db, &bulkReq); err != nil {
		return domain.KelasPeserta{}, err
	}

	var genResult struct {
		GeneratedId int32
	}
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	errId := s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASPESERTAID", nowStr, 1).Scan(&genResult).Error
	if errId != nil {
		return domain.KelasPeserta{}, fmt.Errorf("failed to generate ID: %w", errId)
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

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if payload.IdPeserta != nil && *payload.IdPeserta != 0 {
			if err := updateUmatSDPemula(tx, *payload.IdPeserta, payload.TrxId, payload.DetailId, payload.Lulus, payload.Umat, userID, now); err != nil {
				return err
			}
		}

		if err := tx.Omit("Umat", "Kelas").Create(&payload).Error; err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
		return nil
	})

	if err != nil {
		return domain.KelasPeserta{}, err
	}

	if payload.IdPeserta != nil && *payload.IdPeserta != 0 {
		var umat domain.Umat
		if err := s.db.Where("id = ?", *payload.IdPeserta).Take(&umat).Error; err == nil {
			payload.Umat = &umat
		}
	}

	return payload, nil
}

// CreateBulk registers multiple participants in a class session in a single database transaction.
//
// Parameters:
//   - payload: domain.KelasPesertaBulkRequest containing participant IDs and shared properties
//   - c: *gin.Context containing user auth session
//
// Returns:
//   - []domain.KelasPeserta: slice of created participant records
//   - error: validation, ID generation, or transaction error if failed
func (s *KelasPesertaService) CreateBulk(payload domain.KelasPesertaBulkRequest, c *gin.Context) ([]domain.KelasPeserta, error) {
	if err := ValidateStruct(payload); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}

	// Deduplicate IdPeserta preserving order
	uniqueIds := make([]int32, 0, len(payload.IdPeserta))
	seen := make(map[int32]bool, len(payload.IdPeserta))
	for _, id := range payload.IdPeserta {
		if id > 0 && !seen[id] {
			seen[id] = true
			uniqueIds = append(uniqueIds, id)
		}
	}
	if len(uniqueIds) == 0 {
		return nil, &ValidationError{Details: map[string][]string{"id_peserta": {"id_peserta harus memiliki minimal 1 id valid"}}}
	}
	payload.IdPeserta = uniqueIds

	if err := validatePesertaLookups(s.db, &payload); err != nil {
		return nil, err
	}

	userID := getUserID(c)
	statusTrue := true
	modActI := "I"
	now := domain.NowDateTime()
	nowStr := time.Now().Format("2006-01-02 15:04:05")

	results := make([]domain.KelasPeserta, 0, len(payload.IdPeserta))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range payload.IdPeserta {
			var genResult struct {
				GeneratedId int32
			}
			errId := tx.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "KELASPESERTAID", nowStr, 1).Scan(&genResult).Error
			if errId != nil {
				return fmt.Errorf("failed to generate ID: %w", errId)
			}

			idCopy := id
			item := domain.KelasPeserta{
				DetailId:        genResult.GeneratedId,
				TrxId:           payload.TrxId,
				IdPeserta:       &idCopy,
				Sumbangan:       payload.Sumbangan,
				Barang:          payload.Barang,
				TimKerja:        payload.TimKerja,
				Keterangan:      payload.Keterangan,
				Status:          &statusTrue,
				ModAct:          &modActI,
				ModBy:           &userID,
				ModDate:         &now,
				Lulus:           payload.Lulus,
				KeteranganLulus: payload.KeteranganLulus,
				Anak:            payload.Anak,
				Suster:          payload.Suster,
				Menginap:        payload.Menginap,
				MakananPagi:     payload.MakananPagi,
				MakananSiang:    payload.MakananSiang,
				MakananMalam:    payload.MakananMalam,
			}

			if idCopy != 0 {
				if err := updateUmatSDPemula(tx, idCopy, item.TrxId, item.DetailId, item.Lulus, nil, userID, now); err != nil {
					return err
				}
			}

			if err := tx.Omit("Umat", "Kelas").Create(&item).Error; err != nil {
				return fmt.Errorf("failed to create record for id_peserta %d: %w", idCopy, err)
			}

			results = append(results, item)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		var umats []domain.Umat
		if err := s.db.Where("id IN (?)", payload.IdPeserta).Find(&umats).Error; err == nil {
			umatMap := make(map[int32]*domain.Umat, len(umats))
			for i := range umats {
				umatMap[umats[i].ID] = &umats[i]
			}
			for i := range results {
				if results[i].IdPeserta != nil {
					if u, ok := umatMap[*results[i].IdPeserta]; ok {
						results[i].Umat = u
					}
				}
			}
		}
	}

	return results, nil
}

// Get retrieves a single participant record by detail ID with preloaded Umat data.
//
// Parameters:
//   - id: string representation of the detailid primary key
//
// Returns:
//   - domain.KelasPeserta: participant entity
//   - error: not found or database error
func (s *KelasPesertaService) Get(id string) (domain.KelasPeserta, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasPeserta
	if err := s.db.Preload("Umat").Where("detailid = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPeserta{}, fmt.Errorf("kelas peserta %s not found", id)
		}
		return domain.KelasPeserta{}, err
	}
	return item, nil
}

// GetByIdPerserta retrieves a participant record by umat ID and optional class transaction ID.
//
// Parameters:
//   - idPeserta: string representation of the idpeserta foreign key
//   - trxIds: optional class transaction IDs to filter by
//
// Returns:
//   - domain.KelasPeserta: participant entity
//   - error: not found or database error
func (s *KelasPesertaService) GetByIdPerserta(idPeserta string, trxIds ...string) (domain.KelasPeserta, error) {
	parsedInt, err := strconv.Atoi(idPeserta)
	if err != nil {
		return domain.KelasPeserta{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.KelasPeserta
	query := s.db.Preload("Umat").Where("idpeserta = ?", parsedInt)
	if len(trxIds) > 0 && strings.TrimSpace(trxIds[0]) != "" {
		if trxIdInt, err := strconv.Atoi(strings.TrimSpace(trxIds[0])); err == nil {
			query = query.Where("trxid = ?", trxIdInt)
		}
	}
	if err := query.Order("status DESC, detailid DESC").Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.KelasPeserta{}, fmt.Errorf("kelas peserta with id_peserta %s not found", idPeserta)
		}
		return domain.KelasPeserta{}, err
	}
	return item, nil
}

// GetByIdPeserta is an alias for GetByIdPerserta fixing the typo in method naming.
//
// Parameters:
//   - idPeserta: string representation of the idpeserta foreign key
//   - trxIds: optional class transaction IDs to filter by
//
// Returns:
//   - domain.KelasPeserta: participant entity
//   - error: not found or database error
func (s *KelasPesertaService) GetByIdPeserta(idPeserta string, trxIds ...string) (domain.KelasPeserta, error) {
	return s.GetByIdPerserta(idPeserta, trxIds...)
}

// Update modifies an existing participant record and synchronizes SD Pemula status if applicable.
//
// Parameters:
//   - id: string representation of the detailid primary key
//   - payload: domain.KelasPeserta entity containing updated fields
//   - c: *gin.Context containing user auth session
//
// Returns:
//   - domain.KelasPeserta: updated entity
//   - error: validation, not found, or database error if failed
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

	targetVal := item
	if payload.TrxId != 0 {
		targetVal.TrxId = payload.TrxId
	}
	if payload.IdPeserta != nil {
		targetVal.IdPeserta = payload.IdPeserta
	}
	if payload.TimKerja != nil {
		targetVal.TimKerja = payload.TimKerja
	}

	var ids []int32
	if targetVal.IdPeserta != nil && *targetVal.IdPeserta != 0 {
		ids = []int32{*targetVal.IdPeserta}
	}
	bulkReq := domain.KelasPesertaBulkRequest{
		TrxId:     targetVal.TrxId,
		IdPeserta: ids,
		TimKerja:  targetVal.TimKerja,
	}
	if err := validatePesertaLookups(s.db, &bulkReq); err != nil {
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
	now := domain.NowDateTime()

	item.ModAct = &modActU
	item.ModBy = &userID
	item.ModDate = &now

	err = s.db.Transaction(func(tx *gorm.DB) error {
		idPeserta := int32(0)
		if item.IdPeserta != nil {
			idPeserta = *item.IdPeserta
		}
		if idPeserta != 0 {
			if err := updateUmatSDPemula(tx, idPeserta, item.TrxId, item.DetailId, item.Lulus, payload.Umat, userID, now); err != nil {
				return err
			}
		}

		if err := tx.Omit("Umat", "Kelas").Save(&item).Error; err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
		return nil
	})

	if err != nil {
		return domain.KelasPeserta{}, err
	}

	if item.IdPeserta != nil && *item.IdPeserta != 0 {
		var umat domain.Umat
		if err := s.db.Where("id = ?", *item.IdPeserta).Take(&umat).Error; err == nil {
			item.Umat = &umat
		}
	}

	return item, nil
}

// Delete soft-deletes a participant record (Status = false) and reverts SD Pemula status if appropriate.
//
// Parameters:
//   - id: string representation of the detailid primary key
//   - c: *gin.Context containing user auth session
//
// Returns:
//   - error: not found or database error if failed
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
	now := domain.NowDateTime()

	item.Status = &statusFalse
	item.ModAct = &modActD
	item.ModBy = &userID
	item.ModDate = &now

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return fmt.Errorf("failed to delete record: %w", err)
		}
		if item.IdPeserta != nil && *item.IdPeserta != 0 {
			lulusFalse := false
			if err := updateUmatSDPemula(tx, *item.IdPeserta, item.TrxId, item.DetailId, &lulusFalse, nil, userID, now); err != nil {
				return err
			}
		}
		return nil
	})
}

// LoadPrevious loads participants eligible from previous class levels using stored procedure.
//
// Parameters:
//   - id: string representation of current class TrxId
//   - c: *gin.Context containing user auth session
//   - page: pagination page number
//   - limit: page size limit
//
// Returns:
//   - []domain.KelasPesertaPrevious: list of eligible candidates from earlier classes
//   - int64: total eligible count
//   - error: query error if failed
func (s *KelasPesertaService) LoadPrevious(id string, c *gin.Context, page int, limit int) ([]domain.KelasPesertaPrevious, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	trxID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return []domain.KelasPesertaPrevious{}, 0, fmt.Errorf("invalid ID format: %w", err)
	}

	var (
		kodeKelas string
		subWhId   int64
		wg        sync.WaitGroup
		errKelas  error
	)

	wg.Go(func() {
		errKelas = s.db.Session(&gorm.Session{}).Model(&domain.Kelas{}).Where("trxid = ?", trxID).Pluck("kodekelas", &kodeKelas).Error
	})
	wg.Go(func() {
		if c != nil {
			_ = s.db.Session(&gorm.Session{}).Model(&domain.AdminMatrix{}).
				Where("LOGINID = ?", getUserID(c)).
				Limit(1).
				Pluck("SUBWHID", &subWhId).Error
		}
	})
	wg.Wait()

	if errKelas != nil {
		return []domain.KelasPesertaPrevious{}, 0, fmt.Errorf("database query error: %w", errKelas)
	}
	if kodeKelas == "" {
		return []domain.KelasPesertaPrevious{}, 0, errors.New("kelas tidak ditemukan")
	}

	rows, err := s.db.Raw("EXEC [dbo].[SP_TRX_KELAS_GET_PESERTA_BY_CODE_AND_LEVEL] @TrxId = ?, @KodeKelas = ?, @SubWhId = ?",
		trxID,
		kodeKelas,
		subWhId,
	).Rows()
	if err != nil {
		return []domain.KelasPesertaPrevious{}, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return []domain.KelasPesertaPrevious{}, 0, fmt.Errorf("database error: %w", err)
	}

	all := make([]domain.KelasPesertaPrevious, 0, 64)
	for rows.Next() {
		var item domain.KelasPesertaPrevious
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))
			matched := false
			for j := 0; j < t.NumField(); j++ {
				field := t.Field(j)
				gormTag := field.Tag.Get("gorm")

				isColMatch := false
				for _, part := range strings.Split(gormTag, ";") {
					part = strings.TrimSpace(strings.ToLower(part))
					if part == "column:"+cleanCol {
						isColMatch = true
						break
					}
				}

				if isColMatch || strings.ToLower(field.Name) == cleanCol {
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

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		if item.Id == 0 && item.IdPeserta != 0 {
			item.Id = item.IdPeserta
		}
		item.FotangAktif = strings.TrimSpace(item.FotangAktif)
		item.FotangCiuTao = strings.TrimSpace(item.FotangCiuTao)
		item.FotangAktifDesc = strings.TrimSpace(item.FotangAktifDesc)
		item.FotangCiuTaoDesc = strings.TrimSpace(item.FotangCiuTaoDesc)
		all = append(all, item)
	}

	total := int64(len(all))
	start := (page - 1) * limit
	if start >= len(all) {
		return []domain.KelasPesertaPrevious{}, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}

	return all[start:end], total, nil
}

