package models

import "time"

// Expense records a business cost entered by a user (table: expenses).
//
// Expenses are account-wide (not linked to a specific vendo) and owned by the
// user who created them via the scalar UserID column. No User association
// struct is declared on purpose: AutoMigrate would otherwise walk into the
// existing users table and attempt a schema rebuild (see CLAUDE.md GORM/SQLite
// note). expenses is a brand-new table so sqliteCreateMissing creates it cleanly.
//
// Semantics that drive the profit report and forecast:
//   - One-time expense (IsRecurring=false): counts once, in its ExpenseDate
//     month. Recoverable capital (router, Starlink, UTP, labor).
//   - Recurring expense (IsRecurring=true): Amount is a fixed monthly cost that
//     applies to every month from ExpenseDate onward (subscription, electricity).
type Expense struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"column:user_id;index" json:"user_id"`
	Category    string     `gorm:"column:category" json:"category"` // subscription|electricity|materials|consumables|labor|other
	Description string     `gorm:"column:description" json:"description"`
	Amount      float64    `gorm:"column:amount" json:"amount"`
	IsRecurring bool       `gorm:"column:is_recurring;default:0" json:"is_recurring"`
	ExpenseDate time.Time  `gorm:"column:expense_date;index" json:"expense_date"` // back-datable
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func (Expense) TableName() string { return "expenses" }
