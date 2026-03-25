package model

import (
	"time"

	"github.com/google/uuid"
)

type PharmacistProfile struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_pharmacist_user" json:"user_id"`
	LicenseNumber  string    `gorm:"not null;size:100;uniqueIndex:idx_pharmacist_license" json:"license_number"`
	BranchLocation string    `gorm:"size:200;index:idx_pharmacist_branch" json:"branch_location,omitempty"`
	Shift          string    `gorm:"size:20" json:"shift,omitempty"`
	IsOnDuty       bool      `gorm:"default:true;index:idx_pharmacist_duty" json:"is_on_duty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (PharmacistProfile) TableName() string { return "pharmacist_profiles" }
