package models

import "time"

// User represents the users table.
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex" json:"username"`
	Password  string     `gorm:"column:password" json:"-"` // never serialised to JSON
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	Roles     []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Vendos    []Vendo    `gorm:"many2many:user_vendos;" json:"vendos,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// HasPermission reports whether the user holds the given permission across any of their roles.
// Users without any role fall back to the system defaults (dashboard + account).
func (u *User) HasPermission(perm string) bool {
	for _, r := range u.Roles {
		if r.HasPermission(perm) {
			return true
		}
	}
	return defaultPermissions[perm]
}
