package model

import "time"

// Invoice stores billing records generated at OPD completion or IPD discharge.
// Status: draft, finalized, paid, partially_paid, cancelled.
type Invoice struct {
	InvoiceID            uint       `gorm:"primaryKey;autoIncrement" json:"invoice_id"`
	PatientID            uint       `gorm:"not null;index" json:"patient_id"`
	AdmissionID          *uint      `gorm:"index" json:"admission_id,omitempty"`
	ApptID               *uint      `gorm:"index" json:"appt_id,omitempty"`
	ConsultationCharges  float64    `gorm:"type:decimal(10,2);not null;default:0" json:"consultation_charges"`
	ProcedureCharges     float64    `gorm:"type:decimal(10,2);not null;default:0" json:"procedure_charges"`
	LabCharges           float64    `gorm:"type:decimal(10,2);not null;default:0" json:"lab_charges"`
	PharmacyCharges      float64    `gorm:"type:decimal(10,2);not null;default:0" json:"pharmacy_charges"`
	BedCharges           float64    `gorm:"type:decimal(10,2);not null;default:0" json:"bed_charges"`
	MiscellaneousCharges float64    `gorm:"type:decimal(10,2);not null;default:0" json:"miscellaneous_charges"`
	SubTotal             float64    `gorm:"type:decimal(10,2);not null;default:0" json:"sub_total"`
	TaxAmount            float64    `gorm:"type:decimal(10,2);not null;default:0" json:"tax_amount"`
	TotalAmount          float64    `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	AmountPaid           float64    `gorm:"type:decimal(10,2);not null;default:0" json:"amount_paid"`
	BalanceDue           float64    `gorm:"type:decimal(10,2);not null;default:0" json:"balance_due"`
	PaymentMode          string     `gorm:"size:30" json:"payment_mode,omitempty"`
	InsuranceClaimID     *uint      `gorm:"index" json:"insurance_claim_id,omitempty"`
	Status               string     `gorm:"not null;default:'draft';size:20" json:"status"`
	CreatedBy            *uint      `gorm:"index" json:"created_by,omitempty"`
	CreatedAt            time.Time  `gorm:"not null;default:now()" json:"created_at"`
	FinalizedAt          *time.Time `json:"finalized_at,omitempty"`

	// Relations
	Patient        Patient          `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Admission      *Admission       `gorm:"foreignKey:AdmissionID;references:AdmissionID" json:"admission,omitempty"`
	Appointment    *Appointment     `gorm:"foreignKey:ApptID;references:ApptID" json:"appointment,omitempty"`
	InsuranceClaim *InsuranceClaim  `gorm:"foreignKey:InsuranceClaimID;references:ClaimID" json:"insurance_claim,omitempty"`
	CreatedByStaff *Staff           `gorm:"foreignKey:CreatedBy;references:StaffID" json:"created_by_staff,omitempty"`
}

func (Invoice) TableName() string { return "invoices" }
