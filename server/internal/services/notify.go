package services

import (
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// NotifyVendoUsers creates one notification per user assigned to vendoID (via
// user_vendos), each scoped with that user's ID. Vendo-related events must
// only reach users assigned to that vendo — notification_repository.Search
// treats a NULL user_id as visible to every non-admin user, so a vendo with
// no assigned users intentionally gets no notification at all rather than
// falling back to a "global" one that would leak to unrelated users.
func NotifyVendoUsers(db *gorm.DB, vendoID uint, message string) error {
	var userIDs []uint
	if err := db.Table("user_vendos").
		Where("vendo_id = ?", vendoID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return err
	}

	for _, uid := range userIDs {
		uid := uid
		if err := db.Create(&models.Notification{
			Message:   message,
			UserID:    &uid,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
