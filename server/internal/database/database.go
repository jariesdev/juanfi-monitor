// Package database provides the GORM database connection and optional auto-migration.
package database

import (
	"fmt"
	"log"
	"strings"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM connection using the specified driver ("sqlite" or "mysql").
// Pass migrate=true to run AutoMigrate — use this for fresh or test databases.
// Pass migrate=false when connecting to an existing Alembic-managed SQLite database
// to prevent GORM's column-recreation quirks from altering live data.
func Connect(dsn, driver string, migrate bool) (*gorm.DB, error) {
	dial, err := dialector(driver, dsn)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dial, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", driver, err)
	}

	if migrate {
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	log.Printf("database: connected (%s)", driver)
	return db, nil
}

// dialector returns the GORM dialector for the given driver name.
func dialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case "sqlite":
		// _busy_timeout=5000: wait up to 5 s instead of immediately returning SQLITE_BUSY
		// when the scheduler and HTTP handlers write concurrently.
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		return sqlite.Open(dsn + sep + "_busy_timeout=5000"), nil
	case "mysql":
		// Ensure parseTime=True is present so time.Time fields deserialise correctly.
		return mysql.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q — valid values: sqlite, mysql", driver)
	}
}

// autoMigrate creates or updates tables to match the model structs.
// Safe to call on MySQL. Do NOT call on the Alembic-managed SQLite app.db.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Vendo{},
		&models.VendoLog{},
		&models.VendoSale{},
		&models.VendoStatus{},
		&models.Withdrawal{},
		&models.Notification{},
	)
}
