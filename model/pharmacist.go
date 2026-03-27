package model

import "time"

// Pharmacist stores authentication and profile data for pharmacists.
type Pharmacist struct {
	PharmacistID   uint      `gorm:"primaryKey;autoIncrement" json:"pharmacist_id"`
	FullName       string    `gorm:"not null;size:150" json:"full_name"`
	Email          string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	HashedPassword string    `gorm:"not null;size:255" json:"-"`
	LicenseNumber  string    `gorm:"uniqueIndex;not null;size:50" json:"license_number"`
	Phone          string    `gorm:"size:20" json:"phone,omitempty"`
	Status         string    `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Pharmacist) TableName() string { return "pharmacists" }
