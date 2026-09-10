package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KelasMasterService struct {
	db       *gorm.DB
	resource string
}

func NewKelasMasterService(db *gorm.DB) *KelasMasterService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &KelasMasterService{db: db, resource: "kelas_master"}
}

// getNextLookupValue calculates the next sequential LookupValue and LookupId for B_KELASKHUSUS.
// Example: existing max "022" -> nextVal "023", nextId "B_KELASKHUSUS023"
func (s *KelasMasterService) getNextLookupValue() (string, string, error) {
	var values []string
	if err := s.db.Table("T_APP_LOOKUP").
		Where("CategoryId = ?", "B_KELASKHUSUS").
		Pluck("LookupValue", &values).Error; err != nil {
		return "", "", fmt.Errorf("failed to query lookup values: %w", err)
	}

	maxNum := 0
	for _, v := range values {
		v = strings.TrimSpace(v)
		if num, err := strconv.Atoi(v); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	nextNum := maxNum + 1
	nextVal := fmt.Sprintf("%03d", nextNum)
	nextId := fmt.Sprintf("B_KELASKHUSUS%s", nextVal)

	// In case nextId already exists in DB, increment until an unused ID is found
	var count int64
	_ = s.db.Table("T_APP_LOOKUP").Where("LookupId = ?", nextId).Count(&count).Error
	for count > 0 {
		nextNum++
		nextVal = fmt.Sprintf("%03d", nextNum)
		nextId = fmt.Sprintf("B_KELASKHUSUS%s", nextVal)
		_ = s.db.Table("T_APP_LOOKUP").Where("LookupId = ?", nextId).Count(&count).Error
	}

	return nextVal, nextId, nil
}

func (s *KelasMasterService) validateDescription(description string, excludeId string) error {
	name := strings.TrimSpace(description)
	if name == "" {
		return NewValidationError(map[string][]string{
			"lookup_description": {"Nama kelas (LookupDescription) wajib diisi"},
		})
	}
	if len(name) > 150 {
		return NewValidationError(map[string][]string{
			"lookup_description": {"Nama kelas (LookupDescription) maksimal 150 karakter"},
		})
	}

	var count int64
	query := s.db.Table("T_APP_LOOKUP").
		Where("CategoryId = ? AND LookupDescription = ? AND Status = ?", "B_KELASKHUSUS", name, true)
	if excludeId != "" {
		query = query.Where("LookupId <> ?", excludeId)
	}

	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("database count error: %w", err)
	}

	if count > 0 {
		return NewValidationError(map[string][]string{
			"lookup_description": {fmt.Sprintf("Nama kelas '%s' sudah ada", name)},
		})
	}
	return nil
}

func (s *KelasMasterService) validateKelasMaster(payload domain.AppLookup, excludeId string) error {
	details := make(map[string][]string)

	if err := ValidateStruct(payload); err != nil {
		var vErr *ValidationError
		if errors.As(err, &vErr) && vErr != nil {
			for k, v := range vErr.Details {
				details[k] = append(details[k], v...)
			}
		} else {
			details["general"] = append(details["general"], err.Error())
		}
	}

	desc := ""
	if payload.LookupDescription != nil {
		desc = *payload.LookupDescription
	}
	if err := s.validateDescription(desc, excludeId); err != nil {
		var vErr *ValidationError
		if errors.As(err, &vErr) && vErr != nil {
			for k, v := range vErr.Details {
				details[k] = append(details[k], v...)
			}
		} else {
			details["general"] = append(details["general"], err.Error())
		}
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}
	return nil
}

func (s *KelasMasterService) List(page int, filters map[string]string, limit int) ([]domain.AppLookup, int64, error) {
	var items []domain.AppLookup
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_APP_LOOKUP").Where("CategoryId = ?", "B_KELASKHUSUS")

	// Status filter: default active (Status = 1) unless specified otherwise
	if statusVal, exists := filters["status"]; exists && statusVal != "" {
		if statusVal == "all" {
			// no status filtering
		} else if b, err := strconv.ParseBool(statusVal); err == nil {
			query = query.Where("Status = ?", b)
		} else if statusVal == "1" || statusVal == "0" {
			query = query.Where("Status = ?", statusVal == "1")
		}
	} else {
		query = query.Where("Status = ?", true)
	}

	type FilterRule struct {
		Column string
		IsLike bool
	}

	allowedFilters := map[string]FilterRule{
		"lookup_id":          {Column: "LookupId", IsLike: true},
		"id":                 {Column: "LookupId", IsLike: true},
		"code":               {Column: "LookupValue", IsLike: true},
		"lookup_value":       {Column: "LookupValue", IsLike: true},
		"lookup_description": {Column: "LookupDescription", IsLike: true},
		"name":               {Column: "LookupDescription", IsLike: true},
		"nama":               {Column: "LookupDescription", IsLike: true},
	}

	for field, value := range filters {
		val := strings.TrimSpace(value)
		if val == "" || field == "status" {
			continue
		}
		if rule, exists := allowedFilters[field]; exists {
			if rule.IsLike {
				query = query.Where(fmt.Sprintf("[%s] LIKE ?", rule.Column), "%"+val+"%")
			} else {
				query = query.Where(fmt.Sprintf("[%s] = ?", rule.Column), val)
			}
		}
	}

	var (
		countErr error
		findErr  error
		wg       sync.WaitGroup
	)

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			countErr = fmt.Errorf("database count error: %w", err)
		}
	})

	wg.Go(func() {
		if err := query.Session(&gorm.Session{}).
			Limit(limit).
			Offset(offset).
			Order("LookupValue ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("kelas master tidak ditemukan")
			} else {
				findErr = fmt.Errorf("database error: %w", err)
			}
		}
	})

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if findErr != nil {
		return []domain.AppLookup{}, 0, findErr
	}

	return items, total, nil
}

