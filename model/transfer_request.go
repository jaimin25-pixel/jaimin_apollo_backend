package model

import "time"

// TransferRequest tracks inter-pharmacy transfer requests.
// Status: pending, accepted, rejected, completed.
type TransferRequest struct {
	TransferID      uint      `gorm:"primaryKey;autoIncrement" json:"transfer_id"`
	MedicineID      uint      `gorm:"not null;index" json:"medicine_id"`
	PartnerID       uint      `gorm:"not null;index" json:"partner_id"`
	PharmacistID    uint      `gorm:"not null;index" json:"pharmacist_id"`
	Quantity        float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	Reason          string    `gorm:"type:text" json:"reason,omitempty"`
	Status          string    `gorm:"not null;default:'pending';size:20" json:"status"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
	CompletedAt     *time.Time `gorm:"type:timestamp" json:"completed_at,omitempty"`

	// Relations
	Medicine   Medicine        `gorm:"foreignKey:MedicineID;references:MedicineID" json:"medicine,omitempty"`
	Partner    PartnerPharmacy `gorm:"foreignKey:PartnerID;references:PartnerID" json:"partner,omitempty"`
	Pharmacist Pharmacist      `gorm:"foreignKey:PharmacistID;references:PharmacistID" json:"pharmacist,omitempty"`
}

func (TransferRequest) TableName() string { return "transfer_requests" }
