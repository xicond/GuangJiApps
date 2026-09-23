package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"guangjiapps/gin/internal/api/middleware"
	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
	"guangjiapps/gin/internal/service"
)

type Router struct {
	*gin.Engine
}

func parsePaginationAndFilters(c *gin.Context) (int, int, map[string]string) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 0 {
		limit = 10
	}

	filters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if key == "page" || key == "limit" {
			continue
		}
		if len(values) > 0 && values[0] != "" {
			filters[key] = values[0]
		}
	}

	return page, limit, filters
}

func FilterSuccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Gunakan defer untuk memastikan log tereksekusi meskipun terjadi panic / fatal error
		defer func() {
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			latency := time.Since(start)
			clientIP := c.ClientIP()
			method := c.Request.Method
			referer := c.Request.Referer()

			if rawQuery != "" {
				path = path + "?" + rawQuery
			}

			refStr := ""
			if referer != "" {
				refStr = " | Ref: " + referer
			}

			// 1. Tangkap jika terjadi Panic / Fatal Error
			if err := recover(); err != nil {
				log.Printf("[PANIC] %s | 500 | %13v | %15s | %-7s %s%s | Err: %v\n",
					timestamp, latency, clientIP, method, path, refStr, err,
				)
				// Cetak stack trace lengkap ke terminal/error output
				debug.PrintStack()

				// Kirim respon 500 ke client agar tidak menggantung
				c.AbortWithStatus(500)
				return
			}

			// 2. Tangkap HTTP Error biasa (Status >= 400)
			status := c.Writer.Status()
			if status >= 400 {
				log.Printf("[HTTP] %s | %3d | %13v | %15s | %-7s %s%s\n",
					timestamp, status, latency, clientIP, method, path, refStr,
				)
			}
		}()

		c.Next()
	}
}

func ForwardedHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
			c.Request.URL.Scheme = proto
		} /*  else if scheme := c.GetHeader("X-Forwarded-Scheme"); scheme != "" {
			c.Request.URL.Scheme = scheme
		} else if strings.EqualFold(c.GetHeader("X-Forwarded-Ssl"), "on") || strings.EqualFold(c.GetHeader("Front-End-Https"), "on") {
			c.Request.URL.Scheme = "https"
		} */
		if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			// Ambil IP pertama sebelum koma
			parts := strings.Split(xff, ",")
			firstIP := strings.TrimSpace(parts[0])

			// Hapus port jika ikut terbawa (contoh: 118.136.171.50:53303)
			if strings.Contains(firstIP, ":") {
				if host, _, err := net.SplitHostPort(firstIP); err == nil {
					firstIP = host
				}
			}

			// Timpa header dengan format IP murni yang valid
			c.Request.Header.Set("X-Forwarded-For", firstIP)
		}
		c.Next()
	}
}

