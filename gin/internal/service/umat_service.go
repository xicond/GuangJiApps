package service

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"golang.org/x/image/draw"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UmatService struct {
	db                   *gorm.DB
	resource             string
	reportServerURL      string
	reportServerUsername string
	reportServerPassword string
}

func NewUmatService(db *gorm.DB) *UmatService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &UmatService{
		db:                   db,
		resource:             "umats",
		reportServerURL:      getReportServerURL(),
		reportServerUsername: getReportServerUsername(),
		reportServerPassword: getReportServerPassword(),
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func validateUmatLookups(db *gorm.DB, payload *domain.Umat) error {
	type lookupCheck struct {
		fieldName  string
		categoryID string
		val        string
	}

	checks := [...]lookupCheck{
		{fieldName: "status_umat", categoryID: "B_STATUS", val: derefString(payload.StatusUmat)},
		{fieldName: "tim_kerja", categoryID: "B_TIMKERJA", val: derefString(payload.TimKerja)},
		{fieldName: "kelas_khusus", categoryID: "B_KELAS", val: derefString(payload.KelasKhusus)},
		{fieldName: "kelas_umum", categoryID: "B_KELASUMUM", val: derefString(payload.KelasUmum)},
		{fieldName: "tempat_sd2", categoryID: "B_FOTHANG", val: derefString(payload.TempatSd2)},
		{fieldName: "tempat_sd3", categoryID: "B_FOTHANG", val: derefString(payload.TempatSd3)},
		{fieldName: "fotang_aktif", categoryID: "B_FOTHANG", val: payload.FotangAktif},
		{fieldName: "fotang_chiutao", categoryID: "B_FOTHANG", val: payload.FotangChiutao},
		{fieldName: "tcs", categoryID: "B_TCS", val: payload.Tcs},
		{fieldName: "waktu_chiutao_mandarin", categoryID: "B_WAKTUCIUTAO", val: payload.WaktuChiutaoMandarin},
		{fieldName: "pendidikan", categoryID: "B_PENDIDIKAN", val: derefString(payload.Pendidikan)},
		{fieldName: "pekerjaan", categoryID: "B_PEKERJAAN", val: derefString(payload.Pekerjaan)},
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

func (s *UmatService) List(c *gin.Context, page int, filters map[string]string, limit int) ([]domain.Umat, int64, error) { //[]domain.Umat {
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

	var subWhVal int64
	userID := getUserID(c)
	row := s.db.Model(&domain.AdminMatrix{}).
		Where("LOGINID = ?", userID).
		Select("SUBWHID").
		Row()
	if row != nil {
		_ = row.Scan(&subWhVal)
	}

	subQueryFotang := s.db.Table("T_APP_LOOKUP").
		Where("T_APP_LOOKUP.CategoryId = 'B_FOTHANG'").
		Where("T_APP_LOOKUP.LookupValue = T_BUS_UMAT.fotangaktif").
		Where("T_APP_LOOKUP.Status = ?", true)

	subQueryJK := s.db.Table("T_APP_LOOKUP").
		Where("T_APP_LOOKUP.CategoryId = 'B_JENISKELAMIN'").
		Where("T_APP_LOOKUP.LookupValue = T_BUS_UMAT.jeniskelamin").
		Where("T_APP_LOOKUP.Status = ?", true)

	// Base query
	query := s.db.Table("T_BUS_UMAT").
		Where("fotangaktif = ?", subWhVal).
		Where("EXISTS (?)", subQueryFotang).
		Where("EXISTS (?)", subQueryJK).
		Where("status = ?", true).
		Where("namaindonesia <> ''").
		Where("namaindonesia is not null ").
		Order("namaindonesia ASC")

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
			if field == "tahunchiutaomandarin" {
				query = query.Where("YEAR([TanggalChiuTaoInt]) = ?", value)
			} else if rule.IsLike {
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
			u := int32(now.Year() - items[i].TanggalLahir.Year())
			items[i].Usia = &u
		}
		if items[i].JenisKelaminInfo != nil && items[i].JenisKelaminInfo.LookupDescription != nil && *items[i].JenisKelaminInfo.LookupDescription != "" {
			items[i].JenisKelamin = *items[i].JenisKelaminInfo.LookupDescription
		}
	}

	return items, total, nil
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		name = "image_" + time.Now().Format("20060102150405") + ".jpg"
	}

	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	clean := re.ReplaceAllString(name, "_")
	clean = strings.ReplaceAll(clean, " ", "_")

	ext := filepath.Ext(clean)
	base := strings.TrimSuffix(clean, ext)
	upperBase := strings.ToUpper(base)
	reservedNames := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
		"COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
		"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
	if reservedNames[upperBase] {
		clean = "_" + clean
	}

	if len(clean) > 150 {
		if len(ext) < 150 {
			clean = clean[:150-len(ext)] + ext
		} else {
			clean = clean[:150]
		}
	}
	return clean
}

func validateImageHeader(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return nil
	}

	var errorsList []string

	// 1. File size check (< 8MB)
	maxSize := int64(8 * 1024 * 1024)
	if fileHeader.Size > maxSize {
		errorsList = append(errorsList, "Ukuran file tidak boleh melebihi 8MB")
	}

	// 2. Open file & decode config for type, resolution, and aspect ratio
	src, err := fileHeader.Open()
	if err != nil {
		errorsList = append(errorsList, "Gagal membuka file gambar")
	} else {
		defer src.Close()

		// Read first 512 bytes for MIME type check
		buffer := make([]byte, 512)
		n, _ := src.Read(buffer)
		contentType := http.DetectContentType(buffer[:n])
		if !strings.HasPrefix(contentType, "image/") {
			errorsList = append(errorsList, "File harus berupa gambar (JPEG, PNG)")
			return NewValidationError(map[string][]string{
				"foto": errorsList,
			})
		}

		// Reset file pointer to beginning for image.DecodeConfig
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			errorsList = append(errorsList, "Gagal memproses file gambar")
		} else {
			cfg, format, decodeErr := image.DecodeConfig(src)
			if decodeErr != nil {
				errorsList = append(errorsList, "Format file gambar tidak dapat dibaca (harus berupa gambar JPEG, PNG)")
			} else {
				if !slices.Contains([]string{"image/jpeg", "image/png"}, format) {
					errorsList = append(errorsList, "Format file gambar tidak didukung (harus berupa gambar JPEG, PNG)")
				}
				// Resolution bounds: width 150..2000, height 200..4000
				if cfg.Width < 150 || cfg.Width > 3000 {
					errorsList = append(errorsList, fmt.Sprintf("Lebar gambar (%dpx) harus di antara 150 hingga 3000 piksel", cfg.Width))
				}
				if cfg.Height < 200 || cfg.Height > 4000 {
					errorsList = append(errorsList, fmt.Sprintf("Tinggi gambar (%dpx) harus di antara 200 hingga 4000 piksel", cfg.Height))
				}

				// Aspect ratio 3:4 (ratio = 0.75)
				ratio := float64(cfg.Width) / float64(cfg.Height)
				if math.Abs(ratio-0.75) > 0.03 {
					errorsList = append(errorsList, fmt.Sprintf("Rasio gambar (saat ini %.2f) harus 3:4 (lebar : tinggi)", ratio))
				}
			}
		}
	}

	if len(errorsList) > 0 {
		return NewValidationError(map[string][]string{
			"foto": errorsList,
		})
	}
	return nil
}

