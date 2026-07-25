package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/authz"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// NotificationController handles notification listing endpoints.
type NotificationController struct {
	notificationRepo repository.NotificationRepositoryInterface
}

func NewNotificationController(notificationRepo repository.NotificationRepositoryInterface) *NotificationController {
	return &NotificationController{notificationRepo: notificationRepo}
}

// Search handles GET /notifications — paginated notification history, always
// scoped to the authenticated caller (from the JWT via CurrentUserKey), never
// a client-supplied user ID. Non-admins see only their own notifications;
// admins additionally see system notifications (NULL user_id).
// Query params: q (message filter), type (info|alert), page, size.
func (n *NotificationController) Search(c *gin.Context) {
	q := c.Query("q")
	notifType := c.Query("type")
	page, size := paginationParams(c)

	var qPtr *string
	if q != "" {
		qPtr = &q
	}

	var typePtr *string
	if notifType != "" {
		typePtr = &notifType
	}

	currentUser := c.MustGet(middleware.CurrentUserKey).(*models.User)
	includeGlobal := authz.IsAdmin(c)

	result, err := n.notificationRepo.Search(currentUser.ID, includeGlobal, qPtr, typePtr, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
