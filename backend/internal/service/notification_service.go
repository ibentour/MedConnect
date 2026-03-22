package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"medconnect-oriental/backend/internal/models"
)

// NotificationService manages in-app notifications for MedConnect users.
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService creates a new notification service.
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// Create sends a notification to a specific user.
func (ns *NotificationService) Create(userID, referralID uuid.UUID, message string) error {
	notification := models.Notification{
		UserID:     userID,
		ReferralID: referralID,
		Message:    message,
		IsRead:     false,
	}

	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("notification: failed to create: %w", err)
	}
	return nil
}

// GetForUser retrieves all notifications for a user, newest first.
func (ns *NotificationService) GetForUser(userID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	err := ns.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("notification: query failed: %w", err)
	}
	return notifications, nil
}

// MarkAsRead marks a notification as read.
func (ns *NotificationService) MarkAsRead(notificationID, userID uuid.UUID) error {
	result := ns.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("notification: mark read failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification: not found or not owned by user")
	}
	return nil
}

// GetUnreadCount returns the number of unread notifications for a user.
func (ns *NotificationService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := ns.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