func processAndSaveUmatFoto(db *gorm.DB, umatID int32, namaIndo string, fileHeader *multipart.FileHeader, c *gin.Context, isUpdate bool) error {
	if fileHeader == nil {
		return nil
	}

	uploadBaseDir := os.Getenv("UPLOAD_DIR")
	if uploadBaseDir == "" {
		uploadBaseDir = "uploads"
	}

	docPathStr := fmt.Sprintf("%d", umatID)
	targetDir := filepath.Join(uploadBaseDir, docPathStr)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	safeFileName := sanitizeFilename(fileHeader.Filename)
	dstPath := filepath.Join(targetDir, safeFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("gagal membuka file upload: %w", err)
	}
	defer src.Close()

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("gagal menyimpan file ke disk: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("gagal menulis file ke disk: %w", err)
	}

	docFileStr := docPathStr + "/" + safeFileName
	userID := getUserID(c)
	statusTrue := true
	nowTime := time.Now()

	if isUpdate {
		var existing domain.UmatFoto
		if err := db.Where("Id = ?", umatID).Take(&existing).Error; err == nil {
			modAct := "U"
			existing.DocPath = &docPathStr
			existing.DocFile = &docFileStr
			existing.DocFileName = &safeFileName
			existing.DocDesc = &namaIndo
			existing.Status = &statusTrue
			existing.ModBy = &userID
			existing.ModAct = &modAct
			existing.ModDate = &nowTime
			return db.Save(&existing).Error
		}
	}

	// Create new record in UmatFoto
	var genResult struct {
		GeneratedId int32
	}
	nowStr := nowTime.Format("2006-01-02 15:04:05")
	if errId := db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "FILE", nowStr, 1).Scan(&genResult).Error; errId != nil {
		return fmt.Errorf("failed to generate foto FileID: %w", errId)
	}

	modAct := "I"
	foto := domain.UmatFoto{
		FileID:      genResult.GeneratedId,
		Id:          &umatID,
		DocPath:     &docPathStr,
		DocFile:     &docFileStr,
		DocFileName: &safeFileName,
		DocDesc:     &namaIndo,
		Status:      &statusTrue,
		ModBy:       &userID,
		ModAct:      &modAct,
		ModDate:     &nowTime,
	}

	return db.Create(&foto).Error
}

