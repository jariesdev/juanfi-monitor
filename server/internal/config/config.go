// Package config loads application configuration from environment variables.
package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration values loaded from the .env file.
type Config struct {
	DBDriver    string   // "sqlite" or "mysql"
	DBPath      string   // DSN — file path for SQLite, connection string for MySQL
	AppPort     string   // HTTP server port (default: 8000)
	CORSOrigins []string // Allowed CORS origins
	JWTSecret   string   // HS256 signing secret for JWT tokens
	AppEnv      string   // "development" or "production"
}

// IsProduction returns true when APP_ENV=production.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// ShouldMigrate returns true for MySQL (fresh schema) and false for SQLite
// (schema is managed by Alembic — GORM must not alter it).
func (c *Config) ShouldMigrate() bool {
	return c.DBDriver == "mysql"
}

var instance *Config

// Load reads the .env file from the given path and returns a Config.
// Call this once at startup before any other package uses the config.
func Load(envPath string) *Config {
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("warning: could not load %s, falling back to environment: %v", envPath, err)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	originsRaw := os.Getenv("CORS_ORIGINS")
	var origins []string
	if originsRaw != "" {
		for _, o := range strings.Split(originsRaw, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("warning: JWT_SECRET is not set — using insecure default. Set it in .env for production.")
		jwtSecret = "insecure-default-secret-change-me"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}
	if dbDriver != "sqlite" && dbDriver != "mysql" {
		log.Printf("warning: unknown DB_DRIVER %q, falling back to sqlite", dbDriver)
		dbDriver = "sqlite"
	}

	instance = &Config{
		DBDriver:    dbDriver,
		DBPath:      os.Getenv("DB_DSN"),
		AppPort:     port,
		CORSOrigins: origins,
		JWTSecret:   jwtSecret,
		AppEnv:      appEnv,
	}
	return instance
}

// Get returns the singleton Config. Panics if Load was not called first.
func Get() *Config {
	if instance == nil {
		panic("config.Load() must be called before config.Get()")
	}
	return instance
}
