package model

import "time"

// Bed represents an individual bed within a ward.
// Types: general, icu, isolation, maternity, pediatric.
// Status: available, occupied, maintenance, reserved.
type Bed struct {
	BedID         uint       `gorm:"primaryKey;autoIncrement" json:"bed_id"`
	WardID        uint       `gorm:"not null;index" json:"ward_id"`
	BedNumber     string     `gorm:"not null;size:20" json:"bed_number"`
	BedType       string     `gorm:"not null;size:30" json:"bed_type"`
	Status        string     `gorm:"not null;default:'available';size:30" json:"status"`
	LastCleanedAt *time.Time `json:"last_cleaned_at,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	Ward Ward `gorm:"foreignKey:WardID;references:WardID" json:"ward,omitempty"`
}

func (Bed) TableName() string { return "beds" }
