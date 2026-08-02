package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
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

func validateUmatLookups(db *gorm.DB, payload *domain.Umat) error {
	type lookupCheck struct {
		fieldName  string
		categoryID string
		val        string
	}

	checks := [...]lookupCheck{
		{fieldName: "status_umat", categoryID: "B_STATUS", val: payload.StatusUmat},
		{fieldName: "tim_kerja", categoryID: "B_TIMKERJA", val: payload.TimKerja},
		{fieldName: "kelas_khusus", categoryID: "B_KELAS", val: payload.KelasKhusus},
		{fieldName: "kelas_umum", categoryID: "B_KELASUMUM", val: payload.KelasUmum},
		{fieldName: "tempat_sd2", categoryID: "B_FOTHANG", val: payload.TempatSd2},
		{fieldName: "tempat_sd3", categoryID: "B_FOTHANG", val: payload.TempatSd3},
		{fieldName: "fotang_aktif", categoryID: "B_FOTHANG", val: payload.FotangAktif},
		{fieldName: "fotang_chiutao", categoryID: "B_FOTHANG", val: payload.FotangChiutao},
		{fieldName: "tcs", categoryID: "B_TCS", val: payload.Tcs},
		{fieldName: "waktu_chiutao_mandarin", categoryID: "B_WAKTUCIUTAO", val: payload.WaktuChiutaoMandarin},
		{fieldName: "pendidikan", categoryID: "B_PENDIDIKAN", val: payload.Pendidikan},
		{fieldName: "pekerjaan", categoryID: "B_PEKERJAAN", val: payload.Pekerjaan},
	}

	activeChecks := make([]lookupCheck, 0, len(checks))
	for _, c := range checks {
		if strings.TrimSpace(c.val) != "" {
			activeChecks = append(activeChecks, c)
		}
	}

	if len(activeChecks) == 0 {
		return nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		details = make(map[string][]string)
	)

	wg.Add(len(activeChecks))
	for _, check := range activeChecks {
		go func(c lookupCheck) {
			defer wg.Done()
			var count int64
			if err := db.Session(&gorm.Session{}).Model(&domain.AppLookup{}).
				Where("CategoryId = ? AND (LookupValue = ? OR LookupId = ?) AND Status = ?", c.categoryID, c.val, c.val, true).
				Count(&count).Error; err != nil {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], fmt.Sprintf("gagal memvalidasi %s: %v", c.fieldName, err))
				mu.Unlock()
				return
			}
			if count == 0 {
				mu.Lock()
				details[c.fieldName] = append(details[c.fieldName], fmt.Sprintf("field %s nilai '%s' tidak valid", c.categoryID, c.val))
				mu.Unlock()
			}
		}(check)
	}

	wg.Wait()

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
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

	// Execute Count and List queries concurrently for optimal latency
	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	// Goroutine 1: Concurrent Count query
	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	}()

	// Goroutine 2: Concurrent Find items query
	go func() {
		defer wg.Done()
		if err := query.Session(&gorm.Session{}).
			Preload("JenisKelaminInfo", "CategoryId = ? AND Status = ?", "B_JENISKELAMIN", true).
			Limit(limit).
			Offset(offset).
			Order("ID ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("umat tidak ditemukan")
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
		return []domain.Umat{}, 0, findErr
	}

	// Dynamic post-processing for lowest latency
	now := time.Now()
	for i := range items {
		if !items[i].TanggalLahir.IsZero() && items[i].TanggalLahir.Year() > 1900 {
			items[i].Usia = int32(now.Year() - items[i].TanggalLahir.Year())
		}
		if items[i].JenisKelaminInfo != nil && items[i].JenisKelaminInfo.LookupDescription != nil && *items[i].JenisKelaminInfo.LookupDescription != "" {
			items[i].JenisKelamin = *items[i].JenisKelaminInfo.LookupDescription
		}
	}

	return items, total, nil
}

