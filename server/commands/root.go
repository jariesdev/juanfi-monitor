// Package commands provides CLI sub-commands via Cobra.
// The root command starts the HTTP server; sub-commands are administrative tools.
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jariesdev/vendoreport/internal/config"
	"github.com/jariesdev/vendoreport/internal/database"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var db *gorm.DB

// RootCmd is the top-level Cobra command. Running it without sub-commands
// prints the usage. The server is started from cmd/main.go directly.
var RootCmd = &cobra.Command{
	Use:   "vendoreport",
	Short: "VendoReport CLI – manage vendo machines and users",
}

func init() {
	RootCmd.AddCommand(dbSeedCmd)
	RootCmd.AddCommand(userAddCmd)
	RootCmd.AddCommand(userAssignRoleCmd)
	RootCmd.AddCommand(vendoAddCmd)
	RootCmd.AddCommand(vendoLoggerCmd)
	RootCmd.AddCommand(vendoStatusLogCmd)
	RootCmd.AddCommand(voucherFailureCheckCmd)
	RootCmd.AddCommand(notificationsClearCmd)
}

// initDB loads config and opens the database. Called by sub-commands that need DB access.
// Migrations are intentionally skipped here — schema changes are handled by server startup.
func initDB() {
	// Load .env from the working directory (server/).
	envPath := filepath.Join(".", ".env")

	cfg := config.Load(envPath)
	var err error
	db, err = database.Connect(cfg.DBPath, cfg.DBDriver, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: database: %v\n", err)
		os.Exit(1)
	}
}
