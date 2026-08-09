package middleware

import (
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
