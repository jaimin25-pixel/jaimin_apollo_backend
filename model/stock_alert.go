package model

import "time"

// StockAlert tracks low stock and expiry alerts for medicines.
// Status: triggered, acknowledged, resolved.
type StockAlert struct {
	AlertID     uint      `gorm:"primaryKey;autoIncrement" json:"alert_id"`
	MedicineID  uint      `gorm:"not null;index" json:"medicine_id"`
	BatchID     *uint     `gorm:"index" json:"batch_id,omitempty"`
	AlertType   string    `gorm:"not null;size:30" json:"alert_type"` // low_stock, expiring
	Severity    string    `gorm:"not null;size:20" json:"severity"`   // critical, warning, info
	Status      string    `gorm:"not null;default:'triggered';size:20" json:"status"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	TriggeredAt time.Time `gorm:"not null;default:now()" json:"triggered_at"`
	ResolvedAt  *time.Time `gorm:"type:timestamp" json:"resolved_at,omitempty"`

	// Relations
	Medicine Medicine `gorm:"foreignKey:MedicineID;references:MedicineID" json:"medicine,omitempty"`
	Batch    *MedicineBatch `gorm:"foreignKey:BatchID;references:BatchID" json:"batch,omitempty"`
}

func (StockAlert) TableName() string { return "stock_alerts" }
