package model

import "time"

// Doctor stores credentials and profile data for all hospital doctors.
type Doctor struct {
	DoctorID       uint      `gorm:"primaryKey;autoIncrement" json:"doctor_id"`
	DocCode        string    `gorm:"uniqueIndex;not null;size:20" json:"doc_code"`
	FullName       string    `gorm:"not null;size:150" json:"full_name"`
	Email          string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	HashedPassword string    `gorm:"not null;size:255" json:"-"`
	DeptID         uint      `gorm:"not null;index" json:"dept_id"`
	Specialization string    `gorm:"not null;size:100" json:"specialization"`
	Qualification  string    `gorm:"not null;size:200" json:"qualification"`
	Phone          string    `gorm:"size:20" json:"phone,omitempty"`
	JoiningDate    time.Time `gorm:"type:date;not null" json:"joining_date"`
	Status         string    `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Department Department `gorm:"foreignKey:DeptID;references:DeptID" json:"department,omitempty"`
}

func (Doctor) TableName() string { return "doctors" }
