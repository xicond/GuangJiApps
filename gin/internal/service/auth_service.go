package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// menuCacheEntry stores cached menu items for an admin group with timestamp tracking.
type menuCacheEntry struct {
	menus     []domain.MainMenuItem
	fetchedAt time.Time
}

// AuthService handles authentication, legacy password encryption, JWT issuance, and permission menu resolution.
type AuthService struct {
	cfg           config.Config
	db            *gorm.DB
	menuCache     sync.Map
	webauthnCache sync.Map
}

// convertToDotNetASCIIString replicates .NET System.Text.ASCIIEncoding.ASCII.GetString(result),
// replacing non-ASCII binary bytes (>127) with '?' to match legacy .NET password hashing behavior.
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

// EncryptPassword replicates Common.SecurityHelper.Base64Encode(Common.SecurityHelper.Encrypt(rawpassword))
// from the legacy .NET GuangJiWeb application using MD5 + ASCII quirk + Base64.
//
// Parameters:
//   - rawPassword: plain text password string.
//
// Returns:
//   - string: encrypted and base64-encoded password string matching database storage.
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

// NewAuthService constructs an AuthService instance with configuration and database connection.
//
// Parameters:
//   - cfg: application configuration.
//   - db: GORM database handle.
//
// Returns:
//   - *AuthService: initialized AuthService instance.
func NewAuthService(cfg config.Config, db *gorm.DB) *AuthService {
	return &AuthService{cfg: cfg, db: db}
}

// GenerateToken creates a signed JWT access token for the given user ID valid for 24 hours.
//
// Parameters:
//   - userID: administrative user ID.
//
// Returns:
//   - string: signed JWT token string.
//   - error: non-nil if key loading or token signing fails.
func (s *AuthService) GenerateToken(userID int32) (string, error) {
	key, method, err := s.cfg.GetJWTSigningKey()
	if err != nil {
		return "", fmt.Errorf("failed to load JWT signing key: %w", err)
	}
	token := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(key)
}

// executeWithRetry executes a database operation up to 3 times with backoff if transient TCP errors occur.
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

// Login authenticates credentials via SP_Login, fetches authorized menu items, and generates a JWT token.
//
// Parameters:
//   - username: admin login username.
//   - password: raw plaintext password.
//
// Returns:
//   - domain.Admin: authenticated admin user details (password cleared).
//   - string: signed JWT token.
//   - []domain.MainMenuItem: tree structure of authorized menu items.
//   - error: non-nil if authentication or database lookup fails.
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
	wgUser.Go(func() {
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
	})

	wgUser.Wait()

	if userErr != nil {
		return domain.Admin{}, "", nil, userErr
	}

	// Async fire-and-forget update LastLogin timestamp
	go func(userID int32) {
		if err := s.db.Table("T_Login_Mst").
			Where("LoginId = ?", userID).
			Update("LastLogin", domain.NowDateTime()).Error; err != nil {
			log.Printf("[WARNING] Failed to update LastLogin for LoginId %d: %v", userID, err)
		}
	}(user.ID)

	// Synchronously generate token
	token, err = s.GenerateToken(user.ID)
	if err != nil {
		return domain.Admin{}, "", nil, err
	}

	mainMenus = s.fetchMenusForUser(username, result.GroupId)

	user.Password = ""
	// user.GroupId = nil
	s.CacheJWTToken(user.ID, token)
	return user, token, mainMenus, nil
}

// fetchMenusForUser retrieves the main menus and submenus for a given username and groupID,
// utilizing the 5-minute in-RAM cache to avoid repeated SP executions.
func (s *AuthService) fetchMenusForUser(username string, groupID int) []domain.MainMenuItem {
	var mainMenus []domain.MainMenuItem
	cacheKey := fmt.Sprintf("%s:%d", username, groupID)
	if val, ok := s.menuCache.Load(cacheKey); ok {
		entry := val.(menuCacheEntry)
		if time.Since(entry.fetchedAt) < 5*time.Minute {
			return entry.menus
		}
	}

	xmlErr := s.db.Raw("EXEC SP_Login_Create_Xml ?", username).Scan(&mainMenus).Error
	if xmlErr != nil {
		return []domain.MainMenuItem{}
	}

	if len(mainMenus) > 0 {
		var wgSubMenu sync.WaitGroup
		for i := range mainMenus {
			idx := i
			wgSubMenu.Go(func() {
				parentID := mainMenus[idx].MenuID

				rows, err := s.db.Raw("EXEC [dbo].[SP_Login_View_Mapping_Group] @ParentId = ?, @groupid = ?", parentID, groupID).Rows()
				if err != nil {
					log.Printf("[WARNING] Failed to fetch sub menu for ParentId %d, GroupId %d: %v", parentID, groupID, err)
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
			})
		}

		wgSubMenu.Wait()

		s.menuCache.Store(cacheKey, menuCacheEntry{
			menus:     mainMenus,
			fetchedAt: time.Now(),
		})
	}

	return mainMenus
}

// CacheJWTToken stores claims and token mapping in Redis with calculated TTL to enable fast validation.
//
// Parameters:
//   - userID: administrative user ID.
//   - tokenString: issued JWT token string.
func (s *AuthService) CacheJWTToken(userID int32, tokenString string) {
	if tokenString == "" {
		return
	}
	rdb := database.GetRedisClient(s.cfg)
	if rdb == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var claims jwt.MapClaims
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, &claims)
	if err != nil {
		claims = jwt.MapClaims{
			"sub": userID,
			"exp": time.Now().Add(24 * time.Hour).Unix(),
		}
	}

	expVal, ok := claims["exp"]
	if !ok {
		return
	}

	var expUnix int64
	switch v := expVal.(type) {
	case float64:
		expUnix = int64(v)
	case int64:
		expUnix = v
	case json.Number:
		expUnix, _ = v.Int64()
	default:
		expUnix = time.Now().Add(24 * time.Hour).Unix()
	}

	expTime := time.Unix(expUnix, 0)
	ttl := time.Until(expTime) - 200*time.Millisecond
	if ttl <= 0 {
		return
	}

	userTokensKey := fmt.Sprintf("jwt:user:%d:tokens", userID)
	oldTokenKeys, err := rdb.SMembers(ctx, userTokensKey).Result()
	if err == nil && len(oldTokenKeys) > 0 {
		rdb.Del(ctx, oldTokenKeys...)
	}
	rdb.Del(ctx, userTokensKey)

	tokenKey := fmt.Sprintf("jwt:token:%s", tokenString)
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return
	}

	rdb.Set(ctx, tokenKey, claimsBytes, ttl)
	rdb.SAdd(ctx, userTokensKey, tokenKey)
	rdb.Expire(ctx, userTokensKey, ttl)
}

