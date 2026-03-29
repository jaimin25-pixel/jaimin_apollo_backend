package model

import "time"

// HospitalConfig stores system-wide hospital configuration (single row).
type HospitalConfig struct {
	ConfigID       uint      `gorm:"primaryKey;autoIncrement" json:"config_id"`
	HospitalName   string    `gorm:"not null;size:255" json:"hospital_name"`
	Address        string    `gorm:"type:text" json:"address"`
	GSTNumber      string    `gorm:"size:50" json:"gst_number,omitempty"`
	NABHNumber     string    `gorm:"size:50" json:"nabh_number,omitempty"`
	ContactPhone   string    `gorm:"size:20" json:"contact_phone,omitempty"`
	ContactEmail   string    `gorm:"size:255" json:"contact_email,omitempty"`
	Website        string    `gorm:"size:255" json:"website,omitempty"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (HospitalConfig) TableName() string { return "hospital_config" }
