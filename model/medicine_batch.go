package model

import "time"

// MedicineBatch tracks individual stock batches for FIFO/FEFO dispensing.
type MedicineBatch struct {
	BatchID           uint      `gorm:"primaryKey;autoIncrement" json:"batch_id"`
	MedicineID        uint      `gorm:"not null;index" json:"medicine_id"`
	BatchNumber       string    `gorm:"not null;size:50" json:"batch_number"`
	ExpiryDate        time.Time `gorm:"type:date;not null" json:"expiry_date"`
	Quantity          float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	QuantityRemaining float64   `gorm:"type:decimal(10,2);not null" json:"quantity_remaining"`
	Supplier          string    `gorm:"size:200" json:"supplier,omitempty"`
	PurchaseDate      time.Time `gorm:"type:date;not null" json:"purchase_date"`
	PurchasePrice     *float64  `gorm:"type:decimal(10,2)" json:"purchase_price,omitempty"`
	CreatedAt         time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Medicine Medicine `gorm:"foreignKey:MedicineID;references:MedicineID" json:"medicine,omitempty"`
}

func (MedicineBatch) TableName() string { return "medicine_batches" }
