package repository

import (
	"apollo-backend/model"
	"time"

	"gorm.io/gorm"
)

type NursingRepo interface {
	GetDashboardStats(wardID *uint) (map[string]interface{}, error)
	ListWards(deptID *uint) ([]model.Ward, error)
	GetBedsByWard(wardID uint, status string) ([]model.Bed, error)
	UpdateBedStatus(bedID uint, status string) error
	InitiateTransfer(req *model.BedTransferRequest) error
	ConfirmTransfer(transferID uint, notes string) error
	MarkBedClean(bedID uint) error
	
	ListVitals(patientID uint, fromDate string) ([]model.Vital, error)
	RecordVital(vital *model.Vital) error
	
	GetMAR(patientID uint, date string) ([]model.MedicationAdminRecord, error)
	UpdateMARStatus(marID uint, status, reason, notes string, nurseID uint) error
	
	AddNursingNote(note *model.NursingNote) error
}

type nursingRepo struct {
	db *gorm.DB
}

func NewNursingRepo(db *gorm.DB) NursingRepo {
	return &nursingRepo{db}
}

func (r *nursingRepo) GetDashboardStats(wardID *uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	var totalBeds, occupiedBeds, availableBeds, pendingMARs int64

	bedQ := r.db.Model(&model.Bed{})
	if wardID != nil && *wardID > 0 {
		bedQ = bedQ.Where("ward_id = ?", *wardID)
	}
	bedQ.Count(&totalBeds)
	bedQ.Where("status = ?", "occupied").Count(&occupiedBeds)
	bedQ.Where("status = ?", "available").Count(&availableBeds)

	// Pending MARs for today
	today := time.Now().Truncate(24 * time.Hour)
	marQ := r.db.Model(&model.MedicationAdminRecord{}).Where("status = ?", "scheduled").Where("scheduled_at >= ? AND scheduled_at < ?", today, today.Add(24*time.Hour))
	if wardID != nil && *wardID > 0 {
		marQ = marQ.Joins("JOIN admissions ON admissions.admission_id = medication_admin_records.admission_id").Where("admissions.ward_id = ?", *wardID)
	}
	marQ.Count(&pendingMARs)

	stats["total_beds"] = totalBeds
	stats["occupied_beds"] = occupiedBeds
	stats["available_beds"] = availableBeds
	stats["pending_mars_today"] = pendingMARs

	return stats, nil
}

func (r *nursingRepo) ListWards(deptID *uint) ([]model.Ward, error) {
	var wards []model.Ward
	q := r.db.Preload("Department").Where("status = ?", "active")
	if deptID != nil && *deptID > 0 {
		q = q.Where("dept_id = ?", *deptID)
	}
	err := q.Find(&wards).Error
	return wards, err
}

func (r *nursingRepo) GetBedsByWard(wardID uint, status string) ([]model.Bed, error) {
	var beds []model.Bed
	q := r.db.Preload("Ward").Where("ward_id = ?", wardID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&beds).Error
	return beds, err
}

func (r *nursingRepo) UpdateBedStatus(bedID uint, status string) error {
	return r.db.Model(&model.Bed{}).Where("bed_id = ?", bedID).Update("status", status).Error
}

func (r *nursingRepo) InitiateTransfer(req *model.BedTransferRequest) error {
	return r.db.Create(req).Error
}

func (r *nursingRepo) ConfirmTransfer(transferID uint, notes string) error {
	// A bit simplified: actually needs a tx to update old bed, new bed, and admission
	return r.db.Transaction(func(tx *gorm.DB) error {
		var req model.BedTransferRequest
		if err := tx.First(&req, transferID).Error; err != nil {
			return err
		}
		
		req.Status = "completed"
		req.ApprovalNotes = notes
		now := time.Now()
		req.UpdatedAt = now
		if err := tx.Save(&req).Error; err != nil {
			return err
		}

		// Update old bed
		if req.FromBedID != nil {
			if err := tx.Model(&model.Bed{}).Where("bed_id = ?", *req.FromBedID).Update("status", "maintenance").Error; err != nil {
				return err
			}
		}
		// Update new bed
		if req.ToBedID != 0 {
			if err := tx.Model(&model.Bed{}).Where("bed_id = ?", req.ToBedID).Update("status", "occupied").Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *nursingRepo) MarkBedClean(bedID uint) error {
	now := time.Now()
	return r.db.Model(&model.Bed{}).Where("bed_id = ?", bedID).Updates(map[string]interface{}{
		"status":          "available",
		"last_cleaned_at": now,
	}).Error
}

func (r *nursingRepo) ListVitals(patientID uint, fromDate string) ([]model.Vital, error) {
	var vitals []model.Vital
	q := r.db.Preload("Nurse").Preload("Patient").Where("patient_id = ?", patientID)
	if fromDate != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			q = q.Where("recorded_at >= ?", t)
		}
	}
	err := q.Order("recorded_at desc").Find(&vitals).Error
	return vitals, err
}

func (r *nursingRepo) RecordVital(vital *model.Vital) error {
	return r.db.Create(vital).Error
}

func (r *nursingRepo) GetMAR(patientID uint, date string) ([]model.MedicationAdminRecord, error) {
	var mars []model.MedicationAdminRecord
	q := r.db.Preload("PrescriptionItem.Medicine").Preload("Nurse").Where("patient_id = ?", patientID)
	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			q = q.Where("scheduled_at >= ? AND scheduled_at < ?", t, t.Add(24*time.Hour))
		}
	}
	err := q.Order("scheduled_at asc").Find(&mars).Error
	return mars, err
}

func (r *nursingRepo) UpdateMARStatus(marID uint, status, reason, notes string, nurseID uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":   status,
		"reason":   reason,
		"notes":    notes,
		"nurse_id": nurseID,
	}
	if status == "administered" {
		updates["administered_at"] = now
	}
	return r.db.Model(&model.MedicationAdminRecord{}).Where("mar_id = ?", marID).Updates(updates).Error
}

func (r *nursingRepo) AddNursingNote(note *model.NursingNote) error {
	return r.db.Create(note).Error
}
