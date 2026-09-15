package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByUserID(userID uint64) ([]models.Notification, error)
	MarkAsRead(id uint64) error
	CountUnreadByUserID(userID uint64) (int64, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByUserID(userID uint64) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(id uint64) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *notificationRepository) CountUnreadByUserID(userID uint64) (int64, error) {
	var total int64
	if err := r.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