func (s *UmatService) Create(payload domain.Umat, fileHeader *multipart.FileHeader, c *gin.Context) (domain.Umat, error) {
	if fileHeader != nil {
		if err := validateImageHeader(fileHeader); err != nil {
			return domain.Umat{}, err
		}
	}

	if err := ValidateStruct(payload); err != nil {
		return domain.Umat{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := validateUmatLookups(s.db, &payload); err != nil {
		return domain.Umat{}, err
	}

	type SubWhInfo struct {
		SubWhId  string `gorm:"column:SUBWHID"`
		FullName string `gorm:"column:FULL_NAME"`
	}

	var (
		genResult struct {
			GeneratedId int32
		}
		subWhResult   SubWhInfo
		errId         error
		errSubWh      error
		errPenanggung error
		errPengajak   error
		wg            sync.WaitGroup
	)

	userID := getUserID(c)

	needPenanggung := (payload.Penanggung != nil && string(*payload.Penanggung) != "") && (payload.PenanggungManual == nil || *payload.PenanggungManual == "")
	needPengajak := (payload.Pengajak != nil && string(*payload.Pengajak) != "") && (payload.PengajakManual == nil || *payload.PengajakManual == "")

	tasksCount := 2
	if needPenanggung {
		tasksCount++
	}
	if needPengajak {
		tasksCount++
	}

	wg.Add(tasksCount)

	// Goroutine 1: Execute SP_APP_GenerateId concurrently
	go func() {
		defer wg.Done()
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		errId = s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "UMAT", nowStr, 1).Scan(&genResult).Error
	}()

	// Goroutine 2: Execute SubWhInfo query concurrently
	go func() {
		defer wg.Done()
		subQuery := s.db.Table("T_WH_USER_MATRIX_MST AS A").
			Select("1").
			Where("A.LOGINID = ?", userID).
			Where("A.SUBWHID = T_WH_SUBWH_MST.SUBWHID")

		errSubWh = s.db.Table("T_WH_SUBWH_MST").
			Select("SUBWHID, FULL_NAME").
			Where("EXISTS (?)", subQuery).
			Limit(1).
			Scan(&subWhResult).Error
	}()

	// Goroutine 3: Lookup Penanggung concurrently if needed
	if needPenanggung {
		go func() {
			defer wg.Done()
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Penanggung).Limit(1).Find(&temp).Select("namaindonesia, namamandarin, alias").Error; err != nil {
				errPenanggung = fmt.Errorf("Penanggung %s not found: %w", *payload.Penanggung, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PenanggungManual = &pn
			}
		}()
	}

	// Goroutine 4: Lookup Pengajak concurrently if needed
	if needPengajak {
		go func() {
			defer wg.Done()
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Pengajak).Limit(1).Find(&temp).Select("namaindonesia, namamandarin, alias").Error; err != nil {
				errPengajak = fmt.Errorf("Pengajak %s not found: %w", *payload.Pengajak, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PengajakManual = &pn
			}
		}()
	}

	wg.Wait()

	if errId != nil {
		return domain.Umat{}, fmt.Errorf("failed to generate id: %w", errId)
	}
	if errSubWh != nil {
		log.Printf("[UmatService.Create] warning querying sub warehouse info for user %d: %v", userID, errSubWh)
	}
	if errPenanggung != nil {
		return domain.Umat{}, errPenanggung
	}
	if errPengajak != nil {
		return domain.Umat{}, errPengajak
	}

	// payload.FotangAktif = subWhResult.SubWhId
	payload.ID = genResult.GeneratedId
	kode := fmt.Sprintf("%s-%d", subWhResult.FullName, genResult.GeneratedId)
	payload.Kode = &kode

	// 3. Populate matching schema structural constraints
	payload.Status = true                  // Active status mapping
	payload.ModAct = "I"                   // 'I' standard legacy flag for Insert
	payload.ModBy = getUserID(c)           // Default system user ID matching INT type
	payload.ModDate = domain.NowDateTime() // Local server time object

	// 4. Persist the new entity to the database pool
	if err := s.db.Create(&payload).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to create record: %w", err)
	}

	if fileHeader != nil {
		if err := processAndSaveUmatFoto(s.db, payload.ID, payload.NamaIndonesia, fileHeader, c, false); err != nil {
			log.Printf("[UmatService.Create] Save UmatFoto error: %v", err)
		}
	}

	return payload, nil
}

func (s *UmatService) Get(id string) (domain.Umat, error) {
	var item domain.Umat
	if err := s.db.Preload("JenisKelaminInfo", "CategoryId = ? AND Status = ?", "B_JENISKELAMIN", true).First(&item, "id = ?", id).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("umat %s not found", id)
	}
	if !item.TanggalLahir.IsZero() && item.TanggalLahir.Year() > 1900 {
		u := int32(time.Now().Year() - item.TanggalLahir.Year())
		item.Usia = &u
	}
	// Dont activate this, get from JenisKelaminInfo
	/* if item.JenisKelaminInfo != nil && item.JenisKelaminInfo.LookupDescription != nil && *item.JenisKelaminInfo.LookupDescription != "" {
		item.JenisKelamin = *item.JenisKelaminInfo.LookupDescription
	} */
	return item, nil
}

