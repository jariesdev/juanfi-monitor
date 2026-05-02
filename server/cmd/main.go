// Package main is the entry point for the VendoReport Go server.
// Run without arguments to start the HTTP server.
// Use sub-commands (user-add, vendo-add, etc.) for CLI administration tasks.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/jariesdev/vendoreport/app"
	"github.com/jariesdev/vendoreport/commands"
	"github.com/jariesdev/vendoreport/internal/config"
	"github.com/jariesdev/vendoreport/internal/database"
)

func main() {
	// Delegate to Cobra CLI when sub-commands are provided.
	if len(os.Args) > 1 {
		commands.RootCmd.Execute()
		return
	}

	// ── Configuration ────────────────────────────────────────────────────────
	envPath := filepath.Join(".", ".env")
	cfg := config.Load(envPath)

	// ── Database ─────────────────────────────────────────────────────────────
	// SQLite: migrate=false — schema is managed by Alembic, GORM must not alter it.
	// MySQL:  migrate=true  — GORM creates the schema on a fresh database.
	db, err := database.Connect(cfg.DBPath, cfg.DBDriver, cfg.ShouldMigrate())
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	// ── App (router + scheduler + websocket hub) ──────────────────────────────
	application := app.New(db, cfg.CORSOrigins, cfg.JWTSecret, cfg.IsProduction(), true)
	defer application.StopScheduler()

	// ── Start Server ──────────────────────────────────────────────────────────
	addr := ":" + cfg.AppPort
	log.Printf("server: listening on %s", addr)

	srv := &http.Server{Addr: addr, Handler: application.Router}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("server: shutting down...")
		srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
