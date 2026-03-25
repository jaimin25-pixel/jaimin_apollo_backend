package model

import (
	"time"

	"github.com/google/uuid"
)

type DoctorProfile struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_doctor_user" json:"user_id"`
	LicenseNumber    string    `gorm:"not null;size:100;uniqueIndex:idx_doctor_license" json:"license_number"`
	Specialization   string    `gorm:"not null;size:100;index:idx_doctor_specialization" json:"specialization"`
	Qualification    string    `gorm:"size:200" json:"qualification,omitempty"`
	ExperienceYears  int       `gorm:"default:0" json:"experience_years"`
	ConsultationFee  float64   `gorm:"type:decimal(10,2);default:0" json:"consultation_fee"`
	Bio              string    `gorm:"type:text" json:"bio,omitempty"`
	IsAvailable      bool      `gorm:"default:true;index:idx_doctor_available" json:"is_available"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (DoctorProfile) TableName() string { return "doctor_profiles" }
