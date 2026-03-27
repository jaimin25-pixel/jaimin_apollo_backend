package model

import "time"

// StaffShift tracks shift schedules and attendance for non-doctor staff.
type StaffShift struct {
	ShiftID          uint       `gorm:"primaryKey;autoIncrement" json:"shift_id"`
	StaffID          uint       `gorm:"not null;index" json:"staff_id"`
	ShiftDate        time.Time  `gorm:"type:date;not null" json:"shift_date"`
	ShiftType        string     `gorm:"not null;size:20" json:"shift_type"`
	StartTime        string     `gorm:"type:time;not null" json:"start_time"`
	EndTime          string     `gorm:"type:time;not null" json:"end_time"`
	AttendanceStatus string     `gorm:"size:20" json:"attendance_status,omitempty"`
	CheckInAt        *time.Time `json:"check_in_at,omitempty"`
	CheckOutAt       *time.Time `json:"check_out_at,omitempty"`
	Notes            string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	StaffMember Staff `gorm:"foreignKey:StaffID;references:StaffID" json:"staff_member,omitempty"`
}

func (StaffShift) TableName() string { return "staff_shifts" }
