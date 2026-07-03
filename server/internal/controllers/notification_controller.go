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

// Search handles GET /notifications — paginated notification history. Admins
// see all notifications; everyone else sees global ones plus their own.
// Query params: page, size.
func (n *NotificationController) Search(c *gin.Context) {
	page, size := paginationParams(c)

	var userIDPtr *uint
	if !authz.IsAdmin(c) {
		currentUser := c.MustGet(middleware.CurrentUserKey).(*models.User)
		userIDPtr = &currentUser.ID
	}

	result, err := n.notificationRepo.Search(userIDPtr, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
