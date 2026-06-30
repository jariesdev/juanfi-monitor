package models

import "time"

// VendoLog records a single system log entry for a vendo machine (table: vendo_logs).
// The pair (vendo_id, log_time) is unique as a backstop against duplicate
// inserts; the primary dedup is a windowed (vendo_id, description) check in
// JuanfiLogger.storeLogs, since log_time drifts slightly between polls.
type VendoLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	VendoID     uint      `gorm:"column:vendo_id;uniqueIndex:idx_vendo_log_time" json:"vendo_id"`
	LogTime     time.Time `gorm:"column:log_time;uniqueIndex:idx_vendo_log_time" json:"log_time"`
	Description string    `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`

	Vendo *Vendo `gorm:"foreignKey:VendoID" json:"vendo,omitempty"`
}

func (VendoLog) TableName() string { return "vendo_logs" }
