package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// VendoStatusController handles vendo status history endpoints.
type VendoStatusController struct {
	statusRepo repository.VendoStatusRepositoryInterface
	userRepo   repository.UserRepositoryInterface
}

func NewVendoStatusController(statusRepo repository.VendoStatusRepositoryInterface, userRepo repository.UserRepositoryInterface) *VendoStatusController {
	return &VendoStatusController{statusRepo: statusRepo, userRepo: userRepo}
}

// Search handles GET /vendo-status-history.
// Query params: vendo_id, from_date (YYYY-MM-DD), to_date, active_only (bool).
func (v *VendoStatusController) Search(c *gin.Context) {
	vendoIDStr := c.Query("vendo_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	activeOnly, _ := strconv.ParseBool(c.Query("active_only"))

	var vendoIDPtr *uint
	if vendoIDStr != "" {
		if id, err := strconv.ParseUint(vendoIDStr, 10, 64); err == nil {
			uid := uint(id)
			vendoIDPtr = &uid
		}
	}
	var fromPtr, toPtr *string
	if fromDate != "" {
		fromPtr = &fromDate
	}
	if toDate != "" {
		toPtr = &toDate
	}

	ids, err := assignedVendoIDs(c, v.userRepo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	rows, err := v.statusRepo.GetHourlyStatus(vendoIDPtr, fromPtr, toPtr, activeOnly, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}