func NewRouter(authService *service.AuthService, db *gorm.DB, cfg config.Config) *Router {
	adminService := service.NewAdminService(db)
	adminGroupService := service.NewAdminGroupService(db)
	groupMenuService := service.NewGroupMenuService(db)
	adminSubWarehouseService := service.NewAdminSubWarehouseService(db)
	umatService := service.NewUmatService(db, cfg)
	topicService := service.NewTopicService(db)
	kelasMasterService := service.NewKelasMasterService(db)
	activityService := service.NewActivityService(db)
	timKerjaService := service.NewTimKerjaService(db)
	tahunCiuTaoService := service.NewTahunCiuTaoService(db)
	penggalangDanaService := service.NewPenggalangDanaService(db)
	sxyDonaturService := service.NewSxyDonaturService(db)
	kelasService := service.NewKelasService(db)
	kelasPesertaService := service.NewKelasPesertaService(db)
	kelasPengabdiService := service.NewKelasPengabdiService(db)
	kelasTopikService := service.NewKelasTopikService(db)
	kelasKendaraanService := service.NewKelasKendaraanService(db)
	kelasDonasiService := service.NewKelasDonasiService(db)
	kelasDonasiBarangService := service.NewKelasDonasiBarangService(db)
	kelasPengeluaranService := service.NewKelasPengeluaranService(db)
	kelasMusikService := service.NewKelasMusikService(db)
	kelasAbsensiService := service.NewKelasAbsensiService(db)
	fotangService := service.NewFotangService(db)
	donasiSxyService := service.NewDonasiSxyService(db)
	lookupService := service.NewLookupService(db)

	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		service.ConfigureValidator(v)
	}

	r := gin.New()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	// r.SetTrustedProxies([]string{"0.0.0.0/0"})
	// r.SetTrustedProxies(nil)
	r.Use(gin.Recovery())
	r.Use(ForwardedHeaderMiddleware())
	if cfg.GinMode != "release" {
		r.Use(gin.Logger())
	} else {
		r.Use(FilterSuccessLogMiddleware())
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "User-Agent", "X-Requested-With", "X-Request-Id"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, User-Agent, X-Requested-With, X-Request-Id")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	fail2ban := middleware.GetFail2Ban()
	r.Use(fail2ban.Middleware())

	r.NoRoute(func(c *gin.Context) {
		candidates := []string{
			"errors/404.htm",
			"./errors/404.htm",
			"dist/errors/404.htm",
			"gin/dist/errors/404.htm",
			"/app/errors/404.htm",
		}
		for _, path := range candidates {
			if data, err := os.ReadFile(path); err == nil {
				c.Data(http.StatusNotFound, "text/html; charset=utf-8", data)
				return
			}
		}

		c.Header("Content-Type", "application/json")
		if cfg.GinMode == "debug" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":                "Rute tidak cocok di Gin",
				"url_yang_diterima_go": c.Request.URL.Path,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Not Found",
			})
		}
	})

	r.GET(cfg.BaseURL+"/ping", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"message": "pong", "baseUrl": cfg.BaseURL})
	})

	if cfg.GinMode == "debug" {
		r.GET(cfg.BaseURL+"/debug-headers", func(c *gin.Context) {
			// Print semua header yang diterima (dengan redaksi untuk data sensitif)
			sensitiveHeaders := map[string]struct{}{
				"authorization":       {},
				"proxy-authorization": {},
				"cookie":              {},
				"set-cookie":          {},
				"x-api-key":           {},
				"x-auth-token":        {},
			}

			sanitizedHeaders := make(map[string][]string, len(c.Request.Header))
			for name, headers := range c.Request.Header {
				lowerName := strings.ToLower(name)
				if _, isSensitive := sensitiveHeaders[lowerName]; isSensitive {
					sanitizedHeaders[name] = []string{"[REDACTED]"}
					fmt.Printf("%v: %v\n", name, "[REDACTED]")
					continue
				}

				sanitizedHeaders[name] = headers
				for _, h := range headers {
					fmt.Printf("%v: %v\n", name, h)
				}
			}
			c.JSON(200, sanitizedHeaders)
		})
	}

	rdb := database.GetRedisClient(cfg)

	loginLimiter := middleware.NewLoginRateLimiter(cfg)
	ocrLimiter := middleware.NewUserRateLimiter(cfg, 1, time.Minute/60*18, "ocr_limiter:")

	r.POST(cfg.BaseURL+"/login", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")

		limCtx, err := loginLimiter.Peek(c)
		if err == nil && limCtx.Reached {
			retryAfter := loginLimiter.SetHeaders(c, limCtx)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts.",
				"retry_after": retryAfter,
			})
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			respondValidationError(c, err)
			return
		}

		user, token, mainMenus, err := authService.Login(req.Username, req.Password)
		if err != nil {
			updatedLimCtx, _ := loginLimiter.Increment(c)
			retryAfter := loginLimiter.SetHeaders(c, updatedLimCtx)

			if updatedLimCtx.Reached {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Too many attempt. Login access locked for 5 mins.",
					"retry_after": retryAfter,
				})
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		loginLimiter.Reset(c)
		limCtxAfterReset, _ := loginLimiter.Peek(c)
		loginLimiter.SetHeaders(c, limCtxAfterReset)

		c.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token, "user": user, "main_menu": mainMenus})
	})

	protected := r.Group(cfg.BaseURL + "/v1")
	protected.Use(middleware.AuthMiddleware(cfg))

	protected.POST("/change-password", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var req struct {
			Username    string `json:"username"`
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			respondValidationError(c, err)
			return
		}

		var userID int32
		if userIDVal, exists := c.Get("userID"); exists {
			userID = service.ToInt32(userIDVal)
		}

		err := authService.ChangePassword(userID, req.Username, req.OldPassword, req.NewPassword)
		if err != nil {
			respondValidationError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
	})

	// Admins
	protected.GET("/admins", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := adminService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Admin"})
	})
	protected.GET("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := adminService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Admin"})
	})
	protected.POST("/admins", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Admin
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := adminService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin"})
	})
	protected.PATCH("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.Admin
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := adminService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin"})
	})
	protected.DELETE("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := adminService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Department
	protected.GET("/department", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		items, err := adminService.ListDepartments()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "resource": "Department"})
	})

	// Admin Groups
	protected.GET("/admin-groups", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := adminGroupService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Admin Group"})
	})
	protected.GET("/admin-group/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := adminGroupService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "AdminGroup"})
	})
	protected.POST("/admin-groups", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.AdminGroup
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := adminGroupService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin Group"})
	})
	protected.PATCH("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.AdminGroup
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := adminGroupService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin Group"})
	})
	protected.DELETE("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := adminGroupService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Group Menu Mappings
	protected.GET("/group-menu-mappings", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := groupMenuService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Group Menu Mapping"})
	})
	protected.GET("/group-menu-mapping/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := groupMenuService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "GroupMenuMapping"})
	})
	protected.POST("/group-menu-mappings", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.GroupMenuMapping
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := groupMenuService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Group Menu Mapping"})
	})
	protected.PATCH("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.GroupMenuMapping
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := groupMenuService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Group Menu Mapping"})
	})
	protected.DELETE("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := groupMenuService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Admin Sub Warehouses
	protected.GET("/admin-sub-warehouses", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := adminSubWarehouseService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Admin Sub Warehouse"})
	})
	protected.GET("/admin-sub-warehouse/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := adminSubWarehouseService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "AdminSubWarehouse"})
	})
	protected.POST("/admin-sub-warehouses", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.AdminSubWarehouse
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := adminSubWarehouseService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin Sub Warehouse"})
	})
	protected.PATCH("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.AdminSubWarehouse
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := adminSubWarehouseService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin Sub Warehouse"})
	})
	protected.DELETE("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := adminSubWarehouseService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Umats
	protected.GET("/umats", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_BUS_UMAT", UpdatedColumn: "moddate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := umatService.List(c, page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Umat"})
	})
	protected.GET("/umats/report", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := umatService.Report(page, filters, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": items,
			"meta": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
			"resource": "UmatReport",
		})
	})
	protected.GET("/umats/popup", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_BUS_UMAT", UpdatedColumn: "moddate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := umatService.PopUp(c, page, filters, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": items,
			"meta": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
			"resource": "UmatPopUp",
		})
	})
	protected.GET("/umats/report/excel", func(c *gin.Context) {
		_, _, filters := parsePaginationAndFilters(c)
		if err := umatService.ReportExcel(filters, c); err != nil {
			respondError(c, err)
			return
		}
	})
	protected.POST("/umats/ocr", ocrLimiter.Middleware(), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		fileHeader, err := c.FormFile("file")
		if err != nil {
			fileHeader, err = c.FormFile("image")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "File gambar wajib diunggah (field 'file' atau 'image')"})
				return
			}
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membuka file gambar"})
			return
		}
		defer file.Close()

		result, err := umatService.Ocr(file, fileHeader, c)
		if err != nil {
			respondError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":     result,
			"resource": "UmatOCR",
		})
	})
	protected.POST("/umats/verify-qr", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var req domain.VerifyQRRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respondValidationError(c, err)
			return
		}
		if err := service.ValidateStruct(req); err != nil {
			respondValidationError(c, err)
			return
		}
		result, err := umatService.VerifyQR(req.QRToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":     result,
			"resource": "Umat",
		})
	})
	protected.GET("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := umatService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Umat"})
	})
	protected.POST("/umats", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Umat
		var (
			created domain.Umat
			err     error
		)
		contentType := c.ContentType()
		if contentType == "application/json" {
			if err := c.ShouldBindJSON(&payload); err != nil {
				respondValidationError(c, err)
				return
			}
			created, err = umatService.Create(payload, nil, c)
		} else if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memproses multipart form"})
				return
			}
			// Ambil data JSON dari form field "data" (jika dikirim)
			rawJSON := c.PostForm("data")
			if rawJSON != "" {
				if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON di dalam form field tidak valid"})
					return
				}
			}
			var fileHeader *multipart.FileHeader
			fileHeader, _ = c.FormFile("foto")
			if fileHeader == nil {
				fileHeader, _ = c.FormFile("file")
			}
			created, err = umatService.Create(payload, fileHeader, c)
		} else {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type tidak didukung"})
			return
		}
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Umat"})
	})
	protected.PATCH("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.Umat
		var (
			updated domain.Umat
			errUpd  error
		)
		contentType := c.ContentType()
		if contentType == "application/json" {
			if err := c.ShouldBindJSON(&payload); err != nil {
				respondValidationError(c, err)
				return
			}
			updated, errUpd = umatService.Update(c.Param("id"), payload, nil, c)
		} else if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memproses multipart form"})
				return
			}
			// Ambil data JSON dari form field "data" (jika dikirim)
			rawJSON := c.PostForm("data")
			if rawJSON != "" {
				if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON di dalam form field tidak valid"})
					return
				}
			}
			var fileHeader *multipart.FileHeader
			fileHeader, _ = c.FormFile("foto")
			if fileHeader == nil {
				fileHeader, _ = c.FormFile("file")
			}
			updated, errUpd = umatService.Update(c.Param("id"), payload, fileHeader, c)
		} else {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type tidak didukung"})
			return
		}
		if errUpd != nil {
			respondError(c, errUpd)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Umat"})
	})
	protected.DELETE("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := umatService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Topics
	protected.GET("/topics", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_BUS_TOPIC", UpdatedColumn: "ModDate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := topicService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Topic"})
	})
	// protected.GET("/topik", topicHandler)
	protected.GET("/topic", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		code := c.Query("code")
		if code == "" {
			code = c.Query("id")
		}
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query param 'code' is required"})
			return
		}
		item, err := topicService.Get(code)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Topic"})
	})
	protected.POST("/topic", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Topic
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid: " + err.Error()})
			return
		}
		created, err := topicService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Topic"})
	})
	protected.PATCH("/topic", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		code := c.Query("code")
		if code == "" {
			code = c.Query("id")
		}
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query param 'code' is required"})
			return
		}
		var payload domain.Topic
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid: " + err.Error()})
			return
		}
		updated, err := topicService.Update(code, payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Topic"})
	})
	protected.DELETE("/topic", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		code := c.Query("code")
		if code == "" {
			code = c.Query("id")
		}
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query param 'code' is required"})
			return
		}
		if err := topicService.Delete(code, c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Master (Master Data Kelas - T_APP_LOOKUP CategoryId = B_KELASKHUSUS)
	kelasMasterMetadata := middleware.ResourceMetadata{
		TableName:   "T_APP_LOOKUP",
		WhereClause: "CategoryId = ?",
		WhereArgs:   []interface{}{"B_KELASKHUSUS"},
		CacheKey:    "lookup:B_KELASKHUSUS",
	}

	kelasMasterListHandler := func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := kelasMasterService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasMaster"})
	}

	protected.GET("/kelas-masters", middleware.StatusNotModifiedHeader(db, rdb, kelasMasterMetadata), kelasMasterListHandler)

	protected.GET("/kelas-master", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		code := c.Query("code")
		if code == "" {
			code = c.Query("id")
		}
		if code == "" {
			code = c.Query("lookup_id")
		}
		if code == "" {
			kelasMasterListHandler(c)
			return
		}
		item, err := kelasMasterService.Get(code)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasMaster"})
	})

	protected.GET("/kelas-masters/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		id := c.Param("id")
		item, err := kelasMasterService.Get(id)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasMaster"})
	})

	protected.POST("/kelas-master", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.AppLookup
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid: " + err.Error()})
			return
		}
		created, err := kelasMasterService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		if rdb != nil {
			rdb.Del(c.Request.Context(), "lookup:B_KELASKHUSUS")
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasMaster"})
	})

	protected.PATCH("/kelas-master", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		code := c.Query("code")
		if code == "" {
			code = c.Query("id")
		}
		if code == "" {
			code = c.Query("lookup_id")
		}
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query param 'id' atau 'code' wajib diisi"})
			return
		}
		var payload domain.AppLookup
		if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid: " + err.Error()})
			return
		}
		updated, err := kelasMasterService.Update(code, payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		if rdb != nil {
			rdb.Del(c.Request.Context(), "lookup:B_KELASKHUSUS")
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasMaster"})
	})

	protected.DELETE("/kelas-master", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		code := c.Query("code")
		if code == "" {
			code = c.Query("id")
		}
		if code == "" {
			code = c.Query("lookup_id")
		}
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query param 'id' atau 'code' wajib diisi"})
			return
		}
		if err := kelasMasterService.Delete(code, c); err != nil {
			respondError(c, err)
			return
		}
		if rdb != nil {
			rdb.Del(c.Request.Context(), "lookup:B_KELASKHUSUS")
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
	protected.GET("/activities", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_BUS_EVENT", UpdatedColumn: "ModDate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := activityService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Activity"})
	})
	protected.GET("/activity", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := activityService.Get(c.Query("code"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Activity"})
	})
	protected.POST("/activity", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Activity
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := activityService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Activity"})
	})
	protected.PATCH("/activity", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Activity
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := activityService.Update(c.Query("code"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Activity"})
	})
	deleteActivityHandler := func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		id := c.Param("id")
		if id == "" {
			id = c.Query("code")
		}
		if id == "" {
			id = c.Query("id")
		}
		if err := activityService.Delete(id, c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
	protected.DELETE("/activity", deleteActivityHandler)
	protected.DELETE("/activity/:id", deleteActivityHandler)

	// Tim Kerja
	protected.GET("/tim-kerja", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_POSISI"}, CacheKey: "lookup:B_POSISI"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := timKerjaService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Tim Kerja"})
	})
	protected.GET("/tim-kerja/lookup", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_POSISI"}, CacheKey: "lookup:B_POSISI"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := timKerjaService.Lookup(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "TimKerjaLookup"})
	})
	protected.GET("/tim-kerja/lookup/:id/sub", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := timKerjaService.LookupSub(c.Param("id"), filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "TimKerjaLookupSub"})
	})
	protected.GET("/tim-kerja/lookup-report", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_POSISI"}, CacheKey: "lookup:B_POSISI_REPORT"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := timKerjaService.LookupReport(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "TimKerjaLookupReport"})
	})
	protected.GET("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := timKerjaService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "TimKerja"})
	})
	protected.POST("/tim-kerja", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.TimKerja
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := timKerjaService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Tim Kerja"})
	})
	protected.PATCH("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.TimKerja
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := timKerjaService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Tim Kerja"})
	})
	protected.DELETE("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := timKerjaService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Tahun Ciu Tao
	protected.GET("/tahun-ciu-tao/list", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_BUS_TAHUN_CIUTAO", UpdatedColumn: "ModDate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := tahunCiuTaoService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Tahun Ciu Tao"})
	})
	protected.GET("/tahun-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		tahun := c.Query("tahun")
		if tahun == "" {
			tahun = c.Query("tahun_mandarin")
		}
		item, err := tahunCiuTaoService.Get(tahun)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "TahunCiuTao"})
	})
	protected.POST("/tahun-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.TahunCiuTao
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := tahunCiuTaoService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Tahun Ciu Tao"})
	})
	protected.PATCH("/tahun-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.TahunCiuTao
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		tahun := c.Query("tahun")
		if tahun == "" {
			tahun = c.Query("tahun_mandarin")
		}
		updated, err := tahunCiuTaoService.Update(tahun, payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Tahun Ciu Tao"})
	})
	deleteTahunCiuTaoHandler := func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		id := c.Param("id")
		if id == "" {
			id = c.Query("tahun")
		}
		if id == "" {
			id = c.Query("tahun_mandarin")
		}
		if err := tahunCiuTaoService.Delete(id, c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
	protected.DELETE("/tahun-ciu-tao/:id", deleteTahunCiuTaoHandler)
	protected.DELETE("/tahun-ciu-tao", deleteTahunCiuTaoHandler)

	// Penggalang Dana
	protected.GET("/penggalang-dana", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_SXY_MST_PENGGALANG", UpdatedColumn: "updateddate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := penggalangDanaService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Penggalang Dana"})
	})
	protected.GET("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := penggalangDanaService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "PenggalangDana"})
	})
	protected.POST("/penggalang-dana", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.PenggalangDana
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := penggalangDanaService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Penggalang Dana"})
	})
	protected.PATCH("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.PenggalangDana
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := penggalangDanaService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Penggalang Dana"})
	})
	protected.DELETE("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := penggalangDanaService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Sxy Donatur
	protected.GET("/sxy-donatur", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_SXY_MST_DONATUR", UpdatedColumn: "updateddate"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := sxyDonaturService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Sxy Donatur"})
	})
	protected.GET("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := sxyDonaturService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "SxyDonatur"})
	})
	protected.POST("/sxy-donatur", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.SxyDonatur
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := sxyDonaturService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Sxy Donatur"})
	})
	protected.PATCH("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.SxyDonatur
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := sxyDonaturService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Sxy Donatur"})
	})
	protected.DELETE("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := sxyDonaturService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/fotang/lookup", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_FOTHANG"}, CacheKey: "lookup:B_FOTHANG"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := fotangService.Lookup(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "FotangLookup"})
	})

	protected.GET("/fotang/lookup-sxy", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"SXY_FOTHANG"}, CacheKey: "lookup:SXY_FOTHANG"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := fotangService.LookupSxy(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "FotangSxyLookup"})
	})

	// Lookup Endpoints
	protected.GET("/lookup/waktu-ciu-tao", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_WAKTUCIUTAO"}, CacheKey: "lookup:B_WAKTUCIUTAO"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupWaktuCiuTao(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupWaktuCiuTao"})
	})
	protected.GET("/lookup/gender", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_GENDER"}, CacheKey: "lookup:B_GENDER"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupGender(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupGender"})
	})
	protected.GET("/lookup/tcs", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_TCS"}, CacheKey: "lookup:B_TCS"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupTcs(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupTcs"})
	})
	protected.GET("/lookup/kelas-level", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KLS_LEVEL"}, CacheKey: "lookup:B_KLS_LEVEL"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKelasLevel(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupTcs"})
	})
	protected.GET("/lookup/fotang", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_FOTHANG"}, CacheKey: "lookup:B_FOTHANG"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupFotang(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupFotang"})
	})
	protected.GET("/lookup/kelas", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KELASKHUSUS"}, CacheKey: "lookup:B_KELASKHUSUS"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKelas(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKelas"})
	})
	protected.GET("/lookup/pendidikan", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_PENDIDIKAN"}, CacheKey: "lookup:B_PENDIDIKAN"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupPendidikan(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupPendidikan"})
	})
	protected.GET("/lookup/kelas-umum", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KELASUMUM"}, CacheKey: "lookup:B_KELASUMUM"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKelasUmum(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKelasUmum"})
	})
	protected.GET("/lookup/pekerjaan", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_PEKERJAAN"}, CacheKey: "lookup:B_PEKERJAAN"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupPekerjaan(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupPekerjaan"})
	})
	protected.GET("/lookup/keluarga", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KELUARGA"}, CacheKey: "lookup:B_KELUARGA"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKeluarga(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKeluarga"})
	})
	protected.GET("/lookup/status", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_STATUS"}, CacheKey: "lookup:B_STATUS"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupStatus(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupStatus"})
	})
	protected.GET("/lookup/kategori-topic", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KATEGORI_TOPIK"}, CacheKey: "lookup:B_KATEGORI_TOPIK"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKategoriTopic(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKategoriTopic"})
	})
	protected.GET("/lookup/kategori-event", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KATEGORI_EVENT"}, CacheKey: "lookup:B_KATEGORI_EVENT"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKategoriEvent(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKategoriEvent"})
	})
	protected.GET("/lookup/tipe-sumbangan", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"SXY_TIPESUMBANGAN"}, CacheKey: "lookup:SXY_TIPESUMBANGAN"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupTipeSumbangan(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupTipeSumbangan"})
	})
	/* protected.GET("/lookup/category/:category", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, filters := parsePaginationAndFilters(c)
		categoryID := c.Param("category")
		items, total, err := lookupService.Lookup(categoryID, page, limit, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupCategory"})
	}) */

	// Kelas
	protected.GET("/kelas", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := kelasService.List(page, filters, c, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Kelas"})
	})
	protected.GET("/kelas/lookup", middleware.StatusNotModifiedHeader(db, rdb, middleware.ResourceMetadata{TableName: "T_APP_LOOKUP", WhereClause: "CategoryId = ?", WhereArgs: []interface{}{"B_KELASKHUSUS"}, CacheKey: "lookup:B_KELASKHUSUS"}), func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := kelasService.Lookup(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasLookup"})
	})
	/* protected.GET("/kelas/report", func(c *gin.Context) {
		trxId := c.Query("trx_id")
		if trxId == "" {
			trxId = c.Query("TrxId")
		}
		if err := kelasService.Report(trxId /* subWhId,  *-/, c); err != nil {
			respondError(c, err)
			return
		}
	}) */
	protected.GET("/kelas/:id/report", func(c *gin.Context) {
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		trxId := c.Param("id")
		if err := kelasService.Report(trxId /* subWhId,  */, c); err != nil {
			respondError(c, err)
			return
		}
	})
	protected.GET("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Kelas"})
	})
	protected.POST("/kelas", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Kelas
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Kelas"})
	})
	protected.PATCH("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.Kelas
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Kelas"})
	})
	protected.DELETE("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
	protected.GET("/kelas/:id/peserta", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasPesertaService.List(c.Param("id"), c, page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasPeserta"})
	})
	protected.GET("/kelas/:id/peserta/load-previous", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasPesertaService.LoadPrevious(c.Param("id"), c, page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasPesertaPrevious"})
	})
	protected.GET("/kelas/:id/peserta/by-idpeserta/:id_peserta", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("id_peserta"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasPesertaService.GetByIdPerserta(c.Param("id_peserta"), c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPeserta"})
	})
	protected.GET("/kelas/:id/peserta/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasPesertaService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPeserta"})
	})
	protected.POST("/kelas/:id/peserta", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPeserta
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		payload.TrxId = service.ToInt32(c.Param("id"))
		created, err := kelasPesertaService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPeserta"})
	})
	protected.POST("/kelas/:id/peserta/bulk", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPesertaBulkRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		payload.TrxId = service.ToInt32(c.Param("id"))
		created, err := kelasPesertaService.CreateBulk(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPeserta"})
	})
	protected.PATCH("/kelas/:id/peserta/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPeserta
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasPesertaService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasPeserta"})
	})
	protected.DELETE("/kelas/:id/peserta/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasPesertaService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Pengabdi
	protected.GET("/kelas/:id/pengabdi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasPengabdiService.List(c, c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasPengabdi"})
	})
	protected.GET("/kelas/:id/pengabdi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasPengabdiService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPengabdi"})
	})
	protected.POST("/kelas/:id/pengabdi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPengabdi
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasPengabdiService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPengabdi"})
	})
	protected.PATCH("/kelas/:id/pengabdi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPengabdi
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasPengabdiService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasPengabdi"})
	})
	protected.DELETE("/kelas/:id/pengabdi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasPengabdiService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Topik
	protected.GET("/kelas/:id/topik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasTopikService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasTopik"})
	})
	protected.GET("/kelas/:id/topik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasTopikService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasTopik"})
	})
	protected.POST("/kelas/:id/topik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasTopik
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasTopikService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasTopik"})
	})
	protected.PATCH("/kelas/:id/topik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasTopik
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasTopikService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasTopik"})
	})
	protected.DELETE("/kelas/:id/topik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasTopikService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Kendaraan
	protected.GET("/kelas/:id/kendaraan", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasKendaraanService.List(c, c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasKendaraan"})
	})
	protected.GET("/kelas/:id/kendaraan/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasKendaraanService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasKendaraan"})
	})
	protected.POST("/kelas/:id/kendaraan", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasKendaraan
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasKendaraanService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasKendaraan"})
	})
	protected.PATCH("/kelas/:id/kendaraan/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasKendaraan
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasKendaraanService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasKendaraan"})
	})
	protected.DELETE("/kelas/:id/kendaraan/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasKendaraanService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Donasi
	protected.GET("/kelas/:id/donasi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasDonasiService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasDonasi"})
	})
	protected.GET("/kelas/:id/donasi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasDonasiService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasDonasi"})
	})
	protected.POST("/kelas/:id/donasi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasDonasi
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasDonasiService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasDonasi"})
	})
	protected.PATCH("/kelas/:id/donasi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasDonasi
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasDonasiService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasDonasi"})
	})
	protected.DELETE("/kelas/:id/donasi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasDonasiService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Donasi Barang
	protected.GET("/kelas/:id/donasi-barang", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasDonasiBarangService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasDonasiBarang"})
	})
	protected.GET("/kelas/:id/donasi-barang/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasDonasiBarangService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasDonasiBarang"})
	})
	protected.POST("/kelas/:id/donasi-barang", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasDonasiBarang
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasDonasiBarangService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasDonasiBarang"})
	})
	protected.PATCH("/kelas/:id/donasi-barang/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasDonasiBarang
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasDonasiBarangService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasDonasiBarang"})
	})
	protected.DELETE("/kelas/:id/donasi-barang/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasDonasiBarangService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Pengeluaran
	protected.GET("/kelas/:id/pengeluaran", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasPengeluaranService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasPengeluaran"})
	})
	protected.GET("/kelas/:id/pengeluaran/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasPengeluaranService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPengeluaran"})
	})
	protected.POST("/kelas/:id/pengeluaran", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPengeluaran
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasPengeluaranService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPengeluaran"})
	})
	protected.PATCH("/kelas/:id/pengeluaran/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasPengeluaran
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasPengeluaranService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasPengeluaran"})
	})
	protected.DELETE("/kelas/:id/pengeluaran/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasPengeluaranService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Musik
	protected.GET("/kelas/:id/musik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasMusikService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasMusik"})
	})
	protected.GET("/kelas/:id/musik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasMusikService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasMusik"})
	})
	protected.POST("/kelas/:id/musik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasMusik
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasMusikService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasMusik"})
	})
	protected.PATCH("/kelas/:id/musik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasMusik
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasMusikService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasMusik"})
	})
	protected.DELETE("/kelas/:id/musik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasMusikService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Absensi
	protected.GET("/kelas/:id/absensi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasAbsensiService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasAbsensi"})
	})
	protected.GET("/kelas/:id/absensi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := kelasAbsensiService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasAbsensi"})
	})
	protected.POST("/kelas/:id/absensi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasAbsensi
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasAbsensiService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasAbsensi"})
	})
	protected.PATCH("/kelas/:id/absensi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.KelasAbsensi
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasAbsensiService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasAbsensi"})
	})
	protected.DELETE("/kelas/:id/absensi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		_, err = strconv.ParseUint(c.Param("detail_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := kelasAbsensiService.Delete(c.Param("detail_id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Donasi Sxy
	protected.GET("/donasi-sxy", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := donasiSxyService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Donasi Sxy"})
	})
	protected.GET("/donasi-sxy/report", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, totalJumlah, total, err := donasiSxyService.Report(page, filters, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": items,
			"meta": gin.H{
				"page":         page,
				"limit":        limit,
				"total":        total,
				"total_jumlah": totalJumlah,
			},
			"resource": "SxyDonasiReport",
		})
	})
	protected.GET("/donasi-sxy/report/excel", func(c *gin.Context) {
		_, _, filters := parsePaginationAndFilters(c)
		if err := donasiSxyService.ReportExcel(filters, c); err != nil {
			respondError(c, err)
			return
		}
	})
	protected.GET("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		item, err := donasiSxyService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "DonasiSxy"})
	})
	protected.POST("/donasi-sxy", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.DonasiSxy
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := donasiSxyService.Create(payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Donasi Sxy"})
	})
	protected.PATCH("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		var payload domain.DonasiSxy
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := donasiSxyService.Update(c.Param("id"), payload, c)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Donasi Sxy"})
	})
	protected.DELETE("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Url"})
			return
		}
		if err := donasiSxyService.Delete(c.Param("id"), c); err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	return &Router{r}
}

