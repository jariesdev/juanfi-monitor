package models

import "time"

// Withdrawal records a cash withdrawal event where the operator resets current
// sales and records how much was collected (table: withdrawals).
type Withdrawal struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	VendoID   uint       `gorm:"column:vendo_id" json:"vendo_id"`
	UserID    *uint      `gorm:"column:user_id" json:"user_id"`
	Amount    float64    `gorm:"column:amount" json:"amount"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`

	Vendo *Vendo `gorm:"foreignKey:VendoID" json:"vendo,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
