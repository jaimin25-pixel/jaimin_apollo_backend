package model

import (
	"time"

	"gorm.io/gorm"
)

// User is the unified auth table used by the existing authentication system.
// It coexists with the role-specific tables (admins, doctors, pharmacists, staff).
type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username     string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `gorm:"not null;size:200" json:"full_name"`
	Phone        string         `gorm:"size:20" json:"phone"`
	Role         string         `gorm:"not null;default:'admin';size:50;index:idx_users_role" json:"role"`
	IsActive     bool           `gorm:"default:true;index:idx_users_active" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }
