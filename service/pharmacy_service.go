package service

import (
	"errors"
	"fmt"
	"time"

	"apollo-backend/model"
	"apollo-backend/repository"
)

type PharmacyService struct {
	repo *repository.PharmacyRepo
}

func NewPharmacyService(repo *repository.PharmacyRepo) *PharmacyService {
	return &PharmacyService{repo: repo}
}

// ── Dashboard ──────────────────────────────────────────────────────────

func (s *PharmacyService) GetDashboard() (map[string]interface{}, error) {
	stats, err := s.repo.GetDashboardStats()
	if err != nil {
		return nil, err
	}

	prescriptions, _ := s.repo.ListPrescriptions("", "", "")
	if prescriptions == nil {
		prescriptions = []model.Prescription{}
	}

	alerts, _ := s.repo.GetLowStockAlerts()
	if alerts == nil {
		alerts = []model.StockAlert{}
	}

	return map[string]interface{}{
		"stats":          stats,
		"recent_pending": prescriptions,
		"low_stock_alerts": alerts,
	}, nil
}

// ── Prescriptions ──────────────────────────────────────────────────────

func (s *PharmacyService) ListPrescriptions(status, doctorID, date string) ([]model.Prescription, error) {
	if len(status) == 0 {
		status = ""
	}
	return s.repo.ListPrescriptions(status, doctorID, date)
}

func (s *PharmacyService) GetPrescription(rxID uint) (*model.Prescription, error) {
	if rxID == 0 {
		return nil, errors.New("invalid prescription ID")
	}
	return s.repo.GetPrescription(rxID)
}

// DispenseRequest represents a dispensing request
type DispenseRequest struct {
	Items []DispenseItem `json:"items"`
}

type DispenseItem struct {
	ItemID       uint    `json:"item_id" binding:"required"`
	QuantityDispensed float64 `json:"quantity_dispensed" binding:"required"`
	BatchID      uint    `json:"batch_id" binding:"required"`
}

// DispensePrescription handles partial/full dispensing
func (s *PharmacyService) DispensePrescription(rxID uint, pharmacistID uint, req DispenseRequest) error {
	if rxID == 0 || pharmacistID == 0 {
		return errors.New("invalid request")
	}

	rx, err := s.repo.GetPrescription(rxID)
	if err != nil {
		return errors.New("prescription not found")
	}

	if rx.Status != "pending" && rx.Status != "partially_dispensed" {
		return errors.New("prescription cannot be dispensed in current status")
	}

	// Track if prescription becomes fully dispensed
	allFullyDispensed := true
	rxStatus := "pending"

	for _, dispenseItem := range req.Items {
		// Get item details
		item, err := s.repo.GetPrescriptionItem(dispenseItem.ItemID)
		if err != nil {
			return fmt.Errorf("item %d not found", dispenseItem.ItemID)
		}

		// Validate quantity
		if dispenseItem.QuantityDispensed <= 0 || dispenseItem.QuantityDispensed > item.QuantityPrescribed {
			return fmt.Errorf("invalid quantity for item %d", dispenseItem.ItemID)
		}

		// Get batch and validate
		batch, err := s.repo.GetBatch(dispenseItem.BatchID)
		if err != nil {
			return fmt.Errorf("batch %d not found", dispenseItem.BatchID)
		}

		if batch.QuantityRemaining < dispenseItem.QuantityDispensed {
			return fmt.Errorf("insufficient stock in batch %s", batch.BatchNumber)
		}

		// Create dispense record
		record := &model.DispenseRecord{
			ItemID:            dispenseItem.ItemID,
			PharmacistID:      pharmacistID,
			BatchID:           dispenseItem.BatchID,
			QuantityDispensed: dispenseItem.QuantityDispensed,
			DispensedAt:       time.Now(),
		}
		if err := s.repo.CreateDispenseRecord(record); err != nil {
			return fmt.Errorf("failed to create dispense record: %v", err)
		}

		// Update batch quantity
		newBatchQty := batch.QuantityRemaining - dispenseItem.QuantityDispensed
		if err := s.repo.UpdateBatchQuantity(dispenseItem.BatchID, newBatchQty); err != nil {
			return fmt.Errorf("failed to update batch quantity: %v", err)
		}

		// Update item status
		newDispensedQty := item.QuantityDispensed + dispenseItem.QuantityDispensed
		itemStatus := "pending"
		if newDispensedQty >= item.QuantityPrescribed {
			itemStatus = "fully_dispensed"
		} else if newDispensedQty > 0 {
			itemStatus = "partially_dispensed"
		}

		if err := s.repo.UpdatePrescriptionItemStatus(dispenseItem.ItemID, itemStatus, newDispensedQty); err != nil {
			return fmt.Errorf("failed to update item status: %v", err)
		}

		if itemStatus != "fully_dispensed" {
			allFullyDispensed = false
		}
	}

	// Update prescription status
	if allFullyDispensed {
		rxStatus = "fully_dispensed"
	} else {
		rxStatus = "partially_dispensed"
	}

	return s.repo.UpdatePrescriptionStatus(rxID, rxStatus)
}

