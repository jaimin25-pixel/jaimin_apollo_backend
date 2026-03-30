package service

import (
	"errors"
	"time"

	"apollo-backend/model"
	"apollo-backend/repository"
)

type DoctorService struct {
	repo      *repository.DoctorRepo
	auditRepo *repository.AuditRepo
}

func NewDoctorService(repo *repository.DoctorRepo, ar *repository.AuditRepo) *DoctorService {
	return &DoctorService{repo: repo, auditRepo: ar}
}

// ── Dashboard ─────────────────────────────────────────────────────────

type DoctorDashboardData struct {
	TodayAppointmentsCount int64               `json:"today_appointments_count"`
	TodayAppointments      []model.Appointment `json:"today_appointments"`
	PendingPrescriptions   int64               `json:"pending_prescriptions"`
	CompletedLabOrders     []model.LabOrder    `json:"completed_lab_orders"`
}

func (s *DoctorService) GetDashboard(doctorID uint) (*DoctorDashboardData, error) {
	appts, count, err := s.repo.TodayAppointments(doctorID)
	if err != nil {
		return nil, err
	}
	pendingRx, _ := s.repo.PendingPrescriptionsCount(doctorID)
	labAlerts, _ := s.repo.CompletedLabOrders(doctorID)

	return &DoctorDashboardData{
		TodayAppointmentsCount: count,
		TodayAppointments:      appts,
		PendingPrescriptions:   pendingRx,
		CompletedLabOrders:     labAlerts,
	}, nil
}

// ── Appointments ──────────────────────────────────────────────────────

type AppointmentFilters struct {
	Date      string `form:"date"`
	Status    string `form:"status"`
	PatientID uint   `form:"patient_id"`
}

type UpdateAppointmentStatusInput struct {
	Status            string `json:"status" binding:"required,oneof=scheduled checked_in in_consultation completed cancelled"`
	ConsultationNotes string `json:"consultation_notes"`
}

func (s *DoctorService) ListAppointments(doctorID uint, f AppointmentFilters) ([]model.Appointment, error) {
	return s.repo.ListAppointments(doctorID, f.Date, f.Status, f.PatientID)
}

func (s *DoctorService) GetAppointment(apptID, doctorID uint, role string) (*model.Appointment, error) {
	appt, err := s.repo.GetAppointment(apptID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}
	if role == "doctor" && appt.DoctorID != doctorID {
		return nil, errors.New("access denied: appointment belongs to another doctor")
	}
	return appt, nil
}

func (s *DoctorService) UpdateAppointmentStatus(apptID, doctorID uint, input UpdateAppointmentStatusInput) error {
	appt, err := s.repo.GetAppointment(apptID)
	if err != nil {
		return errors.New("appointment not found")
	}
	if appt.DoctorID != doctorID {
		return errors.New("access denied: appointment belongs to another doctor")
	}
	return s.repo.UpdateAppointmentStatus(apptID, input.Status, input.ConsultationNotes)
}

// ── Patients ──────────────────────────────────────────────────────────

type PatientFilters struct {
	Search string `form:"search"`
	Status string `form:"status"`
}

func (s *DoctorService) ListPatients(doctorID uint, role string, f PatientFilters) ([]model.Patient, error) {
	id := doctorID
	if role == "admin" {
		id = 0
	}
	return s.repo.ListPatients(id, f.Search, f.Status)
}

func (s *DoctorService) GetPatientEHR(patientID, doctorID uint, role string) (*repository.PatientEHR, error) {
	if role == "doctor" && !s.repo.DoctorHasPatient(doctorID, patientID) {
		return nil, errors.New("access denied: no appointments with this patient")
	}
	return s.repo.GetPatientEHR(patientID)
}

// ── Vitals ────────────────────────────────────────────────────────────

