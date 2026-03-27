package model

import "time"

// Appointment stores OPD appointment records.
// Status lifecycle: scheduled → checked_in → in_consultation → completed / cancelled.
type Appointment struct {
	ApptID            uint       `gorm:"primaryKey;autoIncrement" json:"appt_id"`
	PatientID         uint       `gorm:"not null;index" json:"patient_id"`
	DoctorID          uint       `gorm:"not null;index" json:"doctor_id"`
	DeptID            uint       `gorm:"not null;index" json:"dept_id"`
	ScheduledAt       time.Time  `gorm:"not null" json:"scheduled_at"`
	QueueToken        string     `gorm:"size:20" json:"queue_token,omitempty"`
	Status            string     `gorm:"not null;default:'scheduled';size:30" json:"status"`
	ChiefComplaint    string     `gorm:"type:text" json:"chief_complaint,omitempty"`
	ConsultationNotes string     `gorm:"type:text" json:"consultation_notes,omitempty"`
	CreatedByStaffID  *uint      `gorm:"index" json:"created_by_staff_id,omitempty"`
	CreatedAt         time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Patient        Patient     `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Doctor         Doctor      `gorm:"foreignKey:DoctorID;references:DoctorID" json:"doctor,omitempty"`
	Department     Department  `gorm:"foreignKey:DeptID;references:DeptID" json:"department,omitempty"`
	CreatedByStaff *Staff      `gorm:"foreignKey:CreatedByStaffID;references:StaffID" json:"created_by_staff,omitempty"`
}

func (Appointment) TableName() string { return "appointments" }
