package model

import (
	"time"

	"github.com/google/uuid"
)

// PasswordReset is immutable — no DeletedAt, no gorm.Model
type PasswordReset struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_reset_user" json:"user_id"`
	Code      string    `gorm:"not null;size:6;index:idx_reset_code" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index:idx_reset_expires" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func (PasswordReset) TableName() string { return "password_resets" }