// toInt coerces arbitrary primitive types (numbers, byte slices, numeric strings) into an int.
//
// Parameters:
//   - val: interface value of scalar type.
//
// Returns:
//   - int: converted integer value.
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

// BoolPtr allocates and returns a pointer to a boolean value.
//
// Parameters:
//   - b: boolean literal.
//
// Returns:
//   - *bool: pointer to boolean.
func BoolPtr(b bool) *bool {
	return &b
}

// toBool converts arbitrary scalar database values (int, string, bool, byte) into a boolean.
//
// Parameters:
//   - val: raw interface value.
//
// Returns:
//   - bool: true if represents a truthy value (1, "true", "t", "y", "ya").
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

// toString converts arbitrary interface values or byte slices safely into string.
//
// Parameters:
//   - val: raw interface value.
//
// Returns:
//   - string: string representation.
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

// ChangePassword verifies the user's current password and updates it with the new encrypted password.
//
// Parameters:
//   - userID: administrative user ID (optional if username is provided).
//   - username: admin login username.
//   - oldPassword: existing plaintext password to verify.
//   - newPassword: new plaintext password to encrypt and persist.
//
// Returns:
//   - error: non-nil if validation fails, user is not found, or old password does not match.
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

// GenerateRefreshToken creates a signed JWT refresh token for the given user ID valid for 30 days.
//
// Parameters:
//   - userID: administrative user ID.
//
// Returns:
//   - string: signed refresh JWT token string.
//   - time.Time: expiration timestamp.
//   - error: non-nil if key loading or token signing fails.
func (s *AuthService) GenerateRefreshToken(userID int32) (string, time.Time, error) {
	key, method, err := s.cfg.GetJWTRefreshSigningKey()
	if err != nil {
		key, method, err = s.cfg.GetJWTSigningKey()
		if err != nil {
			return "", time.Time{}, fmt.Errorf("failed to load refresh signing key: %w", err)
		}
	}
	expiredAt := time.Now().Add(30 * 24 * time.Hour)
	token := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub":  userID,
		"exp":  expiredAt.Unix(),
		"type": "refresh",
	})
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return signedToken, expiredAt, nil
}

// AddPasskeyRequest encapsulates payload sent to POST /v1/add-passkey.
type AddPasskeyRequest struct {
	ID                string                      `json:"id"`
	RawID             string                      `json:"rawId"`
	Type              string                      `json:"type"`
	PublicKey         string                      `json:"publicKey,omitempty"`
	Response          WebAuthnAttestationResponse `json:"response"`
	ClientDataJSON    string                      `json:"clientDataJSON,omitempty"`
	AttestationObject string                      `json:"attestationObject,omitempty"`
	Transports        []string                    `json:"transports,omitempty"`
	Credential        *AddPasskeyRequest          `json:"credential,omitempty"`
	RefreshToken      string                      `json:"refresh_token,omitempty"`
}

// Normalize normalizes nested or flat payload structures into standard response attributes.
func (r *AddPasskeyRequest) Normalize() {
	if r.Credential != nil {
		if r.ID == "" {
			r.ID = r.Credential.ID
		}
		if r.RawID == "" {
			r.RawID = r.Credential.RawID
		}
		if r.Type == "" {
			r.Type = r.Credential.Type
		}
		if r.PublicKey == "" {
			r.PublicKey = r.Credential.PublicKey
		}
		if r.RefreshToken == "" {
			r.RefreshToken = r.Credential.RefreshToken
		}
		if r.Response.ClientDataJSON == "" {
			r.Response = r.Credential.Response
		}
		if r.ClientDataJSON == "" {
			r.ClientDataJSON = r.Credential.ClientDataJSON
		}
		if r.AttestationObject == "" {
			r.AttestationObject = r.Credential.AttestationObject
		}
		if len(r.Transports) == 0 && len(r.Credential.Transports) > 0 {
			r.Transports = r.Credential.Transports
		}
	}
	if r.Response.ClientDataJSON == "" && r.ClientDataJSON != "" {
		r.Response.ClientDataJSON = r.ClientDataJSON
	}
	if r.Response.AttestationObject == "" && r.AttestationObject != "" {
		r.Response.AttestationObject = r.AttestationObject
	}
	if r.Response.PublicKey == "" && r.PublicKey != "" {
		r.Response.PublicKey = r.PublicKey
	}
	if len(r.Response.Transports) == 0 && len(r.Transports) > 0 {
		r.Response.Transports = r.Transports
	}
}

// WebAuthnAttestationResponse holds raw responses from navigator.credentials.create().
type WebAuthnAttestationResponse struct {
	ClientDataJSON    string   `json:"clientDataJSON"`
	AttestationObject string   `json:"attestationObject"`
	AuthenticatorData string   `json:"authenticatorData,omitempty"`
	PublicKey         string   `json:"publicKey,omitempty"`
	Transports        []string `json:"transports,omitempty"`
}

// LoginOptionsRequest represents payload for POST /login-options.
type LoginOptionsRequest struct {
	Username string `json:"username" binding:"required"`
}

// LoginOptionsCredential represents individual allowed credential in WebAuthn assertion options.
type LoginOptionsCredential struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Transports []string `json:"transports,omitempty"`
}

// LoginOptionsResponse represents the response for POST /login-options.
type LoginOptionsResponse struct {
	Challenge        string                   `json:"challenge"`
	RpID             string                   `json:"rpId"`
	AllowCredentials []LoginOptionsCredential `json:"allowCredentials"`
	UserVerification string                   `json:"userVerification"`
}

// LoginRenewRequest encapsulates payload sent to POST /login-renew.
type LoginRenewRequest struct {
	ID                string                    `json:"id"`
	RawID             string                    `json:"rawId"`
	Type              string                    `json:"type"`
	Response          WebAuthnAssertionResponse `json:"response"`
	ClientDataJSON    string                    `json:"clientDataJSON,omitempty"`
	AuthenticatorData string                    `json:"authenticatorData,omitempty"`
	Signature         string                    `json:"signature,omitempty"`
	UserHandle        string                    `json:"userHandle,omitempty"`
	Credential        *LoginRenewRequest        `json:"credential,omitempty"`
	RefreshToken      string                    `json:"refresh_token,omitempty"`
}

