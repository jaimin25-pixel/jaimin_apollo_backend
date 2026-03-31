package model

import "time"

// BedTransferRequest tracks patient bed transfers within the hospital.
type BedTransferRequest struct {
	TransferID   uint       `gorm:"primaryKey;autoIncrement" json:"transfer_id"`
	PatientID    uint       `gorm:"not null;index" json:"patient_id"`
	FromBedID    *uint      `gorm:"index" json:"from_bed_id,omitempty"`
	ToBedID      uint       `gorm:"not null;index" json:"to_bed_id"`
	Reason       string     `gorm:"type:text;not null" json:"reason"`
	Status       string     `gorm:"not null;default:'pending';size:20" json:"status"` // pending, approved, rejected
	ApprovalNotes string    `gorm:"type:text" json:"approval_notes,omitempty"`
	RequestedBy  *uint      `gorm:"index" json:"requested_by,omitempty"`
	ConfirmedBy  *uint      `gorm:"index" json:"confirmed_by,omitempty"`
	CreatedAt    time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Patient     Patient `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	FromBed     *Bed    `gorm:"foreignKey:FromBedID;references:BedID" json:"from_bed,omitempty"`
	ToBed       Bed     `gorm:"foreignKey:ToBedID;references:BedID" json:"to_bed,omitempty"`
	Requestor   *Staff  `gorm:"foreignKey:RequestedBy;references:StaffID" json:"requestor,omitempty"`
	Confirmer   *Staff  `gorm:"foreignKey:ConfirmedBy;references:StaffID" json:"confirmer,omitempty"`
}

func (BedTransferRequest) TableName() string { return "bed_transfer_requests" }
