package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"guangjiapps/gin/internal/api/middleware"
	"guangjiapps/gin/internal/config"
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

func parseUserID(val interface{}) int32 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int32(v)
	case float32:
		return int32(v)
	case int32:
		return v
	case int:
		return int32(v)
	case int64:
		return int32(v)
	case string:
		i, _ := strconv.Atoi(v)
		return int32(i)
	default:
		return 0
	}
}

func NewRouter(authService *service.AuthService, db *gorm.DB, cfg config.Config) *Router {
	adminService := service.NewAdminService(db)
	adminGroupService := service.NewAdminGroupService(db)
	groupMenuService := service.NewGroupMenuService(db)
	adminSubWarehouseService := service.NewAdminSubWarehouseService(db)
	umatService := service.NewUmatService(db)
	topicService := service.NewTopicService(db)
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

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	r.GET(cfg.BaseURL+"/ping", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"message": "pong", "baseUrl": cfg.BaseURL})
	})

	r.NoRoute(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusNotFound, gin.H{
			"error":                "Rute tidak cocok di Gin",
			"url_yang_diterima_go": c.Request.URL.Path,
		})
	})

	r.POST(cfg.BaseURL+"/login", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token, "user": user, "main_menu": mainMenus})
	})

	changePasswordHandler := func(c *gin.Context) {
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
			userID = parseUserID(userIDVal)
		}

		err := authService.ChangePassword(userID, req.Username, req.OldPassword, req.NewPassword)
		if err != nil {
			respondValidationError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
	}

	r.POST(cfg.BaseURL+"/change-password", middleware.AuthMiddleware(cfg), changePasswordHandler)

	protected := r.Group(cfg.BaseURL + "/v1")
	protected.Use(middleware.AuthMiddleware(cfg))

	protected.POST("/change-password", changePasswordHandler)

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
		var payload domain.Admin
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := adminService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin"})
	})
	protected.DELETE("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := adminService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin Group"})
	})
	protected.PATCH("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.AdminGroup
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := adminGroupService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin Group"})
	})
	protected.DELETE("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := adminGroupService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Group Menu Mapping"})
	})
	protected.PATCH("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.GroupMenuMapping
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := groupMenuService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Group Menu Mapping"})
	})
	protected.DELETE("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := groupMenuService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin Sub Warehouse"})
	})
	protected.PATCH("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.AdminSubWarehouse
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := adminSubWarehouseService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin Sub Warehouse"})
	})
	protected.DELETE("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := adminSubWarehouseService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Umats
	protected.GET("/umats", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := umatService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Umat"})
	})
	protected.GET("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := umatService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Umat"})
	})
	protected.POST("/umats", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Umat
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := umatService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Umat"})
	})
	protected.PATCH("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Umat
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := umatService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Umat"})
	})
	protected.DELETE("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := umatService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Topics
	protected.GET("/topics", func(c *gin.Context) {
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
	protected.GET("/topic/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := topicService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Topic"})
	})
	protected.POST("/topics", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Topic
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := topicService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Topic"})
	})
	protected.PATCH("/topics/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Topic
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := topicService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Topic"})
	})
	protected.DELETE("/topics/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := topicService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Activities
	protected.GET("/activities", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := activityService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Activity"})
	})
	protected.GET("/activity/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := activityService.Get(c.Param("id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Activity"})
	})
	protected.POST("/activities", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Activity
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := activityService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Activity"})
	})
	protected.PATCH("/activities/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Activity
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := activityService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Activity"})
	})
	protected.DELETE("/activities/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := activityService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Tim Kerja
	protected.GET("/tim-kerja", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := timKerjaService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Tim Kerja"})
	})
	protected.GET("/tim-kerja/lookup", func(c *gin.Context) {
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
		idInt, _ := strconv.Atoi(c.Param("id"))
		items, total, err := timKerjaService.LookupSub(idInt, filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "TimKerjaLookupSub"})
	})
	protected.GET("/tim-kerja/lookup-report", func(c *gin.Context) {
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
			respondValidationError(c, err)
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
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Tim Kerja"})
	})
	protected.DELETE("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := timKerjaService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Tahun Ciu Tao
	protected.GET("/tahun-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := tahunCiuTaoService.List(page, filters, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "Tahun Ciu Tao"})
	})
	protected.GET("/tahun-ciu-tao/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := tahunCiuTaoService.Get(c.Param("id"))
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Tahun Ciu Tao"})
	})
	protected.PATCH("/tahun-ciu-tao/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.TahunCiuTao
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := tahunCiuTaoService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Tahun Ciu Tao"})
	})
	protected.DELETE("/tahun-ciu-tao/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := tahunCiuTaoService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Penggalang Dana
	protected.GET("/penggalang-dana", func(c *gin.Context) {
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Penggalang Dana"})
	})
	protected.PATCH("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.PenggalangDana
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := penggalangDanaService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Penggalang Dana"})
	})
	protected.DELETE("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := penggalangDanaService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Sxy Donatur
	protected.GET("/sxy-donatur", func(c *gin.Context) {
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Sxy Donatur"})
	})
	protected.PATCH("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.SxyDonatur
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := sxyDonaturService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Sxy Donatur"})
	})
	protected.DELETE("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := sxyDonaturService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/fotang/lookup", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := fotangService.Lookup(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "FotangLookup"})
	})

	// Lookup Endpoints
	protected.GET("/lookup/waktu-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupWaktuCiuTao(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupWaktuCiuTao"})
	})
	protected.GET("/lookup/gender", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupGender(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupGender"})
	})
	protected.GET("/lookup/tcs", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupTcs(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupTcs"})
	})
	protected.GET("/lookup/kelas-level", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKelasLevel(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupTcs"})
	})
	protected.GET("/lookup/fotang", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupFotang(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupFotang"})
	})
	protected.GET("/lookup/kelas", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKelas(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKelas"})
	})
	protected.GET("/lookup/pendidikan", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupPendidikan(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupPendidikan"})
	})
	protected.GET("/lookup/kelas-umum", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKelasUmum(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKelasUmum"})
	})
	protected.GET("/lookup/pekerjaan", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupPekerjaan(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupPekerjaan"})
	})
	protected.GET("/lookup/keluarga", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupKeluarga(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupKeluarga"})
	})
	protected.GET("/lookup/status", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := lookupService.LookupStatus(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupStatus"})
	})
	protected.GET("/lookup/category/:category", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		categoryID := c.Param("category")
		items, total, err := lookupService.Lookup(categoryID, page, limit, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "LookupCategory"})
	})

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
	protected.GET("/kelas/lookup", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, filters := parsePaginationAndFilters(c)
		items, total, err := kelasService.Lookup(filters, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasLookup"})
	})
	protected.GET("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Kelas"})
	})
	protected.PATCH("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Kelas
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Kelas"})
	})
	protected.DELETE("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
	protected.GET("/kelas/:id/peserta", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasPesertaService.List(c.Param("id"), c, page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasPeserta"})
	})
	protected.GET("/kelas/:id/peserta/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := kelasPesertaService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPeserta"})
	})
	protected.POST("/kelas/:id/peserta", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasPeserta
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		payload.TrxId = service.ToInt32(c.Param("id"))
		created, err := kelasPesertaService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPeserta"})
	})
	protected.PATCH("/kelas/:id/peserta/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasPeserta
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasPesertaService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasPeserta"})
	})
	protected.DELETE("/kelas/:id/peserta/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasPesertaService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Pengabdi
	protected.GET("/kelas/:id/pengabdi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasPengabdiService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasPengabdi"})
	})
	protected.GET("/kelas/:id/pengabdi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := kelasPengabdiService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPengabdi"})
	})
	protected.POST("/kelas/:id/pengabdi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasPengabdi
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasPengabdiService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPengabdi"})
	})
	protected.PATCH("/kelas/:id/pengabdi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasPengabdi
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasPengabdiService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasPengabdi"})
	})
	protected.DELETE("/kelas/:id/pengabdi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasPengabdiService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Topik
	protected.GET("/kelas/:id/topik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
		item, err := kelasTopikService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasTopik"})
	})
	protected.POST("/kelas/:id/topik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasTopik
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasTopikService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasTopik"})
	})
	protected.PATCH("/kelas/:id/topik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasTopik
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasTopikService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasTopik"})
	})
	protected.DELETE("/kelas/:id/topik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasTopikService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Kendaraan
	protected.GET("/kelas/:id/kendaraan", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		page, limit, _ := parsePaginationAndFilters(c)
		items, total, err := kelasKendaraanService.List(c.Param("id"), page, limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "meta": gin.H{"page": page, "limit": limit, "total": total}, "resource": "KelasKendaraan"})
	})
	protected.GET("/kelas/:id/kendaraan/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := kelasKendaraanService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasKendaraan"})
	})
	protected.POST("/kelas/:id/kendaraan", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasKendaraan
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasKendaraanService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasKendaraan"})
	})
	protected.PATCH("/kelas/:id/kendaraan/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasKendaraan
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasKendaraanService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasKendaraan"})
	})
	protected.DELETE("/kelas/:id/kendaraan/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasKendaraanService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Donasi
	protected.GET("/kelas/:id/donasi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
		item, err := kelasDonasiService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasDonasi"})
	})
	protected.POST("/kelas/:id/donasi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasDonasi
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasDonasiService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasDonasi"})
	})
	protected.PATCH("/kelas/:id/donasi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasDonasi
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasDonasiService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasDonasi"})
	})
	protected.DELETE("/kelas/:id/donasi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasDonasiService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Donasi Barang
	protected.GET("/kelas/:id/donasi-barang", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
		item, err := kelasDonasiBarangService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasDonasiBarang"})
	})
	protected.POST("/kelas/:id/donasi-barang", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasDonasiBarang
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasDonasiBarangService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasDonasiBarang"})
	})
	protected.PATCH("/kelas/:id/donasi-barang/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasDonasiBarang
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasDonasiBarangService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasDonasiBarang"})
	})
	protected.DELETE("/kelas/:id/donasi-barang/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasDonasiBarangService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Pengeluaran
	protected.GET("/kelas/:id/pengeluaran", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
		item, err := kelasPengeluaranService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasPengeluaran"})
	})
	protected.POST("/kelas/:id/pengeluaran", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasPengeluaran
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasPengeluaranService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasPengeluaran"})
	})
	protected.PATCH("/kelas/:id/pengeluaran/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasPengeluaran
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasPengeluaranService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasPengeluaran"})
	})
	protected.DELETE("/kelas/:id/pengeluaran/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasPengeluaranService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Musik
	protected.GET("/kelas/:id/musik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
		item, err := kelasMusikService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasMusik"})
	})
	protected.POST("/kelas/:id/musik", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasMusik
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasMusikService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasMusik"})
	})
	protected.PATCH("/kelas/:id/musik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasMusik
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasMusikService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasMusik"})
	})
	protected.DELETE("/kelas/:id/musik/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasMusikService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Kelas Absensi
	protected.GET("/kelas/:id/absensi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
		item, err := kelasAbsensiService.Get(c.Param("detail_id"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "KelasAbsensi"})
	})
	protected.POST("/kelas/:id/absensi", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasAbsensi
		payload.TrxId = service.ToInt32(c.Param("id"))
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		created, err := kelasAbsensiService.Create(payload, c)
		if err != nil {
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "KelasAbsensi"})
	})
	protected.PATCH("/kelas/:id/absensi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.KelasAbsensi
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := kelasAbsensiService.Update(c.Param("detail_id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "KelasAbsensi"})
	})
	protected.DELETE("/kelas/:id/absensi/:detail_id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasAbsensiService.Delete(c.Param("detail_id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
	protected.GET("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
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
			respondValidationError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Donasi Sxy"})
	})
	protected.PATCH("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.DonasiSxy
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondValidationError(c, err)
			return
		}
		updated, err := donasiSxyService.Update(c.Param("id"), payload, c)
		if err != nil {
			if strings.Contains(err.Error(), "validasi gagal") {
				respondValidationError(c, err)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Donasi Sxy"})
	})
	protected.DELETE("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := donasiSxyService.Delete(c.Param("id"), c); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	return &Router{r}
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

	errorsMap := make(map[string][]string)
	errStr := strings.TrimPrefix(err.Error(), "validasi gagal: ")
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

func respondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Validasi gagal",
		"details": FormatValidationError(err),
	})
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var vErr *service.ValidationError
	if errors.As(err, &vErr) {
		respondValidationError(c, err)
		return
	}
	errStr := err.Error()
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "tidak ditemukan") {
		c.JSON(http.StatusNotFound, gin.H{"error": errStr})
		return
	}
	if strings.Contains(errStr, "invalid ID format") || strings.Contains(errStr, "is required") || strings.Contains(errStr, "validasi gagal") || strings.Contains(errStr, "not found in lookup") {
		respondValidationError(c, err)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": errStr})
}
