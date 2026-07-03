package services

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// Voucher-failure detection tuning. The device is polled every 5 minutes and
// ComputeLogTime drifts ~1s between polls, so all matching is window-based.
const (
	// coinInsertDedupWindow treats a re-fetched type-18 entry from an
	// overlapping poll as the same physical coin insert (mirrors storeSales).
	coinInsertDedupWindow = 10 * time.Second
	// coinInsertMaxAge ignores stale buffer replays so a first deploy never
	// floods notifications for old failures.
	coinInsertMaxAge = 2 * time.Hour
	// failureGraceWindow is how long after the last coin insert we wait for a
	// purchase before flagging the group — long enough for a purchase whose
	// log lands only on the next 5-minute poll.
	failureGraceWindow = 10 * time.Minute
	// saleMatchTolerance absorbs sale/insert log_time drift between polls.
	saleMatchTolerance = 5 * time.Second
)

// StoreCoinInserts persists type-18 "Inserted coin" raw log entries as pending
// coin_inserts rows so voucher-failure detection can span multiple polls.
func StoreCoinInserts(db *gorm.DB, api *JuanfiAPI, vendo *models.Vendo, rawLogs []RawLog) error {
	now := time.Now()
	for i, raw := range rawLogs {
		if raw.LogTypeIndex != coinInsertLogTypeIndex || len(raw.LogParams) < 2 {
			continue
		}
		mac := raw.LogParams[0]
		amount, err := strconv.ParseFloat(raw.LogParams[1], 64)
		if err != nil {
			log.Printf("store coin inserts: parse amount %q: %v", raw.LogParams[1], err)
			continue
		}
		t := api.ComputeLogTime(raw.Time)
		if t.Before(now.Add(-coinInsertMaxAge)) {
			continue // stale replay / first-deploy history
		}

		cancelled := i+1 < len(rawLogs) && isCancelTopupFor(rawLogs[i+1], mac)

		// Windowed dedup across ANY status so re-presented raw logs from
		// overlapping polls never resurrect a purchased/failed insert.
		var existing models.CoinInsert
		err = db.Where("vendo_id = ? AND mac_address = ? AND amount = ? AND insert_time BETWEEN ? AND ?",
			vendo.ID, mac, amount,
			t.Add(-coinInsertDedupWindow), t.Add(coinInsertDedupWindow)).
			First(&existing).Error
		if err == nil {
			// Already recorded — the cancel may only have appeared on this
			// poll if the pair straddled the previous poll's boundary.
			if cancelled && !existing.Cancelled {
				db.Model(&existing).Update("cancelled", true)
			}
			continue
		}

		record := models.CoinInsert{
			VendoID:    vendo.ID,
			MacAddress: mac,
			Amount:     amount,
			InsertTime: t,
			Status:     models.CoinInsertPending,
			Cancelled:  cancelled,
			CreatedAt:  time.Now(),
		}
		if err := db.Create(&record).Error; err != nil {
			log.Printf("store coin insert: %v", err)
		}
	}
	return nil
}

// isCancelTopupFor reports whether raw is a "Cancel Topup" entry for mac —
// used to detect a coin insert immediately followed by the user backing out.
func isCancelTopupFor(raw RawLog, mac string) bool {
	return raw.LogTypeIndex == cancelTopupLogTypeIndex && len(raw.LogParams) >= 1 && raw.LogParams[0] == mac
}

// FailureGroup describes a per-MAC group of coin inserts that was never
// followed by a purchase within the grace window.
type FailureGroup struct {
	MacAddress    string
	FirstInsertAt time.Time
	LastInsertAt  time.Time
	CoinTotal     float64
	InsertIDs     []uint
	// Cancelled is true when any insert in the group was immediately followed
	// by a "Cancel Topup" log — the user backed out rather than the machine
	// failing to generate a voucher.
	Cancelled bool
}

// DetectVoucherFailures is the pure detection core shared by the resolver and
// the CLI checker. Given coin inserts and the latest sale time per MAC, it
// returns the IDs of inserts resolved by a purchase (a sale at/after the
// insert time, within saleMatchTolerance) and the per-MAC groups of unresolved
// inserts whose last insert is older than failureGraceWindow relative to now.
func DetectVoucherFailures(inserts []models.CoinInsert, lastSale map[string]time.Time, now time.Time) (purchasedIDs []uint, failures []FailureGroup) {
	groups := map[string][]models.CoinInsert{}
	for _, ci := range inserts {
		if sale, ok := lastSale[ci.MacAddress]; ok && !sale.Before(ci.InsertTime.Add(-saleMatchTolerance)) {
			purchasedIDs = append(purchasedIDs, ci.ID)
		} else {
			groups[ci.MacAddress] = append(groups[ci.MacAddress], ci)
		}
	}

	for mac, rows := range groups {
		group := FailureGroup{
			MacAddress:    mac,
			FirstInsertAt: rows[0].InsertTime,
			LastInsertAt:  rows[0].InsertTime,
		}
		for _, ci := range rows {
			if ci.InsertTime.Before(group.FirstInsertAt) {
				group.FirstInsertAt = ci.InsertTime
			}
			if ci.InsertTime.After(group.LastInsertAt) {
				group.LastInsertAt = ci.InsertTime
			}
			group.CoinTotal += ci.Amount
			group.InsertIDs = append(group.InsertIDs, ci.ID)
			if ci.Cancelled {
				group.Cancelled = true
			}
		}
		if group.LastInsertAt.After(now.Add(-failureGraceWindow)) {
			continue // still inside grace window
		}
		failures = append(failures, group)
	}
	return purchasedIDs, failures
}

