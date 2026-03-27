package model

import "time"

// PasswordReset stores 6-digit OTP codes with 10 min expiry. Immutable.
type PasswordReset struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_reset_user" json:"user_id"`
	Code      string    `gorm:"not null;size:6;index:idx_reset_code" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index:idx_reset_expires" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func (PasswordReset) TableName() string { return "password_resets" }
