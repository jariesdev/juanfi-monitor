package repository

import (
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// NotificationRepository is the concrete implementation of NotificationRepositoryInterface.
type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// PullUnread fetches all notifications where read_at IS NULL and immediately
// marks them as read by setting read_at to now. This atomic pull-and-mark
// pattern prevents the same notification from being broadcast twice.
func (r *NotificationRepository) PullUnread() ([]models.Notification, error) {
	var notifications []models.Notification

	if err := r.db.Where("read_at IS NULL").Find(&notifications).Error; err != nil {
		return nil, err
	}

	if len(notifications) == 0 {
		return notifications, nil
	}

	// Collect IDs to update in a single query.
	ids := make([]uint, len(notifications))
	for i, n := range notifications {
		ids[i] = n.ID
	}

	now := time.Now()
	if err := r.db.Model(&models.Notification{}).
		Where("id IN ?", ids).
		Update("read_at", now).Error; err != nil {
		return nil, err
	}

	return notifications, nil
}

// Add inserts a new notification message into the queue.
func (r *NotificationRepository) Add(message string, userID *uint) (*models.Notification, error) {
	n := &models.Notification{
		Message: message,
		UserID:  userID,
	}
	if err := r.db.Create(n).Error; err != nil {
		return nil, err
	}
	return n, nil
}