func (s *UmatService) Update(id string, payload domain.Umat, fileHeader *multipart.FileHeader, c *gin.Context) (domain.Umat, error) {
	if fileHeader != nil {
		if err := validateImageHeader(fileHeader); err != nil {
			return domain.Umat{}, err
		}
	}

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

	var (
		item          domain.Umat
		errFetchItem  error
		errPenanggung error
		errPengajak   error
		wg            sync.WaitGroup
	)

	// Goroutine 1: Fetch existing item using .Take() to avoid default sorting bugs
	if err := s.db.Where("id = ?", userIDInt32).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errFetchItem = fmt.Errorf("umat %s not found", id)
		} else {
			errFetchItem = err
		}
	}

	// 3. Map values onto the actual field variables present in your legacy schema
	// item.Kode = payload.Kode
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

	needPenanggung := (payload.Penanggung != nil && string(*payload.Penanggung) != "") && (payload.PenanggungManual == nil || *payload.PenanggungManual == "")
	needPengajak := (payload.Pengajak != nil && string(*payload.Pengajak) != "") && (payload.PengajakManual == nil || *payload.PengajakManual == "")

	tasksCount := 1
	if needPenanggung {
		tasksCount++
	}
	if needPengajak {
		tasksCount++
	}

	wg.Add(tasksCount)

	// Goroutine 2: Lookup Penanggung concurrently if needed
	if needPenanggung {
		go func() {
			defer wg.Done()
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Penanggung).Limit(1).Select("namaindonesia, namamandarin, alias").Find(&temp).Error; err != nil {
				errPenanggung = fmt.Errorf("Penanggung %s not found: %w", *payload.Penanggung, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PenanggungManual = &pn
			}
		}()
	}

	// Goroutine 3: Lookup Pengajak concurrently if needed
	if needPengajak {
		go func() {
			defer wg.Done()
			var temp []domain.Umat
			if err := s.db.Where("id = ?", *payload.Pengajak).Limit(1).Select("namaindonesia, namamandarin, alias").Find(&temp).Error; err != nil {
				errPengajak = fmt.Errorf("Pengajak %s not found: %w", *payload.Pengajak, err)
				return
			}
			if len(temp) > 0 {
				pn := temp[0].NamaIndonesia
				if pn == "" && temp[0].NamaMandarin != nil {
					pn = *temp[0].NamaMandarin
				}
				if pn == "" && temp[0].Alias != nil {
					pn = *temp[0].Alias
				}
				payload.PengajakManual = &pn
			}
		}()
	}

	wg.Wait()

	if errFetchItem != nil {
		return domain.Umat{}, errFetchItem
	}
	if errPenanggung != nil {
		return domain.Umat{}, errPenanggung
	}
	if errPengajak != nil {
		return domain.Umat{}, errPengajak
	}

	// Legacy metadata mappings
	item.ModAct = "U"                   // 'U' standard legacy flag for Update
	item.ModBy = getUserID(c)           // System user ID (int32)
	item.ModDate = domain.NowDateTime() // Actual DateTime object expected by DATETIME column

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return domain.Umat{}, fmt.Errorf("failed to update record: %w", err)
	}

	if fileHeader != nil {
		if err := processAndSaveUmatFoto(s.db, item.ID, item.NamaIndonesia, fileHeader, c, true); err != nil {
			log.Printf("[UmatService.Update] Save UmatFoto error: %v", err)
		}
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
	item.ModAct = "D"                   // 'D' standard legacy flag for Delete
	item.ModBy = getUserID(c)           // Default system user ID matching INT type
	item.ModDate = domain.NowDateTime() // Local server time object

	// 4. Save updates back to SQL Server
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}

func getFilterOrDefault(filters map[string]string, keys []string, defaultVal string) string {
	for _, key := range keys {
		if val, exists := filters[key]; exists && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return defaultVal
}

type nullFieldScanner struct {
	target interface{}
}

func (s *nullFieldScanner) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	targetVal := reflect.ValueOf(s.target)
	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	elem := targetVal.Elem()

	for elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			if !elem.CanSet() {
				return nil
			}
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		elem = elem.Elem()
	}

	if elem.CanAddr() {
		if scanner, ok := elem.Addr().Interface().(sql.Scanner); ok {
			return scanner.Scan(value)
		}
	}
	if scanner, ok := elem.Interface().(sql.Scanner); ok {
		return scanner.Scan(value)
	}

	switch v := value.(type) {
	case string:
		vTrim := strings.TrimSpace(v)
		if elem.Kind() == reflect.String {
			elem.SetString(v)
			return nil
		}
		if elem.Kind() == reflect.Bool {
			vLower := strings.ToLower(vTrim)
			elem.SetBool(vLower == "1" || vLower == "true" || vLower == "y" || vLower == "ya")
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			if parsed, err := strconv.ParseInt(vTrim, 10, 64); err == nil {
				elem.SetInt(parsed)
			}
			return nil
		}
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			if parsed, err := strconv.ParseFloat(vTrim, 64); err == nil {
				elem.SetFloat(parsed)
			}
			return nil
		}
	case []byte:
		vStr := string(v)
		vTrim := strings.TrimSpace(vStr)
		if elem.Kind() == reflect.String {
			elem.SetString(vStr)
			return nil
		}
		if elem.Kind() == reflect.Bool {
			vLower := strings.ToLower(vTrim)
			elem.SetBool(vLower == "1" || vLower == "true" || vLower == "y" || vLower == "ya")
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			if parsed, err := strconv.ParseInt(vTrim, 10, 64); err == nil {
				elem.SetInt(parsed)
			}
			return nil
		}
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			if parsed, err := strconv.ParseFloat(vTrim, 64); err == nil {
				elem.SetFloat(parsed)
			}
			return nil
		}
	case int64:
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			elem.SetInt(v)
			return nil
		}
		if elem.Kind() == reflect.Bool {
			elem.SetBool(v != 0)
			return nil
		}
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			elem.SetFloat(float64(v))
			return nil
		}
		if elem.Kind() == reflect.String {
			elem.SetString(strconv.FormatInt(v, 10))
			return nil
		}
	case int32:
		return s.Scan(int64(v))
	case int:
		return s.Scan(int64(v))
	case bool:
		if elem.Kind() == reflect.Bool {
			elem.SetBool(v)
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			if v {
				elem.SetInt(1)
			} else {
				elem.SetInt(0)
			}
			return nil
		}
		if elem.Kind() == reflect.String {
			if v {
				elem.SetString("true")
			} else {
				elem.SetString("false")
			}
			return nil
		}
	case float64:
		if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
			elem.SetFloat(v)
			return nil
		}
		if elem.Kind() == reflect.Int || elem.Kind() == reflect.Int32 || elem.Kind() == reflect.Int64 {
			elem.SetInt(int64(v))
			return nil
		}
		if elem.Kind() == reflect.String {
			elem.SetString(fmt.Sprintf("%v", v))
			return nil
		}
	case time.Time:
		if elem.Kind() == reflect.String {
			elem.SetString(v.Format("2006-01-02"))
			return nil
		}
	}

	return nil
}

