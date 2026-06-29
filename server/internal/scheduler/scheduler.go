// Package scheduler sets up the three recurring cron jobs.
package scheduler

import (
	"encoding/json"
	"log"
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
		safeRun("refresh_vendo_logs", func() { refreshVendoLogs(db) })
	})

	// Every 5 minutes: snapshot current system status for all active vendos.
	c.AddFunc("*/5 * * * *", func() {
		safeRun("update_vendo_status", func() { updateVendoStatus(db) })
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

// refreshVendoLogs iterates all active vendos and runs the JuanfiLogger to
// pull new logs and sales from each machine.
func refreshVendoLogs(db *gorm.DB) {
	vendos := activeVendos(db)
	for i := range vendos {
		v := &vendos[i]
		logger := services.NewJuanfiLogger(v, db)
		if err := logger.Run(); err != nil {
			log.Printf("scheduler: log refresh [%s]: %v", v.Name, err)
		}
	}
}

// updateVendoStatus iterates all active vendos and records a status snapshot.
func updateVendoStatus(db *gorm.DB) {
	vendos := activeVendos(db)
	for i := range vendos {
		v := &vendos[i]
		api := services.NewJuanfiAPI(v)
		status, err := api.GetSystemStatus()
		if err != nil {
			log.Printf("scheduler: status update [%s]: %v", v.Name, err)
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
			log.Printf("scheduler: save status [%s]: %v", v.Name, err)
		}
	}
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
