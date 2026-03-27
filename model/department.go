package model

import "time"

// Department stores the 20 clinical departments of the hospital.
type Department struct {
	DeptID      uint   `gorm:"primaryKey;autoIncrement" json:"dept_id"`
	Name        string `gorm:"uniqueIndex;not null;size:100" json:"name"`
	HODDoctorID *uint  `gorm:"index" json:"hod_doctor_id,omitempty"`
	BedCount    int    `gorm:"not null;default:0" json:"bed_count"`
	OTCount     int    `gorm:"not null;default:0" json:"ot_count"`
	HasICU      bool   `gorm:"not null;default:false" json:"has_icu"`
	Status      string `gorm:"not null;default:'active';size:20" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Relations (belongs-to only; has-many loaded via Preload)
	HODDoctor *Doctor `gorm:"foreignKey:HODDoctorID;references:DoctorID" json:"hod_doctor,omitempty"`
}

func (Department) TableName() string { return "departments" }