// Normalize normalizes nested or flat payload structures into standard response attributes.
func (r *LoginRenewRequest) Normalize() {
	if r.Credential != nil {
		if r.ID == "" {
			r.ID = r.Credential.ID
		}
		if r.RawID == "" {
			r.RawID = r.Credential.RawID
		}
		if r.Type == "" {
			r.Type = r.Credential.Type
		}
		if r.RefreshToken == "" {
			r.RefreshToken = r.Credential.RefreshToken
		}
		if r.Response.ClientDataJSON == "" {
			r.Response = r.Credential.Response
		}
		if r.ClientDataJSON == "" {
			r.ClientDataJSON = r.Credential.ClientDataJSON
		}
		if r.AuthenticatorData == "" {
			r.AuthenticatorData = r.Credential.AuthenticatorData
		}
		if r.Signature == "" {
			r.Signature = r.Credential.Signature
		}
	}
	if r.Response.ClientDataJSON == "" && r.ClientDataJSON != "" {
		r.Response.ClientDataJSON = r.ClientDataJSON
	}
	if r.Response.AuthenticatorData == "" && r.AuthenticatorData != "" {
		r.Response.AuthenticatorData = r.AuthenticatorData
	}
	if r.Response.Signature == "" && r.Signature != "" {
		r.Response.Signature = r.Signature
	}
}

// WebAuthnAssertionResponse holds raw responses from navigator.credentials.get().
type WebAuthnAssertionResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle,omitempty"`
}

// PasskeyUser adapts internal admin identity to the webauthn.User interface.
type PasskeyUser struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

func (u *PasskeyUser) WebAuthnID() []byte                         { return u.ID }
func (u *PasskeyUser) WebAuthnName() string                       { return u.Name }
func (u *PasskeyUser) WebAuthnDisplayName() string                { return u.DisplayName }
func (u *PasskeyUser) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }

// decodeBase64OrURL decodes string encoded with Standard Base64 or URL Base64 (with or without padding).
func decodeBase64OrURL(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return nil, errors.New("failed to decode base64 / base64url data")
}

