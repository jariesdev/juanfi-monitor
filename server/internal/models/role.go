package models

import (
	"encoding/json"
	"time"
)

// Role represents a named set of permissions that can be assigned to users.
type Role struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"uniqueIndex;not null" json:"name"`
	Permissions string     `gorm:"type:text;not null;default:'[]'" json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// GetPermissions deserialises the JSON-encoded permissions column into a string slice.
func (r *Role) GetPermissions() []string {
	var perms []string
	if err := json.Unmarshal([]byte(r.Permissions), &perms); err != nil {
		return []string{}
	}
	return perms
}

// SetPermissions serialises a string slice into the JSON permissions column.
func (r *Role) SetPermissions(perms []string) {
	if perms == nil {
		perms = []string{}
	}
	b, _ := json.Marshal(perms)
	r.Permissions = string(b)
}

// HasPermission reports whether this role includes the given permission.
func (r *Role) HasPermission(perm string) bool {
	for _, p := range r.GetPermissions() {
		if p == perm {
			return true
		}
	}
	return false
}

// MarshalJSON includes the decoded permissions array in JSON output.
func (r Role) MarshalJSON() ([]byte, error) {
	type Alias Role
	return json.Marshal(&struct {
		Alias
		PermissionsList []string `json:"permissions"`
	}{
		Alias:           (Alias)(r),
		PermissionsList: r.GetPermissions(),
	})
}
