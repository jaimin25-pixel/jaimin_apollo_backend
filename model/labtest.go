package model

// LabTest is the master catalog of available laboratory tests.
type LabTest struct {
	TestID          uint    `gorm:"primaryKey;autoIncrement" json:"test_id"`
	TestName        string  `gorm:"not null;size:150" json:"test_name"`
	TestCode        string  `gorm:"uniqueIndex;not null;size:30" json:"test_code"`
	SampleType      string  `gorm:"not null;size:50" json:"sample_type"`
	NormalRange     string  `gorm:"type:text" json:"normal_range,omitempty"`
	TurnaroundHours *int    `json:"turnaround_hours,omitempty"`
	Charges         float64 `gorm:"type:decimal(10,2);not null;default:0" json:"charges"`
	DeptID          *uint   `gorm:"index" json:"dept_id,omitempty"`
	Status          string  `gorm:"not null;default:'active';size:20" json:"status"`

	// Relations
	Department *Department `gorm:"foreignKey:DeptID;references:DeptID" json:"department,omitempty"`
}

func (LabTest) TableName() string { return "lab_tests" }
