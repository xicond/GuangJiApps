package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ResourceMetadata struct mapping route to DB table and caching details
type ResourceMetadata struct {
	TableName     string
	IdParam       string
	UpdatedColumn string        // defaults to "ModDate" or "updated_at"
	CacheKey      string        // explicit cache key override
	WhereClause   string        // SQL filter condition (e.g. "CategoryId = ?")
	WhereArgs     []interface{} // SQL filter arguments
	CacheTTL      time.Duration // defaults to 30 seconds
}

func parseScannedTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05.999999999 -0700 -07",
		"2006-01-02 15:04:05.999999999+07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700 -07",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05+07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// StatusNotModifiedHeader middleware using DB & Redis caching to avoid frequent table scans
func StatusNotModifiedHeader(db *gorm.DB, rdb *redis.Client, resource ResourceMetadata) gin.HandlerFunc {
	updatedCol := resource.UpdatedColumn
	if updatedCol == "" {
		if resource.TableName == "T_SXY_MST_PENGGALANG" || resource.TableName == "T_SXY_MST_DONATUR" {
			updatedCol = "updateddate"
		} else {
			updatedCol = "ModDate"
		}
	}

	ttl := resource.CacheTTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}

	return func(c *gin.Context) {
		id := ""
		if resource.IdParam != "" {
			id = c.Param(resource.IdParam)
			if id == "" {
				id = c.Query(resource.IdParam)
			}
		}

		var cacheKey string
		if resource.CacheKey != "" {
			cacheKey = "etag:" + resource.CacheKey
			if id != "" {
				cacheKey += ":" + id
			}
		} else {
			cacheKey = "etag:" + resource.TableName
			if resource.WhereClause != "" {
				cacheKey += ":" + resource.WhereClause
				for _, arg := range resource.WhereArgs {
					cacheKey += fmt.Sprintf(":%v", arg)
				}
			}
			if id != "" {
				cacheKey += ":" + id
			} else {
				cacheKey += ":max"
			}
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		var serverETag string
		var updatedAt time.Time

		// 1. Check Redis Cache
		if rdb != nil {
			val, err := rdb.Get(ctx, cacheKey).Result()
			if err == nil && val != "" {
				serverETag = val
			}
		}

		// 2. Cache miss -> query DB
		if serverETag == "" {
			query := db.Table(resource.TableName)

			if resource.WhereClause != "" {
				query = query.Where(resource.WhereClause, resource.WhereArgs...)
			}

			var rawStr sql.NullString
			if id != "" {
				query = query.Where("id = ?", id)
				err := query.Select(updatedCol).Scan(&rawStr).Error
				if err != nil || !rawStr.Valid {
					_ = db.Table(resource.TableName).Where("id = ?", id).Select("updated_at").Scan(&rawStr)
				}
			} else {
				err := query.Select(fmt.Sprintf("MAX(%s)", updatedCol)).Scan(&rawStr).Error
				if err != nil || !rawStr.Valid {
					_ = db.Table(resource.TableName).Select("MAX(updated_at)").Scan(&rawStr)
				}
			}

			if rawStr.Valid {
				updatedAt = parseScannedTime(rawStr.String)
			}

			if updatedAt.IsZero() {
				serverETag = "\"none\""
			} else {
				serverETag = "\"" + updatedAt.Format(time.RFC3339Nano) + "\""
			}

			// Store in Redis with TTL
			if rdb != nil && serverETag != "" {
				_ = rdb.Set(ctx, cacheKey, serverETag, ttl).Err()
			}
		}

		// 3. Extract updatedAt time for Last-Modified header & If-Modified-Since validation
		if updatedAt.IsZero() && serverETag != "" && serverETag != "\"none\"" {
			raw := strings.Trim(serverETag, "\"")
			updatedAt = parseScannedTime(raw)
		}

		var lastModifiedStr string
		if !updatedAt.IsZero() {
			lastModifiedStr = updatedAt.UTC().Format(http.TimeFormat)
		}

		// Set HTTP Cache Headers (Cache-Control: public, 1 week)
		// c.Header("Cache-Control", "public, max-age=604800, must-revalidate")
		c.Header("ETag", serverETag)
		if lastModifiedStr != "" {
			c.Header("Last-Modified", lastModifiedStr)
		}

		// 4. Client conditional validation (304 Not Modified)
		clientIfNoneMatch := c.GetHeader("If-None-Match")
		clientIfModifiedSince := c.GetHeader("If-Modified-Since")

		isNotModified := false
		if clientIfNoneMatch != "" && clientIfNoneMatch == serverETag {
			isNotModified = true
		} else if clientIfModifiedSince != "" && !updatedAt.IsZero() {
			if t, err := time.Parse(http.TimeFormat, clientIfModifiedSince); err == nil {
				if !updatedAt.Truncate(time.Second).After(t.Truncate(time.Second)) {
					isNotModified = true
				}
			}
		}

		if isNotModified {
			c.Status(http.StatusNotModified)
			c.Abort()
			return
		}

		c.Next()
	}
}