type RecordVitalInput struct {
	TemperatureC           *float64 `json:"temperature_c"`
	BloodPressureSystolic  *int     `json:"blood_pressure_systolic"`
	BloodPressureDiastolic *int     `json:"blood_pressure_diastolic"`
	PulseBPM               *int     `json:"pulse_bpm"`
	SpO2Percent            *float64 `json:"spo2_percent"`
	RespiratoryRate        *int     `json:"respiratory_rate"`
	BloodGlucoseMgDL       *float64 `json:"blood_glucose_mgdl"`
	AdmissionID            *uint    `json:"admission_id"`
}

func (s *DoctorService) RecordVital(patientID, recorderID uint, role string, input RecordVitalInput) (*model.Vital, error) {
	if input.TemperatureC == nil && input.BloodPressureSystolic == nil &&
		input.BloodPressureDiastolic == nil && input.PulseBPM == nil &&
		input.SpO2Percent == nil && input.RespiratoryRate == nil &&
		input.BloodGlucoseMgDL == nil {
		return nil, errors.New("at least one vital measurement is required")
	}

	isCritical := false
	if input.SpO2Percent != nil && *input.SpO2Percent < 90 {
		isCritical = true
	}
	if input.BloodPressureSystolic != nil && (*input.BloodPressureSystolic > 180 || *input.BloodPressureSystolic < 90) {
		isCritical = true
	}
	if input.PulseBPM != nil && (*input.PulseBPM > 150 || *input.PulseBPM < 40) {
		isCritical = true
	}
	if input.TemperatureC != nil && (*input.TemperatureC > 39.5 || *input.TemperatureC < 35.0) {
		isCritical = true
	}

	var nurseID *uint
	recordedByRole := "nurse"
	if role == "doctor" {
		recordedByRole = "doctor"
	} else {
		nurseID = &recorderID
	}

	v := &model.Vital{
		PatientID:              patientID,
		AdmissionID:            input.AdmissionID,
		NurseID:                nurseID,
		RecordedByRole:         recordedByRole,
		RecordedAt:             time.Now(),
		TemperatureC:           input.TemperatureC,
		BloodPressureSystolic:  input.BloodPressureSystolic,
		BloodPressureDiastolic: input.BloodPressureDiastolic,
		PulseBPM:               input.PulseBPM,
		SpO2Percent:            input.SpO2Percent,
		RespiratoryRate:        input.RespiratoryRate,
		BloodGlucoseMgDL:       input.BloodGlucoseMgDL,
		IsCritical:             isCritical,
	}

	if err := s.repo.CreateVital(v); err != nil {
		return nil, errors.New("failed to record vital")
	}
	return v, nil
}

// ── Clinical Notes ────────────────────────────────────────────────────

type CreateClinicalNoteInput struct {
	Notes       string `json:"notes" binding:"required"`
	ICD10Code   string `json:"icd10_code"`
	ApptID      *uint  `json:"appt_id"`
	AdmissionID *uint  `json:"admission_id"`
}

func (s *DoctorService) CreateClinicalNote(patientID, doctorID uint, input CreateClinicalNoteInput) (*model.ClinicalNote, error) {
	if input.ApptID != nil && input.AdmissionID != nil {
		return nil, errors.New("specify either appt_id or admission_id, not both")
	}

	cn := &model.ClinicalNote{
		PatientID:   patientID,
		DoctorID:    doctorID,
		ApptID:      input.ApptID,
		AdmissionID: input.AdmissionID,
		Notes:       input.Notes,
		ICD10Code:   input.ICD10Code,
	}

	if err := s.repo.CreateClinicalNote(cn); err != nil {
		return nil, errors.New("failed to save clinical note")
	}
	return cn, nil
}

// ── Prescriptions ─────────────────────────────────────────────────────

type PrescriptionItemInput struct {
	MedicineID         uint    `json:"medicine_id" binding:"required"`
	Dose               string  `json:"dose" binding:"required"`
	Frequency          string  `json:"frequency" binding:"required"`
	DurationDays       int     `json:"duration_days" binding:"required,min=1"`
	Instructions       string  `json:"instructions"`
	QuantityPrescribed float64 `json:"quantity_prescribed" binding:"required,gt=0"`
}