func (s *KelasMasterService) Get(id string) (domain.AppLookup, error) {
	cleanID := strings.TrimSpace(id)
	var item domain.AppLookup
	if err := s.db.Where("CategoryId = ? AND (LookupId = ? OR LookupValue = ?)", "B_KELASKHUSUS", cleanID, cleanID).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AppLookup{}, fmt.Errorf("kelas master '%s' tidak ditemukan", id)
		}
		return domain.AppLookup{}, err
	}
	return item, nil
}

func (s *KelasMasterService) Create(payload domain.AppLookup, c *gin.Context) (domain.AppLookup, error) {
	category := "B_KELASKHUSUS"
	payload.CategoryId = &category

	// Auto-generate LookupValue and LookupId if not provided
	if payload.LookupValue == nil || strings.TrimSpace(*payload.LookupValue) == "" {
		nextVal, nextId, err := s.getNextLookupValue()
		if err != nil {
			return domain.AppLookup{}, err
		}
		payload.LookupValue = &nextVal
		if payload.LookupId == "" {
			payload.LookupId = nextId
		}
	} else {
		val := strings.TrimSpace(*payload.LookupValue)
		// If 1 or 2 digits numeric, pad with leading zeros up to 3 digits
		if num, err := strconv.Atoi(val); err == nil && len(val) < 3 {
			val = fmt.Sprintf("%03d", num)
		}
		payload.LookupValue = &val
		if payload.LookupId == "" {
			payload.LookupId = "B_KELASKHUSUS" + val
		}
	}

	if err := s.validateKelasMaster(payload, ""); err != nil {
		return domain.AppLookup{}, err
	}

	// Check if LookupId already exists
	var count int64
	if err := s.db.Table("T_APP_LOOKUP").Where("LookupId = ?", payload.LookupId).Count(&count).Error; err != nil {
		return domain.AppLookup{}, fmt.Errorf("database count error: %w", err)
	}
	if count > 0 {
		return domain.AppLookup{}, NewValidationError(map[string][]string{
			"lookup_id": {fmt.Sprintf("LookupId '%s' sudah ada", payload.LookupId)},
		})
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	if userIDStr == "0" {
		userIDStr = "admin"
	}

	desc := strings.TrimSpace(*payload.LookupDescription)
	payload.LookupDescription = &desc

	actStatus := true
	if payload.Status != nil {
		actStatus = *payload.Status
	}
	payload.Status = &actStatus

	modAct := "I"
	payload.ModAct = &modAct
	payload.ModBy = &userIDStr
	now := domain.NowDateTime()
	payload.ModDate = &now

	if err := s.db.Create(&payload).Error; err != nil {
		return domain.AppLookup{}, fmt.Errorf("failed to create record: %w", err)
	}
	return payload, nil
}

func (s *KelasMasterService) Update(id string, payload domain.AppLookup, c *gin.Context) (domain.AppLookup, error) {
	cleanID := strings.TrimSpace(id)
	var item domain.AppLookup
	if err := s.db.Where("CategoryId = ? AND (LookupId = ? OR LookupValue = ?)", "B_KELASKHUSUS", cleanID, cleanID).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AppLookup{}, fmt.Errorf("kelas master '%s' tidak ditemukan", id)
		}
		return domain.AppLookup{}, err
	}

	validatePayload := payload
	if validatePayload.LookupId == "" {
		validatePayload.LookupId = item.LookupId
	}
	if validatePayload.LookupDescription == nil || strings.TrimSpace(*validatePayload.LookupDescription) == "" {
		validatePayload.LookupDescription = item.LookupDescription
	}

	if err := s.validateKelasMaster(validatePayload, item.LookupId); err != nil {
		return domain.AppLookup{}, err
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	if userIDStr == "0" {
		userIDStr = "admin"
	}

	if payload.LookupDescription != nil && strings.TrimSpace(*payload.LookupDescription) != "" {
		desc := strings.TrimSpace(*payload.LookupDescription)
		item.LookupDescription = &desc
	}
	if payload.Status != nil {
		item.Status = payload.Status
	}

	modAct := "U"
	item.ModAct = &modAct
	item.ModBy = &userIDStr
	now := domain.NowDateTime()
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return domain.AppLookup{}, fmt.Errorf("failed to update record: %w", err)
	}
	return item, nil
}

func (s *KelasMasterService) Delete(id string, c *gin.Context) error {
	cleanID := strings.TrimSpace(id)
	var item domain.AppLookup
	if err := s.db.Where("CategoryId = ? AND (LookupId = ? OR LookupValue = ?)", "B_KELASKHUSUS", cleanID, cleanID).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("kelas master '%s' tidak ditemukan", id)
		}
		return err
	}

	userIDStr := strconv.Itoa(int(getUserID(c)))
	if userIDStr == "0" {
		userIDStr = "admin"
	}

	falseVal := false
	item.Status = &falseVal
	modAct := "D"
	item.ModAct = &modAct
	item.ModBy = &userIDStr
	now := domain.NowDateTime()
	item.ModDate = &now

	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}
