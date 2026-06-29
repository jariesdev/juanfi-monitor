package models

import "time"

// User represents the users table.
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex" json:"username"`
	Password  string     `gorm:"column:password" json:"-"` // never serialised to JSON
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	RoleID    *uint      `gorm:"column:role_id" json:"role_id"`
	Role      *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Vendos    []Vendo    `gorm:"many2many:user_vendos;" json:"vendos,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// HasPermission reports whether the user holds the given permission.
// Users without a role fall back to the system defaults (dashboard + account).
func (u *User) HasPermission(perm string) bool {
	if u.Role != nil {
		return u.Role.HasPermission(perm)
	}
	return defaultPermissions[perm]
}
