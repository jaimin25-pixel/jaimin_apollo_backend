package model

import "time"

// DispenseRecord tracks prescription dispensing details.
type DispenseRecord struct {
	DispenseID     uint      `gorm:"primaryKey;autoIncrement" json:"dispense_id"`
	ItemID         uint      `gorm:"not null;index" json:"item_id"`
	PharmacistID   uint      `gorm:"not null;index" json:"pharmacist_id"`
	BatchID        uint      `gorm:"not null;index" json:"batch_id"`
	QuantityDispensed float64 `gorm:"type:decimal(10,2);not null" json:"quantity_dispensed"`
	DispensedAt    time.Time `gorm:"not null;default:now()" json:"dispensed_at"`
	Notes          string    `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Item       PrescriptionItem `gorm:"foreignKey:ItemID;references:ItemID" json:"item,omitempty"`
	Pharmacist Pharmacist       `gorm:"foreignKey:PharmacistID;references:PharmacistID" json:"pharmacist,omitempty"`
	Batch      MedicineBatch    `gorm:"foreignKey:BatchID;references:BatchID" json:"batch,omitempty"`
}

func (DispenseRecord) TableName() string { return "dispense_records" }
