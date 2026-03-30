package model

import "time"

// VisitorLog tracks hospital visitors against admitted or OPD patients.
type VisitorLog struct {
	VisitorID       uint       `gorm:"primaryKey;autoIncrement" json:"visitor_id"`
	PatientID       uint       `gorm:"not null;index" json:"patient_id"`
	VisitorName     string     `gorm:"not null;size:150" json:"visitor_name"`
	Relation        string     `gorm:"not null;size:50" json:"relation"`
	TimeIn          time.Time  `gorm:"not null" json:"time_in"`
	TimeOut         *time.Time `json:"time_out,omitempty"`
	LoggedByStaffID uint       `gorm:"not null;index" json:"logged_by_staff_id"`
	CreatedAt       time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Patient       Patient `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	LoggedByStaff Staff   `gorm:"foreignKey:LoggedByStaffID;references:StaffID" json:"logged_by_staff,omitempty"`
}

func (VisitorLog) TableName() string { return "visitor_logs" }
