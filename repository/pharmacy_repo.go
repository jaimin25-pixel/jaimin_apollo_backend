package repository

import (
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

type PharmacyRepo struct{ DB *gorm.DB }

func NewPharmacyRepo(db *gorm.DB) *PharmacyRepo { return &PharmacyRepo{DB: db} }

// ── Prescriptions ──────────────────────────────────────────────────────

// ListPrescriptions returns prescriptions filterable by status, doctor, date
func (r *PharmacyRepo) ListPrescriptions(status, doctorIDStr string, date string) ([]model.Prescription, error) {
	q := r.DB.Preload("Patient").Preload("Doctor").Preload("Items").Preload("Items.Medicine")

	if status != "" {
		q = q.Where("status IN (?)", []string{"pending", "partially_dispensed"})
	}
	if doctorIDStr != "" {
		q = q.Where("doctor_id = ?", doctorIDStr)
	}
	if date != "" {
		d, err := time.Parse("2006-01-02", date)
		if err == nil {
			q = q.Where("created_at >= ? AND created_at < ?", d, d.Add(24*time.Hour))
		}
	}

	var prescriptions []model.Prescription
	err := q.Order("created_at DESC").Find(&prescriptions).Error
	return prescriptions, err
}

// GetPrescription returns detailed prescription with all items
func (r *PharmacyRepo) GetPrescription(rxID uint) (*model.Prescription, error) {
	var rx model.Prescription
	err := r.DB.Preload("Patient").Preload("Doctor").Preload("Items").Preload("Items.Medicine").
		First(&rx, rxID).Error
	return &rx, err
}

// UpdatePrescriptionStatus updates prescription status
func (r *PharmacyRepo) UpdatePrescriptionStatus(rxID uint, status string) error {
	return r.DB.Model(&model.Prescription{}).
		Where("rx_id = ?", rxID).
		Update("status", status).Error
}

// ── Prescription Items ─────────────────────────────────────────────────

// GetPrescriptionItem returns a single prescription item
func (r *PharmacyRepo) GetPrescriptionItem(itemID uint) (*model.PrescriptionItem, error) {
	var item model.PrescriptionItem
	err := r.DB.Preload("Medicine").First(&item, itemID).Error
	return &item, err
}

// UpdatePrescriptionItemStatus updates item status
func (r *PharmacyRepo) UpdatePrescriptionItemStatus(itemID uint, status string, quantityDispensed float64) error {
	return r.DB.Model(&model.PrescriptionItem{}).
		Where("item_id = ?", itemID).
		Updates(map[string]interface{}{
			"status":             status,
			"quantity_dispensed": quantityDispensed,
		}).Error
}

// ── Medicines ──────────────────────────────────────────────────────────

// ListMedicines returns medicines filterable by category, stock level
func (r *PharmacyRepo) ListMedicines(category, search string) ([]model.Medicine, error) {
	q := r.DB.Where("status = 'active'")

	if category != "" {
		q = q.Where("category = ?", category)
	}
	if search != "" {
		q = q.Where("generic_name ILIKE ? OR brand_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var medicines []model.Medicine
	err := q.Order("generic_name").Find(&medicines).Error
	return medicines, err
}

// GetMedicine returns medicine with all batches
func (r *PharmacyRepo) GetMedicine(medicineID uint) (*model.Medicine, error) {
	var med model.Medicine
	err := r.DB.Preload("Batches", func(db *gorm.DB) *gorm.DB {
		return db.Where("quantity_remaining > 0").Order("expiry_date ASC")
	}).First(&med, medicineID).Error
	return &med, err
}

// CreateMedicine creates a new medicine
func (r *PharmacyRepo) CreateMedicine(med *model.Medicine) error {
	return r.DB.Create(med).Error
}

// UpdateMedicine updates medicine catalog
func (r *PharmacyRepo) UpdateMedicine(medicineID uint, updates map[string]interface{}) error {
	return r.DB.Model(&model.Medicine{}).
		Where("medicine_id = ?", medicineID).
		Updates(updates).Error
}

// ── Medicine Batches ───────────────────────────────────────────────────

// ListBatchesForMedicine returns all batches for a medicine
func (r *PharmacyRepo) ListBatchesForMedicine(medicineID uint) ([]model.MedicineBatch, error) {
	var batches []model.MedicineBatch
	err := r.DB.Where("medicine_id = ? AND quantity_remaining > 0", medicineID).
		Order("expiry_date ASC").
		Find(&batches).Error
	return batches, err
}

// CreateBatch creates a new stock batch
func (r *PharmacyRepo) CreateBatch(batch *model.MedicineBatch) error {
	return r.DB.Create(batch).Error
}

// GetBatch returns a specific batch
func (r *PharmacyRepo) GetBatch(batchID uint) (*model.MedicineBatch, error) {
	var batch model.MedicineBatch
	err := r.DB.Preload("Medicine").First(&batch, batchID).Error
	return &batch, err
}

// UpdateBatchQuantity updates remaining quantity in a batch
func (r *PharmacyRepo) UpdateBatchQuantity(batchID uint, newQuantity float64) error {
	return r.DB.Model(&model.MedicineBatch{}).
		Where("batch_id = ?", batchID).
		Update("quantity_remaining", newQuantity).Error
}

// ── Stock Adjustments ──────────────────────────────────────────────────

// CreateStockAdjustment logs a manual stock adjustment
func (r *PharmacyRepo) CreateStockAdjustment(adj *model.StockAdjustment) error {
	if err := r.DB.Create(adj).Error; err != nil {
		return err
	}

	// Update medicine current_stock
	medicine, err := r.GetMedicine(adj.MedicineID)
	if err != nil {
		return err
	}

	var newStock float64
	switch adj.AdjustmentType {
	case "add":
		newStock = medicine.CurrentStock + adj.Quantity
	case "write-off", "return":
		newStock = medicine.CurrentStock - adj.Quantity
		if newStock < 0 {
			newStock = 0
		}
	}

	return r.UpdateMedicine(adj.MedicineID, map[string]interface{}{
		"current_stock": newStock,
	})
}

// ListStockAdjustments returns adjustment history
func (r *PharmacyRepo) ListStockAdjustments(medicineID uint) ([]model.StockAdjustment, error) {
	var adjustments []model.StockAdjustment
	q := r.DB.Preload("Medicine").Preload("Pharmacist")

	if medicineID > 0 {
		q = q.Where("medicine_id = ?", medicineID)
	}

	err := q.Order("created_at DESC").Find(&adjustments).Error
	return adjustments, err
}

// ── Stock Alerts ───────────────────────────────────────────────────────

// CreateStockAlert creates a low stock or expiry alert
func (r *PharmacyRepo) CreateStockAlert(alert *model.StockAlert) error {
	return r.DB.Create(alert).Error
}

// GetLowStockAlerts returns medicines below reorder level
func (r *PharmacyRepo) GetLowStockAlerts() ([]model.StockAlert, error) {
	var alerts []model.StockAlert
	err := r.DB.
		Joins("INNER JOIN medicines ON medicines.medicine_id = stock_alerts.medicine_id").
		Where("stock_alerts.alert_type = ? AND stock_alerts.status = ?", "low_stock", "triggered").
		Preload("Medicine").
		Order("stock_alerts.triggered_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// GetExpiringAlerts returns medicines expiring within specified days
func (r *PharmacyRepo) GetExpiringAlerts(days int) ([]model.StockAlert, error) {
	var alerts []model.StockAlert
	expireThreshold := time.Now().AddDate(0, 0, days)
	err := r.DB.
		Joins("INNER JOIN medicine_batches ON medicine_batches.batch_id = stock_alerts.batch_id").
		Where("stock_alerts.alert_type = ? AND stock_alerts.status = ? AND medicine_batches.expiry_date <= ?",
			"expiring", "triggered", expireThreshold).
		Preload("Medicine").
		Preload("Batch").
		Order("stock_alerts.triggered_at DESC").
		Find(&alerts).Error
	return alerts, err
}

// ResolveAlert marks an alert as resolved
func (r *PharmacyRepo) ResolveAlert(alertID uint) error {
	return r.DB.Model(&model.StockAlert{}).
		Where("alert_id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": time.Now(),
		}).Error
}

// ── Dispense Records ───────────────────────────────────────────────────

// CreateDispenseRecord logs a dispensing event
func (r *PharmacyRepo) CreateDispenseRecord(record *model.DispenseRecord) error {
	return r.DB.Create(record).Error
}

// GetDispenseHistoryForItem returns dispensing records for a prescription item
func (r *PharmacyRepo) GetDispenseHistoryForItem(itemID uint) ([]model.DispenseRecord, error) {
	var records []model.DispenseRecord
	err := r.DB.Where("item_id = ?", itemID).
		Preload("Pharmacist").
		Preload("Batch").
		Order("dispensed_at DESC").
		Find(&records).Error
	return records, err
}

// ── Partner Pharmacies ─────────────────────────────────────────────────

// ListPartnerPharmacies returns active partner pharmacies
func (r *PharmacyRepo) ListPartnerPharmacies() ([]model.PartnerPharmacy, error) {
	var partners []model.PartnerPharmacy
	err := r.DB.Where("status = 'active'").Order("name").Find(&partners).Error
	return partners, err
}

// ── Transfer Requests ──────────────────────────────────────────────────

// CreateTransferRequest creates an inter-pharmacy transfer request
func (r *PharmacyRepo) CreateTransferRequest(req *model.TransferRequest) error {
	return r.DB.Create(req).Error
}

// ListTransferRequests returns transfer requests with optional filtering
func (r *PharmacyRepo) ListTransferRequests(status string) ([]model.TransferRequest, error) {
	var requests []model.TransferRequest
	q := r.DB.Preload("Medicine").Preload("Partner").Preload("Pharmacist")

	if status != "" {
		q = q.Where("status = ?", status)
	}

	err := q.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// UpdateTransferRequestStatus updates transfer request status
func (r *PharmacyRepo) UpdateTransferRequestStatus(transferID uint, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == "completed" {
		updates["completed_at"] = time.Now()
	}
	return r.DB.Model(&model.TransferRequest{}).
		Where("transfer_id = ?", transferID).
		Updates(updates).Error
}

// ── Dashboard & Analytics ─────────────────────────────────────────────

// GetDashboardStats returns pharmacy dashboard statistics
func (r *PharmacyRepo) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Pending prescriptions
	var pendingCount int64
	r.DB.Model(&model.Prescription{}).Where("status IN ?", []string{"pending", "partially_dispensed"}).Count(&pendingCount)
	stats["pending_prescriptions"] = pendingCount

	// Low stock medications
	var lowStockCount int64
	r.DB.Model(&model.Medicine{}).
		Where("current_stock <= reorder_level AND status = 'active'").
		Count(&lowStockCount)
	stats["low_stock_medicines"] = lowStockCount

	// Expiring medicines (within 60 days)
	var expiringCount int64
	threshold := time.Now().AddDate(0, 0, 60)
	r.DB.Model(&model.MedicineBatch{}).
		Where("expiry_date <= ? AND quantity_remaining > 0", threshold).
		Count(&expiringCount)
	stats["expiring_medicines"] = expiringCount

	// Total medicines
	var totalMedicines int64
	r.DB.Model(&model.Medicine{}).Where("status = 'active'").Count(&totalMedicines)
	stats["total_medicines"] = totalMedicines

	return stats, nil
}

// GetLowStockMedicines returns medicines below reorder level
func (r *PharmacyRepo) GetLowStockMedicines() ([]model.Medicine, error) {
	var medicines []model.Medicine
	err := r.DB.
		Where("current_stock <= reorder_level AND status = 'active'").
		Order("current_stock ASC").
		Find(&medicines).Error
	return medicines, err
}

// GetExpiringMedicines returns batches expiring within specified days
func (r *PharmacyRepo) GetExpiringMedicines(days int) ([]model.MedicineBatch, error) {
	var batches []model.MedicineBatch
	threshold := time.Now().AddDate(0, 0, days)
	err := r.DB.
		Where("expiry_date <= ? AND quantity_remaining > 0", threshold).
		Preload("Medicine").
		Order("expiry_date ASC").
		Find(&batches).Error
	return batches, err
}
