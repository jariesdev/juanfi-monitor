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
	rawLogs, err := l.api.LoadSystemLogs()
	if err != nil {
		return fmt.Errorf("juanfi logger [%s]: load logs: %w", l.vendo.Name, err)
	}

	if err := l.storeSales(rawLogs); err != nil {
		log.Printf("juanfi logger [%s]: store sales: %v", l.vendo.Name, err)
	}

	formatted := l.api.GetFormattedLogs(rawLogs)
	if err := l.storeLogs(formatted); err != nil {
		log.Printf("juanfi logger [%s]: store logs: %v", l.vendo.Name, err)
	}

	log.Printf("juanfi logger [%s]: done", l.vendo.Name)
	return nil
}

// storeLogs inserts formatted log entries, skipping duplicates via the
// (vendo_id, log_time) unique index (INSERT OR IGNORE semantics via OnConflict).
func (l *JuanfiLogger) storeLogs(logs []FormattedLog) error {
	for _, entry := range logs {
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

		// Queue a notification for the WebSocket broadcast cron.
		notification := models.Notification{
			Message:   fmt.Sprintf("%s: %s bought %s for %.2f", l.vendo.Name, macAddress, voucher, amount),
			CreatedAt: time.Now(),
		}
		if err := l.db.Create(&notification).Error; err != nil {
			log.Printf("create notification: %v", err)
		}
	}
	return nil
}
