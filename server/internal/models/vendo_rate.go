package models

import "time"

// VendoRate represents a pricing tier for a vendo machine (table: vendo_rates).
// VendoID is nullable: rows with VendoID == nil form the global default rate-plan
// template used as the source for "apply to all vendos".
type VendoRate struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	VendoID      *uint      `gorm:"column:vendo_id;index" json:"vendo_id"`
	Name         string     `gorm:"column:name" json:"name"`
	Price        float64    `gorm:"column:price" json:"price"`
	Minutes      int        `gorm:"column:minutes" json:"minutes"`
	ValidityMins int        `gorm:"column:validity_minutes" json:"validity_minutes"`
	DataLimitMB  *int       `gorm:"column:data_limit_mb" json:"data_limit_mb"`
	UserProfile  *string    `gorm:"column:user_profile" json:"user_profile"`
	SortOrder    int        `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

func (VendoRate) TableName() string { return "vendo_rates" }
