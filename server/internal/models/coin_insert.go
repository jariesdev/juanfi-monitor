package models

import "time"

// CoinInsert records a single coin-insert event pulled from a vendo's device
// logs. Rows act as a durable state machine so the voucher-failure grace
// window survives across 5-minute log polls: inserts start as pending, become
// purchased when a matching sale arrives, or failed when flagged as a
// coin-inserted-but-no-voucher instance (table: coin_inserts).
type CoinInsert struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	VendoID    uint      `gorm:"column:vendo_id;index:idx_coin_insert_lookup" json:"vendo_id"`
	MacAddress string    `gorm:"column:mac_address;index:idx_coin_insert_lookup" json:"mac_address"`
	Amount     float64   `gorm:"column:amount" json:"amount"`
	InsertTime time.Time `gorm:"column:insert_time" json:"insert_time"`
	Status     string    `gorm:"column:status;index:idx_coin_insert_lookup" json:"status"` // pending | purchased | failed
	CreatedAt  time.Time `json:"created_at"`
}

func (CoinInsert) TableName() string {
	return "coin_inserts"
}

// CoinInsert status values.
const (
	CoinInsertPending   = "pending"
	CoinInsertPurchased = "purchased"
	CoinInsertFailed    = "failed"
)