func (s *UmatService) Create(payload domain.Umat, c *gin.Context) (domain.Umat, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Umat{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validateUmatLookups(s.db, &payload); err != nil {
		return domain.Umat{}, err
	}

	// item.Id = int.Parse(DalBrand.GenerateId_UmatId(DateTime.Now, 1)[0]);
	//item.Kode = DalBrand.GenerateId_UmatId(DateTime.Now, 1)[0];
	// item.Kode = CurrentLogin.SubWhName + "-" + item.Id.ToString();

	// 2. IMPORTANT: Do NOT generate a random UnixNano string for ID!
	// Your SQL Server schema defines [id] INT NOT NULL.
	// If it is NOT an IDENTITY column, we calculate the next integer sequence manually.
	var maxID int32
	s.db.Table("T_BUS_UMAT").Select("ISNULL(MAX(id), 0)").Row().Scan(&maxID)
	payload.ID = maxID + 1

	// item.Id = int.Parse(DalBrand.GenerateId_UmatId(DateTime.Now, 1)[0]);
	//item.Kode = DalBrand.GenerateId_UmatId(DateTime.Now, 1)[0];
	// item.Kode = CurrentLogin.SubWhName + "-" + item.Id.ToString();

	// 3. Populate matching schema structural constraints
	payload.Status = true        // Active status mapping
	payload.ModAct = "I"         // 'I' standard legacy flag for Insert
	payload.ModBy = getUserID(c) // Default system user ID matching INT type
	payload.ModDate = time.Now() // Local server time object

	// 4. Persist the new entity to the database pool
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *UmatService) Get(id string) (domain.Umat, error) {
	var item domain.Umat
	if err := s.db.Preload("JenisKelaminInfo", "CategoryId = ? AND Status = ?", "B_JENISKELAMIN", true).First(&item, "id = ?", id).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("umat %s not found", id)
	}
	if !item.TanggalLahir.IsZero() && item.TanggalLahir.Year() > 1900 {
		item.Usia = int32(time.Now().Year() - item.TanggalLahir.Year())
	}
	// Dont activate this, get from JenisKelaminInfo
	/* if item.JenisKelaminInfo != nil && item.JenisKelaminInfo.LookupDescription != nil && *item.JenisKelaminInfo.LookupDescription != "" {
		item.JenisKelamin = *item.JenisKelaminInfo.LookupDescription
	} */
	return item, nil
}

func (s *UmatService) Update(id string, payload domain.Umat, c *gin.Context) (domain.Umat, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Umat{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validateUmatLookups(s.db, &payload); err != nil {
		return domain.Umat{}, err
	}

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
	item.Alias = payload.Alias
	item.NamaIndonesia = payload.NamaIndonesia
	item.Marga = payload.Marga
	item.NamaMandarin = payload.NamaMandarin
	item.Alamat = payload.Alamat
	item.Alamat2 = payload.Alamat2
	item.Telepon = payload.Telepon
	item.Mobile = payload.Mobile
	item.TempatLahir = payload.TempatLahir
	item.TanggalLahir = payload.TanggalLahir
	item.Usia = payload.Usia
	item.Wilayah = payload.Wilayah
	item.JenisKelamin = payload.JenisKelamin
	item.Pekerjaan = payload.Pekerjaan
	item.Pendidikan = payload.Pendidikan
	item.TanggalChiutaoInt = payload.TanggalChiutaoInt
	item.TanggalChiutaoMan = payload.TanggalChiutaoMan
	item.TahunChiutaoMandarin = payload.TahunChiutaoMandarin
	item.WaktuChiutaoMandarin = payload.WaktuChiutaoMandarin
	item.Pengajak = payload.Pengajak
	item.PengajakManual = payload.PengajakManual
	item.Penanggung = payload.Penanggung
	item.PenanggungManual = payload.PenanggungManual
	item.Tcs = payload.Tcs
	item.UangPahala = payload.UangPahala
	item.FotangChiutao = payload.FotangChiutao
	item.FotangAktif = payload.FotangAktif
	item.Sd2 = payload.Sd2
	item.TempatSd2 = payload.TempatSd2
	item.TanggalSd2 = payload.TanggalSd2
	item.Sd3 = payload.Sd3
	item.TempatSd3 = payload.TempatSd3
	item.TanggalSd3 = payload.TanggalSd3
	item.KelasUmum = payload.KelasUmum
	item.KelasKhusus = payload.KelasKhusus
	item.ChingKhou = payload.ChingKhou
	item.TanggalChingKhou = payload.TanggalChingKhou
	item.TanggalAncuo = payload.TanggalAncuo
	item.NamaCetyaRumah = payload.NamaCetyaRumah
	item.Meninggal = payload.Meninggal
	item.TanggalMeninggal = payload.TanggalMeninggal
	item.TimKerja = payload.TimKerja
	item.Posisi = payload.Posisi
	item.StatusUmat = payload.StatusUmat
	item.Keterangan = payload.Keterangan
	item.Email = payload.Email
	item.ImagePath = payload.ImagePath

	// Legacy metadata mappings
	item.ModAct = "U"        // 'U' standard legacy flag for Update
	item.ModBy = getUserID(c) // System user ID (int32)
	item.ModDate = time.Now() // Actual time.Time object expected by DATETIME column

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
	item.ModAct = "D"        // 'D' standard legacy flag for Delete
	item.ModBy = getUserID(c) // Default system user ID matching INT type
	item.ModDate = time.Now() // Local server time object

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}
