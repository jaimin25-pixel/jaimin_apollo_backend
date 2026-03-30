package model

import "time"

// StockAdjustment tracks manual stock changes (additions, write-offs, returns).
// Immutable audit trail (no soft delete).
type StockAdjustment struct {
	AdjustmentID   uint      `gorm:"primaryKey;autoIncrement" json:"adjustment_id"`
	MedicineID     uint      `gorm:"not null;index" json:"medicine_id"`
	BatchID        *uint     `gorm:"index" json:"batch_id,omitempty"`
	PharmacistID   uint      `gorm:"not null;index" json:"pharmacist_id"`
	AdjustmentType string    `gorm:"not null;size:30" json:"adjustment_type"` // add, write-off, return
	Quantity       float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	Reason         string    `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Medicine   Medicine     `gorm:"foreignKey:MedicineID;references:MedicineID" json:"medicine,omitempty"`
	Batch      *MedicineBatch `gorm:"foreignKey:BatchID;references:BatchID" json:"batch,omitempty"`
	Pharmacist Pharmacist   `gorm:"foreignKey:PharmacistID;references:PharmacistID" json:"pharmacist,omitempty"`
}

func (StockAdjustment) TableName() string { return "stock_adjustments" }
