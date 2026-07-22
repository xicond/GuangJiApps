package service

import (
	"errors"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/domain"

	"gorm.io/gorm"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	cfg config.Config
	db  *gorm.DB
}

func NewAuthService(cfg config.Config, db *gorm.DB) *AuthService {
	return &AuthService{cfg: cfg, db: db}
}

func (s *AuthService) GenerateToken(userID string) (string, error) {
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
	err := s.db.Raw("EXEC SP_Login ?, ?", username, password).Scan(&result).Error
	if err != nil {
		return domain.User{}, "", err
	}
	if result.FlagWarn == 0 {
		return domain.User{}, "", errors.New(result.Warn)
	}

	// Fetch user from T_Login_Mst
	var user domain.User
	err = s.db.Table("T_Login_Mst").
		Select("LoginId as id, email, Username as name").
		Where("Username = ?", username).
		First(&user).Error
	if err != nil {
		return domain.User{}, "", errors.New("user not found")
	}
	user.Role = "member" // Default role

	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return domain.User{}, "", err
	}
	return user, token, nil
}

func (s *AuthService) Profile(userID string) (domain.User, error) {
	if userID == "" {
		return domain.User{}, errors.New("user id is required")
	}
	return domain.User{ID: userID, Email: "admin@example.com", Name: "Admin", Role: "admin"}, nil
}
