package database

import (
	"time"

	"guangjiapps/gin/internal/domain"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string) (*gorm.DB, error) {
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	// 1. Open the GORM connection with silent logging to save I/O overhead
	db, err := gorm.Open(sqlserver.Open(dsn), cfg)
	if err != nil {
		return nil, err
	}

	// 2. CONFIGURE CONNECTION POOLING HERE
	sqlDB, err := db.DB() // Get the underlying generic sql.DB instance
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections kept alive in the background.
	// Having 10-20 connections pre-warmed prevents the 100ms TCP handshake delay on new requests.
	sqlDB.SetMaxIdleConns(2)

	// SetMaxOpenConns sets the maximum number of simultaneous open connections to MSSQL.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection can be reused
	// before it's cleanly closed and remade (prevents memory/socket leaks).
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	// SetConnMaxIdleTime closes connections that have been sitting completely unused.
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	// log.Println("Database connection pool initialized successfully")
	return db, nil

	// return gorm.Open(sqlserver.Open(dsn), cfg)
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
