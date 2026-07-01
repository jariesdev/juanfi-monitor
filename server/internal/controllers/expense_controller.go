package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
	"gorm.io/gorm"
)

// expenseCategories is the closed set of allowed expense categories.
var expenseCategories = map[string]bool{
	"subscription": true,
	"electricity":  true,
	"materials":    true,
	"consumables":  true,
	"labor":        true,
	"other":        true,
}

// ExpenseController handles expense CRUD plus the profit report and forecast.
// All operations are scoped to the current (owning) user.
type ExpenseController struct {
	expenseRepo   repository.ExpenseRepositoryInterface
	profitService *services.ProfitService
}

func NewExpenseController(expenseRepo repository.ExpenseRepositoryInterface, profitService *services.ProfitService) *ExpenseController {
	return &ExpenseController{expenseRepo: expenseRepo, profitService: profitService}
}

// expenseRequest is the create/update payload.
type expenseRequest struct {
	Category    string  `json:"category" binding:"required"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	IsRecurring bool    `json:"is_recurring"`
	ExpenseDate string  `json:"expense_date" binding:"required"` // YYYY-MM-DD
	EndDate     string  `json:"end_date"`                        // YYYY-MM-DD, recurring only, optional
}

// currentUserID returns the authenticated user's ID.
func currentUserID(c *gin.Context) uint {
	return c.MustGet(middleware.CurrentUserKey).(*models.User).ID
}

// ownVendoIDs returns the current user's own assigned vendo IDs (from the
// user preloaded by the auth middleware). Unlike authz.AssignedVendoIDs this
// never returns nil for admins — the profit report is strictly per-owner.
func ownVendoIDs(c *gin.Context) []uint {
	u := c.MustGet(middleware.CurrentUserKey).(*models.User)
	ids := make([]uint, len(u.Vendos))
	for i, v := range u.Vendos {
		ids[i] = v.ID
	}
	return ids
}

// List handles GET /expenses — the current user's expenses, newest first.
// Query params: from_date, to_date (YYYY-MM-DD), category (all optional).
func (e *ExpenseController) List(c *gin.Context) {
	var from, to *time.Time
	if v := c.Query("from_date"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, phtLocation); err == nil {
			from = &t
		}
	}
	if v := c.Query("to_date"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, phtLocation); err == nil {
			end := t.Add(24 * time.Hour)
			to = &end
		}
	}
	var category *string
	if v := c.Query("category"); v != "" {
		category = &v
	}

	expenses, err := e.expenseRepo.Search(currentUserID(c), from, to, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": expenses})
}

// Create handles POST /expenses.
func (e *ExpenseController) Create(c *gin.Context) {
	var req expenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	date, endDate, ok := parseExpenseRequest(c, &req)
	if !ok {
		return
	}

	exp := &models.Expense{
		UserID:      currentUserID(c),
		Category:    req.Category,
		Description: req.Description,
		Amount:      req.Amount,
		IsRecurring: req.IsRecurring,
		ExpenseDate: date,
		EndDate:     endDate,
	}
	if err := e.expenseRepo.Create(exp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exp)
}

// Update handles PUT /expenses/:id.
func (e *ExpenseController) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	var req expenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	date, endDate, ok := parseExpenseRequest(c, &req)
	if !ok {
		return
	}

	userID := currentUserID(c)
	exp, err := e.expenseRepo.Get(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": "expense not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	exp.Category = req.Category
	exp.Description = req.Description
	exp.Amount = req.Amount
	exp.IsRecurring = req.IsRecurring
	exp.ExpenseDate = date
	exp.EndDate = endDate
	if err := e.expenseRepo.Update(exp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exp)
}

// Delete handles DELETE /expenses/:id.
func (e *ExpenseController) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := e.expenseRepo.Delete(id, currentUserID(c)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": "expense not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "deleted"})
}

// Report handles GET /profit-report. Defaults to an all-time window so the
// cumulative net reflects the true break-even trajectory.
// Query params: from_date, to_date (YYYY-MM-DD, both optional).
func (e *ExpenseController) Report(c *gin.Context) {
	from := time.Date(2000, 1, 1, 0, 0, 0, 0, phtLocation)
	to := time.Now().In(phtLocation)
	if v := c.Query("from_date"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, phtLocation); err == nil {
			from = t
		}
	}
	if v := c.Query("to_date"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, phtLocation); err == nil {
			to = t.Add(24 * time.Hour)
		}
	}

	report, err := e.profitService.Report(ownVendoIDs(c), currentUserID(c), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// Forecast handles GET /profit-forecast.
func (e *ExpenseController) Forecast(c *gin.Context) {
	forecast, err := e.profitService.Forecast(ownVendoIDs(c), currentUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, forecast)
}

// parseExpenseRequest validates the category and parses the expense date plus
// the optional recurring end date, writing a 400 response and returning
// ok=false on failure. The end date is honored only for recurring expenses and
// must be on or after the expense date.
func parseExpenseRequest(c *gin.Context, req *expenseRequest) (date time.Time, endDate *time.Time, ok bool) {
	if !expenseCategories[req.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid category"})
		return time.Time{}, nil, false
	}
	date, err := time.ParseInLocation("2006-01-02", req.ExpenseDate, phtLocation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid expense_date (want YYYY-MM-DD)"})
		return time.Time{}, nil, false
	}
	if req.IsRecurring && req.EndDate != "" {
		ed, err := time.ParseInLocation("2006-01-02", req.EndDate, phtLocation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid end_date (want YYYY-MM-DD)"})
			return time.Time{}, nil, false
		}
		if ed.Before(date) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "end_date must be on or after expense_date"})
			return time.Time{}, nil, false
		}
		endDate = &ed
	}
	return date, endDate, true
}