func (s *UmatService) Report(page int, filters map[string]string, limit int) ([]domain.UmatReport, int64, error) {
	if page <= 0 {
		page = 1
	}

	fotangAktif := getFilterOrDefault(filters, []string{"fotang_aktif", "FotangAktif", "fotangAktif"}, "0")
	fotangChiuTao := getFilterOrDefault(filters, []string{"fotang_chiutao", "fotang_chiu_tao", "FotangChiuTao", "fotangChiuTao"}, "0")
	namaMandarin := getFilterOrDefault(filters, []string{"nama_mandarin", "NamaMandarin", "namaMandarin"}, "")
	startDate := getFilterOrDefault(filters, []string{"start_date", "startDate", "StartDate"}, "1970-01-01")
	endDate := getFilterOrDefault(filters, []string{"end_date", "endDate", "EndDate"}, "1970-01-01")
	namaIndo := getFilterOrDefault(filters, []string{"nama_indo", "nama_indonesia", "NamaIndo", "namaIndo"}, "")
	pengajak := getFilterOrDefault(filters, []string{"pengajak", "Pengajak"}, "")
	alias := getFilterOrDefault(filters, []string{"alias", "Alias"}, "")
	usiaDari := toInt(getFilterOrDefault(filters, []string{"usia_dari", "UsiaDari", "usiaDari"}, "0"))
	usiaSampai := toInt(getFilterOrDefault(filters, []string{"usia_sampai", "UsiaSampai", "usiaSampai"}, "0"))
	isLulusSd := getFilterOrDefault(filters, []string{"is_lulus_sd", "IsLulusSd", "isLulusSd"}, "0")
	isVege := getFilterOrDefault(filters, []string{"is_vege", "IsVege", "isVege"}, "0")
	statusUmat := getFilterOrDefault(filters, []string{"status_umat", "StatusUmat", "statusUmat"}, "0")

	var findErr error

	rows, err := s.db.Raw("EXEC SP_BUS_RPT_UMAT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?",
		fotangAktif,
		fotangChiuTao,
		namaMandarin,
		startDate,
		endDate,
		namaIndo,
		pengajak,
		alias,
		usiaDari,
		usiaSampai,
		isLulusSd,
		isVege,
		statusUmat,
		limit, // @PageSize
		page,  // @CurrentPage
	).Rows()

	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	// log.Printf("[SQL] %s | Query:\n%s\n", timestamp, s.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
	// 	return tx.Raw("EXEC SP_BUS_RPT_UMAT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?",
	// 		fotangAktif,
	// 		fotangChiuTao,
	// 		namaMandarin,
	// 		startDate,
	// 		endDate,
	// 		namaIndo,
	// 		pengajak,
	// 		alias,
	// 		usiaDari,
	// 		usiaSampai,
	// 		isLulusSd,
	// 		isVege,
	// 		statusUmat,
	// 		limit, // @PageSize
	// 		page,  // @CurrentPage
	// 	)
	// }))

	if err != nil {
		return nil, 0, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			findErr = errors.New("umat report tidak ditemukan")
		} else {
			findErr = fmt.Errorf("database error: %w", err)
		}
		return nil, 0, findErr
	}

	var items []domain.UmatReport
	var total int64

	for rows.Next() {
		var item domain.UmatReport
		v := reflect.ValueOf(&item).Elem()
		t := v.Type()

		valuePtrs := make([]interface{}, len(cols))
		var totalRowScan int64

		for i, colName := range cols {
			cleanCol := strings.ToLower(strings.TrimSpace(colName))

			switch cleanCol {
			case "totalrow":
				valuePtrs[i] = &nullFieldScanner{target: &totalRowScan}
			default:
				matched := false
				for j := 0; j < t.NumField(); j++ {
					field := t.Field(j)
					gormTag := field.Tag.Get("gorm")

					if strings.Contains(strings.ToLower(gormTag), "column:"+cleanCol) ||
						strings.ToLower(field.Name) == cleanCol {
						fieldVal := v.Field(j)
						valuePtrs[i] = &nullFieldScanner{target: fieldVal.Addr().Interface()}
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
			log.Printf("[ERROR] Umat Report rows.Scan error: %v", err)
			continue
		}

		if totalRowScan != 0 {
			total = totalRowScan
		}

		items = append(items, item)
	}

	return items, total, nil
}

func (s *UmatService) ReportExcel(filters map[string]string, c *gin.Context) error {
	baseURL := s.reportServerURL
	if baseURL == "" {
		baseURL = getReportServerURL()
	}
	baseURL = strings.TrimRight(baseURL, "/")

	username := s.reportServerUsername
	if username == "" {
		username = getReportServerUsername()
	}
	password := s.reportServerPassword
	if password == "" {
		password = getReportServerPassword()
	}

	if username != "" {
		parsedURL, err := url.Parse(baseURL)
		if err == nil {
			parsedURL.User = url.UserPassword(username, password)
			baseURL = parsedURL.String()
		} else {
			schemeParts := strings.SplitN(baseURL, "://", 2)
			if len(schemeParts) == 2 {
				baseURL = fmt.Sprintf("%s://%s:%s@%s", schemeParts[0], url.QueryEscape(username), url.QueryEscape(password), schemeParts[1])
			}
		}
	}

	fotangAktif := getFilterOrDefault(filters, []string{"fotang_aktif", "FotangAktif", "fotangAktif"}, "0")
	fotangChiuTao := getFilterOrDefault(filters, []string{"fotang_chiutao", "fotang_chiu_tao", "FotangChiuTao", "fotangChiuTao"}, "0")
	namaMandarin := getFilterOrDefault(filters, []string{"nama_mandarin", "NamaMandarin", "namaMandarin"}, "")
	startDate := getFilterOrDefault(filters, []string{"start_date", "startDate", "StartDate"}, "1970-01-01")
	endDate := getFilterOrDefault(filters, []string{"end_date", "endDate", "EndDate"}, "1970-01-01")
	namaIndo := getFilterOrDefault(filters, []string{"nama_indo", "nama_indonesia", "NamaIndo", "namaIndo"}, "")
	pengajak := getFilterOrDefault(filters, []string{"pengajak", "Pengajak"}, "")
	alias := getFilterOrDefault(filters, []string{"alias", "Alias"}, "")
	usiaDari := getFilterOrDefault(filters, []string{"usia_dari", "UsiaDari", "usiaDari"}, "0")
	usiaSampai := getFilterOrDefault(filters, []string{"usia_sampai", "UsiaSampai", "usiaSampai"}, "0")
	isLulusSd := getFilterOrDefault(filters, []string{"is_lulus_sd", "IsLulusSd", "isLulusSd"}, "0")
	isVege := getFilterOrDefault(filters, []string{"is_vege", "IsVege", "isVege"}, "0")
	statusUmat := getFilterOrDefault(filters, []string{"status_umat", "StatusUmat", "statusUmat"}, "0")

	reportURL := fmt.Sprintf("%s/ReportServer?%%2fGuangJiReport%%2frpt_bus_umat_list&FotangAktif=%s&FotangChiuTao=%s&NamaMandarin=%s&StartDate=%s&EndDate=%s&NamaIndo=%s&Pengajak=%s&Alias=%s&UsiaDari=%s&UsiaSampai=%s&IsLulusSd=%s&IsVege=%s&StatusUmat=%s&rs:Command=Render&rs:Format=EXCELOPENXML",
		baseURL,
		url.QueryEscape(fotangAktif),
		url.QueryEscape(fotangChiuTao),
		url.QueryEscape(namaMandarin),
		url.QueryEscape(startDate),
		url.QueryEscape(endDate),
		url.QueryEscape(namaIndo),
		url.QueryEscape(pengajak),
		url.QueryEscape(alias),
		url.QueryEscape(usiaDari),
		url.QueryEscape(usiaSampai),
		url.QueryEscape(isLulusSd),
		url.QueryEscape(isVege),
		url.QueryEscape(statusUmat),
	)

	var reqCtx = c.Request.Context()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, reportURL, nil)
	if err != nil {
		return fmt.Errorf("gagal membuat request report: %w", err)
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := defaultReportClient

	startTime := time.Now()
	resp, err := client.Do(req)
	fetchDuration := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("gagal mengambil report dari SSRS (%v): %w", fetchDuration, err)
	}

	if resp.StatusCode == http.StatusUnauthorized && username != "" {
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.HasPrefix(authHeader, "Digest ") {
			resp.Body.Close()

			req2, err := http.NewRequestWithContext(reqCtx, http.MethodGet, reportURL, nil)
			if err != nil {
				return fmt.Errorf("gagal membuat digest request: %w", err)
			}

			digestVal := formatDigestAuth(authHeader, username, password, http.MethodGet, req2.URL.RequestURI())
			req2.Header.Set("Authorization", digestVal)

			resp, err = client.Do(req2)
			if err != nil {
				return fmt.Errorf("gagal mengambil report dengan Digest auth: %w", err)
			}
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("report server mengembalikan HTTP %d (fetch %v)", resp.StatusCode, fetchDuration)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition == "" {
		hash := sha256.Sum256([]byte(fotangAktif + fotangChiuTao + namaMandarin + startDate + endDate + namaIndo + statusUmat))
		filename := fmt.Sprintf("rpt_bus_umat_list_%x.xlsx", hash[:4])
		contentDisposition = fmt.Sprintf("attachment; filename=\"%s\"", filename)
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", contentDisposition)
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	c.Status(http.StatusOK)

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	bufPtr := copyBufferPool.Get().(*[]byte)
	defer copyBufferPool.Put(bufPtr)

	// streamStart := time.Now()
	_, err = io.CopyBuffer(c.Writer, resp.Body, *bufPtr)
	// streamDuration := time.Since(streamStart)
	// totalDuration := time.Since(startTime)

	// log.Printf("[REPORT PERF] SSRS Fetch: %v, Client Stream (%d bytes): %v, Total: %v",
	// 	fetchDuration, nBytes, streamDuration, totalDuration)

	if err != nil {
		return fmt.Errorf("gagal stream report ke client: %w", err)
	}

	return nil
}

var ocrBufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 512*1024))
	},
}