type CreatePrescriptionInput struct {
	PatientID   uint                    `json:"patient_id" binding:"required"`
	ApptID      *uint                   `json:"appt_id"`
	AdmissionID *uint                   `json:"admission_id"`
	Diagnosis   string                  `json:"diagnosis"`
	Items       []PrescriptionItemInput `json:"items" binding:"required,min=1,dive"`
}

type PrescriptionFilters struct {
	PatientID uint   `form:"patient_id"`
	Status    string `form:"status"`
	FromDate  string `form:"from_date"`
}

func (s *DoctorService) ListPrescriptions(doctorID uint, role string, f PrescriptionFilters) ([]model.Prescription, error) {
	id := doctorID
	if role == "admin" {
		id = 0
	}
	return s.repo.ListPrescriptions(id, f.PatientID, f.Status, f.FromDate)
}

func (s *DoctorService) GetPrescription(rxID, doctorID uint, role string) (*model.Prescription, error) {
	rx, err := s.repo.GetPrescription(rxID)
	if err != nil {
		return nil, errors.New("prescription not found")
	}
	if role == "doctor" && rx.DoctorID != doctorID {
		return nil, errors.New("access denied: prescription belongs to another doctor")
	}
	return rx, nil
}

func (s *DoctorService) CreatePrescription(doctorID uint, input CreatePrescriptionInput) (*model.Prescription, error) {
	medicineIDs := make([]uint, len(input.Items))
	for i, it := range input.Items {
		medicineIDs[i] = it.MedicineID
	}
	if err := s.repo.ValidateMedicineIDs(medicineIDs); err != nil {
		return nil, err
	}

	rx := &model.Prescription{
		DoctorID:    doctorID,
		PatientID:   input.PatientID,
		ApptID:      input.ApptID,
		AdmissionID: input.AdmissionID,
		Diagnosis:   input.Diagnosis,
		Status:      "pending",
	}

	items := make([]model.PrescriptionItem, len(input.Items))
	for i, it := range input.Items {
		items[i] = model.PrescriptionItem{
			MedicineID:         it.MedicineID,
			Dose:               it.Dose,
			Frequency:          it.Frequency,
			DurationDays:       it.DurationDays,
			Instructions:       it.Instructions,
			QuantityPrescribed: it.QuantityPrescribed,
			Status:             "pending",
		}
	}

	if err := s.repo.CreatePrescriptionTx(rx, items); err != nil {
		return nil, errors.New("failed to create prescription")
	}
	return rx, nil
}

// ── Lab Orders ────────────────────────────────────────────────────────

type CreateLabOrderInput struct {
	PatientID          uint   `json:"patient_id" binding:"required"`
	TestID             uint   `json:"test_id" binding:"required"`
	ApptID             *uint  `json:"appt_id"`
	AdmissionID        *uint  `json:"admission_id"`
	ClinicalIndication string `json:"clinical_indication"`
}

type LabOrderFilters struct {
	PatientID uint   `form:"patient_id"`
	Status    string `form:"status"`
	FromDate  string `form:"from_date"`
}

func (s *DoctorService) ListLabOrders(doctorID uint, f LabOrderFilters) ([]model.LabOrder, error) {
	return s.repo.ListLabOrders(doctorID, f.PatientID, f.Status, f.FromDate)
}

func (s *DoctorService) GetLabOrder(orderID, doctorID uint) (*model.LabOrder, error) {
	order, err := s.repo.GetLabOrder(orderID)
	if err != nil {
		return nil, errors.New("lab order not found")
	}
	if order.DoctorID != doctorID {
		return nil, errors.New("access denied: lab order belongs to another doctor")
	}
	return order, nil
}

