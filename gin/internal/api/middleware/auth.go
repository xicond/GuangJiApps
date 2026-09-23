package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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
