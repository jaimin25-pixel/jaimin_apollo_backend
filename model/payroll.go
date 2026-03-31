package model

import "time"

// Payroll stores monthly salary calculations and history per staff member
type Payroll struct {
	PayrollID   uint       `gorm:"primaryKey;autoIncrement" json:"payroll_id"`
	StaffID     uint       `gorm:"not null;index" json:"staff_id"`
	Month       int        `gorm:"not null" json:"month"`
	Year        int        `gorm:"not null" json:"year"`
	BasicSalary float64    `gorm:"type:decimal(10,2);not null" json:"basic_salary"`
	Allowances  float64    `gorm:"type:decimal(10,2);not null;default:0" json:"allowances"`
	Deductions  float64    `gorm:"type:decimal(10,2);not null;default:0" json:"deductions"`
	NetSalary   float64    `gorm:"type:decimal(10,2);not null" json:"net_salary"`
	Status      string     `gorm:"not null;default:'draft';size:20" json:"status"` // draft, generated, paid
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	GeneratedBy *uint      `gorm:"index" json:"generated_by,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Staff            Staff `gorm:"foreignKey:StaffID;references:StaffID" json:"staff,omitempty"`
	GeneratedByAdmin *User `gorm:"foreignKey:GeneratedBy;references:ID" json:"generated_by_admin,omitempty"`
}

func (Payroll) TableName() string { return "payrolls" }
