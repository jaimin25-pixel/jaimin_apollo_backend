package model

import "time"

// PartnerPharmacy stores external partner pharmacy registry.
type PartnerPharmacy struct {
	PartnerID     uint      `gorm:"primaryKey;autoIncrement" json:"partner_id"`
	Name          string    `gorm:"not null;size:200" json:"name"`
	LicenseNumber string    `gorm:"uniqueIndex;not null;size:50" json:"license_number"`
	Address       string    `gorm:"type:text;not null" json:"address"`
	ContactPhone  string    `gorm:"size:20" json:"contact_phone,omitempty"`
	ContactEmail  string    `gorm:"size:255" json:"contact_email,omitempty"`
	Status        string    `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt     time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (PartnerPharmacy) TableName() string { return "partner_pharmacies" }
