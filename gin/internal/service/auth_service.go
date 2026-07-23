package service

import (
	"errors"
	"fmt"
	"time"

	"crypto/md5"
	"encoding/base64"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"

	"github.com/golang-jwt/jwt/v5"
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

func (s *AuthService) Login(username, password string) (domain.User, string, error) {
	if username == "" || password == "" {
		return domain.User{}, "", errors.New("username and password are required")
	}

	// Call SP_Login for validation
	var result struct {
		Warn     string
		FlagWarn int
		GroupId  int
	}
	err := s.db.Raw("EXEC SP_Login ?, ?", username, EncryptPassword(password)).Scan(&result).Error
	if err != nil {
		return domain.User{}, "", err
	}
	if result.FlagWarn == 0 {
		return domain.User{}, "", errors.New(result.Warn)
	}

	// Fetch user from T_Login_Mst
	var user domain.User
	err = s.db.Table("T_Login_Mst").
		Select("LoginId as [id], email, Username as name").
		Where("Username = ?", username).
		First(&user).Error
	if err != nil {
		// 1. Jika error murni karena username tidak terdaftar di DB
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, "", errors.New("username tidak ditemukan")
		}

		// 2. Jika error karena masalah MSSQL (misal: "invalid column name", "connection timeout")
		// Mengembalikan pesan error asli dari sistem SQL Server secara dinamis
		return domain.User{}, "", fmt.Errorf("database error: %w", err)
	}
	// user.Role = "member" // Default role

	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return domain.User{}, "", err
	}
	return user, token, nil
}
