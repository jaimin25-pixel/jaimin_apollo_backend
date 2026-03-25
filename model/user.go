package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleDoctor     Role = "doctor"
	RolePatient    Role = "patient"
	RolePharmacist Role = "pharmacist"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username     string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `gorm:"not null;size:200" json:"full_name"`
	Phone        string         `gorm:"size:20" json:"phone"`
	Role         Role           `gorm:"type:varchar(20);not null;index:idx_users_role" json:"role"`
	IsActive     bool           `gorm:"default:true;index:idx_users_active" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations (loaded via Preload)
	DoctorProfile     *DoctorProfile     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"doctor_profile,omitempty"`
	PatientProfile    *PatientProfile    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"patient_profile,omitempty"`
	PharmacistProfile *PharmacistProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"pharmacist_profile,omitempty"`
	AdminProfile      *AdminProfile      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"admin_profile,omitempty"`
}

func (User) TableName() string { return "users" }
