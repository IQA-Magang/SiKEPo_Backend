package utils

import (
	"fmt"

	"backend/models"

	"gorm.io/gorm"
)

// SendNotification creates a generic notification for a single user.
func SendNotification(db *gorm.DB, userID uint64, notificationType, title, message string) error {
	if db == nil || userID == 0 {
		return nil
	}

	notification := &models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
	}

	return db.Create(notification).Error
}

// SendBulkNotification creates the same notification for multiple users.
func SendBulkNotification(db *gorm.DB, userIDs []uint64, notificationType, title, message string) error {
	if db == nil || len(userIDs) == 0 {
		return nil
	}

	notifications := make([]models.Notification, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID == 0 {
			continue
		}
		notifications = append(notifications, models.Notification{
			UserID:  userID,
			Type:    notificationType,
			Title:   title,
			Message: message,
		})
	}

	if len(notifications) == 0 {
		return nil
	}

	return db.Create(&notifications).Error
}

// NotifyNewPeralatanCreated sends a notification to the PIC assigned to the equipment.
func NotifyNewPeralatanCreated(db *gorm.DB, peralatan *models.Peralatan) error {
	if db == nil || peralatan == nil || peralatan.PICID == 0 {
		return nil
	}

	return SendNotification(
		db,
		uint64(peralatan.PICID),
		"peralatan_created",
		"Peralatan baru ditambahkan",
		fmt.Sprintf("Peralatan %s (%s) telah ditambahkan dan menunggu proses verifikasi.", peralatan.NamaPeralatan, peralatan.NomorAset),
	)
}

// NotifyManagerOnVerificationStatusChange sends a notification to the lab manager when a verification status changes.
func NotifyManagerOnVerificationStatusChange(db *gorm.DB, peralatan *models.Peralatan, status string) error {
	if db == nil || peralatan == nil {
		return nil
	}

	var room models.Ruangan
	if err := db.Preload("Labs").First(&room, peralatan.RuanganID).Error; err != nil {
		return err
	}

	if room.Labs == nil || room.Labs.ManagerID == nil || *room.Labs.ManagerID == 0 {
		return nil
	}

	if status == "" {
		status = "diperbarui"
	}

	return SendNotification(
		db,
		*room.Labs.ManagerID,
		"peralatan_verification_updated",
		"Status verifikasi peralatan berubah",
		fmt.Sprintf("Status verifikasi peralatan %s (%s) berubah menjadi %s.", peralatan.NamaPeralatan, peralatan.NomorAset, status),
	)
}
