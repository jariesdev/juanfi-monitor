package models

import "time"

// User represents the users table.
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex" json:"username"`
	Password  string     `gorm:"column:password" json:"-"` // never serialised to JSON
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
