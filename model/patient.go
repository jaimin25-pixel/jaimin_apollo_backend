package model

import (
	"time"

	"github.com/google/uuid"
)

type BloodGroup string

const (
	BloodAPos  BloodGroup = "A+"
	BloodANeg  BloodGroup = "A-"
	BloodBPos  BloodGroup = "B+"
	BloodBNeg  BloodGroup = "B-"
	BloodABPos BloodGroup = "AB+"
	BloodABNeg BloodGroup = "AB-"
	BloodOPos  BloodGroup = "O+"
	BloodONeg  BloodGroup = "O-"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type PatientProfile struct {
	ID                uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_patient_user" json:"user_id"`
	DateOfBirth       *time.Time `gorm:"type:date" json:"date_of_birth,omitempty"`
	Gender            Gender     `gorm:"type:varchar(10);index:idx_patient_gender" json:"gender,omitempty"`
	BloodGroup        BloodGroup `gorm:"type:varchar(5);index:idx_patient_blood" json:"blood_group,omitempty"`
	InsuranceID       string     `gorm:"size:100;index:idx_patient_insurance" json:"insurance_id,omitempty"`
	InsuranceProvider string     `gorm:"size:200" json:"insurance_provider,omitempty"`
	EmergencyContact  string     `gorm:"size:200" json:"emergency_contact,omitempty"`
	EmergencyPhone    string     `gorm:"size:20" json:"emergency_phone,omitempty"`
	Address           string     `gorm:"type:text" json:"address,omitempty"`
	Allergies         string     `gorm:"type:text" json:"allergies,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (PatientProfile) TableName() string { return "patient_profiles" }
