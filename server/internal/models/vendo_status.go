package models

import "time"

// VendoStatus is a periodic snapshot of a vendo machine's runtime metrics
// captured every 5 minutes by the cron scheduler (table: vendo_status).
type VendoStatus struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	VendoID          uint      `gorm:"column:vendo_id" json:"vendo_id"`
	TotalSales       *float64  `gorm:"column:total_sales" json:"total_sales"`
	CurrentSales     *float64  `gorm:"column:current_sales" json:"current_sales"`
	CustomerCount    *int      `gorm:"column:customer_count" json:"customer_count"`
	FreeHeap         *int      `gorm:"column:free_heap" json:"free_heap"`
	WirelessStrength *float64  `gorm:"column:wireless_strength" json:"wireless_strength"`
	ActiveUsers      *int      `gorm:"column:active_users" json:"active_users"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`

	Vendo *Vendo `gorm:"foreignKey:VendoID" json:"vendo,omitempty"`
}

func (VendoStatus) TableName() string { return "vendo_status" }
