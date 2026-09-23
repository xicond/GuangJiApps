package service

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminService manages admin user accounts and related configuration.
type AdminService struct {
	db       *gorm.DB
	resource string
}

// NewAdminService initializes a new instance of AdminService with the provided database connection.
//
// Parameters:
//   - db: Database connection handle (*gorm.DB). If nil, the default connection is used.
//
// Returns:
//   - *AdminService: An initialized instance of AdminService.
func NewAdminService(db *gorm.DB) *AdminService {
	if db == nil {
		db = database.MustOpen("")
	}
	return &AdminService{db: db, resource: "admins"}
}

// List retrieves a paginated list of admin users matching the given filter criteria.
//
// Parameters:
//   - page: The target page number (1-based index).
//   - filters: Key-value map of filter conditions (e.g. "username", "group_name").
//   - limit: Maximum number of records to return per page.
//
// Returns:
//   - []domain.Admin: Slice of admin records retrieved from the database.
//   - int64: Total count of records matching the filters.
//   - error: Error if the database query fails.
func (s *AdminService) List(page int, filters map[string]string, limit int) ([]domain.Admin, int64, error) {
	var items []domain.Admin
	var total int64

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := s.db.Table("T_Login_Mst") //.Where("FlagUse = ?", true)

	type FilterRule struct {
		Column   string
		IsLike   bool
		SubQuery string
	}

	allowedFilters := map[string]FilterRule{
		"username":   {Column: "Username", IsLike: true},
		"group_name": {SubQuery: "EXISTS (SELECT 1 FROM T_Login_Group g WHERE g.GroupId = T_Login_Mst.GroupId AND g.GroupName LIKE ?)"},
		// "email":         {Column: "email", IsLike: true},
		// "phone_number":  {Column: "PhoneNumber", IsLike: true},
		// "department_id": {Column: "DepartmentId", IsLike: false},
	}

	for field, value := range filters {
		if value == "" {
			continue
		}
		if rule, exists := allowedFilters[field]; exists {
			if rule.SubQuery != "" {
				query = query.Where(rule.SubQuery, "%"+value+"%")
			} else if rule.IsLike {
				query = query.Where(fmt.Sprintf("[%s] LIKE ?", rule.Column), "%"+value+"%")
			} else {
				query = query.Where(fmt.Sprintf("[%s] = ?", rule.Column), value)
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
			Preload("AdminGroup").
			Preload("Department").
			Offset(offset).
			Order("LoginId ASC").
			Find(&items).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				findErr = errors.New("admin tidak ditemukan")
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
		return []domain.Admin{}, 0, findErr
	}

	for i := range items {
		items[i].Password = ""
	}

	return items, total, nil
}

// Create inserts a new admin user record into the database with password encryption and default validity periods.
//
// Parameters:
//   - payload: The admin record to create (domain.Admin).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - domain.Admin: The newly created admin record (with password cleared).
//   - error: Error if validation fails or the insert query fails.
func (s *AdminService) Create(payload domain.Admin, c *gin.Context) (domain.Admin, error) {
	if err := ValidateStruct(payload); err != nil {
		return domain.Admin{}, fmt.Errorf("Validation failed: %w", err)
	}

	// var maxID int32
	// s.db.Table("T_Login_Mst").Select("ISNULL(MAX(LoginId), 0)").Row().Scan(&maxID)
	// payload.ID = maxID + 1

	if payload.Password != "" {
		payload.Password = EncryptPassword(payload.Password)
	}

	payload.FlagUse = true
	if payload.DateStart.IsZero() {
		payload.DateStart = domain.DateOnly{Time: time.Now()}
	}
	if payload.DateEnd.IsZero() {
		payload.DateEnd = domain.DateOnly{Time: time.Now().AddDate(10, 0, 0)}
	}
	if payload.LastLogin.IsZero() {
		payload.LastLogin = domain.DateTime{Time: time.Now()}
	}

	if err := s.db.Omit("AdminGroup", "Department").Create(&payload).Error; err != nil {
		return domain.Admin{}, fmt.Errorf("failed to create record: %w", err)
	}
	payload.Password = ""
	return payload, nil
}

// Get fetches a single admin user by their ID.
//
// Parameters:
//   - id: The primary key (LoginId) of the admin user as a string.
//
// Returns:
//   - domain.Admin: The retrieved admin record (with password cleared).
//   - error: Error if ID format is invalid or record is not found.
func (s *AdminService) Get(id string) (domain.Admin, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Admin{}, fmt.Errorf("invalid ID format: %w", err)
	}
	var item domain.Admin
	if err := s.db.Where("LoginId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Admin{}, fmt.Errorf("admin %s not found", id)
		}
		return domain.Admin{}, err
	}
	item.Password = ""
	return item, nil
}

