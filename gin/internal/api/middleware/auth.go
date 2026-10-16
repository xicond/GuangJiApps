package middleware

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthMiddleware validates Bearer JWT tokens from the Authorization header,
// checks the Redis token cache to accelerate validation, and sets the authenticated "userID" in Gin Context.
//
// Parameters:
//   - cfg: application configuration containing JWT public verification keys.
//
// Returns:
//   - gin.HandlerFunc: Gin middleware handler function.
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 1. Check Redis cache availability to speedup and skip validation
		rdb := database.GetRedisClient(cfg)
		if rdb != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
			defer cancel()

			tokenKey := "jwt:token:" + tokenString
			val, err := rdb.Get(ctx, tokenKey).Result()
			if err == nil && val != "" {
				var claims jwt.MapClaims
				if err := json.Unmarshal([]byte(val), &claims); err == nil {
					if sub, exists := claims["sub"]; exists {
						c.Set("userID", sub)
						c.Next()
						return
					}
				}
			}
		}

		// 2. Cache miss or Redis unavailable: do normal validation
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			vKey, method, err := cfg.GetJWTVerificationKey()
			if err != nil {
				return nil, err
			}
			if token.Method.Alg() != method.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return vKey, nil
		}, jwt.WithValidMethods([]string{
			jwt.SigningMethodRS256.Alg(),
			jwt.SigningMethodES256.Alg(),
		}))

		if err != nil || !token.Valid {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", claims["sub"])
		c.Next()
	}
}

// RefreshTokenAuthMiddleware validates refresh tokens from Authorization or X-Refresh-Token header or body,
// checks claims and expiration, and verifies existence and active status in T_Auth_Passkey database table.
func RefreshTokenAuthMiddleware(cfg config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If still empty, check request body if JSON
		// if tokenString == "" && c.Request.Body != nil {
		// 	bodyBytes, err := io.ReadAll(c.Request.Body)
		// 	if err == nil && len(bodyBytes) > 0 {
		// 		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// 		var bodyMap map[string]interface{}
		// 		if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
		// 			if rt, ok := bodyMap["refresh_token"].(string); ok && rt != "" {
		// 				tokenString = rt
		// 			}
		// 		}
		// 	}
		// }

		if tokenString == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
			c.Abort()
			return
		}

		// 1. Validate JWT signature and claims
		vKey, method, err := cfg.GetJWTRefreshVerificationKey()
		if err != nil {
			vKey, method, err = cfg.GetJWTVerificationKey()
		}
		if err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load verification key"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != method.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return vKey, nil
		}, jwt.WithValidMethods([]string{
			jwt.SigningMethodRS256.Alg(),
			jwt.SigningMethodES256.Alg(),
		}))

		if err != nil || !token.Valid {
			c.Header("Content-Type", "application/json")
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			}
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Check expiration claim
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
				c.Abort()
				return
			}
		}

		// 2. Validate against T_Auth_Passkey in DB: exist, status=1, not expired
		tokenHash := fmt.Sprintf("%x", md5.Sum([]byte(tokenString)))
		var passkey domain.AuthPasskey
		err = db.Table("T_Auth_Passkey").
			Where("(RefreshTokenHash = ? OR RefreshToken = ?) AND (Status = 1 OR Status IS NULL) AND RefreshToken IS NOT NULL AND RefreshToken != ''", tokenHash, tokenString).
			First(&passkey).Error
		if err != nil || passkey.RefreshToken == nil || *passkey.RefreshToken == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found or revoked"})
			c.Abort()
			return
		}

		if passkey.ExpiredAt != nil && time.Now().After(*passkey.ExpiredAt) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired in database"})
			c.Abort()
			return
		}

		c.Set("userID", passkey.LoginID)
		c.Set("refreshToken", tokenString)
		c.Set("passkey", passkey)
		c.Next()
	}
}
