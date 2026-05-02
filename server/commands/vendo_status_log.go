package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
	"github.com/spf13/cobra"
)

var vendoStatusLogCmd = &cobra.Command{
	Use:   "vendo-status-log",
	Short: "Snapshot the current status of all active vendo machines",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		repo := repository.NewVendoRepository(db)
		vendos, err := repo.AllActive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching vendos: %v\n", err)
			os.Exit(1)
		}

		for i := range vendos {
			v := &vendos[i]
			fmt.Printf("Fetching status for %s...\n", v.Name)

			api := services.NewJuanfiAPI(v)
			status, err := api.GetSystemStatus()
			if err != nil {
				fmt.Fprintf(os.Stderr, "  [%s] error: %v\n", v.Name, err)
				continue
			}

			totalSales := float64(status.TotalCoinCount)
			currentSales := float64(status.CurrentCoinCount)
			freeHeap := status.FreeHeap
			wirelessStrength := float64(status.WirelessStrength)
			activeUsers := status.ActiveUserCount
			customerCount := status.CustomerCount

			record := models.VendoStatus{
				VendoID:          v.ID,
				TotalSales:       &totalSales,
				CurrentSales:     &currentSales,
				CustomerCount:    &customerCount,
				FreeHeap:         &freeHeap,
				WirelessStrength: &wirelessStrength,
				ActiveUsers:      &activeUsers,
				CreatedAt:        time.Now(),
			}
			if err := db.Create(&record).Error; err != nil {
				fmt.Fprintf(os.Stderr, "  [%s] save status error: %v\n", v.Name, err)
			} else {
				fmt.Printf("  [%s] status saved (active_users=%d)\n", v.Name, status.ActiveUserCount)
			}
		}
		fmt.Println("Done.")
	},
}