// getWebAuthn returns a cached or newly initialized WebAuthn instance for the specified origin and referer.
// Gets host from referer header:
//   - if localhost, rpID=HOST, rpOrigin=http://HOST:env(PORT)
//   - if no, rpID=HOST, rpOrigin=https://HOST
func (s *AuthService) getWebAuthn(origin string, optionalReferer ...string) (*webauthn.WebAuthn, string, string, error) {
	rpDisplayName := "GuangJiApps"

	// 1. Get host from referer header (fallback to origin)
	rawHost := ""
	if len(optionalReferer) > 0 && strings.TrimSpace(optionalReferer[0]) != "" {
		rawHost = strings.TrimSpace(optionalReferer[0])
	}
	if rawHost == "" && strings.TrimSpace(origin) != "" {
		rawHost = strings.TrimSpace(origin)
	}

	host := "localhost"
	if rawHost != "" {
		parseTarget := rawHost
		if !strings.Contains(parseTarget, "://") {
			parseTarget = "https://" + parseTarget
		}
		if u, err := url.Parse(parseTarget); err == nil && u.Hostname() != "" {
			host = u.Hostname()
		} else if h, _, err := net.SplitHostPort(rawHost); err == nil && h != "" {
			host = h
		} else if !strings.Contains(rawHost, "/") {
			host = rawHost
		}
	}
	if host == "" {
		host = "localhost"
	}

	// 2. Resolve rpID and rpOrigin:
	//    if localhost: rpID=HOST, rpOrigin=http://HOST:env(PORT)
	//    if no: rpID=HOST, rpOrigin=https://HOST
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = strings.TrimSpace(s.cfg.Port)
	}
	if port == "" {
		port = "8080"
	}

	rpID := host
	var rpOrigin string
	if strings.EqualFold(host, "localhost") || host == "127.0.0.1" {
		rpOrigin = fmt.Sprintf("http://%s:%s", host, port)
	} else {
		rpOrigin = fmt.Sprintf("https://%s", host)
	}

	origins := []string{rpOrigin}
	if origin != "" && origin != rpOrigin {
		origins = append(origins, origin)
	}

	cacheKey := rpOrigin + "|" + origin + "|" + rpID
	if val, ok := s.webauthnCache.Load(cacheKey); ok {
		retOrigin := rpOrigin
		if origin != "" {
			retOrigin = origin
		}
		return val.(*webauthn.WebAuthn), retOrigin, rpID, nil
	}

	w, err := webauthn.New(&webauthn.Config{
		RPDisplayName: rpDisplayName,
		RPID:          rpID,
		RPOrigins:     origins,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	s.webauthnCache.Store(cacheKey, w)
	retOrigin := rpOrigin
	if origin != "" {
		retOrigin = origin
	}
	return w, retOrigin, rpID, nil
}

// encodeECDSAPublicKeyToCOSE converts an *ecdsa.PublicKey to COSE key format using modern key.ECDH() (Go 1.20+),
// completely avoiding deprecated direct coordinate access on key.X and key.Y and guaranteeing fixed 32-byte coordinates.
func encodeECDSAPublicKeyToCOSE(key *ecdsa.PublicKey) ([]byte, error) {
	ecdhKey, err := key.ECDH()
	if err != nil {
		return nil, fmt.Errorf("failed to convert ECDSA key to ECDH: %w", err)
	}
	raw := ecdhKey.Bytes() // 0x04 || X (32 bytes) || Y (32 bytes)
	byteLen := (len(raw) - 1) / 2
	xBytes := raw[1 : 1+byteLen]
	yBytes := raw[1+byteLen:]

	coseMap := map[interface{}]interface{}{
		int64(1):  int64(2),
		int64(3):  int64(-7),
		int64(-1): int64(1),
		int64(-2): xBytes,
		int64(-3): yBytes,
	}
	return cbor.Marshal(coseMap)
}

// synthesizeAttestationObject creates a standard WebAuthn 'none' attestationObject CBOR structure
// from a direct PEM or DER public key string.
func synthesizeAttestationObject(pubKeyStr string, credID []byte, rpID string) (string, error) {
	pubKeyStr = strings.TrimSpace(pubKeyStr)
	var derBytes []byte
	if strings.HasPrefix(pubKeyStr, "-----BEGIN") {
		block, _ := pem.Decode([]byte(pubKeyStr))
		if block != nil {
			derBytes = block.Bytes
		}
	} else {
		b, err := decodeBase64OrURL(pubKeyStr)
		if err == nil {
			derBytes = b
		}
	}

	if len(derBytes) == 0 {
		return "", errors.New("empty or invalid public key")
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse PKIX public key: %w", err)
	}

	var coseBytes []byte
	switch key := pub.(type) {
	case *ecdsa.PublicKey:
		coseBytes, err = encodeECDSAPublicKeyToCOSE(key)
		if err != nil {
			return "", err
		}
	case *rsa.PublicKey:
		coseMap := map[interface{}]interface{}{
			int64(1):  int64(3),
			int64(3):  int64(-257),
			int64(-1): key.N.Bytes(),
			int64(-2): big.NewInt(int64(key.E)).Bytes(),
		}
		coseBytes, err = cbor.Marshal(coseMap)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported public key type: %T", pub)
	}

	rpIDHash := sha256.Sum256([]byte(rpID))
	authData := make([]byte, 37)
	copy(authData[0:32], rpIDHash[:])
	authData[32] = 0x41 // UP + AT flags

	aaguid := make([]byte, 16)
	credIdLenBytes := []byte{byte(len(credID) >> 8), byte(len(credID) & 0xff)}

	authData = append(authData, aaguid...)
	authData = append(authData, credIdLenBytes...)
	authData = append(authData, credID...)
	authData = append(authData, coseBytes...)

	attObjMap := map[string]interface{}{
		"fmt":      "none",
		"authData": authData,
		"attStmt":  map[string]interface{}{},
	}
	attObjBytes, err := cbor.Marshal(attObjMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal attestation object: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(attObjBytes), nil
}

// parseLegacyPublicKeyToCredential converts stored PEM/DER strings into a webauthn.Credential struct.
func parseLegacyPublicKeyToCredential(pubKeyStr string, credID []byte) (*webauthn.Credential, error) {
	pubKeyStr = strings.TrimSpace(pubKeyStr)
	var derBytes []byte
	if strings.HasPrefix(pubKeyStr, "-----BEGIN") {
		block, _ := pem.Decode([]byte(pubKeyStr))
		if block != nil {
			derBytes = block.Bytes
		}
	} else {
		b, err := decodeBase64OrURL(pubKeyStr)
		if err == nil {
			derBytes = b
		}
	}

	if len(derBytes) == 0 {
		return nil, errors.New("empty or invalid public key data")
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}

	var coseBytes []byte
	switch key := pub.(type) {
	case *ecdsa.PublicKey:
		coseBytes, err = encodeECDSAPublicKeyToCOSE(key)
		if err != nil {
			return nil, err
		}
	case *rsa.PublicKey:
		coseMap := map[interface{}]interface{}{
			int64(1):  int64(3),
			int64(3):  int64(-257),
			int64(-1): key.N.Bytes(),
			int64(-2): big.NewInt(int64(key.E)).Bytes(),
		}
		coseBytes, err = cbor.Marshal(coseMap)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pub)
	}

	return &webauthn.Credential{
		ID:        credID,
		PublicKey: coseBytes,
	}, nil
}

// AddPasskey validates the biometric creation credential via github.com/go-webauthn/webauthn/webauthn,
// verifies refresh token uniqueness, persists credential details in T_Auth_Passkey, and returns a new refresh token.
//
// Parameters:
//   - userID: authenticated user ID.
//   - req: passkey creation request containing WebAuthn attestation.
//   - bearerToken: raw JWT bearer token of the authenticated user to validate challenge.
//
// Returns:
//   - string: newly issued refresh token.
//   - error: non-nil if validation, verification, or database persistence fails.
func (s *AuthService) AddPasskey(userID int32, req AddPasskeyRequest, bearerToken string, clientInfo ...string) (string, error) {
	req.Normalize()

	var clientIP, userAgent, referer string
	if len(clientInfo) > 0 {
		clientIP = clientInfo[0]
	}
	if len(clientInfo) > 1 {
		userAgent = clientInfo[1]
	}
	if len(clientInfo) > 2 {
		referer = clientInfo[2]
	}
	var ipPtr, uaPtr *string
	if clientIP != "" {
		ipPtr = &clientIP
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	if req.Response.ClientDataJSON == "" {
		return "", errors.New("clientDataJSON is required")
	}

	// 1. Decode clientDataJSON to extract challenge, type, and origin
	clientDataBytes, err := decodeBase64OrURL(req.Response.ClientDataJSON)
	if err != nil {
		return "", fmt.Errorf("invalid clientDataJSON encoding: %w", err)
	}

	var clientData struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}
	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return "", fmt.Errorf("invalid clientDataJSON payload: %w", err)
	}

	if clientData.Type != "webauthn.create" {
		return "", fmt.Errorf("invalid clientData type: expected webauthn.create, got %s", clientData.Type)
	}

	// 2. Validate challenge in clientDataJSON against bearer token (or refresh token)
	bearerChallengeRaw := base64.RawURLEncoding.EncodeToString([]byte(bearerToken))
	bearerChallengeStd := base64.URLEncoding.EncodeToString([]byte(bearerToken))
	match := (clientData.Challenge == bearerChallengeRaw || clientData.Challenge == bearerChallengeStd || clientData.Challenge == bearerToken)
	if !match && req.RefreshToken != "" {
		rfChallengeRaw := base64.RawURLEncoding.EncodeToString([]byte(req.RefreshToken))
		rfChallengeStd := base64.URLEncoding.EncodeToString([]byte(req.RefreshToken))
		match = (clientData.Challenge == rfChallengeRaw || clientData.Challenge == rfChallengeStd || clientData.Challenge == req.RefreshToken)
	}
	if !match {
		return "", errors.New("challenge mismatch: provided challenge does not match expected authentication token")
	}

	// 3. Validate that client-provided refresh token does not already exist in T_Auth_Passkey
	if req.RefreshToken != "" {
		rfHash := fmt.Sprintf("%x", md5.Sum([]byte(req.RefreshToken)))
		var count int64
		if err := s.db.Table("T_Auth_Passkey").Where("RefreshTokenHash = ? OR RefreshToken = ?", rfHash, req.RefreshToken).Count(&count).Error; err != nil {
			return "", fmt.Errorf("database check error: %w", err)
		}
		if count > 0 {
			return "", errors.New("refresh token already registered")
		}
	}

	// 4. Initialize WebAuthn engine for the ceremony origin
	w, rpOrigin, rpID, err := s.getWebAuthn(clientData.Origin, referer)
	if err != nil {
		return "", err
	}

	var credID []byte
	if b, err := decodeBase64OrURL(req.RawID); err == nil && len(b) > 0 {
		credID = b
	} else if b, err := decodeBase64OrURL(req.ID); err == nil && len(b) > 0 {
		credID = b
	} else if req.RawID != "" {
		credID = []byte(req.RawID)
	} else if req.ID != "" {
		credID = []byte(req.ID)
	} else {
		credID = []byte("cred-" + strconv.FormatInt(time.Now().UnixNano(), 10))
	}

	idStr := base64.RawURLEncoding.EncodeToString(credID)
	rawIDStr := idStr

	// Support synthesized attestationObject from direct public key PEM/DER if omitted
	if req.Response.AttestationObject == "" {
		pubKeyStr := req.Response.PublicKey
		if pubKeyStr == "" {
			pubKeyStr = req.PublicKey
		}
		if pubKeyStr != "" {
			attObjB64, synthErr := synthesizeAttestationObject(pubKeyStr, credID, rpID)
			if synthErr != nil {
				return "", fmt.Errorf("failed to process public key: %w", synthErr)
			}
			req.Response.AttestationObject = attObjB64
		} else {
			return "", errors.New("attestationObject or publicKey is required")
		}
	}

	creationMap := map[string]interface{}{
		"id":    idStr,
		"rawId": rawIDStr,
		"type":  "public-key",
		"response": map[string]interface{}{
			"clientDataJSON":    req.Response.ClientDataJSON,
			"attestationObject": req.Response.AttestationObject,
		},
	}
	var transports []string
	if len(req.Transports) > 0 {
		transports = req.Transports
	} else if len(req.Response.Transports) > 0 {
		transports = req.Response.Transports
	}
	if len(transports) > 0 {
		creationMap["response"].(map[string]interface{})["transports"] = transports
	}
	creationBytes, _ := json.Marshal(creationMap)

	parsedCreation, err := protocol.ParseCredentialCreationResponseBytes(creationBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse credential creation response: %w", err)
	}

	user := &PasskeyUser{
		ID:          []byte(strconv.Itoa(int(userID))),
		Name:        strconv.Itoa(int(userID)),
		DisplayName: "User " + strconv.Itoa(int(userID)),
	}

	session := webauthn.SessionData{
		Challenge:      clientData.Challenge,
		UserID:         user.ID,
		Origin:         rpOrigin,
		RelyingPartyID: rpID,
		CredParams:     webauthn.CredentialParametersDefault(),
	}

	credential, err := w.CreateCredential(user, session, parsedCreation)
	if err != nil {
		return "", fmt.Errorf("webauthn registration verification failed: %w", err)
	}
	if len(credential.Transport) == 0 && len(transports) > 0 {
		for _, tr := range transports {
			credential.Transport = append(credential.Transport, protocol.AuthenticatorTransport(tr))
		}
	}

	credJSON, err := json.Marshal(credential)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}

	// 6. Generate new refresh token
	newRefreshToken, expiredAt, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	var count int64
	newHash := fmt.Sprintf("%x", md5.Sum([]byte(newRefreshToken)))
	_ = s.db.Table("T_Auth_Passkey").Where("RefreshTokenHash = ? OR RefreshToken = ?", newHash, newRefreshToken).Count(&count).Error
	if count > 0 {
		newRefreshToken, expiredAt, err = s.GenerateRefreshToken(userID)
		if err != nil {
			return "", err
		}
		newHash = fmt.Sprintf("%x", md5.Sum([]byte(newRefreshToken)))
	}

	// 7. Store passkey in T_Auth_Passkey
	credIDStr := base64.RawURLEncoding.EncodeToString(credential.ID)
	if credIDStr == "" && idStr != "" {
		credIDStr = idStr
	}

	now := time.Now()
	status := true
	passkey := domain.AuthPasskey{
		LoginID:          userID,
		CredentialID:     credIDStr,
		Credential:       string(credJSON),
		RefreshToken:     &newRefreshToken,
		RefreshTokenHash: newHash,
		IP:               ipPtr,
		UserAgent:        uaPtr,
		CreatedAt:        &now,
		LastUsed:         &now,
		ExpiredAt:        &expiredAt,
		Status:           &status,
	}

	if err := s.db.Table("T_Auth_Passkey").Create(&passkey).Error; err != nil {
		return "", fmt.Errorf("failed to save passkey: %w", err)
	}

	// log.Printf("[Passkey] AddPasskey: UserID=%d, CredentialID=%s, IP=%s, UserAgent=%s", userID, credIDStr, clientIP, userAgent)

	return newRefreshToken, nil
}

