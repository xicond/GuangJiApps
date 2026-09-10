package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestItem struct {
	ID      uint `gorm:"primaryKey"`
	Name    string
	ModDate time.Time `gorm:"column:ModDate"`
}

func (TestItem) TableName() string { return "test_items" }

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file:memdb_snm?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	// if err := db.AutoMigrate(&TestItem{}); err != nil {
	// 	t.Fatalf("failed to automigrate: %v", err)
	// }
	return db
}

func TestStatusNotModifiedHeader_WithoutRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	now := time.Now().Truncate(time.Second)
	db.Create(&TestItem{ID: 1, Name: "Item 1", ModDate: now})

	meta := ResourceMetadata{
		TableName:     "test_items",
		UpdatedColumn: "ModDate",
	}

	r := gin.New()
	r.Use(StatusNotModifiedHeader(db, nil, meta))
	r.GET("/items", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	// 1. Initial request: Expect 200 OK + ETag + Last-Modified + Cache-Control
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/items", nil)
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w1.Code)
	}
	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header, got empty")
	}
	// cacheControl := w1.Header().Get("Cache-Control")
	// if cacheControl != "public, max-age=604800, must-revalidate" {
	// 	t.Fatalf("expected Cache-Control header, got %s", cacheControl)
	// }
	lastModified := w1.Header().Get("Last-Modified")
	if lastModified == "" {
		t.Fatalf("expected Last-Modified header, got empty")
	}

	// 2. Request with If-None-Match matching ETag: Expect 304 StatusNotModified
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/items", nil)
	req2.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotModified {
		t.Fatalf("expected 304 StatusNotModified, got %d", w2.Code)
	}

	// 3. Request with If-Modified-Since: Expect 304 StatusNotModified
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/items", nil)
	req3.Header.Set("If-Modified-Since", lastModified)
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusNotModified {
		t.Fatalf("expected 304 StatusNotModified for If-Modified-Since, got %d", w3.Code)
	}
}

func TestStatusNotModifiedHeader_WhereClause(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	meta := ResourceMetadata{
		TableName:     "test_items",
		UpdatedColumn: "ModDate",
		WhereClause:   "Name = ?",
		WhereArgs:     []interface{}{"NonExistent"},
	}

	r := gin.New()
	r.Use(StatusNotModifiedHeader(db, nil, meta))
	r.GET("/items", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/items", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if etag := w.Header().Get("ETag"); etag != `"none"` {
		t.Fatalf("expected ETag \"none\" for empty table result, got %s", etag)
	}
}
