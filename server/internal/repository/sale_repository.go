package repository

import (
	"math"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// SaleRepository is the concrete implementation of SaleRepositoryInterface.
type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// Search returns paginated sales with optional mac_address/voucher text filter,
// date filter, and vendo filter. Results are ordered by sale_time descending.
func (r *SaleRepository) Search(q *string, date *string, vendoID *uint, page, size int) (*PageResult[models.VendoSale], error) {
	var sales []models.VendoSale
	var total int64

	query := r.db.Model(&models.VendoSale{}).Preload("Vendo").Order("sale_time DESC")

	if q != nil && *q != "" {
		query = query.Where("mac_address LIKE ? OR voucher LIKE ?", "%"+*q+"%", "%"+*q+"%")
	}
	if date != nil && *date != "" {
		query = query.Where("DATE(sale_time) = ?", *date)
	}
	if vendoID != nil {
		query = query.Where("vendo_id = ?", *vendoID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&sales).Error; err != nil {
		return nil, err
	}

	pages := int(math.Ceil(float64(total) / float64(size)))
	return &PageResult[models.VendoSale]{
		Items: sales,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// GetDailySales aggregates sales by date and vendo within the given date range.
// Only sales from active vendos are included.
func (r *SaleRepository) GetDailySales(from, to time.Time) ([]DailySaleRow, error) {
	var rows []DailySaleRow
	err := r.db.
		Table("vendo_sales").
		Select("DATE(sale_time) AS date, SUM(amount) AS total, vendo_sales.vendo_id, vendos.name AS vendo_name").
		Joins("JOIN vendos ON vendos.id = vendo_sales.vendo_id").
		Where("DATE(sale_time) BETWEEN ? AND ? AND vendos.is_active = 1", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Group("DATE(sale_time), vendo_sales.vendo_id").
		Order("sale_time ASC").
		Scan(&rows).Error
	return rows, err
}

// GetMonthlySales aggregates sales by year-month and vendo within the given date range.
// Only sales from active vendos are included.
func (r *SaleRepository) GetMonthlySales(from, to time.Time) ([]MonthlySaleRow, error) {
	var rows []MonthlySaleRow
	err := r.db.
		Table("vendo_sales").
		Select("strftime('%Y-%m', sale_time) AS month, SUM(amount) AS total, vendo_sales.vendo_id, vendos.name AS vendo_name").
		Joins("JOIN vendos ON vendos.id = vendo_sales.vendo_id").
		Where("DATE(sale_time) BETWEEN ? AND ? AND vendos.is_active = 1", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Group("strftime('%Y-%m', sale_time), vendo_sales.vendo_id").
		Order("sale_time ASC").
		Scan(&rows).Error
	return rows, err
}
