package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/authz"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// WithdrawalController handles withdrawal record endpoints.
type WithdrawalController struct {
	withdrawalRepo repository.WithdrawalRepositoryInterface
}

func NewWithdrawalController(withdrawalRepo repository.WithdrawalRepositoryInterface) *WithdrawalController {
	return &WithdrawalController{withdrawalRepo: withdrawalRepo}
}

// Search handles GET /withdrawals — list withdrawal records, optionally filtered by vendo_id.
func (w *WithdrawalController) Search(c *gin.Context) {
	var vendoID *uint
	if raw := c.Query("vendo_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid vendo_id"})
			return
		}
		uid := uint(id)
		vendoID = &uid
	}

	withdrawals, err := w.withdrawalRepo.Search(vendoID, authz.AssignedVendoIDs(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": withdrawals})
}
