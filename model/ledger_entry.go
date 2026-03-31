package model

import "time"

// LedgerEntry stores immutable financial transactions
type LedgerEntry struct {
	EntryID         uint      `gorm:"primaryKey;autoIncrement" json:"entry_id"`
	TransactionType string    `gorm:"not null;size:50" json:"transaction_type"` // OPD_Consultation, IPD_Admission, Pharmacy_Sale, Surgery, Lab, etc.
	EntryType       string    `gorm:"not null;size:10" json:"entry_type"`       // credit, debit
	Amount          float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	ReferenceID     *string   `gorm:"size:50" json:"reference_id,omitempty"` // e.g., InvoiceID, ReceiptID
	Description     string    `gorm:"type:text;not null" json:"description"`
	RecordedBy      *uint     `gorm:"index" json:"recorded_by,omitempty"`
	RecordedAt      time.Time `gorm:"not null;default:now()" json:"recorded_at"`

	// Relations
	RecordedByStaff *Staff `gorm:"foreignKey:RecordedBy;references:StaffID" json:"recorded_by_staff,omitempty"`
}

func (LedgerEntry) TableName() string { return "ledger_entries" }
