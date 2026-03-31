package repository

import (
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

// ─── PatientRepo ──────────────────────────────────────────────────────────────

type PatientRepo struct{ DB *gorm.DB }

func NewPatientRepo(db *gorm.DB) *PatientRepo { return &PatientRepo{DB: db} }

// ── Patient CRUD ─────────────────────────────────────────────────────────────

func (r *PatientRepo) SearchPatients(q string) ([]model.Patient, error) {
	var patients []model.Patient
	tx := r.DB.Model(&model.Patient{})
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("full_name ILIKE ? OR contact_number ILIKE ? OR pat_code ILIKE ?", like, like, like)
	}
	err := tx.Order("created_at DESC").Limit(50).Find(&patients).Error
	return patients, err
}

func (r *PatientRepo) GetPatientByID(id uint) (*model.Patient, error) {
	var p model.Patient
	err := r.DB.Where("patient_id = ?", id).First(&p).Error
	return &p, err
}

func (r *PatientRepo) FindByEmail(email string) (*model.Patient, error) {
	var p model.Patient
	err := r.DB.Where("email = ? AND hashed_password != ''", email).First(&p).Error
	return &p, err
}

func (r *PatientRepo) FindByID(id uint) (*model.Patient, error) {
	var p model.Patient
	err := r.DB.Where("patient_id = ?", id).First(&p).Error
	return &p, err
}

func (r *PatientRepo) CreatePatient(p *model.Patient) error {
	return r.DB.Create(p).Error
}

func (r *PatientRepo) UpdatePatient(id uint, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return r.DB.Model(&model.Patient{}).Where("patient_id = ?", id).Updates(fields).Error
}

func (r *PatientRepo) PatCodeExists(code string) bool {
	var count int64
	r.DB.Model(&model.Patient{}).Where("pat_code = ?", code).Count(&count)
	return count > 0
}

func (r *PatientRepo) DuplicateCheck(name string, dob time.Time, phone string) ([]model.Patient, error) {
	var patients []model.Patient
	err := r.DB.Where(
		"full_name ILIKE ? AND date_of_birth = ? AND contact_number = ?",
		name, dob, phone,
	).Find(&patients).Error
	return patients, err
}

// ── Appointment ──────────────────────────────────────────────────────────────

func (r *PatientRepo) ListAppointments(patientID uint, status string, fromDate string) ([]model.Appointment, error) {
	var appts []model.Appointment
	tx := r.DB.Preload("Patient").Preload("Doctor").Preload("Department").
		Where("patient_id = ?", patientID)
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if fromDate != "" {
		tx = tx.Where("DATE(scheduled_at AT TIME ZONE 'Asia/Kolkata') >= ?", fromDate)
	}
	err := tx.Order("scheduled_at DESC").Find(&appts).Error
	return appts, err
}

func (r *PatientRepo) GetAppointment(apptID uint) (*model.Appointment, error) {
	var a model.Appointment
	err := r.DB.Preload("Patient").Preload("Doctor").Preload("Department").
		Where("appt_id = ?", apptID).First(&a).Error
	return &a, err
}

func (r *PatientRepo) CreateAppointment(a *model.Appointment) error {
	return r.DB.Create(a).Error
}

func (r *PatientRepo) UpdateAppointment(apptID uint, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return r.DB.Model(&model.Appointment{}).Where("appt_id = ?", apptID).Updates(fields).Error
}

func (r *PatientRepo) CheckDoubleBooking(doctorID uint, scheduledAt time.Time) (bool, error) {
	var count int64
	// Check within a 30-minute window
	windowStart := scheduledAt.Add(-30 * time.Minute)
	windowEnd := scheduledAt.Add(30 * time.Minute)
	err := r.DB.Model(&model.Appointment{}).
		Where("doctor_id = ? AND scheduled_at BETWEEN ? AND ? AND status NOT IN ('cancelled')",
			doctorID, windowStart, windowEnd).
		Count(&count).Error
	return count > 0, err
}

