package api

import (
	"net/http"

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
	donasiSxyService := service.NewDonasiSxyService(db)
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET(cfg.BaseURL+"/ping", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"message": "pong", "baseUrl": cfg.BaseURL})
	})

	// Fitur Tambahan: Tangkap rute salah apa pun untuk ditampilkan di log browser
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		user, token, err := authService.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token, "user": user})
	})

	r.POST(cfg.BaseURL+"/forgot-password", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"message": "password reset instructions sent"})
	})

	protected := r.Group(cfg.BaseURL + "/v1")
	protected.Use(middleware.AuthMiddleware(cfg))

	protected.GET("/profile", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		userID, _ := c.Get("userID")
		user, err := authService.Profile(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	r.POST(cfg.BaseURL+"/reset-password", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"message": "password reset complete"})
	})

	protected.GET("/admins", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": adminService.List(), "resource": "Admin"})
	})
	protected.POST("/admins", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := adminService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin"})
	})
	protected.POST("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := adminService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Admin"})
	})
	protected.PATCH("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := adminService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin"})
	})
	protected.DELETE("/admins/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := adminService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/admin-groups", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": adminGroupService.List(), "resource": "Admin Group"})
	})
	protected.POST("/admin-groups", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := adminGroupService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin Group"})
	})
	protected.POST("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := adminGroupService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Admin Group"})
	})
	protected.PATCH("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := adminGroupService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin Group"})
	})
	protected.DELETE("/admin-groups/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := adminGroupService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/group-menu-mappings", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": groupMenuService.List(), "resource": "Group Menu Mapping"})
	})
	protected.POST("/group-menu-mappings", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := groupMenuService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Group Menu Mapping"})
	})
	protected.POST("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := groupMenuService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Group Menu Mapping"})
	})
	protected.PATCH("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := groupMenuService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Group Menu Mapping"})
	})
	protected.DELETE("/group-menu-mappings/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := groupMenuService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/admin-sub-warehouses", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": adminSubWarehouseService.List(), "resource": "Admin Sub Warehouse"})
	})
	protected.POST("/admin-sub-warehouses", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := adminSubWarehouseService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Admin Sub Warehouse"})
	})
	protected.POST("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := adminSubWarehouseService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Admin Sub Warehouse"})
	})
	protected.PATCH("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := adminSubWarehouseService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Admin Sub Warehouse"})
	})
	protected.DELETE("/admin-sub-warehouses/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := adminSubWarehouseService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/umats", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": umatService.List(), "resource": "Umat"})
	})
	protected.POST("/umats", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := umatService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Umat"})
	})
	protected.POST("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := umatService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Umat"})
	})
	protected.PATCH("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := umatService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Umat"})
	})
	protected.DELETE("/umats/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := umatService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/topics", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": topicService.List(), "resource": "Topic"})
	})
	protected.POST("/topics", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := topicService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Topic"})
	})
	protected.POST("/topics/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := topicService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Topic"})
	})
	protected.PATCH("/topics/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := topicService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Topic"})
	})
	protected.DELETE("/topics/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := topicService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/activities", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": activityService.List(), "resource": "Activity"})
	})
	protected.POST("/activities", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := activityService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Activity"})
	})
	protected.POST("/activities/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := activityService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Activity"})
	})
	protected.PATCH("/activities/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := activityService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Activity"})
	})
	protected.DELETE("/activities/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := activityService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/tim-kerja", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": timKerjaService.List(), "resource": "Tim Kerja"})
	})
	protected.POST("/tim-kerja", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := timKerjaService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Tim Kerja"})
	})
	protected.POST("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := timKerjaService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Tim Kerja"})
	})
	protected.PATCH("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := timKerjaService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Tim Kerja"})
	})
	protected.DELETE("/tim-kerja/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := timKerjaService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/tahun-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": tahunCiuTaoService.List(), "resource": "Tahun Ciu Tao"})
	})
	protected.POST("/tahun-ciu-tao", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := tahunCiuTaoService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Tahun Ciu Tao"})
	})
	protected.POST("/tahun-ciu-tao/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := tahunCiuTaoService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Tahun Ciu Tao"})
	})
	protected.PATCH("/tahun-ciu-tao/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := tahunCiuTaoService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Tahun Ciu Tao"})
	})
	protected.DELETE("/tahun-ciu-tao/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := tahunCiuTaoService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/penggalang-dana", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": penggalangDanaService.List(), "resource": "Penggalang Dana"})
	})
	protected.POST("/penggalang-dana", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := penggalangDanaService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Penggalang Dana"})
	})
	protected.POST("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := penggalangDanaService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Penggalang Dana"})
	})
	protected.PATCH("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := penggalangDanaService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Penggalang Dana"})
	})
	protected.DELETE("/penggalang-dana/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := penggalangDanaService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/sxy-donatur", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": sxyDonaturService.List(), "resource": "Sxy Donatur"})
	})
	protected.POST("/sxy-donatur", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := sxyDonaturService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Sxy Donatur"})
	})
	protected.POST("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := sxyDonaturService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Sxy Donatur"})
	})
	protected.PATCH("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := sxyDonaturService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Sxy Donatur"})
	})
	protected.DELETE("/sxy-donatur/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := sxyDonaturService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/kelas", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": kelasService.List(), "resource": "Kelas"})
	})
	protected.POST("/kelas", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := kelasService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Kelas"})
	})
	protected.POST("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := kelasService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Kelas"})
	})
	protected.PATCH("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := kelasService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Kelas"})
	})
	protected.DELETE("/kelas/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := kelasService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	protected.GET("/donasi-sxy", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"data": donasiSxyService.List(), "resource": "Donasi Sxy"})
	})
	protected.POST("/donasi-sxy", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		created, err := donasiSxyService.Create(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created, "resource": "Donasi Sxy"})
	})
	protected.POST("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		item, err := donasiSxyService.Get(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item, "resource": "Donasi Sxy"})
	})
	protected.PATCH("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		var payload domain.Resource
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updated, err := donasiSxyService.Update(c.Param("id"), payload)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated, "resource": "Donasi Sxy"})
	})
	protected.DELETE("/donasi-sxy/:id", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		if err := donasiSxyService.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
	return &Router{r}
}
