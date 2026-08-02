package service

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type menuCacheEntry struct {
	menus     []domain.MainMenuItem
	fetchedAt time.Time
}

type AuthService struct {
	cfg       config.Config
	db        *gorm.DB
	menuCache sync.Map
}

// Replicates .NET System.Text.ASCIIEncoding.ASCII.GetString(result)
// It forcefully drops or masks non-ASCII binary bytes exactly like legacy .NET does.
func convertToDotNetASCIIString(bytes []byte) string {
	runes := make([]rune, len(bytes))
	for i, b := range bytes {
		if b > 127 {
			runes[i] = '?' // .NET ASCII encoder replaces bytes > 127 with '?'
		} else {
			runes[i] = rune(b)
		}
	}
	return string(runes)
}

// Replicates Common.SecurityHelper.Base64Encode(Common.SecurityHelper.Encrypt(rawpassword))
func EncryptPassword(rawPassword string) string {
	// 1. Convert input string to ASCII bytes
	inputBytes := []byte(rawPassword)

	// 2. Compute MD5 Hash
	hasher := md5.New()
	hasher.Write(inputBytes)
	md5Result := hasher.Sum(nil)

	// 3. Convert MD5 bytes back to an ASCII string (This replicates the .NET quirk)
	asciiStringBroken := convertToDotNetASCIIString(md5Result)

	// 4. Encode the broken ASCII string to standard Base64
	finalBase64 := base64.StdEncoding.EncodeToString([]byte(asciiStringBroken))

	return finalBase64
}

func NewAuthService(cfg config.Config, db *gorm.DB) *AuthService {
	return &AuthService{cfg: cfg, db: db}
}

func (s *AuthService) GenerateToken(userID int32) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func executeWithRetry(fn func() error) error {
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		errStr := err.Error()
		if strings.Contains(errStr, "dial tcp") || strings.Contains(errStr, "unable to open tcp") || strings.Contains(errStr, "connectex") {
			time.Sleep(time.Duration(attempt*15) * time.Millisecond)
			continue
		}
		return err
	}
	return err
}