// Update updates an existing admin user's details and encrypts new passwords if provided.
//
// Parameters:
//   - id: The primary key (LoginId) of the admin user to update as a string.
//   - payload: Updated admin fields (domain.Admin).
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - domain.Admin: The updated admin record (with password cleared).
//   - error: Error if the admin is not found, validation fails, or database update fails.
func (s *AdminService) Update(id string, payload domain.Admin, c *gin.Context) (domain.Admin, error) {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return domain.Admin{}, fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Admin
	if err := s.db.Where("LoginId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Admin{}, fmt.Errorf("admin %s not found", id)
		}
		return domain.Admin{}, err
	}

	if payload.Username != "" {
		item.Username = payload.Username
	}
	if payload.Email != nil {
		item.Email = payload.Email
	}
	if payload.GroupId != 0 {
		item.GroupId = payload.GroupId
	}
	if payload.PhoneNumber != nil {
		item.PhoneNumber = payload.PhoneNumber
	}
	if payload.ImgUrl != "" {
		item.ImgUrl = payload.ImgUrl
	}
	item.FlagUse = payload.FlagUse
	if payload.LoginDesc != nil {
		item.LoginDesc = payload.LoginDesc
	}
	if payload.DepartmentId != 0 {
		item.DepartmentId = payload.DepartmentId
	}
	item.IsWarehouse = payload.IsWarehouse

	if !payload.DateStart.IsZero() {
		item.DateStart = payload.DateStart
	}
	if !payload.DateEnd.IsZero() {
		item.DateEnd = payload.DateEnd
	}
	if payload.Password != "" {
		item.Password = EncryptPassword(payload.Password)
	}
	if err := ValidateStruct(item); err != nil {
		return domain.Admin{}, fmt.Errorf("Validation failed: %w", err)
	}

	if err := s.db.Omit("AdminGroup", "Department").Save(&item).Error; err != nil {
		return domain.Admin{}, fmt.Errorf("failed to update record: %w", err)
	}
	item.Password = ""
	return item, nil
}

// Delete marks an admin user as deactivated (soft delete via FlagUse = false).
//
// Parameters:
//   - id: The primary key (LoginId) of the admin user to deactivate.
//   - c: Gin context carrying HTTP request metadata.
//
// Returns:
//   - error: Error if the admin is not found or database update fails.
func (s *AdminService) Delete(id string, c *gin.Context) error {
	parsedInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	var item domain.Admin
	if err := s.db.Where("LoginId = ?", parsedInt).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin %s not found", id)
		}
		return err
	}

	item.FlagUse = false
	if err := s.db.Save(&item).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

// ListDepartments retrieves all department master records sorted by DepartmentId.
//
// Returns:
//   - []domain.DepartmentMst: Slice of department master records.
//   - error: Error if database query fails.
func (s *AdminService) ListDepartments() ([]domain.DepartmentMst, error) {
	var items []domain.DepartmentMst
	if err := s.db.Order("DepartmentId ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return items, nil
}
