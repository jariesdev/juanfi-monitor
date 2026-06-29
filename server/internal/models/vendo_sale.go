package models

import "time"

// VendoSale records a single voucher purchase transaction (table: vendo_sales).
// The pair (vendo_id, sale_time) is unique; the logger also does a ±10-second
// window check before inserting to avoid near-duplicate entries.
type VendoSale struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	VendoID    uint      `gorm:"column:vendo_id;uniqueIndex:idx_vendo_sale_time" json:"vendo_id"`
	SaleTime   time.Time `gorm:"column:sale_time;uniqueIndex:idx_vendo_sale_time" json:"sale_time"`
	MacAddress string    `gorm:"column:mac_address" json:"mac_address"`
	Voucher    string    `gorm:"column:voucher" json:"voucher"`
	Amount     float64   `gorm:"column:amount" json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`

	Vendo *Vendo `gorm:"foreignKey:VendoID" json:"vendo,omitempty"`
}

func (VendoSale) TableName() string { return "vendo_sales" }
