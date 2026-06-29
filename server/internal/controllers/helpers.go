package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// assignedVendoIDs returns nil for admin users (no filter) and the user's
// assigned vendo IDs for everyone else.
func assignedVendoIDs(c *gin.Context, userRepo repository.UserRepositoryInterface) ([]uint, error) {
	currentUser := c.MustGet(middleware.CurrentUserKey).(*models.User)
	if currentUser.HasPermission(models.PermUsers) {
		return nil, nil
	}
	return userRepo.GetVendoIDs(currentUser.ID)
}

// parseUintParam reads a named path parameter as uint, writing a 400 response on error.
func parseUintParam(c *gin.Context, name string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid " + name})
	}
	return uint(val), err
}
