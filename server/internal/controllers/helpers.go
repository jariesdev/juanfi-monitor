package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
)

// assignedVendoIDs returns nil for admin users (no filter) and the user's
// assigned vendo IDs for everyone else. Uses the vendos already preloaded
// by the auth middleware — no extra DB query.
func assignedVendoIDs(c *gin.Context) []uint {
	currentUser := c.MustGet(middleware.CurrentUserKey).(*models.User)
	if currentUser.HasPermission(models.PermUsers) {
		return nil
	}
	ids := make([]uint, len(currentUser.Vendos))
	for i, v := range currentUser.Vendos {
		ids[i] = v.ID
	}
	return ids
}

// parseUintParam reads a named path parameter as uint, writing a 400 response on error.
func parseUintParam(c *gin.Context, name string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid " + name})
	}
	return uint(val), err
}
