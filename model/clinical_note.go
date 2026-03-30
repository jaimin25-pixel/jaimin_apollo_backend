package model

import "time"

// ClinicalNote stores doctor SOAP/narrative notes per patient visit.
// Linked to an appointment or an admission — never both simultaneously.
type ClinicalNote struct {
	NoteID      uint      `gorm:"primaryKey;autoIncrement" json:"note_id"`
	PatientID   uint      `gorm:"not null;index" json:"patient_id"`
	DoctorID    uint      `gorm:"not null;index" json:"doctor_id"`
	ApptID      *uint     `gorm:"index" json:"appt_id,omitempty"`
	AdmissionID *uint     `gorm:"index" json:"admission_id,omitempty"`
	Notes       string    `gorm:"type:text;not null" json:"notes"`
	ICD10Code   string    `gorm:"size:20" json:"icd10_code,omitempty"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Patient     Patient      `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Doctor      Doctor       `gorm:"foreignKey:DoctorID;references:DoctorID" json:"doctor,omitempty"`
	Appointment *Appointment `gorm:"foreignKey:ApptID;references:ApptID" json:"appointment,omitempty"`
	Admission   *Admission   `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
}

func (ClinicalNote) TableName() string { return "clinical_notes" }
