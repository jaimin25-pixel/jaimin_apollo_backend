package model

import "time"

// LeaveRequest tracks leave applications and approval workflow.
// Types: casual, sick, earned, maternity, paternity.
// Status: pending, approved, rejected, cancelled.
type LeaveRequest struct {
	LeaveID    uint       `gorm:"primaryKey;autoIncrement" json:"leave_id"`
	StaffID    uint       `gorm:"not null;index" json:"staff_id"`
	LeaveType  string     `gorm:"not null;size:30" json:"leave_type"`
	FromDate   time.Time  `gorm:"type:date;not null" json:"from_date"`
	ToDate     time.Time  `gorm:"type:date;not null" json:"to_date"`
	Reason     string     `gorm:"type:text" json:"reason,omitempty"`
	Status     string     `gorm:"not null;default:'pending';size:20" json:"status"`
	ApprovedBy *uint      `gorm:"index" json:"approved_by,omitempty"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	CreatedAt  time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	StaffMember  Staff  `gorm:"foreignKey:StaffID;references:StaffID" json:"staff_member,omitempty"`
	ApprovedByHR *Staff `gorm:"foreignKey:ApprovedBy;references:StaffID" json:"approved_by_hr,omitempty"`
}

func (LeaveRequest) TableName() string { return "leave_requests" }
