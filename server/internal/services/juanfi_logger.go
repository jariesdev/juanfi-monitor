package services

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JuanfiLogger fetches logs and sales from a vendo machine and persists them
// to the database, creating notifications for new sale transactions.
type JuanfiLogger struct {
	vendo  *models.Vendo
	api    *JuanfiAPI
	db     *gorm.DB
}

// NewJuanfiLogger constructs a logger for a single vendo machine.
func NewJuanfiLogger(vendo *models.Vendo, db *gorm.DB) *JuanfiLogger {
	return &JuanfiLogger{
		vendo: vendo,
		api:   NewJuanfiAPI(vendo),
		db:    db,
	}
}

// Run fetches the latest logs from the vendo machine, stores sales (with
// notifications), then stores the formatted system logs.
func (l *JuanfiLogger) Run() error {
	return l.RunWithProgress(nil)
}

// RunWithProgress behaves like Run but invokes report (when non-nil) with a
// 0..1 fraction of the logger's work as each stage completes, so callers can
// surface live pull progress.
func (l *JuanfiLogger) RunWithProgress(report func(frac float64)) error {
	rawLogs, err := l.api.LoadSystemLogs()
	if err != nil {
		return fmt.Errorf("juanfi logger [%s]: load logs: %w", l.vendo.Name, err)
	}
	if report != nil {
		report(0.5)
	}

	if err := l.storeSales(rawLogs); err != nil {
		log.Printf("juanfi logger [%s]: store sales: %v", l.vendo.Name, err)
	}
	if report != nil {
		report(0.7)
	}

	if err := StoreCoinInserts(l.db, l.api, l.vendo, rawLogs); err != nil {
		log.Printf("juanfi logger [%s]: store coin inserts: %v", l.vendo.Name, err)
	}
	if report != nil {
		report(0.85)
	}

	formatted := l.api.GetFormattedLogs(rawLogs)
	if err := l.storeLogs(formatted); err != nil {
		log.Printf("juanfi logger [%s]: store logs: %v", l.vendo.Name, err)
	}
	if err := ResolveVoucherFailures(l.db, l.vendo); err != nil {
		log.Printf("juanfi logger [%s]: resolve voucher failures: %v", l.vendo.Name, err)
	}
	if report != nil {
		report(1.0)
	}

	log.Printf("juanfi logger [%s]: done", l.vendo.Name)
	return nil
}

// logDedupWindow is the tolerance for treating a freshly fetched log entry as
// one already stored on an earlier poll. The device keeps a rolling log buffer,
// so overlapping 5-minute poll windows re-deliver the same entries; but their
// absolute LogTime is recomputed each poll by ComputeLogTime (now − reported
// uptime + offset), which drifts by tens of milliseconds to ~1s between polls.
// An exact (vendo_id, log_time) match therefore can't reliably recognize a
// re-fetched entry, so we match on description within this window instead.
const logDedupWindow = 5 * time.Second

// storeLogs inserts formatted log entries, skipping ones already recorded.
// Deduplication mirrors storeSales: a windowed match on (vendo_id, description)
// absorbs the per-poll LogTime drift, with the (vendo_id, log_time) unique index
// (via OnConflict DoNothing) kept as a backstop.
func (l *JuanfiLogger) storeLogs(logs []FormattedLog) error {
	for _, entry := range logs {
		// An identical description within ±logDedupWindow is the same entry
		// re-fetched on an overlapping poll (or already inserted this batch).
		var count int64
		l.db.Model(&models.VendoLog{}).
			Where("vendo_id = ? AND description = ? AND log_time BETWEEN ? AND ?",
				l.vendo.ID, entry.Description,
				entry.LogTime.Add(-logDedupWindow),
				entry.LogTime.Add(logDedupWindow),
			).Count(&count)
		if count > 0 {
			continue // already recorded
		}

		record := models.VendoLog{
			VendoID:     l.vendo.ID,
			LogTime:     entry.LogTime,
			Description: entry.Description,
			CreatedAt:   time.Now(),
		}
		// DoNothing skips the row when the unique constraint fires.
		if err := l.db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&record).Error; err != nil {
			log.Printf("store log entry: %v", err)
		}
	}
	return nil
}

// storeSales filters for purchase log entries (type 14), deduplicates using a
// ±10-second window query, and inserts new sales + matching notifications.
func (l *JuanfiLogger) storeSales(rawLogs []RawLog) error {
	for _, raw := range rawLogs {
		if raw.LogTypeIndex != salesLogTypeIndex {
			continue
		}
		if len(raw.LogParams) < 3 {
			continue
		}

		macAddress := raw.LogParams[0]
		voucher := raw.LogParams[1]
		amount, err := strconv.ParseFloat(raw.LogParams[2], 64)
		if err != nil {
			log.Printf("store sales: parse amount %q: %v", raw.LogParams[2], err)
			continue
		}

		saleTime := l.api.ComputeLogTime(raw.Time)

		// Check for a near-duplicate within ±10 seconds to prevent double-inserts
		// caused by overlapping log poll windows.
		var count int64
		l.db.Model(&models.VendoSale{}).
			Where("vendo_id = ? AND mac_address = ? AND sale_time BETWEEN ? AND ?",
				l.vendo.ID, macAddress,
				saleTime.Add(-10*time.Second),
				saleTime.Add(10*time.Second),
			).Count(&count)

		if count > 0 {
			continue // already recorded
		}

		sale := models.VendoSale{
			VendoID:    l.vendo.ID,
			SaleTime:   saleTime,
			MacAddress: macAddress,
			Voucher:    voucher,
			Amount:     amount,
			CreatedAt:  time.Now(),
		}
		if err := l.db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&sale).Error; err != nil {
			log.Printf("store sale: %v", err)
			continue
		}

		// Queue a notification (WebSocket broadcast cron picks it up) for each
		// user assigned to this vendo — never a global one, so a sale never
		// leaks to users with no relationship to this vendo.
		message := fmt.Sprintf("%s: %s bought %s for %.2f", l.vendo.Name, macAddress, voucher, amount)
		if err := NotifyVendoUsers(l.db, l.vendo.ID, message); err != nil {
			log.Printf("create notification: %v", err)
		}
	}
	return nil
}
