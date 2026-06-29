package repository

import (
	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// WithdrawalRepository is the concrete implementation of WithdrawalRepositoryInterface.
type WithdrawalRepository struct {
	db *gorm.DB
}

func NewWithdrawalRepository(db *gorm.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

// Search returns withdrawal records with their associated vendo preloaded.
// When vendoID is non-nil the results are filtered to that vendo.
func (r *WithdrawalRepository) Search(vendoID *uint, assignedIDs []uint) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	q := r.db.Preload("Vendo")
	if vendoID != nil {
		q = q.Where("vendo_id = ?", *vendoID)
	}
	if len(assignedIDs) > 0 {
		q = q.Where("vendo_id IN ?", assignedIDs)
	}
	err := q.Find(&withdrawals).Error
	return withdrawals, err
}

// Add creates a new withdrawal record for the given vendo and amount.
func (r *WithdrawalRepository) Add(vendoID uint, amount float64, userID *uint) (*models.Withdrawal, error) {
	w := &models.Withdrawal{
		VendoID: vendoID,
		Amount:  amount,
		UserID:  userID,
	}
	if err := r.db.Create(w).Error; err != nil {
		return nil, err
	}
	return w, nil
}
