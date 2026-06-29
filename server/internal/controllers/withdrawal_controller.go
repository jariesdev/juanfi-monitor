package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// WithdrawalController handles withdrawal record endpoints.
type WithdrawalController struct {
	withdrawalRepo repository.WithdrawalRepositoryInterface
}

func NewWithdrawalController(withdrawalRepo repository.WithdrawalRepositoryInterface) *WithdrawalController {
	return &WithdrawalController{withdrawalRepo: withdrawalRepo}
}

// Search handles GET /withdrawals — list all withdrawal records.
func (w *WithdrawalController) Search(c *gin.Context) {
	withdrawals, err := w.withdrawalRepo.Search()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": withdrawals})
}
