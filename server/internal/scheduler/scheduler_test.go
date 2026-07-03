package scheduler

import (
	"testing"

	"github.com/jariesdev/vendoreport/internal/database"
	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

func newSchedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := database.Connect("file::memory:", "sqlite", true)
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	return db
}

func TestRunVendoStatus_NotifiesOnlyOnOnlineToOfflineTransition(t *testing.T) {
	db := newSchedulerTestDB(t)

	// Nothing listens on this port, so the connection is refused immediately
	// rather than waiting out the client's timeout.
	apiURL := "http://127.0.0.1:1"
	apiKey := "test-key"
	v := &models.Vendo{Name: "Offline Test Vendo", APIURL: &apiURL, APIKey: &apiKey, IsActive: 1, IsOnline: true}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create vendo: %v", err)
	}

	user := &models.User{Username: "offline-notify-user", Password: "x", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Model(user).Association("Vendos").Append([]models.Vendo{*v}); err != nil {
		t.Fatalf("assign vendo: %v", err)
	}

	// First poll: the vendo was online, the device is unreachable → exactly
	// one offline notification is queued for the assigned user.
	online := runVendoStatus(db, v, nil)
	if online {
		t.Fatal("expected runVendoStatus to report offline for an unreachable device")
	}

	var notifications []models.Notification
	if err := db.Where("user_id = ?", user.ID).Find(&notifications).Error; err != nil {
		t.Fatalf("load notifications: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected 1 offline notification, got %d", len(notifications))
	}
	if !contains(notifications[0].Message, v.Name) || !contains(notifications[0].Message, "unplug") {
		t.Errorf("expected message to name the vendo and suggest unplugging, got: %q", notifications[0].Message)
	}

	var stored models.Vendo
	if err := db.First(&stored, v.ID).Error; err != nil {
		t.Fatalf("reload vendo: %v", err)
	}
	if stored.IsOnline {
		t.Error("expected is_online to be persisted as false")
	}

	// Second poll: reload the vendo the way activeVendos() would, so
	// IsOnline now correctly reflects "already offline" — no new notification
	// should be queued for a vendo that was already down.
	if err := db.First(&stored, v.ID).Error; err != nil {
		t.Fatalf("reload vendo: %v", err)
	}
	if runVendoStatus(db, &stored, nil) {
		t.Fatal("expected still offline")
	}

	var count int64
	db.Model(&models.Notification{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Errorf("expected no additional notification while the vendo remains offline, got %d total", count)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