func (s *DoctorService) CreateLabOrder(doctorID uint, input CreateLabOrderInput) (*model.LabOrder, error) {
	if err := s.repo.ValidateTestID(input.TestID); err != nil {
		return nil, err
	}

	lo := &model.LabOrder{
		PatientID:          input.PatientID,
		DoctorID:           doctorID,
		TestID:             input.TestID,
		ApptID:             input.ApptID,
		AdmissionID:        input.AdmissionID,
		Notes:              input.ClinicalIndication,
		Status:             "ordered",
		OrderedAt:          time.Now(),
	}

	if err := s.repo.CreateLabOrder(lo); err != nil {
		return nil, errors.New("failed to create lab order")
	}
	return lo, nil
}

// ── Radiology Orders ──────────────────────────────────────────────────

type CreateRadiologyOrderInput struct {
	PatientID          uint   `json:"patient_id" binding:"required"`
	Modality           string `json:"modality" binding:"required,oneof=xray ultrasound ct_scan mri echocardiogram"`
	BodyPart           string `json:"body_part"`
	ClinicalIndication string `json:"clinical_indication"`
	ApptID             *uint  `json:"appt_id"`
	AdmissionID        *uint  `json:"admission_id"`
}

type RadiologyOrderFilters struct {
	PatientID uint   `form:"patient_id"`
	Status    string `form:"status"`
}

func (s *DoctorService) ListRadiologyOrders(doctorID uint, f RadiologyOrderFilters) ([]model.RadiologyOrder, error) {
	return s.repo.ListRadiologyOrders(doctorID, f.PatientID, f.Status)
}

func (s *DoctorService) GetRadiologyOrder(radiologyID, doctorID uint) (*model.RadiologyOrder, error) {
	order, err := s.repo.GetRadiologyOrder(radiologyID)
	if err != nil {
		return nil, errors.New("radiology order not found")
	}
	if order.DoctorID != doctorID {
		return nil, errors.New("access denied: radiology order belongs to another doctor")
	}
	return order, nil
}

func (s *DoctorService) CreateRadiologyOrder(doctorID uint, input CreateRadiologyOrderInput) (*model.RadiologyOrder, error) {
	ro := &model.RadiologyOrder{
		PatientID:          input.PatientID,
		DoctorID:           doctorID,
		Modality:           input.Modality,
		BodyPart:           input.BodyPart,
		ClinicalIndication: input.ClinicalIndication,
		Status:             "ordered",
		OrderedAt:          time.Now(),
	}

	if err := s.repo.CreateRadiologyOrder(ro); err != nil {
		return nil, errors.New("failed to create radiology order")
	}
	return ro, nil
}

// ── Admissions ────────────────────────────────────────────────────────

type CreateAdmissionInput struct {
	PatientID     uint   `json:"patient_id" binding:"required"`
	WardID        uint   `json:"ward_id" binding:"required"`
	BedID         uint   `json:"bed_id" binding:"required"`
	Diagnosis     string `json:"diagnosis"`
	TreatmentPlan string `json:"treatment_plan"`
}

type DischargeInput struct {
	DischargeSummary string `json:"discharge_summary" binding:"required"`
}

func (s *DoctorService) CreateAdmission(doctorID uint, input CreateAdmissionInput) (*model.Admission, error) {
	adm := &model.Admission{
		PatientID:         input.PatientID,
		AdmittingDoctorID: doctorID,
		WardID:            input.WardID,
		BedID:             input.BedID,
		Diagnosis:         input.Diagnosis,
		TreatmentPlan:     input.TreatmentPlan,
		Status:            "admitted",
	}

	if err := s.repo.CreateAdmissionTx(adm); err != nil {
		return nil, err
	}
	return adm, nil
}

func (s *DoctorService) DischargeAdmission(admissionID, doctorID uint, input DischargeInput) error {
	return s.repo.DischargeTx(admissionID, input.DischargeSummary)
}