// CancelPrescription cancels a pending prescription
func (s *PharmacyService) CancelPrescription(rxID uint, reason string) error {
	if rxID == 0 {
		return errors.New("invalid prescription ID")
	}

	rx, err := s.repo.GetPrescription(rxID)
	if err != nil {
		return errors.New("prescription not found")
	}

	if rx.Status != "pending" && rx.Status != "partially_dispensed" {
		return errors.New("only pending/partially dispensed prescriptions can be cancelled")
	}

	return s.repo.UpdatePrescriptionStatus(rxID, "cancelled")
}

// ── Medicines ──────────────────────────────────────────────────────────

func (s *PharmacyService) ListMedicines(category, search string) ([]model.Medicine, error) {
	return s.repo.ListMedicines(category, search)
}

func (s *PharmacyService) GetMedicine(medicineID uint) (*model.Medicine, error) {
	if medicineID == 0 {
		return nil, errors.New("invalid medicine ID")
	}
	return s.repo.GetMedicine(medicineID)
}

type CreateMedicineRequest struct {
	GenericName       string  `json:"generic_name" binding:"required"`
	BrandName         string  `json:"brand_name"`
	Category          string  `json:"category" binding:"required"`
	Unit              string  `json:"unit" binding:"required"`
	HSNCode           string  `json:"hsn_code"`
	ReorderLevel      int     `json:"reorder_level" binding:"required"`
	StorageConditions string  `json:"storage_conditions"`
}

func (s *PharmacyService) CreateMedicine(req CreateMedicineRequest) (*model.Medicine, error) {
	if req.ReorderLevel < 0 {
		return nil, errors.New("reorder level cannot be negative")
	}

	medicine := &model.Medicine{
		GenericName:       req.GenericName,
		BrandName:         req.BrandName,
		Category:          req.Category,
		Unit:              req.Unit,
		HSNCode:           req.HSNCode,
		ReorderLevel:      req.ReorderLevel,
		CurrentStock:      0,
		StorageConditions: req.StorageConditions,
		Status:            "active",
	}

	if err := s.repo.CreateMedicine(medicine); err != nil {
		return nil, err
	}

	return medicine, nil
}

type UpdateMedicineRequest struct {
	GenericName       string `json:"generic_name"`
	BrandName         string `json:"brand_name"`
	Category          string `json:"category"`
	Unit              string `json:"unit"`
	HSNCode           string `json:"hsn_code"`
	ReorderLevel      int    `json:"reorder_level"`
	StorageConditions string `json:"storage_conditions"`
}

func (s *PharmacyService) UpdateMedicine(medicineID uint, req UpdateMedicineRequest) error {
	if medicineID == 0 {
		return errors.New("invalid medicine ID")
	}

	updates := make(map[string]interface{})
	if req.GenericName != "" {
		updates["generic_name"] = req.GenericName
	}
	if req.BrandName != "" {
		updates["brand_name"] = req.BrandName
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Unit != "" {
		updates["unit"] = req.Unit
	}
	if req.HSNCode != "" {
		updates["hsn_code"] = req.HSNCode
	}
	if req.ReorderLevel > 0 {
		updates["reorder_level"] = req.ReorderLevel
	}
	if req.StorageConditions != "" {
		updates["storage_conditions"] = req.StorageConditions
	}

	return s.repo.UpdateMedicine(medicineID, updates)
}

// ── Medicine Batches ───────────────────────────────────────────────────

func (s *PharmacyService) ListBatchesForMedicine(medicineID uint) ([]model.MedicineBatch, error) {
	if medicineID == 0 {
		return nil, errors.New("invalid medicine ID")
	}
	return s.repo.ListBatchesForMedicine(medicineID)
}

type CreateBatchRequest struct {
	BatchNumber   string  `json:"batch_number" binding:"required"`
	ExpiryDate    string  `json:"expiry_date" binding:"required"`
	Quantity      float64 `json:"quantity" binding:"required"`
	Supplier      string  `json:"supplier"`
	PurchaseDate  string  `json:"purchase_date" binding:"required"`
	PurchasePrice *float64 `json:"purchase_price"`
}

func (s *PharmacyService) CreateBatch(medicineID uint, req CreateBatchRequest) (*model.MedicineBatch, error) {
	if medicineID == 0 {
		return nil, errors.New("invalid medicine ID")
	}

	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return nil, errors.New("invalid expiry date format")
	}

	purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return nil, errors.New("invalid purchase date format")
	}

	batch := &model.MedicineBatch{
		MedicineID:        medicineID,
		BatchNumber:       req.BatchNumber,
		ExpiryDate:        expiryDate,
		Quantity:          req.Quantity,
		QuantityRemaining: req.Quantity,
		Supplier:          req.Supplier,
		PurchaseDate:      purchaseDate,
		PurchasePrice:     req.PurchasePrice,
	}

	if err := s.repo.CreateBatch(batch); err != nil {
		return nil, err
	}

	// Update medicine stock
	med, _ := s.repo.GetMedicine(medicineID)
	if med != nil {
		s.repo.UpdateMedicine(medicineID, map[string]interface{}{
			"current_stock": med.CurrentStock + req.Quantity,
		})
	}

	return batch, nil
}