type OCRSpaceResult struct {
	ParsedResults []struct {
		TextOverlay struct {
			Lines []struct {
				Words []struct {
					WordText string  `json:"WordText"`
					Left     float64 `json:"Left"`
					Top      float64 `json:"Top"`
					Width    float64 `json:"Width"`
					Height   float64 `json:"Height"`
				} `json:"Words"`
				MaxHeight float64 `json:"MaxHeight"`
				MinTop    float64 `json:"MinTop"`
			} `json:"Lines"`
			HasOverlay bool   `json:"HasOverlay"`
			Message    string `json:"Message"`
		} `json:"TextOverlay"`
		TextOrientation   string `json:"TextOrientation"`
		FileParseExitCode int    `json:"FileParseExitCode"`
		ParsedText        string `json:"ParsedText"`
		ErrorMessage      string `json:"ErrorMessage"`
		ErrorDetails      string `json:"ErrorDetails"`
	} `json:"ParsedResults"`
	OCRExitCode                  int      `json:"OCRExitCode"`
	IsErroredOnProcessing        bool     `json:"IsErroredOnProcessing"`
	ProcessingTimeInMilliseconds string   `json:"ProcessingTimeInMilliseconds"`
	SearchablePDFURL             string   `json:"SearchablePDFURL"`
	ErrorMessage                 []string `json:"ErrorMessage"`
	ErrorDetails                 []string `json:"ErrorDetails"`
}

