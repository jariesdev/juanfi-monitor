package models

import "time"

// VendoVoucher records a voucher code generated on a vendo machine's hotspot
// (table: vendo_vouchers). Every row belongs to a vendo.
type VendoVoucher struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	VendoID         uint      `gorm:"column:vendo_id;index" json:"vendo_id"`
	Code            string    `gorm:"column:code;index" json:"code"`
	Prefix          string    `gorm:"column:prefix" json:"prefix"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	DurationMinutes int       `gorm:"column:duration_minutes" json:"duration_minutes"`
	AddedToSales    bool      `gorm:"column:added_to_sales" json:"added_to_sales"`
	PrintedThermal  bool      `gorm:"column:printed_thermal" json:"printed_thermal"`
	CreatedAt       time.Time `json:"created_at"`
}

func (VendoVoucher) TableName() string { return "vendo_vouchers" }
