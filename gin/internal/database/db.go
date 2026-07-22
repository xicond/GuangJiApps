package database

import (
	"strings"

	"guangjiapps/gin/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	if strings.TrimSpace(dsn) == "" {
		return gorm.Open(sqlite.Open("file::memory:?cache=shared"), cfg)
	}
	return gorm.Open(sqlserver.Open(dsn), cfg)
}

func MustOpen(dsn string) *gorm.DB {
	db, err := Open(dsn)
	if err != nil {
		panic(err)
	}
	return db
}

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	models := []interface{}{
		&domain.Resource{},
		&domain.LoginGroup{},
		&domain.LoginMenu{},
		&domain.LoginMenuGroup{},
		&domain.LoginMst{},
		&domain.SubWarehouse{},
		&domain.Umat{},
		&domain.Topic{},
		&domain.Activity{},
		&domain.TimKerja{},
		&domain.TahunCiuTao{},
		&domain.PenggalangDana{},
		&domain.SxyDonatur{},
		&domain.Kelas{},
		&domain.DonasiSxy{},
	}
	return db.AutoMigrate(models...)
}
