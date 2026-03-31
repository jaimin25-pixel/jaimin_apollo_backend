package model

import "time"

// MedicationAdminRecord tracks individual scheduled medication doses and actual administration
type MedicationAdminRecord struct {
	MARID          uint       `gorm:"primaryKey;autoIncrement" json:"mar_id"`
	PatientID      uint       `gorm:"not null;index" json:"patient_id"`
	AdmissionID    *uint      `gorm:"index" json:"admission_id,omitempty"`
	RxItemID       uint       `gorm:"not null;index" json:"rx_item_id"`
	NurseID        *uint      `gorm:"index" json:"nurse_id,omitempty"`
	ScheduledAt    time.Time  `gorm:"not null" json:"scheduled_at"`
	AdministeredAt *time.Time `json:"administered_at,omitempty"`
	Status         string     `gorm:"not null;default:'scheduled';size:20" json:"status"` // scheduled, administered, held, refused, missed
	Reason         string     `gorm:"type:text" json:"reason,omitempty"`
	Notes          string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Patient          Patient          `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Admission        *Admission       `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
	PrescriptionItem PrescriptionItem `gorm:"foreignKey:RxItemID;references:ItemID" json:"prescription_item,omitempty"`
	Nurse            *Staff           `gorm:"foreignKey:NurseID;references:StaffID" json:"nurse,omitempty"`
}

func (MedicationAdminRecord) TableName() string { return "medication_admin_records" }