// LoginRenew verifies biometric assertion signature using github.com/go-webauthn/webauthn/webauthn,
// invalidates the used refresh token, generates and saves a new refresh token with the same credential,
// and returns the new refresh token.
//
// Parameters:
//   - passkey: validated existing passkey record from database.
//   - req: assertion request payload from navigator.credentials.get().
//   - currentRefreshToken: active refresh token used for challenge validation.
//
// Returns:
//   - string: new refresh token.
//   - string: renewed signed JWT access token.
//   - error: non-nil if verification fails or database update fails.
func (s *AuthService) LoginRenew(passkey domain.AuthPasskey, req LoginRenewRequest, currentRefreshToken string, clientInfo ...string) (string, string, error) {
	if passkey.RefreshToken == nil || *passkey.RefreshToken == "" {
		return "", "", errors.New("refresh token is required and cannot be null")
	}

	var clientIP, userAgent, referer string
	if len(clientInfo) > 0 {
		clientIP = clientInfo[0]
	}
	if len(clientInfo) > 1 {
		userAgent = clientInfo[1]
	}
	if len(clientInfo) > 2 {
		referer = clientInfo[2]
	}
	var ipPtr, uaPtr *string
	if clientIP != "" {
		ipPtr = &clientIP
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	req.Normalize()

	if req.Response.ClientDataJSON == "" {
		return "", "", errors.New("clientDataJSON is required")
	}
	if req.Response.AuthenticatorData == "" {
		return "", "", errors.New("authenticatorData is required")
	}
	if req.Response.Signature == "" {
		return "", "", errors.New("signature is required")
	}

	// 1. Decode clientDataJSON to verify type and challenge
	clientDataBytes, err := decodeBase64OrURL(req.Response.ClientDataJSON)
	if err != nil {
		return "", "", fmt.Errorf("invalid clientDataJSON encoding: %w", err)
	}

	var clientData struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}
	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return "", "", fmt.Errorf("invalid clientDataJSON payload: %w", err)
	}

	if clientData.Type != "webauthn.get" {
		return "", "", fmt.Errorf("invalid clientData type: expected webauthn.get, got %s", clientData.Type)
	}

	currChallengeRaw := base64.RawURLEncoding.EncodeToString([]byte(currentRefreshToken))
	currChallengeStd := base64.URLEncoding.EncodeToString([]byte(currentRefreshToken))
	pkToken := *passkey.RefreshToken
	pkChallengeRaw := base64.RawURLEncoding.EncodeToString([]byte(pkToken))
	pkChallengeStd := base64.URLEncoding.EncodeToString([]byte(pkToken))

	match := (clientData.Challenge == currChallengeRaw || clientData.Challenge == currChallengeStd || clientData.Challenge == currentRefreshToken ||
		clientData.Challenge == pkChallengeRaw || clientData.Challenge == pkChallengeStd || clientData.Challenge == pkToken)
	if !match {
		return "", "", errors.New("challenge mismatch: provided challenge does not match expected authentication token")
	}

	// 2. Initialize WebAuthn engine for origin
	w, rpOrigin, rpID, err := s.getWebAuthn(clientData.Origin, referer)
	if err != nil {
		return "", "", err
	}

	// 3. Unmarshal stored credential record from T_Auth_Passkey
	var storedCred webauthn.Credential
	if err := json.Unmarshal([]byte(passkey.Credential), &storedCred); err != nil || len(storedCred.PublicKey) == 0 {
		var fallbackID []byte
		if req.RawID != "" {
			fallbackID, _ = decodeBase64OrURL(req.RawID)
		}
		if len(fallbackID) == 0 && req.ID != "" {
			fallbackID, _ = decodeBase64OrURL(req.ID)
		}
		if len(fallbackID) == 0 && req.ID != "" {
			fallbackID = []byte(req.ID)
		}
		converted, convErr := parseLegacyPublicKeyToCredential(passkey.Credential, fallbackID)
		if convErr != nil {
			return "", "", fmt.Errorf("failed to decode stored passkey public key: %w", convErr)
		}
		storedCred = *converted
	}

	credID := storedCred.ID
	if len(credID) == 0 {
		if b, err := decodeBase64OrURL(req.RawID); err == nil && len(b) > 0 {
			credID = b
		} else if req.ID != "" {
			credID = []byte(req.ID)
		}
		storedCred.ID = credID
	}

	idStr := base64.RawURLEncoding.EncodeToString(credID)
	rawIDStr := idStr

	assertionMap := map[string]interface{}{
		"id":    idStr,
		"rawId": rawIDStr,
		"type":  "public-key",
		"response": map[string]interface{}{
			"clientDataJSON":    req.Response.ClientDataJSON,
			"authenticatorData": req.Response.AuthenticatorData,
			"signature":         req.Response.Signature,
			"userHandle":        req.Response.UserHandle,
		},
	}
	assertionBytes, _ := json.Marshal(assertionMap)

	parsedAssertion, err := protocol.ParseCredentialRequestResponseBytes(assertionBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse credential assertion response: %w", err)
	}

	// Match credential ID with parsed assertion ID if stored ID was empty
	if len(storedCred.ID) == 0 {
		storedCred.ID = parsedAssertion.RawID
	}

	user := &PasskeyUser{
		ID:          []byte(strconv.Itoa(int(passkey.LoginID))),
		Name:        strconv.Itoa(int(passkey.LoginID)),
		DisplayName: "User " + strconv.Itoa(int(passkey.LoginID)),
		Credentials: []webauthn.Credential{storedCred},
	}

	loginSession := webauthn.SessionData{
		Challenge:            clientData.Challenge,
		UserID:               user.ID,
		Origin:               rpOrigin,
		RelyingPartyID:       rpID,
		AllowedCredentialIDs: [][]byte{storedCred.ID},
		UserVerification:     protocol.VerificationPreferred,
	}

	// 5. Validate biometric assertion via WebAuthn library
	_, err = w.ValidateLogin(user, loginSession, parsedAssertion)
	if err != nil {
		return "", "", fmt.Errorf("biometric assertion validation failed: %w", err)
	}

	// 6. Invalidate old refresh token (soft-delete status = false, update LastUsed)
	now := time.Now()
	if err := s.db.Table("T_Auth_Passkey").
		Where("PasskeyId = ?", passkey.PasskeyID).
		Updates(map[string]interface{}{
			"Status":   false,
			"LastUsed": now,
		}).Error; err != nil {
		return "", "", fmt.Errorf("failed to invalidate old refresh token: %w", err)
	}

	// 7. Generate new refresh token
	newRefreshToken, expiredAt, err := s.GenerateRefreshToken(passkey.LoginID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// 8. Generate renewed access token
	newAccessToken, err := s.GenerateToken(passkey.LoginID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate renewed access token: %w", err)
	}

	// 9. Store new refresh token with same credential
	status := true
	renewedHash := fmt.Sprintf("%x", md5.Sum([]byte(newRefreshToken)))
	newPasskey := domain.AuthPasskey{
		LoginID:          passkey.LoginID,
		CredentialID:     passkey.CredentialID,
		Credential:       passkey.Credential,
		RefreshToken:     &newRefreshToken,
		RefreshTokenHash: renewedHash,
		IP:               ipPtr,
		UserAgent:        uaPtr,
		CreatedAt:        &now,
		LastUsed:         &now,
		ExpiredAt:        &expiredAt,
		Status:           &status,
	}

	if err := s.db.Table("T_Auth_Passkey").Create(&newPasskey).Error; err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}

	// log.Printf("[Passkey] LoginRenew: LoginID=%d, CredentialID=%s, IP=%s, UserAgent=%s", passkey.LoginID, passkey.CredentialID, clientIP, userAgent)

	return newRefreshToken, newAccessToken, nil
}

