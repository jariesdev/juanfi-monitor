package repository

import (
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// ExpenseRepository is the concrete implementation of ExpenseRepositoryInterface.
type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Search returns the user's expenses (newest first), with optional date-range
// and category filters. Results are always scoped to userID.
func (r *ExpenseRepository) Search(userID uint, from, to *time.Time, category *string) ([]models.Expense, error) {
	var expenses []models.Expense
	q := r.db.Where("user_id = ?", userID).Order("expense_date DESC, id DESC")
	if from != nil {
		q = q.Where("expense_date >= ?", *from)
	}
	if to != nil {
		q = q.Where("expense_date < ?", *to)
	}
	if category != nil && *category != "" {
		q = q.Where("category = ?", *category)
	}
	err := q.Find(&expenses).Error
	return expenses, err
}

// Get returns a single expense owned by userID, or gorm.ErrRecordNotFound.
func (r *ExpenseRepository) Get(id, userID uint) (*models.Expense, error) {
	var e models.Expense
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// Create inserts a new expense.
func (r *ExpenseRepository) Create(e *models.Expense) error {
	return r.db.Create(e).Error
}

// Update persists changes to an existing expense.
func (r *ExpenseRepository) Update(e *models.Expense) error {
	return r.db.Save(e).Error
}

// Delete removes an expense owned by userID. Returns gorm.ErrRecordNotFound
// when no matching row exists (wrong owner or missing), so callers can 404.
func (r *ExpenseRepository) Delete(id, userID uint) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Expense{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MonthlyOneTime sums non-recurring expense amounts grouped by PHT year-month
// within [from, to). The '+8 hours' offset mirrors the sales aggregation so
// month buckets line up with GetMonthlySales.
func (r *ExpenseRepository) MonthlyOneTime(userID uint, from, to time.Time) ([]MonthAmount, error) {
	var rows []MonthAmount
	err := r.db.
		Table("expenses").
		Select("strftime('%Y-%m', expense_date, '+8 hours') AS month, SUM(amount) AS amount").
		Where("user_id = ? AND is_recurring = 0 AND expense_date >= ? AND expense_date < ?", userID, from, to).
		Group("strftime('%Y-%m', expense_date, '+8 hours')").
		Order("month ASC").
		Scan(&rows).Error
	return rows, err
}

// ActiveRecurring returns all recurring expenses for the user, oldest first.
func (r *ExpenseRepository) ActiveRecurring(userID uint) ([]models.Expense, error) {
	var expenses []models.Expense
	err := r.db.
		Where("user_id = ? AND is_recurring = 1", userID).
		Order("expense_date ASC").
		Find(&expenses).Error
	return expenses, err
}
