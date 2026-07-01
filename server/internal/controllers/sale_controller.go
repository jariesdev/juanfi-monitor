package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/authz"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// phtLocation is the Philippine Time zone (UTC+8), used for date filter
// conversions so that queries correctly include entries near midnight.
var phtLocation = time.FixedZone("PHT", 8*60*60)

// SaleController handles voucher sale data endpoints.
type SaleController struct {
	saleRepo repository.SaleRepositoryInterface
}

func NewSaleController(saleRepo repository.SaleRepositoryInterface) *SaleController {
	return &SaleController{saleRepo: saleRepo}
}

// Search handles GET /sales — paginated sale search.
// Query params: q, date (YYYY-MM-DD), vendo_id, page, size, sort_by, sort_dir.
func (s *SaleController) Search(c *gin.Context) {
	q := c.Query("q")
	date := c.Query("date")
	vendoIDStr := c.Query("vendo_id")
	sortBy := c.Query("sort_by")
	sortDir := c.Query("sort_dir")
	page, size := paginationParams(c)

	var qPtr *string
	if q != "" {
		qPtr = &q
	}
	var datePtr *string
	if date != "" {
		datePtr = &date
	}
	var vendoIDPtr *uint
	if vendoIDStr != "" {
		if id, err := strconv.ParseUint(vendoIDStr, 10, 64); err == nil {
			uid := uint(id)
			vendoIDPtr = &uid
		}
	}

	result, err := s.saleRepo.Search(qPtr, datePtr, vendoIDPtr, authz.AssignedVendoIDs(c), page, size, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// DailySales handles GET /daily-sales.
// Query params: from_date, to_date (YYYY-MM-DD, both optional – defaults to last 30 days).
func (s *SaleController) DailySales(c *gin.Context) {
	from, to := parseDateRange(c)
	rows, err := s.saleRepo.GetDailySales(from, to, authz.AssignedVendoIDs(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// MonthlySales handles GET /monthly-sales.
// Query params: from_date, to_date (YYYY-MM-DD, both optional – defaults to last 12 months).
func (s *SaleController) MonthlySales(c *gin.Context) {
	from, to := parseDateRange(c)
	rows, err := s.saleRepo.GetMonthlySales(from, to, authz.AssignedVendoIDs(c), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// parseDateRange reads from_date/to_date query params, defaulting to a
// 30-day window ending today (in PHT) when absent. The returned times are
// PHT-local start-of-day, suitable for conversion to UTC in the repository.
func parseDateRange(c *gin.Context) (from, to time.Time) {
	layout := "2006-01-02"
	now := time.Now().In(phtLocation)
	to = now
	from = now.AddDate(0, -1, 0)

	if f := c.Query("from_date"); f != "" {
		if t, err := time.ParseInLocation(layout, f, phtLocation); err == nil {
			from = t
		}
	}
	if t := c.Query("to_date"); t != "" {
		if parsed, err := time.ParseInLocation(layout, t, phtLocation); err == nil {
			to = parsed
		}
	}
	return
}