func (s *UmatService) Ocr(file multipart.File, header *multipart.FileHeader, c *gin.Context) (interface{}, error) {
	if file == nil || header == nil {
		return nil, errors.New("file upload tidak valid")
	}

	srcImg, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca format gambar: %w", err)
	}

	bounds := srcImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var finalImg image.Image = srcImg
	maxDim := 1600
	if width > maxDim || height > maxDim {
		scale := math.Min(float64(maxDim)/float64(width), float64(maxDim)/float64(height))
		targetW := int(float64(width) * scale)
		targetH := int(float64(height) * scale)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}

		dstRGBA := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		draw.BiLinear.Scale(dstRGBA, dstRGBA.Bounds(), srcImg, bounds, draw.Over, nil)
		finalImg = dstRGBA
	}

	buf := ocrBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer ocrBufferPool.Put(buf)

	quality := 80
	if err := jpeg.Encode(buf, finalImg, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("gagal kompresi gambar: %w", err)
	}

	for buf.Len() > 1000000 && quality > 30 {
		quality -= 20
		buf.Reset()
		if err := jpeg.Encode(buf, finalImg, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("gagal re-encode kompresi gambar: %w", err)
		}
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	apiKey := os.Getenv("OCR_SPACE_API_KEY")
	if apiKey == "" {
		apiKey = "helloworld"
	}

	formData := url.Values{}
	formData.Set("apikey", apiKey)
	formData.Set("base64Image", "data:image/jpeg;base64,"+base64Str)
	formData.Set("language", "cht")
	formData.Set("OCREngine", "3")

	client := &http.Client{
		Timeout: 45 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://api.ocr.space/parse/image", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request OCR.space: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim request ke OCR.space API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OCR.space API error HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result OCRSpaceResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gagal membaca respon JSON dari OCR.space: %w", err)
	}

	var parsedUmat map[string]interface{}
	if len(result.ParsedResults) > 0 && result.ParsedResults[0].ParsedText != "" {
		parsedUmat = ParseUmatOcrText(result.ParsedResults[0].ParsedText)
	} else {
		parsedUmat = make(map[string]interface{})
	}

	// Fetch subWhId from user matrix to populate fotang_aktif & fotang_chiutao
	if c != nil {
		userID := getUserID(c)
		if userID > 0 {
			type SubWhInfo struct {
				SubWhId  string `gorm:"column:SUBWHID"`
				FullName string `gorm:"column:FULL_NAME"`
			}
			var subWhResult SubWhInfo
			subQuery := s.db.Table("T_WH_USER_MATRIX_MST AS A").
				Select("1").
				Where("A.LOGINID = ?", userID).
				Where("A.SUBWHID = T_WH_SUBWH_MST.SUBWHID")

			_ = s.db.Table("T_WH_SUBWH_MST").
				Select("SUBWHID, FULL_NAME").
				Where("EXISTS (?)", subQuery).
				Limit(1).
				Scan(&subWhResult).Error

			if subWhResult.SubWhId != "" {
				parsedUmat["fotang_aktif"] = subWhResult.SubWhId
				parsedUmat["fotang_chiutao"] = subWhResult.SubWhId
			}
		}
	}

	return OCRResponsePayload{
		ParsedUmat: parsedUmat,
		OCRRaw:     result,
	}, nil
}

type OCRResponsePayload struct {
	ParsedUmat map[string]interface{} `json:"parsed_umat"`
	OCRRaw     interface{}            `json:"ocr_raw"`
}

func cleanSymbols(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			builder.WriteRune(r)
		}
	}
	return strings.TrimSpace(builder.String())
}

func deduplicateName(rawName string) string {
	rawName = strings.TrimSpace(rawName)
	if rawName == "" {
		return ""
	}
	parts := strings.Fields(rawName)
	if len(parts) > 1 {
		firstPartRunes := []rune(parts[0])
		lastPartRunes := []rune(parts[len(parts)-1])

		if len(firstPartRunes) > 0 && len(lastPartRunes) > 0 {
			lastCharOfFirst := string(firstPartRunes[len(firstPartRunes)-1])
			if parts[len(parts)-1] == lastCharOfFirst {
				return parts[0]
			}
		}
	}
	return rawName
}

