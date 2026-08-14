package middleware

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	memorystore "github.com/ulule/limiter/v3/drivers/store/memory"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
)

type LoginRateLimiter struct {
	instance *limiter.Limiter
	store    limiter.Store
	rate     limiter.Rate
}

func NewLoginRateLimiter(cfg config.Config) *LoginRateLimiter {
	// 3 attempts per 5 minutes
	rate := limiter.Rate{
		Period: 5 * time.Minute,
		Limit:  3,
	}

	var store limiter.Store
	if rdb := database.GetRedisClient(cfg); rdb != nil {
		rs, err := redisstore.NewStoreWithOptions(rdb, limiter.StoreOptions{
			Prefix: "login_limiter:",
		})
		if err == nil {
			store = rs
			// log.Println("[RateLimiter] Redis store initialized successfully for /login")
		} else {
			log.Printf("[RateLimiter] Failed to create Redis store, fallback to memory store: %v\n", err)
		}
	}

	if store == nil {
		store = memorystore.NewStore()
		log.Println("[RateLimiter] Memory store initialized for /login")
	}

	instance := limiter.New(store, rate)
	return &LoginRateLimiter{
		instance: instance,
		store:    store,
		rate:     rate,
	}
}

func sanitizeIP(ip string) string {
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}

func (l *LoginRateLimiter) GetKey(c *gin.Context) string {
	return c.ClientIP()
}

func (l *LoginRateLimiter) SetHeaders(c *gin.Context, limCtx limiter.Context) int64 {
	c.Header("X-RateLimit-Limit", strconv.FormatInt(limCtx.Limit, 10))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(limCtx.Remaining, 10))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(limCtx.Reset, 10))

	var retryAfterSeconds int64 = 0
	if limCtx.Reached {
		now := time.Now().Unix()
		retryAfterSeconds = limCtx.Reset - now
		if retryAfterSeconds <= 0 {
			retryAfterSeconds = 1
		}
		c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
		c.Header("X-Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	}
	return retryAfterSeconds
}

func (l *LoginRateLimiter) Peek(c *gin.Context) (limiter.Context, error) {
	key := l.GetKey(c)
	return l.instance.Peek(c.Request.Context(), key)
}

func (l *LoginRateLimiter) Increment(c *gin.Context) (limiter.Context, error) {
	key := l.GetKey(c)
	return l.instance.Get(c.Request.Context(), key)
}

func (l *LoginRateLimiter) Reset(c *gin.Context) {
	key := l.GetKey(c)
	ctx := c.Request.Context()
	_, _ = l.store.Reset(ctx, key, l.rate)
}

type UserRateLimiter struct {
	instance *limiter.Limiter
	store    limiter.Store
	rate     limiter.Rate
}

func NewUserRateLimiter(cfg config.Config, limit int64, period time.Duration, prefix string) *UserRateLimiter {
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	var store limiter.Store
	if rdb := database.GetRedisClient(cfg); rdb != nil {
		rs, err := redisstore.NewStoreWithOptions(rdb, limiter.StoreOptions{
			Prefix: prefix,
		})
		if err == nil {
			store = rs
		} else {
			log.Printf("[RateLimiter] Failed to create Redis store for %s, fallback to memory: %v\n", prefix, err)
		}
	}

	if store == nil {
		store = memorystore.NewStore()
	}

	instance := limiter.New(store, rate)
	return &UserRateLimiter{
		instance: instance,
		store:    store,
		rate:     rate,
	}
}

func (u *UserRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		if userIDVal, exists := c.Get("userID"); exists && userIDVal != nil {
			key = fmt.Sprintf("user_%v", userIDVal)
		} else {
			key = c.ClientIP()
		}

		limCtx, err := u.instance.Get(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(limCtx.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(limCtx.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(limCtx.Reset, 10))

		if limCtx.Reached {
			now := time.Now().Unix()
			retryAfterSeconds := limCtx.Reset - now
			if retryAfterSeconds <= 0 {
				retryAfterSeconds = 1
			}
			c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
			c.Header("X-Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
			c.JSON(429, gin.H{
				"error":       fmt.Sprintf("Terlalu banyak permintaan. Harap tunggu %d detik.", retryAfterSeconds),
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
