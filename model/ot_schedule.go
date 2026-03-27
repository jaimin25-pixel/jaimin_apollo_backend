package model

import "time"

// OTSchedule tracks Operation Theatre bookings.
// Status: scheduled, prepared, in_progress, completed, closed, cancelled.
type OTSchedule struct {
	OTID                    uint       `gorm:"primaryKey;autoIncrement" json:"ot_id"`
	RoomID                  uint       `gorm:"not null;index" json:"room_id"`
	PatientID               uint       `gorm:"not null;index" json:"patient_id"`
	SurgeonID               uint       `gorm:"not null;index" json:"surgeon_id"`
	AnesthesiologistID      *uint      `gorm:"index" json:"anesthesiologist_id,omitempty"`
	ProcedureName           string     `gorm:"not null;size:200" json:"procedure_name"`
	ScheduledAt             time.Time  `gorm:"not null" json:"scheduled_at"`
	EstimatedDurationMin    *int       `json:"estimated_duration_min,omitempty"`
	ActualStartAt           *time.Time `json:"actual_start_at,omitempty"`
	ActualEndAt             *time.Time `json:"actual_end_at,omitempty"`
	Status                  string     `gorm:"not null;default:'scheduled';size:20" json:"status"`
	SurgicalNotes           string     `gorm:"type:text" json:"surgical_notes,omitempty"`
	AnesthesiaRecord        string     `gorm:"type:text" json:"anesthesia_record,omitempty"`
	SterilizationConfirmed  bool       `gorm:"not null;default:false" json:"sterilization_confirmed"`
	CreatedAt               time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Room              Ward    `gorm:"foreignKey:RoomID;references:WardID" json:"room,omitempty"`
	Patient           Patient `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Surgeon           Doctor  `gorm:"foreignKey:SurgeonID;references:DoctorID" json:"surgeon,omitempty"`
	Anesthesiologist  *Staff  `gorm:"foreignKey:AnesthesiologistID;references:StaffID" json:"anesthesiologist,omitempty"`
}

func (OTSchedule) TableName() string { return "ot_schedules" }
