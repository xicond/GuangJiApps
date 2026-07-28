package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	if limit < 1 {
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
	fotangService := service.NewFotangService(db)
	donasiSxyService := service.NewDonasiSxyService(db)

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

func respondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Validasi gagal",
		"details": err.Error(),
	})
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	errStr := err.Error()
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "tidak ditemukan") {
		c.JSON(http.StatusNotFound, gin.H{"error": errStr})
		return
	}
	if strings.Contains(errStr, "invalid ID format") || strings.Contains(errStr, "is required") {
		c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": errStr})
}
