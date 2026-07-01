package models

import "time"

// Adjustment is a manual, signed correction to a month's sales revenue, owned
// by the user who created it (table: adjustments).
//
// It exists because vendo devices sometimes miscount coins: a positive
// adjustment adds under-counted revenue, a negative one removes over-counted
// revenue. Adjustments are account-wide (not tied to a single vendo) and target
// a calendar month via the Month column ("YYYY-MM", PHT), matching the profit
// report's monthly buckets.
//
// UserID is a plain scalar (no User association struct) to avoid the SQLite
// AutoMigrate table-rebuild hazard documented in CLAUDE.md.
type Adjustment struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"column:user_id;index" json:"user_id"`
	Month       string     `gorm:"column:month;index" json:"month"` // YYYY-MM
	Amount      float64    `gorm:"column:amount" json:"amount"`     // signed: + adds, - removes
	Description string     `gorm:"column:description" json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func (Adjustment) TableName() string { return "adjustments" }