// GetLoginOptions generates WebAuthn assertion options for user login given a username.
// Returns challenge, relying party ID, user verification requirement, and allowCredentials.
//
// Parameters:
//   - username: user login name
//   - origin: client origin or Host string for resolving rpId
//
// Returns:
//   - *LoginOptionsResponse: WebAuthn assertion configuration
//   - error: non-nil if username not found, has no passkeys, or database fails
func (s *AuthService) GetLoginOptions(username, origin string) (*LoginOptionsResponse, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username is required")
	}

	// 1. Find user by username in T_Login_Mst
	var user domain.Admin
	err := s.db.Table("T_Login_Mst").
		Select("LoginId, Username").
		Where("Username = ?", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2. Fetch active passkeys for this LoginId (1 to Many, filtering out null/empty CredentialId)
	var passkeys []domain.AuthPasskey
	err = s.db.Table("T_Auth_Passkey").
		Where("LoginId = ? AND (Status = 1 OR Status IS NULL) AND CredentialId IS NOT NULL AND CredentialId != ''", user.ID).
		Order("PasskeyId DESC").
		Find(&passkeys).Error
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	now := time.Now()
	var allowCredentials []LoginOptionsCredential
	seenCredID := make(map[string]bool)

	for _, pk := range passkeys {
		if pk.CredentialID == "" {
			continue
		}
		if pk.ExpiredAt != nil && now.After(*pk.ExpiredAt) {
			continue
		}

		credID := pk.CredentialID
		var transports []string

		// Parse stored Credential JSON to extract transports
		if pk.Credential != "" {
			var storedCred webauthn.Credential
			if jsonErr := json.Unmarshal([]byte(pk.Credential), &storedCred); jsonErr == nil {
				for _, t := range storedCred.Transport {
					transports = append(transports, string(t))
				}
			}
		}

		if credID == "" {
			continue
		}

		if seenCredID[credID] {
			continue
		}
		seenCredID[credID] = true

		if len(transports) == 0 {
			transports = []string{"hybrid", "internal"}
		}

		allowCredentials = append(allowCredentials, LoginOptionsCredential{
			Type:       "public-key",
			ID:         credID,
			Transports: transports,
		})
	}

	if len(allowCredentials) == 0 {
		return nil, errors.New("no passkey found for this user")
	}

	// 3. Generate secure challenge
	challengeBytes, err := protocol.CreateChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	challengeStr := base64.RawURLEncoding.EncodeToString(challengeBytes)

	// 4. Resolve rpId from origin/host
	rpID := "localhost"
	origin = strings.TrimSpace(origin)
	if origin != "" {
		if u, err := url.Parse(origin); err == nil && u.Hostname() != "" {
			rpID = u.Hostname()
		} else if host, _, err := net.SplitHostPort(origin); err == nil {
			rpID = host
		} else if !strings.Contains(origin, "/") {
			rpID = origin
		}
	}

	return &LoginOptionsResponse{
		Challenge:        challengeStr,
		RpID:             rpID,
		AllowCredentials: allowCredentials,
		UserVerification: "required",
	}, nil
}

