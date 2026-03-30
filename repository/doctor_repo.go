package repository

import (
	"errors"
	"fmt"
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

type DoctorRepo struct{ DB *gorm.DB }

func NewDoctorRepo(db *gorm.DB) *DoctorRepo { return &DoctorRepo{DB: db} }

// ── Auth ─────────────────────────────────────────────────────────────

func (r *DoctorRepo) FindByEmail(email string) (*model.Doctor, error) {
	var doc model.Doctor
	err := r.DB.Where("email = ? AND status = 'active'", email).First(&doc).Error
	return &doc, err
}

func (r *DoctorRepo) FindByID(id uint) (*model.Doctor, error) {
	var doc model.Doctor
	err := r.DB.Where("doctor_id = ? AND status = 'active'", id).
		Preload("Department").First(&doc).Error
	return &doc, err
}

// ── Dashboard ─────────────────────────────────────────────────────────

func (r *DoctorRepo) TodayAppointments(doctorID uint) ([]model.Appointment, int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	var appts []model.Appointment
	err := r.DB.Where("doctor_id = ? AND scheduled_at >= ? AND scheduled_at < ?", doctorID, today, tomorrow).
		Preload("Patient").
		Order("scheduled_at").
		Find(&appts).Error
	return appts, int64(len(appts)), err
}

func (r *DoctorRepo) PendingPrescriptionsCount(doctorID uint) (int64, error) {
	var c int64
	err := r.DB.Model(&model.Prescription{}).
		Where("doctor_id = ? AND status = 'pending'", doctorID).
		Count(&c).Error
	return c, err
}

func (r *DoctorRepo) CompletedLabOrders(doctorID uint) ([]model.LabOrder, error) {
	var orders []model.LabOrder
	err := r.DB.Where("doctor_id = ? AND status = 'completed'", doctorID).
		Preload("Patient").
		Preload("TestRef").
		Order("result_uploaded_at DESC").
		Limit(10).
		Find(&orders).Error
	return orders, err
}

// ── Appointments ──────────────────────────────────────────────────────

func (r *DoctorRepo) ListAppointments(doctorID uint, date, status string, patientID uint) ([]model.Appointment, error) {
	q := r.DB.Preload("Patient").Preload("Department")

	if doctorID > 0 {
		q = q.Where("doctor_id = ?", doctorID)
	}
	if date != "" {
		d, err := time.Parse("2006-01-02", date)
		if err == nil {
			q = q.Where("scheduled_at >= ? AND scheduled_at < ?", d, d.Add(24*time.Hour))
		}
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if patientID > 0 {
		q = q.Where("patient_id = ?", patientID)
	}

	var appts []model.Appointment
	err := q.Order("scheduled_at DESC").Find(&appts).Error
	return appts, err
}

func (r *DoctorRepo) GetAppointment(apptID uint) (*model.Appointment, error) {
	var appt model.Appointment
	err := r.DB.Preload("Patient").Preload("Doctor").Preload("Department").
		First(&appt, apptID).Error
	return &appt, err
}

func (r *DoctorRepo) UpdateAppointmentStatus(apptID uint, status, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if notes != "" {
		updates["consultation_notes"] = notes
	}
	return r.DB.Model(&model.Appointment{}).Where("appt_id = ?", apptID).Updates(updates).Error
}

// ── Patients ──────────────────────────────────────────────────────────

func (r *DoctorRepo) ListPatients(doctorID uint, search, status string) ([]model.Patient, error) {
	var patients []model.Patient
	q := r.DB.Model(&model.Patient{})

	if doctorID > 0 {
		// Only patients who have had an appointment with this doctor
		q = q.Where("patient_id IN (SELECT DISTINCT patient_id FROM appointments WHERE doctor_id = ?)", doctorID)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("full_name ILIKE ? OR contact_number ILIKE ? OR pat_code ILIKE ?", like, like, like)
	}
	if status != "" {
		// Filter by latest admission status
		q = q.Where("patient_id IN (SELECT DISTINCT patient_id FROM admissions WHERE status = ?)", status)
	}

	err := q.Order("full_name").Find(&patients).Error
	return patients, err
}

// PatientEHR is the full electronic health record for a patient.
type PatientEHR struct {
	Patient         model.Patient           `json:"patient"`
	Appointments    []model.Appointment     `json:"appointments"`
	Admissions      []model.Admission       `json:"admissions"`
	Prescriptions   []model.Prescription    `json:"prescriptions"`
	LabOrders       []model.LabOrder        `json:"lab_orders"`
	RadiologyOrders []model.RadiologyOrder  `json:"radiology_orders"`
	Vitals          []model.Vital           `json:"vitals"`
	ClinicalNotes   []model.ClinicalNote    `json:"clinical_notes"`
}

func (r *DoctorRepo) GetPatientEHR(patientID uint) (*PatientEHR, error) {
	var patient model.Patient
	if err := r.DB.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	ehr := &PatientEHR{Patient: patient}

	r.DB.Where("patient_id = ?", patientID).
		Preload("Doctor").Preload("Department").
		Order("scheduled_at DESC").Find(&ehr.Appointments)

	r.DB.Where("patient_id = ?", patientID).
		Preload("Ward").Preload("Bed").Preload("Department").
		Order("admitted_at DESC").Find(&ehr.Admissions)

	r.DB.Where("patient_id = ?", patientID).
		Preload("Doctor").
		Order("created_at DESC").Find(&ehr.Prescriptions)
	// Load items for each prescription
	for i := range ehr.Prescriptions {
		r.DB.Where("rx_id = ?", ehr.Prescriptions[i].RxID).
			Preload("Medicine").Find(&ehr.Prescriptions[i].Items)
	}

	r.DB.Where("patient_id = ?", patientID).
		Preload("Doctor").Preload("TestRef").
		Order("ordered_at DESC").Find(&ehr.LabOrders)

	r.DB.Where("patient_id = ?", patientID).
		Preload("Doctor").
		Order("ordered_at DESC").Find(&ehr.RadiologyOrders)

	r.DB.Where("patient_id = ?", patientID).
		Order("recorded_at DESC").Find(&ehr.Vitals)

	r.DB.Where("patient_id = ?", patientID).
		Preload("Doctor").
		Order("created_at DESC").Find(&ehr.ClinicalNotes)

	return ehr, nil
}

// DoctorHasPatient checks if the doctor has at least one appointment with this patient.
func (r *DoctorRepo) DoctorHasPatient(doctorID, patientID uint) bool {
	var c int64
	r.DB.Model(&model.Appointment{}).
		Where("doctor_id = ? AND patient_id = ?", doctorID, patientID).
		Count(&c)
	return c > 0
}

// ── Vitals ────────────────────────────────────────────────────────────

func (r *DoctorRepo) CreateVital(v *model.Vital) error {
	return r.DB.Create(v).Error
}

// ── Clinical Notes ────────────────────────────────────────────────────

func (r *DoctorRepo) CreateClinicalNote(cn *model.ClinicalNote) error {
	return r.DB.Create(cn).Error
}

// ── Prescriptions ─────────────────────────────────────────────────────

func (r *DoctorRepo) ListPrescriptions(doctorID uint, patientID uint, status, fromDate string) ([]model.Prescription, error) {
	q := r.DB.Preload("Patient").Preload("Doctor")

	if doctorID > 0 {
		q = q.Where("doctor_id = ?", doctorID)
	}
	if patientID > 0 {
		q = q.Where("patient_id = ?", patientID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if fromDate != "" {
		d, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			q = q.Where("created_at >= ?", d)
		}
	}

	var rxs []model.Prescription
	err := q.Order("created_at DESC").Find(&rxs).Error
	return rxs, err
}

func (r *DoctorRepo) GetPrescription(rxID uint) (*model.Prescription, error) {
	var rx model.Prescription
	err := r.DB.Preload("Patient").Preload("Doctor").
		First(&rx, rxID).Error
	if err != nil {
		return nil, err
	}
	r.DB.Where("rx_id = ?", rxID).Preload("Medicine").Find(&rx.Items)
	return &rx, nil
}

func (r *DoctorRepo) CreatePrescriptionTx(rx *model.Prescription, items []model.PrescriptionItem) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rx).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].RxID = rx.RxID
		}
		return tx.Create(&items).Error
	})
}