// ParseCoinInsertLog parses a stored vendo_logs description of the form
// "<mac> Inserted coin <amount>" produced by log type 18.
func ParseCoinInsertLog(description string) (mac string, amount float64, ok bool) {
	const marker = " Inserted coin "
	idx := strings.Index(description, marker)
	if idx <= 0 {
		return "", 0, false
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(description[idx+len(marker):]), 64)
	if err != nil {
		return "", 0, false
	}
	return description[:idx], amount, true
}

// ParseCancelTopupLog parses a stored vendo_logs description of the form
// "<mac> Cancel Topup" produced by log type 6.
func ParseCancelTopupLog(description string) (mac string, ok bool) {
	const suffix = " Cancel Topup"
	if !strings.HasSuffix(description, suffix) || len(description) == len(suffix) {
		return "", false
	}
	return description[:len(description)-len(suffix)], true
}

// FormatVoucherFailureMessage builds the notification/report text for a
// voucher-failure instance, noting when the user cancelled the top-up.
func FormatVoucherFailureMessage(vendoName, mac string, total float64, last time.Time, cancelled bool) string {
	loc := time.FixedZone("PHT", 8*60*60)
	if cancelled {
		return fmt.Sprintf("%s: %s inserted coins totaling %.2f at %s but no voucher was generated (user cancelled the top-up)",
			vendoName, mac, total, last.In(loc).Format("Jan 2, 3:04 PM"))
	}
	return fmt.Sprintf("%s: %s inserted coins totaling %.2f at %s but no voucher was generated",
		vendoName, mac, total, last.In(loc).Format("Jan 2, 3:04 PM"))
}

// ResolveVoucherFailures resolves pending coin inserts for a vendo: inserts
// followed by a sale from the same MAC become purchased; groups whose last
// insert has passed the grace window with no sale are flagged once as a
// voucher failure and a vendo-scoped notification is queued.
func ResolveVoucherFailures(db *gorm.DB, vendo *models.Vendo) error {
	var pending []models.CoinInsert
	if err := db.Where("vendo_id = ? AND status = ?", vendo.ID, models.CoinInsertPending).
		Order("mac_address, insert_time").
		Find(&pending).Error; err != nil {
		return fmt.Errorf("load pending coin inserts: %w", err)
	}
	if len(pending) == 0 {
		return nil
	}

	macSet := map[string]struct{}{}
	for _, ci := range pending {
		macSet[ci.MacAddress] = struct{}{}
	}
	macs := make([]string, 0, len(macSet))
	for m := range macSet {
		macs = append(macs, m)
	}

	// Latest sale time per MAC for this vendo. Aggregated in Go because
	// SQLite returns MAX(datetime) as a string GORM can't scan into time.Time.
	var sales []models.VendoSale
	if err := db.Select("mac_address, sale_time").
		Where("vendo_id = ? AND mac_address IN ?", vendo.ID, macs).
		Find(&sales).Error; err != nil {
		return fmt.Errorf("load last sales: %w", err)
	}
	lastSale := map[string]time.Time{}
	for _, s := range sales {
		if s.SaleTime.After(lastSale[s.MacAddress]) {
			lastSale[s.MacAddress] = s.SaleTime
		}
	}

	// Phase A: mark purchased every insert with a sale at/after its time
	// (within tolerance) — handles multiple coins resolved by one purchase.
	purchasedIDs, failures := DetectVoucherFailures(pending, lastSale, time.Now())
	if len(purchasedIDs) > 0 {
		if err := db.Model(&models.CoinInsert{}).
			Where("id IN ?", purchasedIDs).
			Update("status", models.CoinInsertPurchased).Error; err != nil {
			return fmt.Errorf("mark purchased: %w", err)
		}
	}

	// Phase B: flag per-MAC groups whose grace window has expired.
	for _, group := range failures {
		mac := group.MacAddress
		err := db.Transaction(func(tx *gorm.DB) error {
			failure := models.VoucherFailure{
				VendoID:       vendo.ID,
				MacAddress:    mac,
				LastInsertAt:  group.LastInsertAt,
				FirstInsertAt: group.FirstInsertAt,
				CoinTotal:     group.CoinTotal,
				Cancelled:     group.Cancelled,
				CreatedAt:     time.Now(),
			}
			if err := tx.Create(&failure).Error; err != nil {
				// Unique index violation → already flagged by a concurrent run.
				return err
			}
			if err := tx.Model(&models.CoinInsert{}).
				Where("id IN ?", group.InsertIDs).
				Update("status", models.CoinInsertFailed).Error; err != nil {
				return err
			}
			message := FormatVoucherFailureMessage(vendo.Name, mac, group.CoinTotal, group.LastInsertAt, group.Cancelled)
			return notifyVendoUsers(tx, vendo.ID, message)
		})
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				continue
			}
			log.Printf("voucher failure [%s/%s]: %v", vendo.Name, mac, err)
		}
	}
	return nil
}
