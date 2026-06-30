// Package scheduler sets up the three recurring cron jobs.
package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
	ws "github.com/jariesdev/vendoreport/internal/websocket"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Start registers all cron jobs and starts the scheduler in the background.
// It returns the cron instance so the caller can stop it on shutdown.
func Start(db *gorm.DB, hub *ws.Hub) *cron.Cron {
	c := cron.New()

	// Every minute: pull unread notifications and broadcast to WebSocket clients.
	c.AddFunc("* * * * *", func() {
		safeRun("broadcast_notifications", func() { broadcastNotifications(db, hub) })
	})

	// Every 5 minutes: fetch system logs and sales from all active vendos.
	c.AddFunc("*/5 * * * *", func() {
		safeRun("refresh_vendo_logs", func() { RefreshVendoLogs(db) })
	})

	// Every 5 minutes: snapshot current system status for all active vendos.
	c.AddFunc("*/5 * * * *", func() {
		safeRun("update_vendo_status", func() { UpdateVendoStatus(db) })
	})

	c.Start()
	log.Println("scheduler: started (notifications @1m, logs @5m, status @5m)")
	return c
}

// broadcastNotifications pulls unread notifications and sends them as JSON
// messages to all connected WebSocket clients, matching the Python format:
// {"type": "notification", "message": "..."}
func broadcastNotifications(db *gorm.DB, hub *ws.Hub) {
	repo := repository.NewNotificationRepository(db)
	notifications, err := repo.PullUnread()
	if err != nil {
		log.Printf("scheduler: pull notifications: %v", err)
		return
	}

	for _, n := range notifications {
		payload := map[string]string{
			"type":    "notification",
			"message": n.Message,
		}
		msg, err := json.Marshal(payload)
		if err != nil {
			log.Printf("scheduler: marshal notification: %v", err)
			continue
		}
		hub.Broadcast(msg)
	}
}

// RefreshVendoLogs iterates all active vendos and runs the JuanfiLogger to
// pull new logs and sales from each machine. Each vendo is refreshed
// concurrently since they hit independent devices over the network.
func RefreshVendoLogs(db *gorm.DB) {
	vendos := activeVendos(db)
	var wg sync.WaitGroup
	for i := range vendos {
		v := &vendos[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer recoverVendoJob("refresh_vendo_logs", v.Name)
			logger := services.NewJuanfiLogger(v, db)
			if err := logger.Run(); err != nil {
				log.Printf("scheduler: log refresh [%s]: %v", v.Name, err)
			}
		}()
	}
	wg.Wait()
}

// UpdateVendoStatus iterates all active vendos, records a status snapshot only
// when metrics have changed, and keeps is_online in sync. Each vendo is
// polled concurrently since they hit independent devices over the network.
func UpdateVendoStatus(db *gorm.DB) {
	vendos := activeVendos(db)
	var wg sync.WaitGroup
	for i := range vendos {
		v := &vendos[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer recoverVendoJob("update_vendo_status", v.Name)
			updateSingleVendoStatus(db, v)
		}()
	}
	wg.Wait()
}

// updateSingleVendoStatus polls one vendo's device status, persisting a new
// snapshot only when metrics changed, and updates its is_online flag.
func updateSingleVendoStatus(db *gorm.DB, v *models.Vendo) {
	api := services.NewJuanfiAPI(v)
	status, err := api.GetSystemStatus()
	if err != nil {
		log.Printf("scheduler: status update [%s]: %v", v.Name, err)
		db.Model(&models.Vendo{}).Where("id = ?", v.ID).Update("is_online", false)
		return
	}

	totalSales := float64(status.TotalCoinCount)
	currentSales := float64(status.CurrentCoinCount)
	freeHeap := status.FreeHeap
	wirelessStrength := float64(status.WirelessStrength)
	activeUsers := status.ActiveUserCount
	customerCount := status.CustomerCount

	incoming := statusHash(totalSales, currentSales, freeHeap, wirelessStrength, activeUsers, customerCount)

	var latest models.VendoStatus
	changed := true
	if db.Where("vendo_id = ?", v.ID).Order("id DESC").First(&latest).Error == nil {
		prev := statusHash(
			derefF64(latest.TotalSales), derefF64(latest.CurrentSales),
			derefInt(latest.FreeHeap), derefF64(latest.WirelessStrength),
			derefInt(latest.ActiveUsers), derefInt(latest.CustomerCount),
		)
		changed = incoming != prev
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
			log.Printf("scheduler: save status [%s]: %v", v.Name, err)
		}
	}

	db.Model(&models.Vendo{}).Where("id = ?", v.ID).Update("is_online", true)
}

// statusHash returns a SHA-256 fingerprint of the metric fields to detect
// whether a new snapshot is identical to the previous one.
func statusHash(totalSales, currentSales float64, freeHeap int, wirelessStrength float64, activeUsers, customerCount int) string {
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

// activeVendos fetches all is_active=1 vendos from the database.
func activeVendos(db *gorm.DB) []models.Vendo {
	repo := repository.NewVendoRepository(db)
	vendos, err := repo.AllActive()
	if err != nil {
		log.Printf("scheduler: fetch active vendos: %v", err)
		return nil
	}
	return vendos
}

// safeRun wraps a job function with panic recovery so one failing job does not
// crash the entire scheduler goroutine.
func safeRun(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("scheduler: panic in %s: %v", name, r)
		}
	}()
	fn()
}

// recoverVendoJob recovers from a panic in a per-vendo goroutine so one
// vendo's failure does not crash the job or take down the others running
// concurrently.
func recoverVendoJob(job, vendoName string) {
	if r := recover(); r != nil {
		log.Printf("scheduler: panic in %s [%s]: %v", job, vendoName, r)
	}
}
