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

// Search returns all withdrawal records with their associated vendo preloaded.
func (r *WithdrawalRepository) Search() ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	err := r.db.Preload("Vendo").Find(&withdrawals).Error
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
