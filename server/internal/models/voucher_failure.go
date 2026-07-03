package models

import "time"

// VoucherFailure records one coin-inserted-but-no-voucher instance: a group of
// coin inserts from the same vendo+MAC that was never followed by a purchase.
// The unique index on (vendo_id, mac_address, last_insert_at) is the durable
// notify-once-per-instance guarantee — last_insert_at comes from stored
// coin_inserts.insert_time values, which are fixed at ingestion and therefore
// stable across polls (table: voucher_failures).
type VoucherFailure struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	VendoID       uint      `gorm:"column:vendo_id;uniqueIndex:idx_voucher_failure_instance" json:"vendo_id"`
	MacAddress    string    `gorm:"column:mac_address;uniqueIndex:idx_voucher_failure_instance" json:"mac_address"`
	LastInsertAt  time.Time `gorm:"column:last_insert_at;uniqueIndex:idx_voucher_failure_instance" json:"last_insert_at"`
	FirstInsertAt time.Time `gorm:"column:first_insert_at" json:"first_insert_at"`
	CoinTotal     float64   `gorm:"column:coin_total" json:"coin_total"`
	// Cancelled is true when the user cancelled the top-up (device logged
	// "Cancel Topup") rather than the voucher generation failing outright.
	Cancelled bool      `gorm:"column:cancelled" json:"cancelled"`
	CreatedAt time.Time `json:"created_at"`
}

func (VoucherFailure) TableName() string {
	return "voucher_failures"
}
