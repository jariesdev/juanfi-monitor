package models

import "time"

// Vendo represents a registered vending machine (table: vendos).
type Vendo struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	Name         string        `gorm:"uniqueIndex" json:"name"`
	MacAddress   *string       `gorm:"column:mac_address" json:"mac_address"`
	APIURL       *string       `gorm:"column:api_url" json:"api_url"`
	APIKey       *string       `gorm:"column:api_key" json:"-"` // sensitive – never serialised
	IsOnline     bool          `gorm:"default:1" json:"is_online"`
	TotalSales   *float64      `gorm:"column:total_sales" json:"total_sales"`
	CurrentSales *float64      `gorm:"column:current_sales" json:"current_sales"`
	IsActive     int           `gorm:"default:1" json:"is_active"`
	// Commission is the percentage (0–100) of this vendo's sales kept by the
	// location/partner. It reduces the vendo's revenue in the profit report.
	Commission   float64       `gorm:"column:commission;default:0" json:"commission"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    *time.Time    `json:"updated_at"`

	// Associations – excluded from default JSON responses
	VendoLogs   []VendoLog   `gorm:"foreignKey:VendoID" json:"-"`
	VendoSales  []VendoSale  `gorm:"foreignKey:VendoID" json:"-"`
	VendoStatus []VendoStatus `gorm:"foreignKey:VendoID" json:"-"`
	Withdrawals []Withdrawal  `gorm:"foreignKey:VendoID" json:"-"`

	// RecentStatus is populated manually in VendoRepository.Search; not a GORM relation.
	RecentStatus *VendoStatus `gorm:"-" json:"recent_status,omitempty"`
}

func (Vendo) TableName() string { return "vendos" }
