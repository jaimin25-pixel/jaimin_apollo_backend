package model

import "time"

// Patient is the master patient registry. Each patient gets a unique PAT-XXXXXX code.
type Patient struct {
	PatientID             uint      `gorm:"primaryKey;autoIncrement" json:"patient_id"`
	PatCode               string    `gorm:"uniqueIndex;not null;size:20" json:"pat_code"`
	FullName              string    `gorm:"not null;size:150" json:"full_name"`
	DateOfBirth           time.Time `gorm:"type:date;not null" json:"date_of_birth"`
	Gender                string    `gorm:"not null;size:10" json:"gender"`
	BloodGroup            string    `gorm:"size:5" json:"blood_group,omitempty"`
	ContactNumber         string    `gorm:"not null;size:20" json:"contact_number"`
	Address               string    `gorm:"type:text" json:"address,omitempty"`
	EmergencyContactName  string    `gorm:"size:150" json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone string    `gorm:"size:20" json:"emergency_contact_phone,omitempty"`
	InsuranceID           string    `gorm:"size:100" json:"insurance_id,omitempty"`
	InsuranceProvider     string    `gorm:"size:100" json:"insurance_provider,omitempty"`
	CreatedAt             time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt             time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Has-many relations loaded via Preload in queries
}

func (Patient) TableName() string { return "patients" }
