package model

import "time"

// LabOrder tracks lab test orders from ordering through result upload.
// Status: ordered, sample_collected, in_progress, completed, cancelled.
type LabOrder struct {
	OrderID           uint       `gorm:"primaryKey;autoIncrement" json:"order_id"`
	PatientID         uint       `gorm:"not null;index" json:"patient_id"`
	DoctorID          uint       `gorm:"not null;index" json:"doctor_id"`
	TestID            uint       `gorm:"not null;index" json:"test_id"`
	ApptID            *uint      `gorm:"index" json:"appt_id,omitempty"`
	AdmissionID       *uint      `gorm:"index" json:"admission_id,omitempty"`
	OrderedAt         time.Time  `gorm:"not null;default:now()" json:"ordered_at"`
	SampleCollectedAt *time.Time `json:"sample_collected_at,omitempty"`
	TechnicianID      *uint      `gorm:"index" json:"technician_id,omitempty"`
	ResultValue       string     `gorm:"type:text" json:"result_value,omitempty"`
	IsAbnormal        *bool      `json:"is_abnormal,omitempty"`
	ResultUploadedAt  *time.Time `json:"result_uploaded_at,omitempty"`
	Status            string     `gorm:"not null;default:'ordered';size:20" json:"status"`
	Notes             string     `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Patient    Patient      `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Doctor     Doctor       `gorm:"foreignKey:DoctorID;references:DoctorID" json:"doctor,omitempty"`
	TestRef    LabTest      `gorm:"foreignKey:TestID;references:TestID" json:"test,omitempty"`
	Appointment *Appointment `gorm:"foreignKey:ApptID;references:ApptID" json:"appointment,omitempty"`
	Admission   *Admission   `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
	Technician  *Staff       `gorm:"foreignKey:TechnicianID;references:StaffID" json:"technician,omitempty"`
}

func (LabOrder) TableName() string { return "lab_orders" }