func (s *AuthService) Login(username, password string) (domain.Admin, string, []domain.MainMenuItem, error) {
	if username == "" || password == "" {
		return domain.Admin{}, "", nil, errors.New("username and password are required")
	}

	// Call SP_Login for validation with transient connection retry
	var result struct {
		Warn     string
		FlagWarn int
		GroupId  int
	}
	err := executeWithRetry(func() error {
		return s.db.Raw("EXEC SP_Login ?, ?", username, EncryptPassword(password)).Scan(&result).Error
	})
	if err != nil {
		return domain.Admin{}, "", nil, err
	}
	if result.FlagWarn == 0 {
		return domain.Admin{}, "", nil, errors.New(result.Warn)
	}

	var (
		user      domain.Admin
		userErr   error
		mainMenus = []domain.MainMenuItem{}
		token     string

		wgUser sync.WaitGroup
	)

	// Fetch user details from T_Login_Mst concurrently
	wgUser.Add(1)
	go func() {
		defer wgUser.Done()
		err := executeWithRetry(func() error {
			return s.db.Table("T_Login_Mst").
				Select("LoginId, DepartmentId, GroupId, Username, ImgUrl, LastLogin, IsWarehouse, GroupId").
				Where("Username = ?", username).
				First(&user).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				userErr = errors.New("username tidak ditemukan")
			} else {
				userErr = fmt.Errorf("database error: %w", err)
			}
		}
	}()

	wgUser.Wait()

	if userErr != nil {
		return domain.Admin{}, "", nil, userErr
	}

	// Async fire-and-forget update LastLogin timestamp
	go func(userID int32) {
		if err := s.db.Table("T_Login_Mst").
			Where("LoginId = ?", userID).
			Update("LastLogin", time.Now()).Error; err != nil {
			log.Printf("[WARNING] Failed to update LastLogin for LoginId %d: %v", userID, err)
		}
	}(user.ID)

	// Synchronously generate token
	token, err = s.GenerateToken(user.ID)
	if err != nil {
		return domain.Admin{}, "", nil, err
	}

	// Check menu cache in RAM (5-minute TTL) to eliminate repetitive SP_Login_Create_Xml calls
	cacheKey := fmt.Sprintf("%s:%d", username, result.GroupId)
	if val, ok := s.menuCache.Load(cacheKey); ok {
		entry := val.(menuCacheEntry)
		if time.Since(entry.fetchedAt) < 5*time.Minute {
			mainMenus = entry.menus
		}
	}

	if len(mainMenus) == 0 {
		xmlErr := s.db.Raw("EXEC SP_Login_Create_Xml ?", username).Scan(&mainMenus).Error
		if xmlErr != nil {
			return domain.Admin{}, "", nil, fmt.Errorf("failed to execute SP Login: %w", xmlErr)
		}

		// Coroutine loop: for each main menu item, fetch sub-menus via SP_Login_View_Mapping_Group
		if len(mainMenus) > 0 {
			var wgSubMenu sync.WaitGroup
			wgSubMenu.Add(len(mainMenus))

			for i := range mainMenus {
				go func(idx int) {
					defer wgSubMenu.Done()
					parentID := mainMenus[idx].MenuID

					rows, err := s.db.Raw("EXEC [dbo].[SP_Login_View_Mapping_Group] @ParentId = ?, @groupid = ?", parentID, result.GroupId).Rows()
					if err != nil {
						log.Printf("[WARNING] Failed to fetch sub menu for ParentId %d, GroupId %d: %v", parentID, result.GroupId, err)
						mainMenus[idx].SubMenu = []domain.SubMenuItem{}
						return
					}
					defer rows.Close()

					cols, err := rows.Columns()
					if err != nil {
						mainMenus[idx].SubMenu = []domain.SubMenuItem{}
						return
					}

					subMenus := []domain.SubMenuItem{}
					for rows.Next() {
						values := make([]interface{}, len(cols))
						valuePtrs := make([]interface{}, len(cols))
						for k := range values {
							valuePtrs[k] = &values[k]
						}

						if err := rows.Scan(valuePtrs...); err != nil {
							continue
						}

						rowMap := make(map[string]interface{})
						var firstParentID int
						gotFirstParent := false

						for k, colName := range cols {
							val := values[k]
							if b, ok := val.([]byte); ok {
								val = string(b)
							}

							cleanName := strings.ToLower(strings.TrimSpace(colName))

							// Capture the VERY FIRST ParentId column value (preventing 2nd ParenId from overwriting it)
							if strings.EqualFold(cleanName, "parentid") && !gotFirstParent {
								firstParentID = toInt(val)
								gotFirstParent = true
							}

							rowMap[cleanName] = val
						}

						if toBool(rowMap["flaguse"]) {
							seqVal := rowMap["squence"]

							subMenus = append(subMenus, domain.SubMenuItem{
								MenuID:   toInt(rowMap["menuid"]),
								ParentID: firstParentID,
								Level2:   toString(rowMap["level2"]),
								Level3:   toString(rowMap["level3"]),
								Sequence: toInt(seqVal),
							})
						}
					}
					mainMenus[idx].SubMenu = subMenus
				}(i)
			}

			wgSubMenu.Wait()

			s.menuCache.Store(cacheKey, menuCacheEntry{
				menus:     mainMenus,
				fetchedAt: time.Now(),
			})
		}
	}

	user.Password = ""
	// user.GroupId = nil
	return user, token, mainMenus, nil
}

func toInt(val interface{}) int {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	case int16:
		return int(v)
	case int8:
		return int(v)
	case uint:
		return int(v)
	case uint64:
		return int(v)
	case uint32:
		return int(v)
	case uint16:
		return int(v)
	case uint8:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case []byte:
		i, _ := strconv.Atoi(strings.TrimSpace(string(v)))
		return i
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(v))
		return i
	default:
		str := strings.TrimSpace(fmt.Sprintf("%v", v))
		i, _ := strconv.Atoi(str)
		return i
	}
}

func BoolPtr(b bool) *bool {
	return &b
}

func toBool(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return toInt(v) != 0
	case float64, float32:
		return toInt(v) != 0
	case []byte:
		str := strings.ToLower(strings.TrimSpace(string(v)))
		return str == "1" || str == "true" || str == "t"
	case string:
		str := strings.ToLower(strings.TrimSpace(v))
		return str == "1" || str == "true" || str == "t"
	default:
		return false
	}
}

func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (s *AuthService) ChangePassword(userID int32, username, oldPassword, newPassword string) error {
	if strings.TrimSpace(oldPassword) == "" || strings.TrimSpace(newPassword) == "" {
		return errors.New("old_password and new_password are required")
	}

	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	var user domain.Admin
	query := s.db.Table("T_Login_Mst")
	if userID > 0 {
		query = query.Where("LoginId = ?", userID)
	} else if strings.TrimSpace(username) != "" {
		query = query.Where("Username = ?", strings.TrimSpace(username))
	} else {
		return errors.New("user identification (userID or username) is required")
	}

	err := query.First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	if user.Password != EncryptPassword(oldPassword) {
		return errors.New("invalid old password")
	}

	encryptedNew := EncryptPassword(newPassword)
	if err := s.db.Table("T_Login_Mst").Where("LoginId = ?", user.ID).Update("Password", encryptedNew).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
