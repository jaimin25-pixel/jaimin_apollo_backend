package model

// PrescriptionItem represents one medicine line in a prescription.
type PrescriptionItem struct {
	ItemID             uint    `gorm:"primaryKey;autoIncrement" json:"item_id"`
	RxID               uint    `gorm:"not null;index" json:"rx_id"`
	MedicineID         uint    `gorm:"not null;index" json:"medicine_id"`
	Dose               string  `gorm:"not null;size:50" json:"dose"`
	Frequency          string  `gorm:"not null;size:50" json:"frequency"`
	DurationDays       int     `gorm:"not null" json:"duration_days"`
	Instructions       string  `gorm:"type:text" json:"instructions,omitempty"`
	QuantityPrescribed float64 `gorm:"type:decimal(10,2);not null" json:"quantity_prescribed"`
	QuantityDispensed  float64 `gorm:"type:decimal(10,2);not null;default:0" json:"quantity_dispensed"`
	Status             string  `gorm:"not null;default:'pending';size:20" json:"status"`

	// Relations
	Prescription Prescription `gorm:"foreignKey:RxID;references:RxID" json:"prescription,omitempty"`
	Medicine     Medicine     `gorm:"foreignKey:MedicineID;references:MedicineID" json:"medicine,omitempty"`
}

func (PrescriptionItem) TableName() string { return "prescription_items" }
