package service

import (
	"apollo-backend/model"
	"apollo-backend/repository"
	"errors"
	"time"
)

type LabService interface {
	GetDashboard() (map[string]interface{}, error)
	ListTests(search, status string) ([]model.LabTest, error)
	AddTest(test *model.LabTest) error
	UpdateTest(testID uint, updates map[string]interface{}) error

	ListLabOrders(status string, fromDate string, patientID *uint) ([]model.LabOrder, error)
	GetLabOrder(id uint) (*model.LabOrder, error)
	CollectSample(id uint, techID uint) error
	UploadResult(id uint, result string, isAbnormal *bool, notes string, techID uint) error
	CancelLabOrder(id uint, reason string) error

	ListRadiologyOrders(status string, modality string) ([]model.RadiologyOrder, error)
	GetRadiologyOrder(id uint) (*model.RadiologyOrder, error)
	UploadRadiologyReport(id uint, reportText string, filePath string, radID uint) error
	AttachImage(id uint, filePath string) error
}

type labService struct {
	repo repository.LabRepo
}

func NewLabService(repo repository.LabRepo) LabService {
	return &labService{repo}
}

func (s *labService) GetDashboard() (map[string]interface{}, error) {
	return s.repo.GetDashboardStats()
}

func (s *labService) ListTests(search, status string) ([]model.LabTest, error) {
	return s.repo.ListTests(search, status)
}

func (s *labService) AddTest(test *model.LabTest) error {
	if test.TestName == "" || test.TestCode == "" {
		return errors.New("test_name and test_code are required")
	}
	return s.repo.AddTest(test)
}

func (s *labService) UpdateTest(testID uint, updates map[string]interface{}) error {
	return s.repo.UpdateTest(testID, updates)
}

func (s *labService) ListLabOrders(status string, fromDate string, patientID *uint) ([]model.LabOrder, error) {
	return s.repo.ListLabOrders(status, fromDate, patientID)
}

func (s *labService) GetLabOrder(id uint) (*model.LabOrder, error) {
	return s.repo.GetLabOrder(id)
}

func (s *labService) CollectSample(id uint, techID uint) error {
	return s.repo.CollectSample(id, time.Now(), techID)
}

func (s *labService) UploadResult(id uint, result string, isAbnormal *bool, notes string, techID uint) error {
	if result == "" {
		return errors.New("result_value is required")
	}
	return s.repo.UploadResult(id, result, isAbnormal, notes, techID)
}

func (s *labService) CancelLabOrder(id uint, reason string) error {
	return s.repo.UpdateLabOrderStatus(id, "cancelled")
}

func (s *labService) ListRadiologyOrders(status string, modality string) ([]model.RadiologyOrder, error) {
	return s.repo.ListRadiologyOrders(status, modality)
}

func (s *labService) GetRadiologyOrder(id uint) (*model.RadiologyOrder, error) {
	return s.repo.GetRadiologyOrder(id)
}

func (s *labService) UploadRadiologyReport(id uint, reportText string, filePath string, radID uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"report_text":        reportText,
		"radiologist_id":     radID,
		"status":             "completed",
		"report_uploaded_at": now,
	}
	if filePath != "" {
		updates["report_file_url"] = filePath
	}
	return s.repo.UpdateRadiologyOrder(id, updates)
}

func (s *labService) AttachImage(id uint, filePath string) error {
	return s.repo.UpdateRadiologyOrder(id, map[string]interface{}{
		"image_file_urls": filePath, // In reality this would be an array append, but sticking to basics
	})
}