// ParseJSONError parses low-level JSON unmarshaling, syntax, and payload errors into user-friendly messages.
func ParseJSONError(err error) map[string]string {
	errorsMap := make(map[string]string)
	if err == nil {
		return errorsMap
	}

	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError

	switch {
	case errors.As(err, &syntaxError) || errors.Is(err, io.ErrUnexpectedEOF):
		errorsMap["general"] = "Format payload JSON tidak valid atau rusak."

	case errors.Is(err, io.EOF) || err.Error() == "EOF":
		errorsMap["general"] = "Payload request tidak boleh kosong."

	// Deteksi spesifik berdasarkan UnmarshalTypeError
	case errors.As(err, &typeError):
		field := typeError.Field
		if field == "" {
			field = "general"
		}

		// Petakan pesan error spesifik berdasarkan nama field-nya
		switch field {
		case "idpeserta", "id_peserta":
			errMsg := err.Error()
			if strings.Contains(errMsg, "cannot unmarshal array") {
				errorsMap[field] = "idpeserta should be single value of umat"
			} else if strings.Contains(errMsg, "cannot unmarshal") {
				errorsMap[field] = "idpeserta must array"
			} else {
				errorsMap[field] = errMsg
			}

		default:
			// Fallback jika ada field lain yang tipe datanya salah
			errorsMap[field] = fmt.Sprintf("Field '%s' memiliki tipe data yang tidak sesuai.", field)
		}

	default:
		errStr := strings.TrimPrefix(err.Error(), "Validation failed: ")
		if idx := strings.Index(errStr, ": "); idx != -1 {
			f := errStr[:idx]
			m := errStr[idx+2:]
			errorsMap[f] = m
		} else {
			errorsMap["general"] = "Terjadi kesalahan pada struktur data."
		}
	}

	return errorsMap
}

