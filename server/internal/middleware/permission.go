package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/models"
)

// RequirePermission returns a middleware that aborts with 403 if the
// authenticated user does not hold the specified permission.
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(CurrentUserKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "unauthenticated"})
			return
		}
		if !user.(*models.User).HasPermission(perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "forbidden"})
			return
		}
		c.Next()
	}
}
