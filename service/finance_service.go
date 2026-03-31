package service

import (
	"apollo-backend/model"
	"apollo-backend/repository"
	"errors"
	"strconv"
)

type FinanceService interface {
	GetDashboard() (map[string]interface{}, error)
	ListInvoices(status string, patientID *uint) ([]model.Invoice, error)
	CreateInvoice(invoice *model.Invoice) error
	GetInvoice(id uint) (*model.Invoice, error)
	MarkInvoiceStatus(id uint, status string, method string, ref string, staffID uint) error

	ListClaims(status string, patientID *uint) ([]model.InsuranceClaim, error)
	CreateClaim(claim *model.InsuranceClaim) error
	UpdateClaimStatus(id uint, status string, notes string) error

	ListLedger(transactionType string) ([]model.LedgerEntry, error)
}

type financeService struct {
	repo repository.FinanceRepo
}

func NewFinanceService(repo repository.FinanceRepo) FinanceService {
	return &financeService{repo}
}

func (s *financeService) GetDashboard() (map[string]interface{}, error) {
	return s.repo.GetDashboardStats()
}

func (s *financeService) ListInvoices(status string, patientID *uint) ([]model.Invoice, error) {
	return s.repo.ListInvoices(status, patientID)
}

func (s *financeService) CreateInvoice(invoice *model.Invoice) error {
	if invoice.PatientID == 0 || invoice.TotalAmount <= 0 {
		return errors.New("patient_id and total_amount > 0 are required")
	}
	invoice.Status = "unpaid"
	return s.repo.CreateInvoice(invoice)
}

func (s *financeService) GetInvoice(id uint) (*model.Invoice, error) {
	return s.repo.GetInvoice(id)
}

func (s *financeService) MarkInvoiceStatus(id uint, status string, method string, ref string, staffID uint) error {
	// If marking paid, add ledger entry
	if status == "paid" {
		inv, err := s.repo.GetInvoice(id)
		if err != nil {
			return err
		}
		if inv.Status == "paid" {
			return errors.New("invoice already paid")
		}

		refStr := strconv.FormatUint(uint64(id), 10)
		entry := model.LedgerEntry{
			TransactionType: "Invoice_Payment",
			EntryType:       "credit",
			Amount:          inv.TotalAmount,
			ReferenceID:     &refStr,
			Description:     "Payment for Invoice #" + ref,
			RecordedBy:      &staffID,
		}
		if err := s.repo.AddLedgerEntry(&entry); err != nil {
			return err
		}
	}
	return s.repo.UpdateInvoiceStatus(id, status)
}

func (s *financeService) ListClaims(status string, patientID *uint) ([]model.InsuranceClaim, error) {
	return s.repo.ListClaims(status, patientID)
}

func (s *financeService) CreateClaim(claim *model.InsuranceClaim) error {
	if claim.InvoiceID == 0 || claim.PatientID == 0 {
		return errors.New("invoice_id and patient_id required")
	}
	claim.Status = "pending"
	return s.repo.CreateClaim(claim)
}

func (s *financeService) UpdateClaimStatus(id uint, status string, notes string) error {
	valid := map[string]bool{"approved": true, "rejected": true, "pending": true}
	if !valid[status] {
		return errors.New("invalid status variant")
	}
	return s.repo.UpdateClaimStatus(id, status, notes)
}

func (s *financeService) ListLedger(transactionType string) ([]model.LedgerEntry, error) {
	return s.repo.ListLedger(transactionType)
}
