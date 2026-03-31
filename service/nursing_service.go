package service

import (
	"apollo-backend/model"
	"apollo-backend/repository"
	"errors"
)

type NursingService interface {
	GetDashboard(wardID *uint) (map[string]interface{}, error)
	ListWards(deptID *uint) ([]model.Ward, error)
	GetBedsByWard(wardID uint, status string) ([]model.Bed, error)
	UpdateBedStatus(bedID uint, status, notes string) error
	TransferBed(req *model.BedTransferRequest) error
	ConfirmTransfer(transferID uint, notes string) error
	MarkBedClean(bedID uint) error

	ListVitals(patientID uint, fromDate string) ([]model.Vital, error)
	RecordVital(vital *model.Vital) error

	GetMAR(patientID uint, date string) ([]model.MedicationAdminRecord, error)
	UpdateMARStatus(marID uint, status, reason, notes string, nurseID uint) error

	AddNursingNote(note *model.NursingNote) error
}

type nursingService struct {
	repo repository.NursingRepo
}

func NewNursingService(repo repository.NursingRepo) NursingService {
	return &nursingService{repo}
}

func (s *nursingService) GetDashboard(wardID *uint) (map[string]interface{}, error) {
	return s.repo.GetDashboardStats(wardID)
}

func (s *nursingService) ListWards(deptID *uint) ([]model.Ward, error) {
	return s.repo.ListWards(deptID)
}

func (s *nursingService) GetBedsByWard(wardID uint, status string) ([]model.Bed, error) {
	return s.repo.GetBedsByWard(wardID, status)
}

func (s *nursingService) UpdateBedStatus(bedID uint, status, notes string) error {
	// Simple wrapper for now
	return s.repo.UpdateBedStatus(bedID, status)
}

func (s *nursingService) TransferBed(req *model.BedTransferRequest) error {
	if req.PatientID == 0 || req.ToBedID == 0 {
		return errors.New("patient_id and to_bed_id are required")
	}
	req.Status = "pending"
	return s.repo.InitiateTransfer(req)
}

func (s *nursingService) ConfirmTransfer(transferID uint, notes string) error {
	return s.repo.ConfirmTransfer(transferID, notes)
}

func (s *nursingService) MarkBedClean(bedID uint) error {
	return s.repo.MarkBedClean(bedID)
}

func (s *nursingService) ListVitals(patientID uint, fromDate string) ([]model.Vital, error) {
	return s.repo.ListVitals(patientID, fromDate)
}

func (s *nursingService) RecordVital(vital *model.Vital) error {
	if vital.PatientID == 0 {
		return errors.New("patient_id is required")
	}
	if vital.RecordedByRole == "" {
		vital.RecordedByRole = "nurse"
	}

	// simple logic to check for critical:
	vital.IsCritical = false
	if vital.SpO2Percent != nil && *vital.SpO2Percent < 90 {
		vital.IsCritical = true
	}
	if vital.PulseBPM != nil && (*vital.PulseBPM > 120 || *vital.PulseBPM < 50) {
		vital.IsCritical = true
	}
	if vital.TemperatureC != nil && (*vital.TemperatureC > 39.0 || *vital.TemperatureC < 35.0) {
		vital.IsCritical = true
	}

	return s.repo.RecordVital(vital)
}

func (s *nursingService) GetMAR(patientID uint, date string) ([]model.MedicationAdminRecord, error) {
	return s.repo.GetMAR(patientID, date)
}

func (s *nursingService) UpdateMARStatus(marID uint, status, reason, notes string, nurseID uint) error {
	validStatuses := map[string]bool{"administered": true, "held": true, "refused": true, "missed": true}
	if !validStatuses[status] {
		return errors.New("invalid status, must be administered, held, refused, or missed")
	}
	return s.repo.UpdateMARStatus(marID, status, reason, notes, nurseID)
}

func (s *nursingService) AddNursingNote(note *model.NursingNote) error {
	if note.PatientID == 0 || note.NurseID == 0 || note.NoteText == "" {
		return errors.New("patient_id, nurse_id, and note_text are required")
	}
	return s.repo.AddNursingNote(note)
}
