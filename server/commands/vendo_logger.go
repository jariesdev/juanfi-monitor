package commands

import (
	"github.com/jariesdev/vendoreport/internal/scheduler"
	"github.com/spf13/cobra"
)

var vendoLoggerCmd = &cobra.Command{
	Use:   "vendo-logger",
	Short: "Pull the latest logs and sales from all active vendo machines",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		scheduler.RefreshVendoLogs(db, nil)
	},
}
