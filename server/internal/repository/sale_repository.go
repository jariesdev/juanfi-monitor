package repository

import (
	"math"
	"strings"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// saleAllowedSort maps client sort_by values to safe SQL column expressions.
var saleAllowedSort = map[string]string{
	"sale_time":   "sale_time",
	"mac_address": "mac_address",
	"amount":      "amount",
	"voucher":     "voucher",
	"vendo_name":  "vendos.name",
}

// SaleRepository is the concrete implementation of SaleRepositoryInterface.
type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// Search returns paginated sales with optional filters and sort order.
func (r *SaleRepository) Search(q *string, date *string, vendoID *uint, assignedIDs []uint, page, size int, sortBy, sortDir string) (*PageResult[models.VendoSale], error) {
	var sales []models.VendoSale
	var total int64

	col, ok := saleAllowedSort[sortBy]
	if !ok {
		col = "sale_time"
	}
	dir := "DESC"
	if strings.ToUpper(sortDir) == "ASC" {
		dir = "ASC"
	}

	query := r.db.Model(&models.VendoSale{}).Preload("Vendo").Order(col + " " + dir)
	if sortBy == "vendo_name" {
		query = query.Joins("LEFT JOIN vendos ON vendos.id = vendo_sales.vendo_id")
	}

	if q != nil && *q != "" {
		query = query.Where("mac_address LIKE ? OR voucher LIKE ?", "%"+*q+"%", "%"+*q+"%")
	}
	if date != nil && *date != "" {
		loc := time.FixedZone("PHT", 8*60*60)
		start, err := time.ParseInLocation(time.DateOnly, *date, loc)
		if err == nil {
			end := start.Add(24 * time.Hour)
			query = query.Where("sale_time >= ? AND sale_time < ?", start, end)
		}
	}
	if vendoID != nil {
		query = query.Where("vendo_id = ?", *vendoID)
	}
	if len(assignedIDs) > 0 {
		query = query.Where("vendo_sales.vendo_id IN ?", assignedIDs)
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
// from and to are PHT-local start-of-day times — they are converted to UTC for
// the query so that entries between midnight and 8 AM PHT are correctly included.
// Only sales from active vendos are included.
func (r *SaleRepository) GetDailySales(from, to time.Time, assignedIDs []uint) ([]DailySaleRow, error) {
	var rows []DailySaleRow
	query := r.db.
		Table("vendo_sales").
		Select("DATE(sale_time, '+8 hours') AS date, SUM(amount) AS total, vendo_sales.vendo_id, vendos.name AS vendo_name").
		Joins("JOIN vendos ON vendos.id = vendo_sales.vendo_id").
		Where("sale_time >= ? AND sale_time < ? AND vendos.is_active = 1",
			from, to)
	if len(assignedIDs) > 0 {
		query = query.Where("vendo_sales.vendo_id IN ?", assignedIDs)
	}
	err := query.
		Group("DATE(sale_time, '+8 hours'), vendo_sales.vendo_id").
		Order("sale_time ASC").
		Scan(&rows).Error
	return rows, err
}

// GetMonthlySales aggregates sales by year-month and vendo within the given date range.
// from and to are PHT-local start-of-day times. When includeInactive is false
// only sales from active vendos are counted; the profit report passes true so a
// deactivated vendo's historical sales still count toward revenue.
func (r *SaleRepository) GetMonthlySales(from, to time.Time, assignedIDs []uint, includeInactive bool) ([]MonthlySaleRow, error) {
	var rows []MonthlySaleRow
	query := r.db.
		Table("vendo_sales").
		Select("strftime('%Y-%m', sale_time, '+8 hours') AS month, SUM(amount) AS total, vendo_sales.vendo_id, vendos.name AS vendo_name").
		Joins("JOIN vendos ON vendos.id = vendo_sales.vendo_id").
		Where("sale_time >= ? AND sale_time < ?", from, to)
	if !includeInactive {
		query = query.Where("vendos.is_active = 1")
	}
	if len(assignedIDs) > 0 {
		query = query.Where("vendo_sales.vendo_id IN ?", assignedIDs)
	}
	err := query.
		Group("strftime('%Y-%m', sale_time, '+8 hours'), vendo_sales.vendo_id").
		Order("sale_time ASC").
		Scan(&rows).Error
	return rows, err
}
