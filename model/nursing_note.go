package model

import "time"

// NursingNote stores nursing observations.
type NursingNote struct {
	NoteID      uint      `gorm:"primaryKey;autoIncrement" json:"note_id"`
	PatientID   uint      `gorm:"not null;index" json:"patient_id"`
	NurseID     uint      `gorm:"not null;index" json:"nurse_id"`
	AdmissionID *uint     `gorm:"index" json:"admission_id,omitempty"`
	NoteText    string    `gorm:"type:text;not null" json:"note_text"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Patient   Patient    `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Nurse     Staff      `gorm:"foreignKey:NurseID;references:StaffID" json:"nurse,omitempty"`
	Admission *Admission `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
}

func (NursingNote) TableName() string { return "nursing_notes" }
