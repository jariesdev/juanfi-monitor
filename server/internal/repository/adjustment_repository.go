package repository

import (
	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// AdjustmentRepository is the concrete implementation of AdjustmentRepositoryInterface.
type AdjustmentRepository struct {
	db *gorm.DB
}

func NewAdjustmentRepository(db *gorm.DB) *AdjustmentRepository {
	return &AdjustmentRepository{db: db}
}

// Search returns the user's adjustments, newest month first.
func (r *AdjustmentRepository) Search(userID uint) ([]models.Adjustment, error) {
	var adjustments []models.Adjustment
	err := r.db.Where("user_id = ?", userID).Order("month DESC, id DESC").Find(&adjustments).Error
	return adjustments, err
}

// Get returns a single adjustment owned by userID, or gorm.ErrRecordNotFound.
func (r *AdjustmentRepository) Get(id, userID uint) (*models.Adjustment, error) {
	var a models.Adjustment
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// Create inserts a new adjustment.
func (r *AdjustmentRepository) Create(a *models.Adjustment) error {
	return r.db.Create(a).Error
}

// Update persists changes to an existing adjustment.
func (r *AdjustmentRepository) Update(a *models.Adjustment) error {
	return r.db.Save(a).Error
}

// Delete removes an adjustment owned by userID. Returns gorm.ErrRecordNotFound
// when no matching row exists (wrong owner or missing).
func (r *AdjustmentRepository) Delete(id, userID uint) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Adjustment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MonthlyTotals sums signed adjustment amounts grouped by month for the user.
func (r *AdjustmentRepository) MonthlyTotals(userID uint) ([]MonthAmount, error) {
	var rows []MonthAmount
	err := r.db.
		Table("adjustments").
		Select("month, SUM(amount) AS amount").
		Where("user_id = ?", userID).
		Group("month").
		Order("month ASC").
		Scan(&rows).Error
	return rows, err
}
