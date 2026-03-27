package model

import "time"

// Staff stores all non-doctor, non-pharmacist hospital staff.
// Role determines access: nurse, receptionist, lab_technician, radiologist,
// billing_staff, hr_manager, it_admin, ward_boy, housekeeping, security, paramedic.
type Staff struct {
	StaffID        uint      `gorm:"primaryKey;autoIncrement" json:"staff_id"`
	FullName       string    `gorm:"not null;size:150" json:"full_name"`
	Email          string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	HashedPassword string    `gorm:"not null;size:255" json:"-"`
	Role           string    `gorm:"not null;size:50" json:"role"`
	DeptID         *uint     `gorm:"index" json:"dept_id,omitempty"`
	Qualification  string    `gorm:"size:200" json:"qualification,omitempty"`
	EmploymentType string    `gorm:"not null;default:'full_time';size:20" json:"employment_type"`
	JoiningDate    time.Time `gorm:"type:date;not null" json:"joining_date"`
	Status         string    `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Department *Department `gorm:"foreignKey:DeptID;references:DeptID" json:"department,omitempty"`
}

func (Staff) TableName() string { return "staff" }
