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

type AuthService struct {
	cfg config.Config
	db  *gorm.DB
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

func (s *AuthService) Login(username, password string) (domain.Admin, string, []domain.MainMenuItem, error) {
	if username == "" || password == "" {
		return domain.Admin{}, "", nil, errors.New("username and password are required")
	}

	// Call SP_Login for validation
	var result struct {
		Warn     string
		FlagWarn int
		GroupId  int
	}
	err := s.db.Raw("EXEC SP_Login ?, ?", username, EncryptPassword(password)).Scan(&result).Error
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
		xmlErr    error
		token     string

		wgPhase1 sync.WaitGroup
		wgPhase2 sync.WaitGroup
	)

	// Phase 1 goroutines: Fetch user from T_Login_Mst and execute SP_Login_Create_Xml concurrently
	wgPhase1.Add(1)
	wgPhase2.Add(1)

	go func() {
		defer wgPhase1.Done()
		err := s.db.Table("T_Login_Mst").
			Select("LoginId as [id], email, Username as name").
			Where("Username = ?", username).
			First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				userErr = errors.New("username tidak ditemukan")
			} else {
				userErr = fmt.Errorf("database error: %w", err)
			}
		}
	}()

	go func() {
		defer wgPhase2.Done()
		err := s.db.Raw("EXEC SP_Login_Create_Xml ?", username).Scan(&mainMenus).Error
		if err != nil {
			xmlErr = fmt.Errorf("failed to execute SP_Login_Create_Xml: %w", err)
		}
	}()

	wgPhase1.Wait()

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

	wgPhase2.Wait()

	if xmlErr != nil {
		return domain.Admin{}, "", nil, xmlErr
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
	}

	user.Password = ""
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
