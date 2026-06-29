package commands

import (
	"crypto/sha256"
	"encoding/hex"
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
				db.Model(&models.Vendo{}).Where("id = ?", v.ID).Update("is_online", false)
				continue
			}

			totalSales := float64(status.TotalCoinCount)
			currentSales := float64(status.CurrentCoinCount)
			freeHeap := status.FreeHeap
			wirelessStrength := float64(status.WirelessStrength)
			activeUsers := status.ActiveUserCount
			customerCount := status.CustomerCount

			incomingHash := vendoStatusHash(totalSales, currentSales, freeHeap, wirelessStrength, activeUsers, customerCount)

			// Only insert when metric values have changed since the last snapshot.
			var latest models.VendoStatus
			changed := true
			if err := db.Where("vendo_id = ?", v.ID).Order("id DESC").First(&latest).Error; err == nil {
				latestHash := vendoStatusHash(
					derefF64(latest.TotalSales),
					derefF64(latest.CurrentSales),
					derefInt(latest.FreeHeap),
					derefF64(latest.WirelessStrength),
					derefInt(latest.ActiveUsers),
					derefInt(latest.CustomerCount),
				)
				changed = incomingHash != latestHash
			}

			if changed {
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
			} else {
				fmt.Printf("  [%s] no change, skipped\n", v.Name)
			}

			db.Model(&models.Vendo{}).Where("id = ?", v.ID).Update("is_online", true)
		}
		fmt.Println("Done.")
	},
}

// vendoStatusHash returns a SHA-256 fingerprint of the metric fields used to
// detect whether a new snapshot is identical to the previous one.
func vendoStatusHash(totalSales, currentSales float64, freeHeap int, wirelessStrength float64, activeUsers, customerCount int) string {
	s := fmt.Sprintf("%v|%v|%v|%v|%v|%v", totalSales, currentSales, freeHeap, wirelessStrength, activeUsers, customerCount)
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func derefF64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
