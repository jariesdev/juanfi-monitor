// Package authz provides centralized authorization helpers for controllers.
// All functions rely on the current user injected into the Gin context by the
// auth middleware under middleware.CurrentUserKey.
package authz

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
)

// AssignedVendoIDs returns nil for admin users (no filter) and the user's
// assigned vendo IDs for everyone else. Uses the vendos already preloaded
// by the auth middleware — no extra DB query.
func AssignedVendoIDs(c *gin.Context) []uint {
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

// CanAccessVendo returns true if the current user is an admin or has the
// given vendo assigned. Uses the vendos already preloaded on the user by
// the auth middleware.
func CanAccessVendo(c *gin.Context, vendoID uint) bool {
	ids := AssignedVendoIDs(c)
	if ids == nil {
		return true
	}
	return slices.Contains(ids, vendoID)
}

// IsAdmin returns true when the current user has unrestricted vendo access
// (holds the PermUsers permission).
func IsAdmin(c *gin.Context) bool {
	return AssignedVendoIDs(c) == nil
}

// CanManageRate returns true if the current user may modify the given rate:
// admins may modify any rate; everyone else may only modify rates belonging
// to a vendo they are assigned to. Default-template rates (nil VendoID) are
// admin-only.
func CanManageRate(c *gin.Context, rate *models.VendoRate) bool {
	if rate.VendoID == nil {
		return IsAdmin(c)
	}
	return CanAccessVendo(c, *rate.VendoID)
}

// AccessDenied writes a 403 response with the given detail message.
func AccessDenied(c *gin.Context, detail string) {
	c.JSON(http.StatusForbidden, gin.H{"detail": detail})
}
