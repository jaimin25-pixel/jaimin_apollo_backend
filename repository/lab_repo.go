package repository

import (
	"apollo-backend/model"
	"time"

	"gorm.io/gorm"
)

type LabRepo interface {
	GetDashboardStats() (map[string]interface{}, error)
	ListTests(search, status string) ([]model.LabTest, error)
	AddTest(test *model.LabTest) error
	UpdateTest(testID uint, updates map[string]interface{}) error

	ListLabOrders(status string, fromDate string, patientID *uint) ([]model.LabOrder, error)
	GetLabOrder(id uint) (*model.LabOrder, error)
	UpdateLabOrderStatus(id uint, status string) error
	CollectSample(id uint, time time.Time, techID uint) error
	UploadResult(id uint, result string, isAbnormal *bool, notes string, techID uint) error

	ListRadiologyOrders(status string, modality string) ([]model.RadiologyOrder, error)
	GetRadiologyOrder(id uint) (*model.RadiologyOrder, error)
	UpdateRadiologyOrder(id uint, updates map[string]interface{}) error
}

type labRepo struct {
	db *gorm.DB
}

func NewLabRepo(db *gorm.DB) LabRepo {
	return &labRepo{db}
}

func (r *labRepo) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	today := time.Now().Truncate(24 * time.Hour)

	var pendingLab, completedLabToday int64
	r.db.Model(&model.LabOrder{}).Where("status IN (?)", []string{"ordered", "sample_collected"}).Count(&pendingLab)
	r.db.Model(&model.LabOrder{}).Where("status = ? AND result_uploaded_at >= ?", "completed", today).Count(&completedLabToday)

	var pendingRad, completedRadToday int64
	r.db.Model(&model.RadiologyOrder{}).Where("status IN (?)", []string{"ordered", "in_progress"}).Count(&pendingRad)
	r.db.Model(&model.RadiologyOrder{}).Where("status = ? AND report_uploaded_at >= ?", "completed", today).Count(&completedRadToday)

	stats["pending_lab"] = pendingLab
	stats["completed_lab_today"] = completedLabToday
	stats["pending_radiology"] = pendingRad
	stats["completed_radiology_today"] = completedRadToday

	return stats, nil
}

func (r *labRepo) ListTests(search, status string) ([]model.LabTest, error) {
	var tests []model.LabTest
	q := r.db.Model(&model.LabTest{})
	if search != "" {
		q = q.Where("test_name ILIKE ? OR test_code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&tests).Error
	return tests, err
}

func (r *labRepo) AddTest(test *model.LabTest) error {
	return r.db.Create(test).Error
}

func (r *labRepo) UpdateTest(testID uint, updates map[string]interface{}) error {
	return r.db.Model(&model.LabTest{}).Where("test_id = ?", testID).Updates(updates).Error
}

func (r *labRepo) ListLabOrders(status string, fromDate string, patientID *uint) ([]model.LabOrder, error) {
	var orders []model.LabOrder
	q := r.db.Preload("Patient").Preload("Doctor").Preload("TestRef").Preload("Technician")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if fromDate != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			q = q.Where("ordered_at >= ?", t)
		}
	}
	if patientID != nil && *patientID > 0 {
		q = q.Where("patient_id = ?", *patientID)
	}
	err := q.Order("ordered_at desc").Find(&orders).Error
	return orders, err
}

func (r *labRepo) GetLabOrder(id uint) (*model.LabOrder, error) {
	var order model.LabOrder
	err := r.db.Preload("Patient").Preload("Doctor").Preload("TestRef").Preload("Technician").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *labRepo) UpdateLabOrderStatus(id uint, status string) error {
	return r.db.Model(&model.LabOrder{}).Where("order_id = ?", id).Update("status", status).Error
}

func (r *labRepo) CollectSample(id uint, t time.Time, techID uint) error {
	return r.db.Model(&model.LabOrder{}).Where("order_id = ?", id).Updates(map[string]interface{}{
		"sample_collected_at": t,
		"status":              "sample_collected",
		"technician_id":       techID,
	}).Error
}

func (r *labRepo) UploadResult(id uint, result string, isAbnormal *bool, notes string, techID uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"result_value":       result,
		"is_abnormal":        isAbnormal,
		"notes":              notes,
		"technician_id":      techID,
		"result_uploaded_at": now,
		"status":             "completed",
	}
	return r.db.Model(&model.LabOrder{}).Where("order_id = ?", id).Updates(updates).Error
}

func (r *labRepo) ListRadiologyOrders(status string, modality string) ([]model.RadiologyOrder, error) {
	var orders []model.RadiologyOrder
	q := r.db.Preload("Patient").Preload("Doctor").Preload("Radiologist")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if modality != "" {
		q = q.Where("modality = ?", modality)
	}
	err := q.Order("ordered_at desc").Find(&orders).Error
	return orders, err
}

func (r *labRepo) GetRadiologyOrder(id uint) (*model.RadiologyOrder, error) {
	var order model.RadiologyOrder
	err := r.db.Preload("Patient").Preload("Doctor").Preload("Radiologist").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *labRepo) UpdateRadiologyOrder(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.RadiologyOrder{}).Where("order_id = ?", id).Updates(updates).Error
}
