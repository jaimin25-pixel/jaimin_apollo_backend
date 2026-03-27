package model

import "time"

// Admin stores the hospital system administrator account.
// Only one row is expected. Email validated against ADMIN_EMAIL env var.
type Admin struct {
	AdminID        uint       `gorm:"primaryKey;autoIncrement" json:"admin_id"`
	Email          string     `gorm:"uniqueIndex;not null;size:255" json:"email"`
	HashedPassword string     `gorm:"not null;size:255" json:"-"`
	LastLogin      *time.Time `json:"last_login,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

func (Admin) TableName() string { return "admins" }
