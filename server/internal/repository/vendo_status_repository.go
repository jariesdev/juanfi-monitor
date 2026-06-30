package repository

import (
	"time"

	"gorm.io/gorm"
)

// VendoStatusRepository is the concrete implementation of VendoStatusRepositoryInterface.
type VendoStatusRepository struct {
	db *gorm.DB
}

func NewVendoStatusRepository(db *gorm.DB) *VendoStatusRepository {
	return &VendoStatusRepository{db: db}
}

// GetHourlyStatus returns status snapshots grouped by hour, vendo, and optionally
// filtered by vendo_id, date range, and active-only flag.
// Mirrors the Python get_hourly_status() aggregation.
func (r *VendoStatusRepository) GetHourlyStatus(vendoID *uint, from, to *string, activeOnly bool, assignedIDs []uint) ([]HourlyStatusRow, error) {
	var rows []HourlyStatusRow

	query := r.db.
		Table("vendo_status").
		Select(`
			strftime('%Y-%m-%d %H:00:00', vendo_status.created_at) AS time,
			MAX(active_users)                                        AS average_active_users,
			AVG(free_heap)                                           AS average_free_heap,
			vendo_status.vendo_id,
			vendos.name                                              AS vendo_name
		`).
		Joins("JOIN vendos ON vendos.id = vendo_status.vendo_id")

	if vendoID != nil {
		query = query.Where("vendo_status.vendo_id = ?", *vendoID)
	}
	if from != nil && *from != "" {
		loc := time.FixedZone("PHT", 8*60*60)
		start, err := time.ParseInLocation("2006-01-02", *from, loc)
		if err == nil {
			query = query.Where("vendo_status.created_at >= ?", start)
		}
	}
	if to != nil && *to != "" {
		loc := time.FixedZone("PHT", 8*60*60)
		end, err := time.ParseInLocation("2006-01-02", *to, loc)
		if err == nil {
			end = end.Add(24 * time.Hour)
			query = query.Where("vendo_status.created_at < ?", end)
		}
	}
	if activeOnly {
		query = query.Where("vendos.is_active = 1")
	}
	if len(assignedIDs) > 0 {
		query = query.Where("vendo_status.vendo_id IN ?", assignedIDs)
	}

	err := query.
		Group("strftime('%Y-%m-%d %H:00:00', vendo_status.created_at), vendo_status.vendo_id").
		Order("time ASC").
		Scan(&rows).Error

	return rows, err
}
