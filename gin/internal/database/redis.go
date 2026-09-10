package database

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"guangjiapps/gin/internal/config"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// GetRedisClient returns a thread-safe singleton *redis.Client instance initialized from config.
// Returns nil if RedisAddr is empty or if Ping fails.
func GetRedisClient(cfg config.Config) *redis.Client {
	redisOnce.Do(func() {
		if cfg.RedisAddr == "" {
			return
		}
		dbIdx := 0
		if d, err := strconv.Atoi(cfg.RedisDB); err == nil {
			dbIdx = d
		}
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       dbIdx,
		})
		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelTimeout()
		if err := client.Ping(ctxTimeout).Err(); err == nil {
			redisClient = client
			log.Println("[Redis] Singleton client initialized successfully")
		} else {
			log.Printf("[Redis] Ping failed (%v), Redis caching disabled", err)
			_ = client.Close()
		}
	})
	return redisClient
}

// CloseRedisClient closes the singleton Redis client connection if initialized.
func CloseRedisClient() {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("[Redis] Error closing Redis client: %v", err)
		} else {
			log.Println("[Redis] Client connection closed")
		}
		redisClient = nil
	}
}