func convertToPinyin(chineseName string) string {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	pyList := pinyin.Pinyin(chineseName, a)
	var words []string
	for _, p := range pyList {
		if len(p) > 0 && p[0] != "" {
			word := strings.Title(strings.ToLower(p[0]))
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}

func ParseUmatOcrText(rawText string) map[string]interface{} {
	parsed := make(map[string]interface{})

	lines := strings.Split(rawText, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		// 1. NAMA / 姓名
		if (strings.Contains(line, "NAMA") || strings.Contains(line, "姓名")) && (strings.Contains(line, ":") || strings.Contains(line, "：")) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				cleaned := deduplicateName(val)
				parsed["nama_indonesia"] = cleaned
				parsed["nama_mandarin"] = cleaned
				if cleaned != "" {
					py := convertToPinyin(cleaned)
					if py != "" {
						parsed["alias"] = py
					}
				}
			}
		}

		// 2. UMUR / 年齡
		if strings.Contains(line, "UMUR") || strings.Contains(line, "年齡") {
			reDigits := regexp.MustCompile(`\d+`)
			if numStr := reDigits.FindString(line); numStr != "" {
				if ageNum, err := strconv.Atoi(numStr); err == nil {
					parsed["usia"] = ageNum
					currentYear := time.Now().Year()
					birthYear := currentYear - ageNum
					parsed["tanggal_lahir"] = fmt.Sprintf("%d-01-01", birthYear)
				}
			}
		}

		// 3. PENDIDIKAN / 教育
		if strings.Contains(line, "PENDIDIKAN") || strings.Contains(line, "教育") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			var val string
			if len(parts) == 2 {
				val = strings.TrimSpace(parts[1])
			} else {
				val = strings.TrimSpace(strings.ReplaceAll(line, "PENDIDIKAN", ""))
				val = strings.TrimSpace(strings.ReplaceAll(val, "教育", ""))
			}
			val = strings.ToUpper(val)
			if val == "SI" || val == "S I" {
				val = "S1"
			}
			parsed["pendidikan"] = val
		}

		// 4. JENIS KELAMIN / 性別
		if strings.Contains(line, "JENIS KELAMIN") || strings.Contains(line, "性別") {
			lineUpper := strings.ToUpper(line)
			if strings.Contains(lineUpper, "坤") || strings.Contains(lineUpper, "WANITA") || strings.Contains(lineUpper, "PEREMPUAN") || strings.Contains(lineUpper, " W ") || strings.HasSuffix(lineUpper, " W") || strings.HasSuffix(lineUpper, "WANITA") {
				parsed["jenis_kelamin"] = "WANITA"
			} else if strings.Contains(lineUpper, "乾") || lineUpper == "PRIA" || lineUpper == "LAKI-LAKI" || lineUpper == "L" {
				parsed["jenis_kelamin"] = "PRIA"
			} else if strings.Contains(lineUpper, "童") || lineUpper == "ANAK PRIA" || lineUpper == "ANAK LAKI-LAKI" || lineUpper == "ANAK L" || lineUpper == "ANAK LAKI" {
				parsed["jenis_kelamin"] = "ANAK PRIA"
			} else if strings.Contains(lineUpper, "女") || lineUpper == "ANAK WANITA" || lineUpper == "ANAK W" {
				parsed["jenis_kelamin"] = "ANAK WANITA"
			}
		}

		// 5. ALAMAT / 地址
		if strings.Contains(line, "ALAMAT") || strings.Contains(line, "地址") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				addr := strings.TrimSpace(parts[1])
				reIl := regexp.MustCompile(`(?i)\bIl\.\s*`)
				addr = reIl.ReplaceAllString(addr, "JL. ")
				reJl := regexp.MustCompile(`(?i)\bJl\.\s*`)
				addr = reJl.ReplaceAllString(addr, "JL. ")
				parsed["alamat"] = addr
			}
		}

		// 6. TELEPON / 電話 / MOBILE
		if strings.Contains(line, "電話") || strings.Contains(line, "TELEPON") || strings.Contains(line, "TELP") || strings.Contains(line, "MOBILE") || strings.Contains(line, "HP") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				mob := strings.TrimSpace(parts[1])
				mob = strings.ReplaceAll(mob, "-", "")
				mob = strings.ReplaceAll(mob, " ", "")
				parsed["mobile"] = mob
			}
		}

		// 7. PERANTARA / 引師 (pengajak_manual)
		if strings.Contains(line, "PERANTARA") || strings.Contains(line, "引師") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				parsed["pengajak_manual"] = cleanSymbols(parts[1])
			}
		}

		// 8. PENANGGUNG / 保師 (penanggung_manual)
		if strings.Contains(line, "PENANGGUNG") || strings.Contains(line, "保師") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				parsed["penanggung_manual"] = cleanSymbols(parts[1])
			}
		}

		// 9. TCS / 點傳師
		if strings.Contains(line, "點傳師") || strings.Contains(line, "TCS") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) == 2 {
				parsed["tcs"] = cleanSymbols(parts[1])
			}
		}

		// 10. 功德費 / uang_pahala & waktu_chiutao_mandarin
		if strings.Contains(line, "功德費") || strings.Contains(line, "UANG PAHALA") {
			reDigits := regexp.MustCompile(`\d[\d\.\,]*`)
			if numMatch := reDigits.FindString(line); numMatch != "" {
				numStr := strings.ReplaceAll(numMatch, ".", "")
				numStr = strings.ReplaceAll(numStr, ",", "")
				if amount, err := strconv.ParseFloat(numStr, 64); err == nil {
					parsed["uang_pahala"] = amount
				}
			}
			lunarHours := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
			for _, lh := range lunarHours {
				if strings.Contains(line, lh) {
					parsed["waktu_chiutao_mandarin"] = lh
					break
				}
			}
		}
	}

	return parsed
}
