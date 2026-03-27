package model

import "time"

// Medicine is the master medicine catalog.
type Medicine struct {
	MedicineID        uint    `gorm:"primaryKey;autoIncrement" json:"medicine_id"`
	GenericName       string  `gorm:"not null;size:200" json:"generic_name"`
	BrandName         string  `gorm:"size:200" json:"brand_name,omitempty"`
	Category          string  `gorm:"not null;size:100" json:"category"`
	Unit              string  `gorm:"not null;size:30" json:"unit"`
	HSNCode           string  `gorm:"size:20" json:"hsn_code,omitempty"`
	ReorderLevel      int     `gorm:"not null;default:0" json:"reorder_level"`
	CurrentStock      float64 `gorm:"type:decimal(10,2);not null;default:0" json:"current_stock"`
	StorageConditions string  `gorm:"type:text" json:"storage_conditions,omitempty"`
	Status            string  `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt         time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Has-many batches loaded via Preload in queries
}

func (Medicine) TableName() string { return "medicines" }