func FormatValidationError(err error) map[string][]string {
	if err == nil {
		return make(map[string][]string)
	}

	var vErr *service.ValidationError
	if errors.As(err, &vErr) && vErr != nil && len(vErr.Details) > 0 {
		return vErr.Details
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return service.FormatValidatorErrors(validationErrs)
	}

	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &syntaxError) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) || err.Error() == "EOF" || errors.As(err, &typeError) {
		jsonErrors := ParseJSONError(err)
		errorsMap := make(map[string][]string, len(jsonErrors))
		for k, v := range jsonErrors {
			errorsMap[k] = []string{v}
		}
		return errorsMap
	}

	errorsMap := make(map[string][]string)
	errStr := strings.TrimPrefix(err.Error(), "Validation failed: ")
	parts := strings.Split(errStr, ", ")
	for _, part := range parts {
		if idx := strings.Index(part, ": "); idx != -1 {
			f := part[:idx]
			m := part[idx+2:]
			errorsMap[f] = append(errorsMap[f], m)
		} else {
			errorsMap["general"] = append(errorsMap["general"], part)
		}
	}
	return errorsMap
}

func getPathAndMethod(c *gin.Context) (string, string) {
	if c == nil || c.Request == nil {
		return "", ""
	}
	path := ""
	if c.Request.URL != nil {
		path = c.Request.URL.Path
	}
	return path, c.Request.Method
}