func (r *PatientRepo) CheckInAppointment(apptID uint, token string) error {
	return r.DB.Model(&model.Appointment{}).
		Where("appt_id = ? AND status = 'scheduled'", apptID).
		Updates(map[string]interface{}{
			"status":      "checked_in",
			"queue_token": token,
			"updated_at":  time.Now(),
		}).Error
}

func (r *PatientRepo) CountTodayCheckedInByDept(deptID uint) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.DB.Model(&model.Appointment{}).
		Where("dept_id = ? AND DATE(scheduled_at AT TIME ZONE 'Asia/Kolkata') = ? AND status IN ('checked_in','in_consultation','completed')",
			deptID, today).
		Count(&count).Error
	return count, err
}

func (r *PatientRepo) GetDepartment(id uint) (*model.Department, error) {
	var d model.Department
	err := r.DB.Where("dept_id = ?", id).First(&d).Error
	return &d, err
}

// ── Admission ────────────────────────────────────────────────────────────────

func (r *PatientRepo) ListAdmissions(patientID uint, status string) ([]model.Admission, error) {
	var admissions []model.Admission
	tx := r.DB.Preload("Patient").Preload("AdmittingDoctor").Preload("Ward").Preload("Bed").Preload("Department").
		Where("patient_id = ?", patientID)
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	err := tx.Order("admitted_at DESC").Find(&admissions).Error
	return admissions, err
}

func (r *PatientRepo) GetAdmission(admissionID uint) (*model.Admission, error) {
	var a model.Admission
	err := r.DB.Preload("Patient").Preload("AdmittingDoctor").Preload("Ward").Preload("Bed").Preload("Department").
		Where("admission_id = ?", admissionID).First(&a).Error
	return &a, err
}

// ── EHR ──────────────────────────────────────────────────────────────────────

type EHRData struct {
	Visits        []model.Appointment  `json:"visits"`
	Diagnoses     []model.ClinicalNote `json:"diagnoses"`
	Prescriptions []model.Prescription `json:"prescriptions"`
	LabResults    []model.LabOrder     `json:"lab_results"`
	Vitals        []model.Vital        `json:"vitals"`
	ClinicalNotes []model.ClinicalNote `json:"clinical_notes"`
}

func (r *PatientRepo) GetFullEHR(patientID uint) (*EHRData, error) {
	ehr := &EHRData{}

	// Visits (appointments)
	if err := r.DB.Preload("Doctor").Preload("Department").
		Where("patient_id = ?", patientID).
		Order("scheduled_at DESC").Limit(50).
		Find(&ehr.Visits).Error; err != nil {
		return nil, err
	}

	// Clinical notes (serve as diagnoses)
	if err := r.DB.Preload("Doctor").
		Where("patient_id = ?", patientID).
		Order("created_at DESC").Limit(50).
		Find(&ehr.ClinicalNotes).Error; err != nil {
		return nil, err
	}
	ehr.Diagnoses = ehr.ClinicalNotes

	// Prescriptions
	if err := r.DB.Preload("Doctor").Preload("Items").
		Where("patient_id = ?", patientID).
		Order("created_at DESC").Limit(50).
		Find(&ehr.Prescriptions).Error; err != nil {
		return nil, err
	}

	// Lab results
	if err := r.DB.Preload("Doctor").Preload("TestRef").
		Where("patient_id = ?", patientID).
		Order("ordered_at DESC").Limit(50).
		Find(&ehr.LabResults).Error; err != nil {
		return nil, err
	}

	// Vitals
	if err := r.DB.Where("patient_id = ?", patientID).
		Order("recorded_at DESC").Limit(100).
		Find(&ehr.Vitals).Error; err != nil {
		return nil, err
	}

	return ehr, nil
}

// ── Invoices ─────────────────────────────────────────────────────────────────

func (r *PatientRepo) ListInvoices(patientID uint, status string) ([]model.Invoice, error) {
	var invoices []model.Invoice
	tx := r.DB.Preload("Patient").Preload("Admission").Preload("Appointment").
		Where("patient_id = ?", patientID)
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	err := tx.Order("created_at DESC").Find(&invoices).Error
	return invoices, err
}
