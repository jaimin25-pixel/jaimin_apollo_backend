package model

import "time"

// Prescription is the header record for doctor-issued prescriptions.
// Status: pending, partially_dispensed, fully_dispensed, cancelled.
type Prescription struct {
	RxID        uint   `gorm:"primaryKey;autoIncrement" json:"rx_id"`
	DoctorID    uint   `gorm:"not null;index" json:"doctor_id"`
	PatientID   uint   `gorm:"not null;index" json:"patient_id"`
	ApptID      *uint  `gorm:"index" json:"appt_id,omitempty"`
	AdmissionID *uint  `gorm:"index" json:"admission_id,omitempty"`
	Diagnosis   string `gorm:"type:text" json:"diagnosis,omitempty"`
	Status      string `gorm:"not null;default:'pending';size:30" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Doctor      Doctor       `gorm:"foreignKey:DoctorID;references:DoctorID" json:"doctor,omitempty"`
	Patient     Patient      `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Appointment *Appointment `gorm:"foreignKey:ApptID;references:ApptID" json:"appointment,omitempty"`
	Admission   *Admission   `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
}

func (Prescription) TableName() string { return "prescriptions" }
