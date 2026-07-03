package commands

import (
	"fmt"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/services"
	"github.com/spf13/cobra"
)

var voucherFailureCheckDate string

// voucherFailureCheckCmd scans a previous day's vendo_logs for coin inserts
// that were never followed by a purchase and prints the result. It is
// read-only: no coin_inserts/voucher_failures rows or notifications are
// created — detection reuses services.DetectVoucherFailures.
var voucherFailureCheckCmd = &cobra.Command{
	Use:   "voucher-failure-check",
	Short: "Check a previous date's vendo logs for coin inserts with no voucher (read-only)",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()

		loc := time.FixedZone("PHT", 8*60*60)
		dateStr := voucherFailureCheckDate
		if dateStr == "" {
			dateStr = time.Now().In(loc).AddDate(0, 0, -1).Format(time.DateOnly)
		}
		start, err := time.ParseInLocation(time.DateOnly, dateStr, loc)
		if err != nil {
			fmt.Printf("invalid --date %q (want YYYY-MM-DD): %v\n", dateStr, err)
			return
		}
		end := start.Add(24 * time.Hour)

		var vendos []models.Vendo
		if err := db.Find(&vendos).Error; err != nil {
			fmt.Printf("load vendos: %v\n", err)
			return
		}

		fmt.Printf("Checking vendo logs for %s\n", dateStr)
		totalFailures := 0
		for i := range vendos {
			v := &vendos[i]

			var logs []models.VendoLog
			if err := db.Where("vendo_id = ? AND log_time >= ? AND log_time < ?", v.ID, start, end).
				Order("log_time").Find(&logs).Error; err != nil {
				fmt.Printf("[%s] load logs: %v\n", v.Name, err)
				continue
			}

			// Parse coin-insert entries out of the stored log descriptions,
			// noting when the immediately following entry is that MAC's
			// "Cancel Topup" (user backed out of the purchase).
			var inserts []models.CoinInsert
			for i, l := range logs {
				mac, amount, ok := services.ParseCoinInsertLog(l.Description)
				if !ok {
					continue
				}
				cancelled := false
				if i+1 < len(logs) {
					if nextMac, ok := services.ParseCancelTopupLog(logs[i+1].Description); ok && nextMac == mac {
						cancelled = true
					}
				}
				inserts = append(inserts, models.CoinInsert{
					ID:         l.ID,
					VendoID:    v.ID,
					MacAddress: mac,
					Amount:     amount,
					InsertTime: l.LogTime,
					Cancelled:  cancelled,
				})
			}
			if len(inserts) == 0 {
				continue
			}

			macs := map[string]struct{}{}
			for _, ci := range inserts {
				macs[ci.MacAddress] = struct{}{}
			}
			macList := make([]string, 0, len(macs))
			for m := range macs {
				macList = append(macList, m)
			}

			// Latest sale per MAC from the day onward, so purchases logged
			// after midnight still resolve late-night inserts.
			var sales []models.VendoSale
			if err := db.Select("mac_address, sale_time").
				Where("vendo_id = ? AND mac_address IN ? AND sale_time >= ?", v.ID, macList, start).
				Find(&sales).Error; err != nil {
				fmt.Printf("[%s] load sales: %v\n", v.Name, err)
				continue
			}
			lastSale := map[string]time.Time{}
			for _, s := range sales {
				if s.SaleTime.After(lastSale[s.MacAddress]) {
					lastSale[s.MacAddress] = s.SaleTime
				}
			}

			_, failures := services.DetectVoucherFailures(inserts, lastSale, time.Now())
			for _, f := range failures {
				totalFailures++
				reason := ""
				if f.Cancelled {
					reason = " (user cancelled the top-up)"
				}
				fmt.Printf("[%s] %s inserted coins totaling %.2f between %s and %s — no voucher generated%s\n",
					v.Name, f.MacAddress, f.CoinTotal,
					f.FirstInsertAt.In(loc).Format(time.TimeOnly),
					f.LastInsertAt.In(loc).Format(time.TimeOnly), reason)
			}
		}

		if totalFailures == 0 {
			fmt.Println("No voucher failures found.")
		} else {
			fmt.Printf("Total: %d voucher failure(s)\n", totalFailures)
		}
	},
}

func init() {
	voucherFailureCheckCmd.Flags().StringVar(&voucherFailureCheckDate, "date", "",
		"Date to check (YYYY-MM-DD, PHT). Defaults to yesterday.")
}
