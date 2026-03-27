package model

// Ward defines ward/unit within each department.
// Types: general, icu, nicu, picu, ccu, ot, emergency.
type Ward struct {
	WardID   uint   `gorm:"primaryKey;autoIncrement" json:"ward_id"`
	DeptID   uint   `gorm:"not null;index" json:"dept_id"`
	Name     string `gorm:"not null;size:100" json:"name"`
	WardType string `gorm:"not null;size:30" json:"ward_type"`
	Capacity int    `gorm:"not null;default:0" json:"capacity"`
	Status   string `gorm:"not null;default:'active';size:20" json:"status"`

	// Relations
	Department Department `gorm:"foreignKey:DeptID;references:DeptID" json:"department,omitempty"`
}

func (Ward) TableName() string { return "wards" }
