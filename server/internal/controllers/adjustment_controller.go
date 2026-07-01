package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"gorm.io/gorm"
)

// AdjustmentController handles manual sales adjustment CRUD, scoped to the
// current (owning) user.
type AdjustmentController struct {
	adjustmentRepo repository.AdjustmentRepositoryInterface
}

func NewAdjustmentController(adjustmentRepo repository.AdjustmentRepositoryInterface) *AdjustmentController {
	return &AdjustmentController{adjustmentRepo: adjustmentRepo}
}

// adjustmentRequest is the create/update payload. Amount is signed; a zero
// amount is rejected via binding:"required".
type adjustmentRequest struct {
	Month       string  `json:"month" binding:"required"`  // YYYY-MM
	Amount      float64 `json:"amount" binding:"required"` // signed, non-zero
	Description string  `json:"description"`
}

// List handles GET /adjustments — the current user's adjustments, newest first.
func (a *AdjustmentController) List(c *gin.Context) {
	adjustments, err := a.adjustmentRepo.Search(currentUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": adjustments})
}

// Create handles POST /adjustments.
func (a *AdjustmentController) Create(c *gin.Context) {
	var req adjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	month, ok := parseAdjustmentMonth(c, req.Month)
	if !ok {
		return
	}

	adj := &models.Adjustment{
		UserID:      currentUserID(c),
		Month:       month,
		Amount:      req.Amount,
		Description: req.Description,
	}
	if err := a.adjustmentRepo.Create(adj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, adj)
}

// Update handles PUT /adjustments/:id.
func (a *AdjustmentController) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	var req adjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	month, ok := parseAdjustmentMonth(c, req.Month)
	if !ok {
		return
	}

	userID := currentUserID(c)
	adj, err := a.adjustmentRepo.Get(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": "adjustment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	adj.Month = month
	adj.Amount = req.Amount
	adj.Description = req.Description
	if err := a.adjustmentRepo.Update(adj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, adj)
}

// Delete handles DELETE /adjustments/:id.
func (a *AdjustmentController) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := a.adjustmentRepo.Delete(id, currentUserID(c)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": "adjustment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "deleted"})
}

// parseAdjustmentMonth validates a YYYY-MM month string, writing a 400 response
// and returning ok=false on failure. The returned value is normalized to YYYY-MM.
func parseAdjustmentMonth(c *gin.Context, raw string) (string, bool) {
	t, err := time.ParseInLocation("2006-01", raw, phtLocation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid month (want YYYY-MM)"})
		return "", false
	}
	return t.Format("2006-01"), true
}
