package model

import "time"

// Vital records patient vitals per shift by nurses.
// Values outside thresholds trigger doctor notification.
type Vital struct {
	VitalID               uint      `gorm:"primaryKey;autoIncrement" json:"vital_id"`
	PatientID             uint      `gorm:"not null;index" json:"patient_id"`
	AdmissionID           *uint     `gorm:"index" json:"admission_id,omitempty"`
	NurseID               *uint     `gorm:"index" json:"nurse_id,omitempty"`
	RecordedByRole        string    `gorm:"not null;size:20;default:'nurse'" json:"recorded_by_role"`
	RecordedAt            time.Time `gorm:"not null;default:now()" json:"recorded_at"`
	TemperatureC          *float64  `gorm:"type:decimal(4,1)" json:"temperature_c,omitempty"`
	BloodPressureSystolic *int      `json:"blood_pressure_systolic,omitempty"`
	BloodPressureDiastolic *int     `json:"blood_pressure_diastolic,omitempty"`
	PulseBPM              *int      `json:"pulse_bpm,omitempty"`
	SpO2Percent           *float64  `gorm:"type:decimal(4,1)" json:"spo2_percent,omitempty"`
	RespiratoryRate       *int      `json:"respiratory_rate,omitempty"`
	BloodGlucoseMgDL      *float64  `gorm:"type:decimal(5,1)" json:"blood_glucose_mgdl,omitempty"`
	IsCritical            bool      `gorm:"not null;default:false" json:"is_critical"`
	Notes                 string    `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Patient   Patient    `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Admission *Admission `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
	Nurse     *Staff     `gorm:"foreignKey:NurseID;references:StaffID" json:"nurse,omitempty"`
}

func (Vital) TableName() string { return "vitals" }
