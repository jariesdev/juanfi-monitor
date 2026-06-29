// Package repository defines interfaces for all data-access objects.
// Controllers depend on these interfaces, not on concrete structs, so
// implementations can be swapped or mocked without changing caller code.
package repository

import (
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
)

// UserRepositoryInterface handles user lookups and authentication checks.
type UserRepositoryInterface interface {
	CheckUser(username, password string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	List(q string, sortBy, sortDir string) ([]models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uint) error
	UpdatePassword(userID uint, newPassword string) error
	AssignRoles(userID uint, roleIDs []uint) error
	AssignVendos(userID uint, vendoIDs []uint) error
	GetVendoIDs(userID uint) ([]uint, error)
}

// RoleRepositoryInterface manages named permission roles.
type RoleRepositoryInterface interface {
	List() ([]models.Role, error)
	GetByID(id uint) (*models.Role, error)
	Create(role *models.Role) error
	Update(role *models.Role) error
	Delete(id uint) error
	CountUsersWithRole(roleID uint) (int64, error)
}

// VendoRepositoryInterface manages vendo machine records.
type VendoRepositoryInterface interface {
	// assignedIDs: when non-empty, results are filtered to only those IDs.
	Search(q *string, isActive *bool, assignedIDs []uint) ([]models.Vendo, error)
	GetByID(id uint) (*models.Vendo, error)
	Create(v *models.Vendo) error
	Delete(id uint) error
	SetStatus(id uint, active bool) error
	AllActive() ([]models.Vendo, error)
}

// LogRepositoryInterface provides access to vendo system logs.
type LogRepositoryInterface interface {
	// assignedIDs: when non-empty, results are filtered to those vendo IDs.
	Search(q *string, date *string, vendoID *uint, assignedIDs []uint, page, size int) (*PageResult[models.VendoLog], error)
}

// SaleRepositoryInterface provides access to voucher sale transactions.
type SaleRepositoryInterface interface {
	// assignedIDs: when non-empty, results are filtered to those vendo IDs.
	Search(q *string, date *string, vendoID *uint, assignedIDs []uint, page, size int, sortBy, sortDir string) (*PageResult[models.VendoSale], error)
	GetDailySales(from, to time.Time, assignedIDs []uint) ([]DailySaleRow, error)
	GetMonthlySales(from, to time.Time, assignedIDs []uint) ([]MonthlySaleRow, error)
}

// VendoStatusRepositoryInterface aggregates historical status snapshots.
type VendoStatusRepositoryInterface interface {
	// assignedIDs: when non-empty, results are filtered to those vendo IDs.
	GetHourlyStatus(vendoID *uint, from, to *string, activeOnly bool, assignedIDs []uint) ([]HourlyStatusRow, error)
}

// WithdrawalRepositoryInterface manages withdrawal records.
type WithdrawalRepositoryInterface interface {
	// assignedIDs: when non-empty, results are filtered to those vendo IDs.
	Search(vendoID *uint, assignedIDs []uint) ([]models.Withdrawal, error)
	Add(vendoID uint, amount float64, userID *uint) (*models.Withdrawal, error)
}

// NotificationRepositoryInterface manages notification delivery state.
type NotificationRepositoryInterface interface {
	PullUnread() ([]models.Notification, error)
	Add(message string, userID *uint) (*models.Notification, error)
}

// ---- Shared result types used across multiple repositories ----

// PageResult is the pagination envelope matching fastapi-pagination's shape.
type PageResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Pages int   `json:"pages"`
}

// DailySaleRow holds an aggregated daily sales summary.
type DailySaleRow struct {
	Date      string  `json:"date"`
	Total     float64 `json:"total"`
	VendoID   uint    `json:"vendo_id"`
	VendoName string  `json:"vendo_name"`
}

// MonthlySaleRow holds an aggregated monthly sales summary.
type MonthlySaleRow struct {
	Month     string  `json:"month"`
	Total     float64 `json:"total"`
	VendoID   uint    `json:"vendo_id"`
	VendoName string  `json:"vendo_name"`
}

// HourlyStatusRow holds an hourly aggregated vendo status snapshot.
type HourlyStatusRow struct {
	Time               string  `json:"time"`
	AverageActiveUsers float64 `json:"average_active_users"`
	AverageFreeHeap    float64 `json:"average_free_heap"`
	VendoID            uint    `json:"vendo_id"`
	VendoName          string  `json:"vendo_name"`
}
