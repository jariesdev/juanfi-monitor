package models

import "time"

// Notification type values, distinguishing routine events from ones needing
// the user's attention.
const (
	NotificationTypeInfo  = "info"
	NotificationTypeAlert = "alert"
)

// Notification holds messages queued for broadcast to connected WebSocket clients.
// Once broadcast, read_at is set to prevent re-delivery (table: notifications).
type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Message   string     `gorm:"column:message" json:"message"`
	Type      string     `gorm:"column:type;default:info" json:"type"`
	UserID    *uint      `gorm:"column:user_id" json:"user_id"`
	ReadAt    *time.Time `gorm:"column:read_at" json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
