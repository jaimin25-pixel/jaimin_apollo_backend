package repository

import (
	"apollo-backend/model"
	"time"

	"gorm.io/gorm"
)

type FinanceRepo interface {
	GetDashboardStats() (map[string]interface{}, error)
	ListInvoices(status string, patientID *uint) ([]model.Invoice, error)
	CreateInvoice(invoice *model.Invoice) error
	GetInvoice(id uint) (*model.Invoice, error)
	UpdateInvoiceStatus(id uint, status string) error

	ListClaims(status string, patientID *uint) ([]model.InsuranceClaim, error)
	CreateClaim(claim *model.InsuranceClaim) error
	UpdateClaimStatus(id uint, status string, notes string) error

	ListLedger(transactionType string) ([]model.LedgerEntry, error)
	AddLedgerEntry(entry *model.LedgerEntry) error
}

type financeRepo struct {
	db *gorm.DB
}

func NewFinanceRepo(db *gorm.DB) FinanceRepo {
	return &financeRepo{db}
}

func (r *financeRepo) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	today := time.Now().Truncate(24 * time.Hour)

	var dailyRevenue float64
	r.db.Model(&model.LedgerEntry{}).Where("transaction_type = 'credit' AND recorded_at >= ?", today).Select("COALESCE(SUM(amount), 0)").Scan(&dailyRevenue)

	var pendingClaims, outstandingInvoices int64
	r.db.Model(&model.InsuranceClaim{}).Where("status = ?", "pending").Count(&pendingClaims)
	r.db.Model(&model.Invoice{}).Where("status = ?", "unpaid").Count(&outstandingInvoices)

	stats["daily_revenue_today"] = dailyRevenue
	stats["pending_claims"] = pendingClaims
	stats["outstanding_invoices"] = outstandingInvoices

	return stats, nil
}

func (r *financeRepo) ListInvoices(status string, patientID *uint) ([]model.Invoice, error) {
	var invoices []model.Invoice
	q := r.db.Preload("Patient")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if patientID != nil && *patientID > 0 {
		q = q.Where("patient_id = ?", *patientID)
	}
	err := q.Order("created_at desc").Find(&invoices).Error
	return invoices, err
}

func (r *financeRepo) CreateInvoice(invoice *model.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *financeRepo) GetInvoice(id uint) (*model.Invoice, error) {
	var invoice model.Invoice
	err := r.db.Preload("Patient").First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *financeRepo) UpdateInvoiceStatus(id uint, status string) error {
	return r.db.Model(&model.Invoice{}).Where("invoice_id = ?", id).Update("status", status).Error
}

func (r *financeRepo) ListClaims(status string, patientID *uint) ([]model.InsuranceClaim, error) {
	var claims []model.InsuranceClaim
	q := r.db.Preload("Patient").Preload("Invoice")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if patientID != nil && *patientID > 0 {
		q = q.Where("patient_id = ?", *patientID)
	}
	err := q.Order("created_at desc").Find(&claims).Error
	return claims, err
}

func (r *financeRepo) CreateClaim(claim *model.InsuranceClaim) error {
	return r.db.Create(claim).Error
}

func (r *financeRepo) UpdateClaimStatus(id uint, status string, notes string) error {
	updates := map[string]interface{}{
		"status": status,
		"notes":  notes,
	}
	if status == "approved" {
		updates["approved_amount"] = gorm.Expr("claim_amount") // simple mock
	}
	return r.db.Model(&model.InsuranceClaim{}).Where("claim_id = ?", id).Updates(updates).Error
}

func (r *financeRepo) ListLedger(transactionType string) ([]model.LedgerEntry, error) {
	var records []model.LedgerEntry
	q := r.db.Preload("RecordedByStaff")
	if transactionType != "" {
		q = q.Where("transaction_type = ?", transactionType)
	}
	err := q.Order("recorded_at desc").Find(&records).Error
	return records, err
}

func (r *financeRepo) AddLedgerEntry(entry *model.LedgerEntry) error {
	return r.db.Create(entry).Error
}
