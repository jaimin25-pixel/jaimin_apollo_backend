package model

import "time"

// Admission tracks IPD (In-Patient Department) records.
// Status: admitted, discharged, transferred, deceased.
type Admission struct {
	AdmissionID       uint       `gorm:"primaryKey;autoIncrement" json:"admission_id"`
	PatientID         uint       `gorm:"not null;index" json:"patient_id"`
	AdmittingDoctorID uint       `gorm:"not null;index" json:"admitting_doctor_id"`
	WardID            uint       `gorm:"not null;index" json:"ward_id"`
	BedID             uint       `gorm:"not null;index" json:"bed_id"`
	DeptID            uint       `gorm:"not null;index" json:"dept_id"`
	AdmittedAt        time.Time  `gorm:"not null;default:now()" json:"admitted_at"`
	Diagnosis         string     `gorm:"type:text" json:"diagnosis,omitempty"`
	TreatmentPlan     string     `gorm:"type:text" json:"treatment_plan,omitempty"`
	DischargedAt      *time.Time `json:"discharged_at,omitempty"`
	Status            string     `gorm:"not null;default:'admitted';size:20" json:"status"`
	CreatedAt         time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Patient         Patient    `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	AdmittingDoctor Doctor     `gorm:"foreignKey:AdmittingDoctorID;references:DoctorID" json:"admitting_doctor,omitempty"`
	Ward            Ward       `gorm:"foreignKey:WardID;references:WardID" json:"ward,omitempty"`
	Bed             Bed        `gorm:"foreignKey:BedID;references:BedID" json:"bed,omitempty"`
	Department      Department `gorm:"foreignKey:DeptID;references:DeptID" json:"department,omitempty"`
}

func (Admission) TableName() string { return "admissions" }
