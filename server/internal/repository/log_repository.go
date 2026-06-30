package repository

import (
	"math"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// LogRepository is the concrete implementation of LogRepositoryInterface.
type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

// Search returns a paginated list of vendo logs with optional filters.
// Logs are ordered by log_time descending (most recent first).
func (r *LogRepository) Search(q *string, date *string, vendoID *uint, assignedIDs []uint, page, size int) (*PageResult[models.VendoLog], error) {
	var logs []models.VendoLog
	var total int64

	query := r.db.Model(&models.VendoLog{}).Preload("Vendo").Order("log_time DESC")

	if q != nil && *q != "" {
		query = query.Where("description LIKE ?", "%"+*q+"%")
	}
	if date != nil && *date != "" {
		loc := time.FixedZone("PHT", 8*60*60)
		start, err := time.ParseInLocation("2006-01-02", *date, loc)
		if err == nil {
			end := start.Add(24 * time.Hour)
			query = query.Where("log_time >= ? AND log_time < ?", start.UTC(), end.UTC())
		}
	}
	if vendoID != nil {
		query = query.Where("vendo_id = ?", *vendoID)
	}
	if len(assignedIDs) > 0 {
		query = query.Where("vendo_id IN ?", assignedIDs)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&logs).Error; err != nil {
		return nil, err
	}

	pages := int(math.Ceil(float64(total) / float64(size)))
	return &PageResult[models.VendoLog]{
		Items: logs,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}
