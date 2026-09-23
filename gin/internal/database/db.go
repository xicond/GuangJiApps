package database

import (
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/domain"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open initializes a GORM connection to Microsoft SQL Server with performance-tuned
// connection pooling parameters (MaxIdleConns, MaxOpenConns, pre-warmed sockets).
//
// Parameters:
//   - dsn: database connection string (MSSQL format). If empty, falls back to DatabaseDSN from config.
//
// Returns:
//   - *gorm.DB: configured GORM database handle.
//   - error: non-nil if connection or pool initialization fails.
func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		dsn = config.Load().DatabaseDSN
	}
	cfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true, // Disable default transaction wrapping for read performance
		PrepareStmt:                              true, // Cache prepared statements to reduce SQL parsing overhead
	}

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

	// SetMaxIdleConns keeps connections pre-warmed to eliminate TCP handshake latency.
	sqlDB.SetMaxIdleConns(8)

	// SetMaxOpenConns sets the maximum number of simultaneous open connections to MSSQL.
	sqlDB.SetMaxOpenConns(50)

	// SetConnMaxLifetime sets the maximum amount of time a connection can be reused.
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// SetConnMaxIdleTime closes connections that have been sitting completely unused.
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// Pre-warm connection pool asynchronously so active TCP sockets are established before traffic spikes
	go func() {
		for i := 0; i < 7; i++ {
			go func() {
				var dummy int
				_ = db.Raw("SELECT 1").Scan(&dummy).Error
			}()
		}
	}()

	return db, nil

	// return gorm.Open(sqlserver.Open(dsn), cfg)
}

// MustOpen calls Open with the provided DSN and panics if the database connection fails.
//
// Parameters:
//   - dsn: database connection string.
//
// Returns:
//   - *gorm.DB: active GORM database handle.
func MustOpen(dsn string) *gorm.DB {
	db, err := Open(dsn)
	if err != nil {
		panic(err)
	}
	return db
}

// AutoMigrate applies schema auto-migrations for all core domain entity models.
//
// Parameters:
//   - db: target GORM database handle.
//
// Returns:
//   - error: non-nil if schema migration fails.
func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	models := []interface{}{
		&domain.GroupMenuMapping{},
		&domain.Admin{},
		&domain.AdminGroup{},
		&domain.AdminSubWarehouse{},
		&domain.Umat{},
		&domain.Topic{},
		&domain.Activity{},
		&domain.TimKerja{},
		&domain.TahunCiuTao{},
		&domain.PenggalangDana{},
		&domain.SxyDonatur{},
		&domain.Kelas{},
		&domain.KelasPeserta{},
		&domain.KelasPengabdi{},
		&domain.KelasTopik{},
		&domain.KelasKendaraan{},
		&domain.KelasDonasi{},
		&domain.KelasDonasiBarang{},
		&domain.KelasPengeluaran{},
		&domain.KelasMusik{},
		&domain.KelasAbsensi{},
		&domain.DonasiSxy{},
		&domain.AppLookupCategory{},
		&domain.AppLookup{},
		&domain.WorkMapping{},
		&domain.DepartmentMst{},
		&domain.AdminMatrix{},
	}
	return db.AutoMigrate(models...)
}