// ── Stock Adjustments ──────────────────────────────────────────────────

type AdjustStockRequest struct {
	AdjustmentType string  `json:"adjustment_type" binding:"required"`
	Quantity       float64 `json:"quantity" binding:"required"`
	Reason         string  `json:"reason"`
	BatchID        *uint   `json:"batch_id"`
}

func (s *PharmacyService) AdjustStock(medicineID, pharmacistID uint, req AdjustStockRequest) error {
	if medicineID == 0 || pharmacistID == 0 {
		return errors.New("invalid request")
	}

	if req.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	adjTypes := map[string]bool{"add": true, "write-off": true, "return": true}
	if !adjTypes[req.AdjustmentType] {
		return errors.New("invalid adjustment type")
	}

	adjustment := &model.StockAdjustment{
		MedicineID:     medicineID,
		BatchID:        req.BatchID,
		PharmacistID:   pharmacistID,
		AdjustmentType: req.AdjustmentType,
		Quantity:       req.Quantity,
		Reason:         req.Reason,
	}

	return s.repo.CreateStockAdjustment(adjustment)
}

func (s *PharmacyService) ListStockAdjustments(medicineID uint) ([]model.StockAdjustment, error) {
	return s.repo.ListStockAdjustments(medicineID)
}

// ── Alerts ─────────────────────────────────────────────────────────────

func (s *PharmacyService) GetLowStockAlerts() ([]model.StockAlert, error) {
	return s.repo.GetLowStockAlerts()
}

func (s *PharmacyService) GetExpiringAlerts(days int) ([]model.StockAlert, error) {
	if days <= 0 {
		days = 60
	}
	return s.repo.GetExpiringAlerts(days)
}

func (s *PharmacyService) GetLowStockMedicines() ([]model.Medicine, error) {
	return s.repo.GetLowStockMedicines()
}

func (s *PharmacyService) GetExpiringMedicines(days int) ([]model.MedicineBatch, error) {
	if days <= 0 {
		days = 60
	}
	return s.repo.GetExpiringMedicines(days)
}

// ── Partner Pharmacies ─────────────────────────────────────────────────

func (s *PharmacyService) ListPartnerPharmacies() ([]model.PartnerPharmacy, error) {
	return s.repo.ListPartnerPharmacies()
}

// ── Transfer Requests ──────────────────────────────────────────────────

type CreateTransferRequest struct {
	PartnerID  uint    `json:"partner_id" binding:"required"`
	MedicineID uint    `json:"medicine_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required"`
	Reason     string  `json:"reason"`
}

func (s *PharmacyService) CreateTransferRequest(pharmacistID uint, req CreateTransferRequest) (*model.TransferRequest, error) {
	if pharmacistID == 0 || req.PartnerID == 0 || req.MedicineID == 0 {
		return nil, errors.New("invalid request")
	}

	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	transfer := &model.TransferRequest{
		MedicineID:   req.MedicineID,
		PartnerID:    req.PartnerID,
		PharmacistID: pharmacistID,
		Quantity:     req.Quantity,
		Reason:       req.Reason,
		Status:       "pending",
	}

	if err := s.repo.CreateTransferRequest(transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

func (s *PharmacyService) ListTransferRequests(status string) ([]model.TransferRequest, error) {
	return s.repo.ListTransferRequests(status)
}

func (s *PharmacyService) UpdateTransferRequestStatus(transferID uint, status string) error {
	if transferID == 0 {
		return errors.New("invalid transfer ID")
	}

	validStatuses := map[string]bool{"pending": true, "accepted": true, "rejected": true, "completed": true}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	return s.repo.UpdateTransferRequestStatus(transferID, status)
}
