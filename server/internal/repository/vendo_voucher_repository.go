package repository

import (
	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// VendoVoucherRepository is the concrete implementation of VendoVoucherRepositoryInterface.
type VendoVoucherRepository struct {
	db *gorm.DB
}

func NewVendoVoucherRepository(db *gorm.DB) *VendoVoucherRepository {
	return &VendoVoucherRepository{db: db}
}

func (r *VendoVoucherRepository) ListByVendo(vendoID uint) ([]models.VendoVoucher, error) {
	var vouchers []models.VendoVoucher
	err := r.db.Where("vendo_id = ?", vendoID).Order("created_at DESC, id DESC").Find(&vouchers).Error
	return vouchers, err
}

func (r *VendoVoucherRepository) CreateBatch(vouchers []models.VendoVoucher) error {
	if len(vouchers) == 0 {
		return nil
	}
	return r.db.Create(&vouchers).Error
}
