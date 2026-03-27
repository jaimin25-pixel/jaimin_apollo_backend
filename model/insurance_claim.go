package model

import "time"

// InsuranceClaim tracks insurance/TPA claim records.
// Status: draft, submitted, under_review, approved, rejected, settled.
type InsuranceClaim struct {
	ClaimID           uint       `gorm:"primaryKey;autoIncrement" json:"claim_id"`
	InvoiceID         uint       `gorm:"not null;index" json:"invoice_id"`
	PatientID         uint       `gorm:"not null;index" json:"patient_id"`
	InsuranceProvider string     `gorm:"not null;size:150" json:"insurance_provider"`
	PolicyNumber      string     `gorm:"not null;size:100" json:"policy_number"`
	ClaimAmount       float64    `gorm:"type:decimal(10,2);not null" json:"claim_amount"`
	ApprovedAmount    *float64   `gorm:"type:decimal(10,2)" json:"approved_amount,omitempty"`
	ICD10Codes        string     `gorm:"type:text" json:"icd10_codes,omitempty"`
	ProcedureCodes    string     `gorm:"type:text" json:"procedure_codes,omitempty"`
	Status            string     `gorm:"not null;default:'draft';size:20" json:"status"`
	SubmittedAt       *time.Time `json:"submitted_at,omitempty"`
	SettledAt         *time.Time `json:"settled_at,omitempty"`
	Remarks           string     `gorm:"type:text" json:"remarks,omitempty"`

	// Relations
	Patient Patient `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
}

func (InsuranceClaim) TableName() string { return "insurance_claims" }