func respondValidationError(c *gin.Context, err error) {
	// path, method := getPathAndMethod(c)
	// log.Printf("[API VALIDATION ERROR 400] path=%s method=%s err=%v", path, method, err)
	details := FormatValidationError(err)
	errMsg := "Validation failed"
	if err != nil && err.Error() != "" {
		errMsg = err.Error()
	}

	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &syntaxError) || errors.Is(err, io.ErrUnexpectedEOF) {
		errMsg = "Format payload JSON tidak valid atau rusak."
	} else if errors.Is(err, io.EOF) || err.Error() == "EOF" {
		errMsg = "Payload request tidak boleh kosong."
	} else if errors.As(err, &typeError) {
		field := typeError.Field
		if field == "" {
			field = "general"
		}
		if msgs, ok := details[field]; ok && len(msgs) > 0 {
			errMsg = fmt.Sprintf("%s: %s", field, msgs[0])
		} else {
			errMsg = fmt.Sprintf("Field '%s' memiliki tipe data yang tidak sesuai.", field)
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error":   errMsg,
		"details": details,
	})
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	path, method := getPathAndMethod(c)
	// log.Printf("[API ERROR HANDLER] path=%s method=%s type=%T err=%+v", path, method, err, err)
	var vErr *service.ValidationError
	if errors.As(err, &vErr) {
		// log.Printf("[API ERROR -> 400 VALIDATION] matched *service.ValidationError: %+v", vErr)
		respondValidationError(c, err)
		return
	}
	errStr := err.Error()
	if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(errStr, "not found") || strings.Contains(errStr, "tidak ditemukan") {
		log.Printf("[API NOT FOUND 404] path=%s method=%s err=%s", path, method, errStr)
		c.JSON(http.StatusNotFound, gin.H{"error": errStr})
		return
	}
	if strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "is required") ||
		strings.Contains(errStr, "Validation failed") ||
		strings.Contains(errStr, "not found in lookup") ||
		strings.Contains(errStr, "tidak valid") ||
		strings.Contains(errStr, "failed to update") ||
		strings.Contains(errStr, "failed to create") ||
		strings.Contains(errStr, "failed to delete") ||
		strings.Contains(errStr, "mssql:") ||
		strings.Contains(errStr, "conflicted") ||
		strings.Contains(errStr, "truncated") ||
		strings.Contains(errStr, "duplicate") {
		// log.Printf("[API ERROR -> 400 BAD REQUEST] matched error string pattern: %s", errStr)
		respondValidationError(c, err)
		return
	}
	log.Printf("[API 500 INTERNAL SERVER ERROR] path=%s method=%s type=%T err=%s", path, method, err, errStr)
	c.JSON(http.StatusInternalServerError, gin.H{"error": errStr})
}
