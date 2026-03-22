// Package models — Notification entity for in-app alerts.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification for a user
// (e.g., when a referral is denied, redirected, or scheduled).
type Notification struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index;index:idx_user_read,priority:1" json:"user_id"`
	ReferralID uuid.UUID `gorm:"type:uuid;not null;index" json:"referral_id"`
	Message    string    `gorm:"type:text;not null" json:"message"`
	IsRead     bool      `gorm:"default:false;not null;index:idx_user_read,priority:2" json:"is_read"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// Relations
	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Referral Referral `gorm:"foreignKey:ReferralID" json:"-"`
}
