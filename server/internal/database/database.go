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
// Pass migrate=true to run AutoMigrate on startup.
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

// allModels lists every model in dependency order.
var allModels = []interface{}{
	&models.Role{},
	&models.User{},
	&models.Vendo{},
	&models.VendoLog{},
	&models.VendoSale{},
	&models.VendoStatus{},
	&models.Withdrawal{},
	&models.Notification{},
}

// autoMigrate runs the appropriate migration strategy for the driver.
// MySQL: full AutoMigrate (adds/alters columns and indexes).
// SQLite: create-only — only creates missing tables and join tables, never
// alters existing ones. Altering SQLite tables requires a full recreation
// which fails against schemas created by a different tool (e.g. Python/Alembic).
func autoMigrate(db *gorm.DB) error {
	if db.Dialector.Name() == "sqlite" {
		return sqliteCreateMissing(db)
	}
	return db.AutoMigrate(allModels...)
}

// sqliteCreateMissing creates tables and join tables that do not yet exist.
// It never touches tables that are already present.
func sqliteCreateMissing(db *gorm.DB) error {
	for _, m := range allModels {
		if db.Migrator().HasTable(m) {
			continue
		}
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}

	// Join tables are not covered by HasTable checks on the parent model,
	// so ensure they exist explicitly.
	joinTables := []struct {
		name string
		ddl  string
	}{
		{
			name: "user_roles",
			ddl:  `CREATE TABLE IF NOT EXISTS user_roles (user_id integer, role_id integer, PRIMARY KEY (user_id, role_id))`,
		},
		{
			name: "user_vendos",
			ddl:  `CREATE TABLE IF NOT EXISTS user_vendos (user_id integer, vendo_id integer, PRIMARY KEY (user_id, vendo_id))`,
		},
	}

	for _, jt := range joinTables {
		if !db.Migrator().HasTable(jt.name) {
			if err := db.Exec(jt.ddl).Error; err != nil {
				return fmt.Errorf("create join table %s: %w", jt.name, err)
			}
		}
	}

	return nil
}