func (r *DoctorRepo) ValidateMedicineIDs(ids []uint) error {
	var count int64
	r.DB.Model(&model.Medicine{}).
		Where("medicine_id IN ? AND status = 'active'", ids).
		Count(&count)
	if int(count) != len(ids) {
		return errors.New("one or more medicines are invalid or inactive")
	}
	return nil
}

// ── Lab Orders ────────────────────────────────────────────────────────

func (r *DoctorRepo) ListLabOrders(doctorID uint, patientID uint, status, fromDate string) ([]model.LabOrder, error) {
	q := r.DB.Preload("Patient").Preload("TestRef")

	if doctorID > 0 {
		q = q.Where("doctor_id = ?", doctorID)
	}
	if patientID > 0 {
		q = q.Where("patient_id = ?", patientID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if fromDate != "" {
		d, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			q = q.Where("ordered_at >= ?", d)
		}
	}

	var orders []model.LabOrder
	err := q.Order("ordered_at DESC").Find(&orders).Error
	return orders, err
}

func (r *DoctorRepo) GetLabOrder(orderID uint) (*model.LabOrder, error) {
	var order model.LabOrder
	err := r.DB.Preload("Patient").Preload("Doctor").Preload("TestRef").
		First(&order, orderID).Error
	return &order, err
}

func (r *DoctorRepo) ValidateTestID(testID uint) error {
	var count int64
	r.DB.Model(&model.LabTest{}).Where("test_id = ? AND status = 'active'", testID).Count(&count)
	if count == 0 {
		return fmt.Errorf("lab test %d not found or inactive", testID)
	}
	return nil
}

func (r *DoctorRepo) CreateLabOrder(lo *model.LabOrder) error {
	return r.DB.Create(lo).Error
}

// ── Radiology Orders ──────────────────────────────────────────────────

func (r *DoctorRepo) ListRadiologyOrders(doctorID uint, patientID uint, status string) ([]model.RadiologyOrder, error) {
	q := r.DB.Preload("Patient")

	if doctorID > 0 {
		q = q.Where("doctor_id = ?", doctorID)
	}
	if patientID > 0 {
		q = q.Where("patient_id = ?", patientID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var orders []model.RadiologyOrder
	err := q.Order("ordered_at DESC").Find(&orders).Error
	return orders, err
}

func (r *DoctorRepo) GetRadiologyOrder(radiologyID uint) (*model.RadiologyOrder, error) {
	var order model.RadiologyOrder
	err := r.DB.Preload("Patient").Preload("Doctor").
		First(&order, radiologyID).Error
	return &order, err
}

func (r *DoctorRepo) CreateRadiologyOrder(ro *model.RadiologyOrder) error {
	return r.DB.Create(ro).Error
}

// ── Admissions ────────────────────────────────────────────────────────

func (r *DoctorRepo) GetBed(bedID uint) (*model.Bed, error) {
	var bed model.Bed
	err := r.DB.Preload("Ward").First(&bed, bedID).Error
	return &bed, err
}

func (r *DoctorRepo) CreateAdmissionTx(adm *model.Admission) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var bed model.Bed
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Preload("Ward").First(&bed, adm.BedID).Error; err != nil {
			return errors.New("bed not found")
		}
		if bed.Status != "available" {
			return errors.New("bed is not available")
		}
		// Infer dept from ward
		if adm.DeptID == 0 {
			adm.DeptID = bed.Ward.DeptID
		}
		if err := tx.Create(adm).Error; err != nil {
			return err
		}
		return tx.Model(&model.Bed{}).Where("bed_id = ?", adm.BedID).
			Update("status", "occupied").Error
	})
}

func (r *DoctorRepo) GetAdmission(admissionID uint) (*model.Admission, error) {
	var adm model.Admission
	err := r.DB.Preload("Patient").Preload("Ward").Preload("Bed").Preload("Department").
		First(&adm, admissionID).Error
	return &adm, err
}

func (r *DoctorRepo) DischargeTx(admissionID uint, summary string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var adm model.Admission
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&adm, admissionID).Error; err != nil {
			return errors.New("admission not found")
		}
		if adm.Status != "admitted" {
			return errors.New("admission is already discharged")
		}

		now := time.Now()
		plan := adm.TreatmentPlan
		if summary != "" {
			plan += "\n\nDISCHARGE SUMMARY:\n" + summary
		}

		if err := tx.Model(&model.Admission{}).Where("admission_id = ?", admissionID).
			Updates(map[string]interface{}{
				"status":         "discharged",
				"discharged_at":  now,
				"treatment_plan": plan,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Bed{}).Where("bed_id = ?", adm.BedID).
			Update("status", "available").Error
	})
}
