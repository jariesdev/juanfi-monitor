package repository

import (
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// VendoRateRepository is the concrete implementation of VendoRateRepositoryInterface.
type VendoRateRepository struct {
	db *gorm.DB
}

func NewVendoRateRepository(db *gorm.DB) *VendoRateRepository {
	return &VendoRateRepository{db: db}
}

func (r *VendoRateRepository) ListByVendo(vendoID uint) ([]models.VendoRate, error) {
	var rates []models.VendoRate
	err := r.db.Where("vendo_id = ?", vendoID).Order("sort_order ASC, id ASC").Find(&rates).Error
	return rates, err
}

func (r *VendoRateRepository) ListDefault() ([]models.VendoRate, error) {
	var rates []models.VendoRate
	err := r.db.Where("vendo_id IS NULL").Order("sort_order ASC, id ASC").Find(&rates).Error
	return rates, err
}

func (r *VendoRateRepository) GetByID(id uint) (*models.VendoRate, error) {
	var rate models.VendoRate
	if err := r.db.First(&rate, id).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *VendoRateRepository) Create(rate *models.VendoRate) error {
	return r.db.Create(rate).Error
}

func (r *VendoRateRepository) Update(rate *models.VendoRate) error {
	return r.db.Save(rate).Error
}

func (r *VendoRateRepository) Delete(id uint) error {
	return r.db.Delete(&models.VendoRate{}, id).Error
}

// ReplaceForVendo deletes vendoID's existing rows and inserts rates, in one transaction.
func (r *VendoRateRepository) ReplaceForVendo(vendoID uint, rates []models.VendoRate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("vendo_id = ?", vendoID).Delete(&models.VendoRate{}).Error; err != nil {
			return err
		}
		return insertClones(tx, rates, &vendoID)
	})
}

// ReplaceDefault deletes existing default (VendoID == nil) rows and inserts rates.
func (r *VendoRateRepository) ReplaceDefault(rates []models.VendoRate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("vendo_id IS NULL").Delete(&models.VendoRate{}).Error; err != nil {
			return err
		}
		return insertClones(tx, rates, nil)
	})
}

// AppendToDefault inserts rates as additional default (VendoID == nil) rows
// without removing the existing default template.
func (r *VendoRateRepository) AppendToDefault(rates []models.VendoRate) error {
	return insertClones(r.db, rates, nil)
}

// ApplyDefaultToVendos clones the current default template onto each vendo ID,
// replacing that vendo's existing rows.
func (r *VendoRateRepository) ApplyDefaultToVendos(vendoIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var defaults []models.VendoRate
		if err := tx.Where("vendo_id IS NULL").Order("sort_order ASC, id ASC").Find(&defaults).Error; err != nil {
			return err
		}
		for _, vendoID := range vendoIDs {
			if err := tx.Where("vendo_id = ?", vendoID).Delete(&models.VendoRate{}).Error; err != nil {
				return err
			}
			if err := insertClones(tx, defaults, &vendoID); err != nil {
				return err
			}
		}
		return nil
	})
}

// insertClones inserts rates as fresh rows owned by vendoID (nil for the default
// template), resetting identity/timestamp fields so stale data from a source row
// (e.g. when cloning from another vendo) isn't carried over.
func insertClones(tx *gorm.DB, rates []models.VendoRate, vendoID *uint) error {
	if len(rates) == 0 {
		return nil
	}
	clones := make([]models.VendoRate, len(rates))
	for i, rate := range rates {
		rate.ID = 0
		rate.VendoID = vendoID
		rate.CreatedAt = time.Time{}
		rate.UpdatedAt = nil
		clones[i] = rate
	}
	return tx.Create(&clones).Error
}
