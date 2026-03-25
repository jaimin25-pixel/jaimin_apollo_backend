package model

import (
	"time"

	"github.com/google/uuid"
)

type AccessLevel string

const (
	AccessSuperAdmin AccessLevel = "super_admin"
	AccessManager    AccessLevel = "manager"
	AccessStaff      AccessLevel = "staff"
)

type AdminProfile struct {
	ID          uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID      uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_admin_user" json:"user_id"`
	EmployeeID  string      `gorm:"not null;size:100;uniqueIndex:idx_admin_employee" json:"employee_id"`
	Department  string      `gorm:"not null;size:100;index:idx_admin_department" json:"department"`
	AccessLevel AccessLevel `gorm:"type:varchar(20);default:'staff';index:idx_admin_access" json:"access_level"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (AdminProfile) TableName() string { return "admin_profiles" }
