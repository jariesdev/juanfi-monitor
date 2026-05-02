package repository

import (
	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// VendoRepository is the concrete implementation of VendoRepositoryInterface.
type VendoRepository struct {
	db *gorm.DB
}

func NewVendoRepository(db *gorm.DB) *VendoRepository {
	return &VendoRepository{db: db}
}

// Search returns vendos ordered by name with optional name filter and active status.
// After loading, it attaches the most recent VendoStatus to each vendo.
func (r *VendoRepository) Search(q *string, isActive *bool) ([]models.Vendo, error) {
	var vendos []models.Vendo

	query := r.db.Order("name ASC")

	if q != nil && *q != "" {
		query = query.Where("name LIKE ?", "%"+*q+"%")
	}
	if isActive != nil {
		val := 0
		if *isActive {
			val = 1
		}
		query = query.Where("is_active = ?", val)
	}

	if err := query.Find(&vendos).Error; err != nil {
		return nil, err
	}

	// Batch-load the most recent status for every returned vendo.
	if err := r.attachRecentStatuses(vendos); err != nil {
		return nil, err
	}

	return vendos, nil
}

// attachRecentStatuses fetches the latest VendoStatus per vendo_id in a single
// query and assigns it to each Vendo's RecentStatus field.
func (r *VendoRepository) attachRecentStatuses(vendos []models.Vendo) error {
	if len(vendos) == 0 {
		return nil
	}

	// Subquery: pick the max id per vendo_id → guarantees the most recent row.
	var statuses []models.VendoStatus
	err := r.db.
		Where("id IN (SELECT MAX(id) FROM vendo_status GROUP BY vendo_id)").
		Find(&statuses).Error
	if err != nil {
		return err
	}

	// Build a lookup map for O(1) access.
	statusByVendoID := make(map[uint]*models.VendoStatus, len(statuses))
	for i := range statuses {
		statusByVendoID[statuses[i].VendoID] = &statuses[i]
	}

	for i := range vendos {
		vendos[i].RecentStatus = statusByVendoID[vendos[i].ID]
	}
	return nil
}

// GetByID returns a single vendo with its most recent status attached.
func (r *VendoRepository) GetByID(id uint) (*models.Vendo, error) {
	var vendo models.Vendo
	if err := r.db.First(&vendo, id).Error; err != nil {
		return nil, err
	}
	if err := r.attachRecentStatuses([]models.Vendo{vendo}); err != nil {
		return nil, err
	}
	return &vendo, nil
}

// Create inserts a new vendo machine record.
func (r *VendoRepository) Create(v *models.Vendo) error {
	return r.db.Create(v).Error
}

// Delete removes a vendo machine by ID.
func (r *VendoRepository) Delete(id uint) error {
	return r.db.Delete(&models.Vendo{}, id).Error
}

// SetStatus updates the is_active flag (1 = active, 0 = inactive).
func (r *VendoRepository) SetStatus(id uint, active bool) error {
	val := 0
	if active {
		val = 1
	}
	return r.db.Model(&models.Vendo{}).Where("id = ?", id).
		Update("is_active", val).Error
}

// AllActive returns all vendos where is_active = 1, used by the cron jobs.
func (r *VendoRepository) AllActive() ([]models.Vendo, error) {
	var vendos []models.Vendo
	err := r.db.Where("is_active = 1").Find(&vendos).Error
	return vendos, err
}