// LoginPasskeyRequest represents the payload for POST /login-passkey.
type LoginPasskeyRequest struct {
	Username   string                 `json:"username" binding:"required"`
	Credential LoginPasskeyCredential `json:"credential" binding:"required"`
	RememberMe *bool                  `json:"rememberme"`
}

// LoginPasskeyCredential represents the assertion credential from navigator.credentials.get().
type LoginPasskeyCredential struct {
	ID       string                    `json:"id"`
	RawID    string                    `json:"rawId"`
	Type     string                    `json:"type"`
	Response WebAuthnAssertionResponse `json:"response"`
}

// LoginPasskeyResponse represents the response returned by POST /login-passkey.
type LoginPasskeyResponse struct {
	Message      string                `json:"message"`
	Token        string                `json:"token"`
	AccessToken  string                `json:"access_token"`
	User         domain.Admin          `json:"user"`
	MainMenu     []domain.MainMenuItem `json:"main_menu"`
	RefreshToken *string               `json:"refresh_token,omitempty"`
	PublicKey    *string               `json:"public_key,omitempty"`
	Credential   *string               `json:"credential,omitempty"`
}

// LoginPasskey authenticates a user using a biometric WebAuthn passkey assertion.
// It verifies the cryptographic signature against the user's registered passkey credential,
// and if rememberme is true, issues a new refresh token and records it in T_Auth_Passkey without invalidating existing passkeys.
func (s *AuthService) LoginPasskey(req LoginPasskeyRequest, clientIP, userAgent string, clientInfo ...string) (*LoginPasskeyResponse, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, errors.New("username is required")
	}

	var referer string
	if len(clientInfo) > 0 {
		referer = clientInfo[0]
	}

	credID := strings.TrimSpace(req.Credential.ID)
	if credID == "" {
		credID = strings.TrimSpace(req.Credential.RawID)
	}
	if credID == "" {
		return nil, errors.New("credential id is required")
	}

	if req.Credential.Response.ClientDataJSON == "" {
		return nil, errors.New("clientDataJSON is required")
	}
	if req.Credential.Response.AuthenticatorData == "" {
		return nil, errors.New("authenticatorData is required")
	}
	if req.Credential.Response.Signature == "" {
		return nil, errors.New("signature is required")
	}

	// 1. Find user by username
	var user domain.Admin
	err := s.db.Table("T_Login_Mst").
		Select("LoginId, DepartmentId, GroupId, Username, ImgUrl, LastLogin, IsWarehouse").
		Where("Username = ?", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2. Decode clientDataJSON to extract challenge and origin
	clientDataBytes, err := decodeBase64OrURL(req.Credential.Response.ClientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid clientDataJSON encoding: %w", err)
	}

	var clientData struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}
	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return nil, fmt.Errorf("invalid clientDataJSON payload: %w", err)
	}

	if clientData.Type != "webauthn.get" {
		return nil, fmt.Errorf("invalid clientData type: expected webauthn.get, got %s", clientData.Type)
	}

	// 3. Find passkey matching user.ID and credential ID
	var passkeys []domain.AuthPasskey
	err = s.db.Table("T_Auth_Passkey").
		Where("LoginId = ? AND (Status = 1 OR Status IS NULL) AND CredentialId IS NOT NULL AND CredentialId != ''", user.ID).
		Order("PasskeyId DESC").
		Find(&passkeys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query passkeys: %w", err)
	}

	var matchedPasskey *domain.AuthPasskey
	rawReqCredID, _ := decodeBase64OrURL(credID)
	credIDAsB64Raw := base64.RawURLEncoding.EncodeToString([]byte(credID))
	credIDAsB64Std := base64.StdEncoding.EncodeToString([]byte(credID))

	for i := range passkeys {
		pk := &passkeys[i]
		if pk.CredentialID == credID || pk.CredentialID == credIDAsB64Raw || pk.CredentialID == credIDAsB64Std {
			matchedPasskey = pk
			break
		}
		if len(rawReqCredID) > 0 && pk.CredentialID == base64.RawURLEncoding.EncodeToString(rawReqCredID) {
			matchedPasskey = pk
			break
		}
		pkRaw, _ := decodeBase64OrURL(pk.CredentialID)
		if len(pkRaw) > 0 {
			if bytes.Equal(pkRaw, []byte(credID)) || (len(rawReqCredID) > 0 && bytes.Equal(rawReqCredID, pkRaw)) {
				matchedPasskey = pk
				break
			}
		}
	}

	if matchedPasskey == nil {
		return nil, errors.New("passkey credential not found or revoked")
	}

	// 4. Initialize WebAuthn engine
	w, rpOrigin, rpID, err := s.getWebAuthn(clientData.Origin, referer)
	if err != nil {
		return nil, err
	}

	// 5. Unmarshal stored credential
	var storedCred webauthn.Credential
	if err := json.Unmarshal([]byte(matchedPasskey.Credential), &storedCred); err != nil || len(storedCred.PublicKey) == 0 {
		var fallbackID []byte
		if len(rawReqCredID) > 0 {
			fallbackID = rawReqCredID
		} else {
			fallbackID = []byte(credID)
		}
		converted, convErr := parseLegacyPublicKeyToCredential(matchedPasskey.Credential, fallbackID)
		if convErr != nil {
			return nil, fmt.Errorf("failed to decode passkey credential: %w", convErr)
		}
		storedCred = *converted
	}

	if len(storedCred.ID) == 0 {
		if len(rawReqCredID) > 0 {
			storedCred.ID = rawReqCredID
		} else {
			storedCred.ID = []byte(credID)
		}
	}

	idStr := base64.RawURLEncoding.EncodeToString(storedCred.ID)
	assertionMap := map[string]interface{}{
		"id":    idStr,
		"rawId": idStr,
		"type":  "public-key",
		"response": map[string]interface{}{
			"clientDataJSON":    req.Credential.Response.ClientDataJSON,
			"authenticatorData": req.Credential.Response.AuthenticatorData,
			"signature":         req.Credential.Response.Signature,
			"userHandle":        req.Credential.Response.UserHandle,
		},
	}
	assertionBytes, _ := json.Marshal(assertionMap)
	parsedAssertion, err := protocol.ParseCredentialRequestResponseBytes(assertionBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential assertion response: %w", err)
	}

	passkeyUser := &PasskeyUser{
		ID:          []byte(strconv.Itoa(int(user.ID))),
		Name:        user.Username,
		DisplayName: user.Username,
		Credentials: []webauthn.Credential{storedCred},
	}

	loginSession := webauthn.SessionData{
		Challenge:            clientData.Challenge,
		UserID:               passkeyUser.ID,
		Origin:               rpOrigin,
		RelyingPartyID:       rpID,
		AllowedCredentialIDs: [][]byte{storedCred.ID},
		UserVerification:     protocol.VerificationPreferred,
	}

	// 6. Validate assertion signature
	_, err = w.ValidateLogin(passkeyUser, loginSession, parsedAssertion)
	if err != nil {
		return nil, fmt.Errorf("passkey signature validation failed: %w", err)
	}

	// 7. Update LastUsed on matched passkey
	now := time.Now()
	_ = s.db.Table("T_Auth_Passkey").
		Where("PasskeyId = ?", matchedPasskey.PasskeyID).
		Update("LastUsed", now).Error

	// 8. Async update LastLogin
	go func(userID int32) {
		_ = s.db.Table("T_Login_Mst").
			Where("LoginId = ?", userID).
			Update("LastLogin", domain.NowDateTime()).Error
	}(user.ID)

	// 9. Generate access token & fetch menus
	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	s.CacheJWTToken(user.ID, token)

	groupID := int(user.GroupId)
	mainMenus := s.fetchMenusForUser(user.Username, groupID)
	user.Password = ""

	var ipPtr, uaPtr *string
	if clientIP != "" {
		ipPtr = &clientIP
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	rememberMe := req.RememberMe != nil && *req.RememberMe

	resp := &LoginPasskeyResponse{
		Message:     "login successful",
		Token:       token,
		AccessToken: token,
		User:        user,
		MainMenu:    mainMenus,
	}

	if rememberMe {
		newRefreshToken, expiredAt, err := s.GenerateRefreshToken(user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate refresh token: %w", err)
		}
		newHash := fmt.Sprintf("%x", md5.Sum([]byte(newRefreshToken)))
		statusTrue := true
		newPasskeyRecord := domain.AuthPasskey{
			LoginID:          user.ID,
			CredentialID:     matchedPasskey.CredentialID,
			Credential:       matchedPasskey.Credential,
			RefreshToken:     &newRefreshToken,
			RefreshTokenHash: newHash,
			IP:               ipPtr,
			UserAgent:        uaPtr,
			CreatedAt:        &now,
			LastUsed:         &now,
			ExpiredAt:        &expiredAt,
			Status:           &statusTrue,
		}
		if err := s.db.Table("T_Auth_Passkey").Create(&newPasskeyRecord).Error; err != nil {
			return nil, fmt.Errorf("failed to persist passkey session: %w", err)
		}
		resp.RefreshToken = &newRefreshToken
		resp.PublicKey = &matchedPasskey.Credential
		resp.Credential = &matchedPasskey.Credential
		// log.Printf("[Passkey] LoginPasskey (RememberMe): User=%s, LoginID=%d, CredentialID=%s, IP=%s, UserAgent=%s", user.Username, user.ID, matchedPasskey.CredentialID, clientIP, userAgent)
	} else {
		// log.Printf("[Passkey] LoginPasskey: User=%s, LoginID=%d, CredentialID=%s, IP=%s, UserAgent=%s", user.Username, user.ID, matchedPasskey.CredentialID, clientIP, userAgent)
	}

	return resp, nil
}
