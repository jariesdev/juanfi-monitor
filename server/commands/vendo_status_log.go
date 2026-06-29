package commands

import (
	"github.com/jariesdev/vendoreport/internal/scheduler"
	"github.com/spf13/cobra"
)

var vendoStatusLogCmd = &cobra.Command{
	Use:   "vendo-status-log",
	Short: "Snapshot the current status of all active vendo machines",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		scheduler.UpdateVendoStatus(db)
	},
}
