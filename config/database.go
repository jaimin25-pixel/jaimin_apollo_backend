package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"apollo-backend/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Kolkata",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	dropLegacyTables(db)
	migrate(db)
	addForeignKeys(db)
	addIndexes(db)
	seed(db)
	log.Println("database connected, migrated & indexed")
	return db
}

// dropLegacyTables removes old UUID-based tables that conflict with the new SERIAL PK schema.
func dropLegacyTables(db *gorm.DB) {
	// Check if old users table has UUID PK
	var colType string
	row := db.Raw(`SELECT data_type FROM information_schema.columns
		WHERE table_name = 'users' AND column_name = 'id' AND table_schema = CURRENT_SCHEMA()`).Row()
	if row != nil && row.Scan(&colType) == nil && colType == "uuid" {
		log.Println("detected legacy UUID-based tables, dropping for re-creation...")
		db.Exec("DROP TABLE IF EXISTS password_resets CASCADE")
		db.Exec("DROP TABLE IF EXISTS audit_logs CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
	}
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		// Admin & Auth
		&model.User{},
		&model.Admin{},
		&model.PasswordReset{},

		// Core entities
		&model.Department{},
		&model.Doctor{},
		&model.Pharmacist{},
		&model.Staff{},
		&model.Patient{},

		// Facility
		&model.Ward{},
		&model.Bed{},

		// Clinical
		&model.Appointment{},
		&model.Admission{},
		&model.Prescription{},
		&model.PrescriptionItem{},
		&model.ClinicalNote{},

		// Pharmacy
		&model.Medicine{},
		&model.MedicineBatch{},

		// Lab & Radiology
		&model.LabTest{},
		&model.LabOrder{},
		&model.RadiologyOrder{},

		// OT & Nursing
		&model.OTSchedule{},
		&model.Vital{},
		&model.NursingNote{},
		&model.MedicationAdminRecord{},
		&model.BedTransferRequest{},

		// Finance
		&model.Invoice{},
		&model.InsuranceClaim{},
		&model.LedgerEntry{},

		// HR
		&model.StaffShift{},
		&model.LeaveRequest{},
		&model.Payroll{},

		// Audit & Partners
		&model.AuditLog{},
		&model.PartnerPharmacy{},

		// Hospital Config
		&model.HospitalConfig{},

		// Receptionist
		&model.VisitorLog{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

// addForeignKeys adds all FK constraints after tables are created.
func addForeignKeys(db *gorm.DB) {
	fks := []struct {
		table, column, refTable, refColumn string
	}{
		// departments
		{"departments", "hod_doctor_id", "doctors", "doctor_id"},

		// doctors
		{"doctors", "dept_id", "departments", "dept_id"},

		// staff
		{"staff", "dept_id", "departments", "dept_id"},

		// wards
		{"wards", "dept_id", "departments", "dept_id"},

		// beds
		{"beds", "ward_id", "wards", "ward_id"},

		// appointments
		{"appointments", "patient_id", "patients", "patient_id"},
		{"appointments", "doctor_id", "doctors", "doctor_id"},
		{"appointments", "dept_id", "departments", "dept_id"},
		{"appointments", "created_by_staff_id", "staff", "staff_id"},

		// admissions
		{"admissions", "patient_id", "patients", "patient_id"},
		{"admissions", "admitting_doctor_id", "doctors", "doctor_id"},
		{"admissions", "ward_id", "wards", "ward_id"},
		{"admissions", "bed_id", "beds", "bed_id"},
		{"admissions", "dept_id", "departments", "dept_id"},

		// prescriptions
		{"prescriptions", "doctor_id", "doctors", "doctor_id"},
		{"prescriptions", "patient_id", "patients", "patient_id"},
		{"prescriptions", "appt_id", "appointments", "appt_id"},
		{"prescriptions", "admission_id", "admissions", "admission_id"},

		// prescription_items
		{"prescription_items", "rx_id", "prescriptions", "rx_id"},
		{"prescription_items", "medicine_id", "medicines", "medicine_id"},

		// medicine_batches
		{"medicine_batches", "medicine_id", "medicines", "medicine_id"},

		// lab_tests
		{"lab_tests", "dept_id", "departments", "dept_id"},

		// lab_orders
		{"lab_orders", "patient_id", "patients", "patient_id"},
		{"lab_orders", "doctor_id", "doctors", "doctor_id"},
		{"lab_orders", "test_id", "lab_tests", "test_id"},
		{"lab_orders", "appt_id", "appointments", "appt_id"},
		{"lab_orders", "admission_id", "admissions", "admission_id"},
		{"lab_orders", "technician_id", "staff", "staff_id"},

		// radiology_orders
		{"radiology_orders", "patient_id", "patients", "patient_id"},
		{"radiology_orders", "doctor_id", "doctors", "doctor_id"},
		{"radiology_orders", "radiologist_id", "staff", "staff_id"},

		// ot_schedules
		{"ot_schedules", "room_id", "wards", "ward_id"},
		{"ot_schedules", "patient_id", "patients", "patient_id"},
		{"ot_schedules", "surgeon_id", "doctors", "doctor_id"},
		{"ot_schedules", "anesthesiologist_id", "staff", "staff_id"},

		// vitals
		{"vitals", "patient_id", "patients", "patient_id"},
		{"vitals", "admission_id", "admissions", "admission_id"},
		{"vitals", "nurse_id", "staff", "staff_id"},

		// nursing_notes
		{"nursing_notes", "patient_id", "patients", "patient_id"},
		{"nursing_notes", "admission_id", "admissions", "admission_id"},
		{"nursing_notes", "nurse_id", "staff", "staff_id"},

		// medication_admin_records
		{"medication_admin_records", "patient_id", "patients", "patient_id"},
		{"medication_admin_records", "admission_id", "admissions", "admission_id"},
		{"medication_admin_records", "rx_item_id", "prescription_items", "rx_item_id"},
		{"medication_admin_records", "nurse_id", "staff", "staff_id"},

		// bed_transfer_requests
		{"bed_transfer_requests", "patient_id", "patients", "patient_id"},
		{"bed_transfer_requests", "from_bed_id", "beds", "bed_id"},
		{"bed_transfer_requests", "to_bed_id", "beds", "bed_id"},
		{"bed_transfer_requests", "requested_by", "staff", "staff_id"},
		{"bed_transfer_requests", "confirmed_by", "staff", "staff_id"},

		// invoices
		{"invoices", "patient_id", "patients", "patient_id"},
		{"invoices", "admission_id", "admissions", "admission_id"},
		{"invoices", "appt_id", "appointments", "appt_id"},
		{"invoices", "insurance_claim_id", "insurance_claims", "claim_id"},
		{"invoices", "created_by", "staff", "staff_id"},

		// insurance_claims
		{"insurance_claims", "invoice_id", "invoices", "invoice_id"},
		{"insurance_claims", "patient_id", "patients", "patient_id"},

		// ledger_entries
		{"ledger_entries", "recorded_by", "staff", "staff_id"},

		// staff_shifts
		{"staff_shifts", "staff_id", "staff", "staff_id"},

		// leave_requests
		{"leave_requests", "staff_id", "staff", "staff_id"},
		{"leave_requests", "approved_by", "staff", "staff_id"},

		// payrolls
		{"payrolls", "staff_id", "staff", "staff_id"},
		{"payrolls", "generated_by", "users", "id"},

		// password_resets
		{"password_resets", "user_id", "users", "id"},

		// clinical_notes
		{"clinical_notes", "patient_id", "patients", "patient_id"},
		{"clinical_notes", "doctor_id", "doctors", "doctor_id"},
		{"clinical_notes", "appt_id", "appointments", "appt_id"},
		{"clinical_notes", "admission_id", "admissions", "admission_id"},

		// visitor_logs
		{"visitor_logs", "patient_id", "patients", "patient_id"},
		{"visitor_logs", "logged_by_staff_id", "staff", "staff_id"},
	}

	for _, fk := range fks {
		constraintName := fmt.Sprintf("fk_%s_%s", fk.table, fk.column)
		sql := fmt.Sprintf(
			`DO $$ BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = '%s') THEN
					ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s);
				END IF;
			END $$;`,
			constraintName, fk.table, constraintName, fk.column, fk.refTable, fk.refColumn,
		)
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("FK %s: %v", constraintName, err)
		}
	}
}

// addIndexes adds performance indexes matching the HMS document.
func addIndexes(db *gorm.DB) {
	indexes := []string{
		// doctors
		`CREATE INDEX IF NOT EXISTS idx_doctors_dept ON doctors(dept_id)`,
		`CREATE INDEX IF NOT EXISTS idx_doctors_status ON doctors(status)`,
		`CREATE INDEX IF NOT EXISTS idx_doctors_specialization ON doctors(specialization)`,

		// staff
		`CREATE INDEX IF NOT EXISTS idx_staff_role ON staff(role)`,
		`CREATE INDEX IF NOT EXISTS idx_staff_dept ON staff(dept_id)`,
		`CREATE INDEX IF NOT EXISTS idx_staff_status ON staff(status)`,

		// patients
		`CREATE INDEX IF NOT EXISTS idx_patients_gender ON patients(gender)`,
		`CREATE INDEX IF NOT EXISTS idx_patients_blood ON patients(blood_group)`,
		`CREATE INDEX IF NOT EXISTS idx_patients_contact ON patients(contact_number)`,

		// appointments
		`CREATE INDEX IF NOT EXISTS idx_appt_patient ON appointments(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appt_doctor ON appointments(doctor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appt_dept ON appointments(dept_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appt_scheduled ON appointments(scheduled_at)`,
		`CREATE INDEX IF NOT EXISTS idx_appt_status ON appointments(status)`,

		// admissions
		`CREATE INDEX IF NOT EXISTS idx_admission_patient ON admissions(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_admission_doctor ON admissions(admitting_doctor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_admission_ward ON admissions(ward_id)`,
		`CREATE INDEX IF NOT EXISTS idx_admission_bed ON admissions(bed_id)`,
		`CREATE INDEX IF NOT EXISTS idx_admission_status ON admissions(status)`,

		// wards
		`CREATE INDEX IF NOT EXISTS idx_ward_dept ON wards(dept_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ward_type ON wards(ward_type)`,
		`CREATE INDEX IF NOT EXISTS idx_ward_status ON wards(status)`,

		// beds
		`CREATE INDEX IF NOT EXISTS idx_bed_ward ON beds(ward_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bed_status ON beds(status)`,
		`CREATE INDEX IF NOT EXISTS idx_bed_type ON beds(bed_type)`,

		// prescriptions
		`CREATE INDEX IF NOT EXISTS idx_rx_doctor ON prescriptions(doctor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rx_patient ON prescriptions(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rx_status ON prescriptions(status)`,

		// prescription_items
		`CREATE INDEX IF NOT EXISTS idx_rx_item_rx ON prescription_items(rx_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rx_item_medicine ON prescription_items(medicine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rx_item_status ON prescription_items(status)`,

		// medicines
		`CREATE INDEX IF NOT EXISTS idx_medicine_category ON medicines(category)`,
		`CREATE INDEX IF NOT EXISTS idx_medicine_status ON medicines(status)`,
		`CREATE INDEX IF NOT EXISTS idx_medicine_generic ON medicines(generic_name)`,

		// medicine_batches
		`CREATE INDEX IF NOT EXISTS idx_batch_medicine ON medicine_batches(medicine_id)`,
		`CREATE INDEX IF NOT EXISTS idx_batch_expiry ON medicine_batches(expiry_date)`,

		// lab_tests
		`CREATE INDEX IF NOT EXISTS idx_labtest_dept ON lab_tests(dept_id)`,
		`CREATE INDEX IF NOT EXISTS idx_labtest_status ON lab_tests(status)`,

		// lab_orders
		`CREATE INDEX IF NOT EXISTS idx_laborder_patient ON lab_orders(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_laborder_doctor ON lab_orders(doctor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_laborder_test ON lab_orders(test_id)`,
		`CREATE INDEX IF NOT EXISTS idx_laborder_status ON lab_orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_laborder_ordered ON lab_orders(ordered_at)`,

		// radiology_orders
		`CREATE INDEX IF NOT EXISTS idx_radiology_patient ON radiology_orders(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_radiology_doctor ON radiology_orders(doctor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_radiology_modality ON radiology_orders(modality)`,
		`CREATE INDEX IF NOT EXISTS idx_radiology_status ON radiology_orders(status)`,

		// ot_schedules
		`CREATE INDEX IF NOT EXISTS idx_ot_patient ON ot_schedules(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ot_surgeon ON ot_schedules(surgeon_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ot_room ON ot_schedules(room_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ot_scheduled ON ot_schedules(scheduled_at)`,
		`CREATE INDEX IF NOT EXISTS idx_ot_status ON ot_schedules(status)`,

		// vitals
		`CREATE INDEX IF NOT EXISTS idx_vital_patient ON vitals(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vital_admission ON vitals(admission_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vital_nurse ON vitals(nurse_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vital_recorded ON vitals(recorded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_vital_critical ON vitals(is_critical)`,

		// nursing_notes
		`CREATE INDEX IF NOT EXISTS idx_nnote_patient ON nursing_notes(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nnote_nurse ON nursing_notes(nurse_id)`,

		// medication_admin_records
		`CREATE INDEX IF NOT EXISTS idx_mar_patient ON medication_admin_records(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mar_rx_item ON medication_admin_records(rx_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mar_nurse ON medication_admin_records(nurse_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mar_scheduled ON medication_admin_records(scheduled_at)`,

		// bed_transfer_requests
		`CREATE INDEX IF NOT EXISTS idx_btransfer_patient ON bed_transfer_requests(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_btransfer_status ON bed_transfer_requests(status)`,

		// invoices
		`CREATE INDEX IF NOT EXISTS idx_invoice_patient ON invoices(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_status ON invoices(status)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_created ON invoices(created_at)`,

		// insurance_claims
		`CREATE INDEX IF NOT EXISTS idx_claim_invoice ON insurance_claims(invoice_id)`,
		`CREATE INDEX IF NOT EXISTS idx_claim_patient ON insurance_claims(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_claim_status ON insurance_claims(status)`,

		// ledger_entries
		`CREATE INDEX IF NOT EXISTS idx_ledger_type ON ledger_entries(transaction_type)`,
		`CREATE INDEX IF NOT EXISTS idx_ledger_recorded ON ledger_entries(recorded_at)`,

		// staff_shifts
		`CREATE INDEX IF NOT EXISTS idx_shift_staff ON staff_shifts(staff_id)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_date ON staff_shifts(shift_date)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_type ON staff_shifts(shift_type)`,

		// leave_requests
		`CREATE INDEX IF NOT EXISTS idx_leave_staff ON leave_requests(staff_id)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_status ON leave_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_dates ON leave_requests(from_date, to_date)`,

		// payrolls
		`CREATE INDEX IF NOT EXISTS idx_payroll_staff ON payrolls(staff_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payroll_period ON payrolls(month, year)`,
		`CREATE INDEX IF NOT EXISTS idx_payroll_status ON payrolls(status)`,

		// audit_logs
		`CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_table ON audit_logs(table_name)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit_logs(ts)`,

		// departments
		`CREATE INDEX IF NOT EXISTS idx_dept_status ON departments(status)`,

		// pharmacists
		`CREATE INDEX IF NOT EXISTS idx_pharmacist_status ON pharmacists(status)`,

		// partner_pharmacies
		`CREATE INDEX IF NOT EXISTS idx_partner_status ON partner_pharmacies(status)`,

		// clinical_notes
		`CREATE INDEX IF NOT EXISTS idx_clinicalnote_patient ON clinical_notes(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clinicalnote_doctor ON clinical_notes(doctor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clinicalnote_appt ON clinical_notes(appt_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clinicalnote_admission ON clinical_notes(admission_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clinicalnote_created ON clinical_notes(created_at)`,
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("index: %v", err)
		}
	}
}

func seed(db *gorm.DB) {
	seedDepartments(db)
	seedAdmin(db)
	seedStaff(db)
	seedMedicines(db)
	seedPatients(db)
}

// seedPatients inserts demo patients, OPD appointments, IPD admissions, and invoices.
// Idempotent: only runs when the patients table is empty.
func seedPatients(db *gorm.DB) {
	var count int64
	db.Model(&model.Patient{}).Count(&count)
	hash, _ := bcrypt.GenerateFromPassword([]byte("Patient@123"), 12)
	hashedPwd := string(hash)

	if count > 0 {
		// Backfill existing patients with email/password if empty
		db.Exec("UPDATE patients SET email = LOWER(REPLACE(full_name, ' ', '.')) || '@patient.apollo.health', hashed_password = ? WHERE email IS NULL OR email = ''", hashedPwd)
		return
	}

	// ── Wards (idempotent) ─────────────────────────────────────────────────────
	var wardCount int64
	db.Model(&model.Ward{}).Count(&wardCount)
	if wardCount == 0 {
		wards := []model.Ward{
			{DeptID: 1, Name: "General Ward A", WardType: "general", Capacity: 20, Status: "active"},
			{DeptID: 2, Name: "Cardiac Ward", WardType: "icu", Capacity: 10, Status: "active"},
			{DeptID: 3, Name: "Neuro Ward", WardType: "general", Capacity: 15, Status: "active"},
			{DeptID: 5, Name: "Pediatric Ward", WardType: "general", Capacity: 12, Status: "active"},
			{DeptID: 19, Name: "Emergency Ward", WardType: "emergency", Capacity: 8, Status: "active"},
		}
		for i := range wards {
			db.Create(&wards[i])
		}
		log.Println("seeded 5 wards")
	}

	// ── Beds ─────────────────────────────────────────────────────────────────────
	var bedCount int64
	db.Model(&model.Bed{}).Count(&bedCount)
	if bedCount == 0 {
		var allWards []model.Ward
		db.Find(&allWards)
		bedIdx := 0
		bedTypes := []string{"general", "icu", "isolation", "general"}
		for _, w := range allWards {
			bedsPerWard := 4
			if w.WardType == "icu" || w.WardType == "emergency" {
				bedsPerWard = 4
			}
			for j := 0; j < bedsPerWard; j++ {
				bedIdx++
				bed := model.Bed{
					WardID:    w.WardID,
					BedNumber: fmt.Sprintf("B-%03d", bedIdx),
					BedType:   bedTypes[j%len(bedTypes)],
					Status:    "available",
				}
				db.Create(&bed)
			}
		}
		log.Printf("seeded %d beds across wards\n", bedIdx)
	}

	// ── Patients ─────────────────────────────────────────────────────────────────
	type patientData struct {
		FullName, Gender, BloodGroup, Phone, Address, EmergencyName, EmergencyPhone, InsID, InsProv string
	}

	patients := []patientData{
		{"Aarav Sharma", "male", "A+", "+91-9876543001", "12, MG Road, Mumbai", "Priya Sharma", "+91-9876543101", "INS-001", "Star Health"},
		{"Diya Patel", "female", "B+", "+91-9876543002", "45, SG Highway, Ahmedabad", "Rajesh Patel", "+91-9876543102", "INS-002", "ICICI Lombard"},
		{"Vihaan Reddy", "male", "O+", "+91-9876543003", "78, Hi-Tech City, Hyderabad", "Lakshmi Reddy", "+91-9876543103", "", ""},
		{"Ananya Gupta", "female", "AB+", "+91-9876543004", "23, Civil Lines, Delhi", "Suresh Gupta", "+91-9876543104", "INS-004", "Max Bupa"},
		{"Arjun Singh", "male", "A-", "+91-9876543005", "56, Rajpath, Jaipur", "Meena Singh", "+91-9876543105", "", ""},
		{"Ishita Nair", "female", "B-", "+91-9876543006", "89, Marine Drive, Kochi", "Gopal Nair", "+91-9876543106", "INS-006", "Bajaj Allianz"},
		{"Kabir Kumar", "male", "O-", "+91-9876543007", "34, Park Street, Kolkata", "Rina Kumar", "+91-9876543107", "", ""},
		{"Saanvi Joshi", "female", "A+", "+91-9876543008", "67, FC Road, Pune", "Amit Joshi", "+91-9876543108", "INS-008", "New India Assurance"},
		{"Reyansh Mehta", "male", "B+", "+91-9876543009", "90, Bandra West, Mumbai", "Neha Mehta", "+91-9876543109", "INS-009", "Star Health"},
		{"Myra Iyer", "female", "AB-", "+91-9876543010", "15, Adyar, Chennai", "Subramanian Iyer", "+91-9876543110", "", ""},
	}

	dobs := []string{
		"1990-05-15", "1985-08-22", "2000-01-10", "1992-11-30", "1978-03-18",
		"1995-07-05", "1988-12-25", "2002-09-14", "1980-06-08", "1998-04-20",
	}

	createdPatients := make([]model.Patient, 0, len(patients))
	for i, p := range patients {
		dob, _ := time.Parse("2006-01-02", dobs[i])
		pat := model.Patient{
			PatCode:               fmt.Sprintf("PAT-%06d", i+1),
			Email:                 strings.ToLower(strings.ReplaceAll(p.FullName, " ", ".")) + "@patient.apollo.health",
			HashedPassword:        hashedPwd,
			FullName:              p.FullName,
			DateOfBirth:           dob,
			Gender:                p.Gender,
			BloodGroup:            p.BloodGroup,
			ContactNumber:         p.Phone,
			Address:               p.Address,
			EmergencyContactName:  p.EmergencyName,
			EmergencyContactPhone: p.EmergencyPhone,
			InsuranceID:           p.InsID,
			InsuranceProvider:     p.InsProv,
		}
		if err := db.Create(&pat).Error; err != nil {
			log.Printf("seed patient %s: %v", p.FullName, err)
			continue
		}
		createdPatients = append(createdPatients, pat)
	}
	log.Printf("seeded %d patients", len(createdPatients))

	// ── Doctors lookup ───────────────────────────────────────────────────────────
	var doctors []model.Doctor
	db.Where("status = 'active'").Limit(5).Find(&doctors)
	if len(doctors) == 0 {
		log.Println("no active doctors found, skipping appointment/admission seed")
		return
	}

	// ── Departments lookup ───────────────────────────────────────────────────────
	var depts []model.Department
	db.Where("status = 'active'").Limit(5).Find(&depts)
	if len(depts) == 0 {
		log.Println("no active departments found, skipping appointment seed")
		return
	}

	// ── Appointments (15) ────────────────────────────────────────────────────────
	statuses := []string{"scheduled", "checked_in", "in_consultation", "completed", "completed",
		"cancelled", "scheduled", "checked_in", "completed", "completed",
		"scheduled", "checked_in", "completed", "scheduled", "cancelled"}

	baseTime := time.Now().Truncate(24 * time.Hour).Add(9 * time.Hour) // today 9 AM

	createdAppts := make([]model.Appointment, 0)
	for i := 0; i < 15 && i < len(statuses); i++ {
		patIdx := i % len(createdPatients)
		docIdx := i % len(doctors)
		deptIdx := i % len(depts)

		scheduledAt := baseTime.Add(time.Duration(i) * 30 * time.Minute)
		if i >= 10 {
			scheduledAt = baseTime.AddDate(0, 0, -1).Add(time.Duration(i-10) * 30 * time.Minute)
		}

		token := ""
		if statuses[i] == "checked_in" || statuses[i] == "in_consultation" || statuses[i] == "completed" {
			token = fmt.Sprintf("QGM-%03d", i+1)
		}

		complaints := []string{
			"Fever and cold", "Chest pain", "Headache", "Back pain", "Skin rash",
			"Stomach ache", "Cough", "Joint pain", "Eye problem", "Breathing difficulty",
			"Follow-up visit", "Annual checkup", "Dizziness", "Fatigue", "Sore throat",
		}

		appt := model.Appointment{
			PatientID:      createdPatients[patIdx].PatientID,
			DoctorID:       doctors[docIdx].DoctorID,
			DeptID:         depts[deptIdx].DeptID,
			ScheduledAt:    scheduledAt,
			QueueToken:     token,
			Status:         statuses[i],
			ChiefComplaint: complaints[i],
		}
		if err := db.Create(&appt).Error; err != nil {
			log.Printf("seed appointment %d: %v", i+1, err)
			continue
		}
		createdAppts = append(createdAppts, appt)
	}
	log.Printf("seeded %d appointments", len(createdAppts))

	// ── Beds lookup for admissions ──────────────────────────────────────────────
	var beds []model.Bed
	db.Where("status = 'available'").Limit(5).Find(&beds)

	var wards []model.Ward
	db.Limit(5).Find(&wards)

	// ── Admissions (5) ──────────────────────────────────────────────────────────
	if len(beds) >= 5 && len(wards) >= 1 {
		admStatuses := []string{"admitted", "admitted", "discharged", "discharged", "admitted"}
		diagnoses := []string{
			"Pneumonia - Community acquired",
			"Acute myocardial infarction",
			"Fracture - Left tibia",
			"Dengue fever with thrombocytopenia",
			"Diabetic ketoacidosis",
		}
		treatments := []string{
			"IV antibiotics, respiratory support",
			"PCI with stent placement, dual antiplatelet therapy",
			"Open reduction internal fixation",
			"Platelet transfusion, fluid management",
			"Insulin drip, electrolyte correction",
		}

		for i := 0; i < 5; i++ {
			patIdx := i % len(createdPatients)
			docIdx := i % len(doctors)
			wardIdx := i % len(wards)
			bedIdx := i % len(beds)
			deptIdx := i % len(depts)

			admittedAt := time.Now().AddDate(0, 0, -(i * 3 + 1))
			var dischargedAt *time.Time
			if admStatuses[i] == "discharged" {
				t := admittedAt.AddDate(0, 0, i+2)
				dischargedAt = &t
			}

			adm := model.Admission{
				PatientID:         createdPatients[patIdx].PatientID,
				AdmittingDoctorID: doctors[docIdx].DoctorID,
				WardID:            wards[wardIdx].WardID,
				BedID:             beds[bedIdx].BedID,
				DeptID:            depts[deptIdx].DeptID,
				AdmittedAt:        admittedAt,
				Diagnosis:         diagnoses[i],
				TreatmentPlan:     treatments[i],
				DischargedAt:      dischargedAt,
				Status:            admStatuses[i],
			}
			if err := db.Create(&adm).Error; err != nil {
				log.Printf("seed admission %d: %v", i+1, err)
				continue
			}

			// Mark bed as occupied for active admissions
			if admStatuses[i] == "admitted" {
				db.Model(&model.Bed{}).Where("bed_id = ?", beds[bedIdx].BedID).
					Update("status", "occupied")
			}
		}
		log.Println("seeded 5 admissions")
	}

	// ── Invoices (8) ──────────────────────────────────────────────────────────
	invStatuses := []string{"paid", "finalized", "draft", "paid", "partially_paid", "paid", "draft", "cancelled"}
	for i := 0; i < 8; i++ {
		patIdx := i % len(createdPatients)

		consultation := float64(500 + (i * 100))
		procedure := float64(i * 250)
		lab := float64(200 + (i * 50))
		pharmacy := float64(150 + (i * 30))
		bed := float64(0)
		misc := float64(50 + (i * 10))

		if i < 5 {
			bed = float64(1000 + (i * 200))
		}

		subTotal := consultation + procedure + lab + pharmacy + bed + misc
		tax := subTotal * 0.05
		total := subTotal + tax
		var paid float64
		var balance float64

		switch invStatuses[i] {
		case "paid":
			paid = total
			balance = 0
		case "partially_paid":
			paid = total * 0.5
			balance = total - paid
		case "finalized", "draft":
			paid = 0
			balance = total
		case "cancelled":
			paid = 0
			balance = 0
		}

		paymentMode := ""
		if paid > 0 {
			modes := []string{"cash", "card", "upi", "insurance"}
			paymentMode = modes[i%len(modes)]
		}

		var apptID *uint
		if i < len(createdAppts) {
			id := createdAppts[i].ApptID
			apptID = &id
		}

		var finalizedAt *time.Time
		if invStatuses[i] == "paid" || invStatuses[i] == "finalized" || invStatuses[i] == "partially_paid" {
			t := time.Now().AddDate(0, 0, -i)
			finalizedAt = &t
		}

		inv := model.Invoice{
			PatientID:            createdPatients[patIdx].PatientID,
			ApptID:               apptID,
			ConsultationCharges:  consultation,
			ProcedureCharges:     procedure,
			LabCharges:           lab,
			PharmacyCharges:      pharmacy,
			BedCharges:           bed,
			MiscellaneousCharges: misc,
			SubTotal:             subTotal,
			TaxAmount:            tax,
			TotalAmount:          total,
			AmountPaid:           paid,
			BalanceDue:           balance,
			PaymentMode:          paymentMode,
			Status:               invStatuses[i],
			FinalizedAt:          finalizedAt,
		}
		if err := db.Create(&inv).Error; err != nil {
			log.Printf("seed invoice %d: %v", i+1, err)
		}
	}
	log.Println("seeded 8 invoices")
}

func seedMedicines(db *gorm.DB) {
	var count int64
	db.Model(&model.Medicine{}).Count(&count)
	if count > 0 {
		return
	}

	type medicineData struct {
		GenericName       string
		BrandName         string
		Category          string
		Unit              string
		HSNCode           string
		ReorderLevel      int
		StorageConditions string
	}

	medicines := []medicineData{
		{"Paracetamol", "Crocin / Dolo-650", "Analgesic / Antipyretic", "Tablet", "30049099", 200, "Store below 25°C, away from moisture"},
		{"Amoxicillin", "Mox / Novamox", "Antibiotic", "Capsule", "30041010", 100, "Store below 25°C in a dry place"},
		{"Metformin Hydrochloride", "Glycomet / Glucophage", "Antidiabetic", "Tablet", "30046020", 150, "Store at room temperature, away from light"},
		{"Atorvastatin", "Storvas / Lipitor", "Antilipidemic", "Tablet", "30049099", 100, "Store below 30°C, protect from moisture"},
		{"Omeprazole", "Omez / Prilosec", "Proton Pump Inhibitor", "Capsule", "30049099", 120, "Store below 25°C, keep container tightly closed"},
		{"Amlodipine Besylate", "Amlogard / Norvasc", "Antihypertensive", "Tablet", "30049099", 80, "Store below 30°C, protect from moisture"},
		{"Aspirin", "Ecosprin / Disprin", "Antiplatelet / Analgesic", "Tablet", "30049099", 200, "Store below 25°C in a dry place"},
		{"Losartan Potassium", "Repace / Cozaar", "ARB Antihypertensive", "Tablet", "30049099", 80, "Store below 30°C, protect from light"},
		{"Cetirizine Hydrochloride", "Cetzine / Zyrtec", "Antihistamine", "Tablet", "30045090", 100, "Store below 30°C"},
		{"Pantoprazole Sodium", "Pantocid / Pantop", "Proton Pump Inhibitor", "Tablet", "30049099", 120, "Store below 25°C, protect from moisture"},
		{"Azithromycin", "Azithral / Zithromax", "Antibiotic (Macrolide)", "Tablet", "30041010", 80, "Store below 30°C"},
		{"Ciprofloxacin Hydrochloride", "Ciplox / Cipro", "Antibiotic (Fluoroquinolone)", "Tablet", "30041010", 80, "Store below 25°C, protect from light"},
		{"Ibuprofen", "Brufen / Combiflam", "NSAID", "Tablet", "30049099", 150, "Store below 25°C, away from moisture"},
		{"Metronidazole", "Metrogyl / Flagyl", "Antibiotic / Antiprotozoal", "Tablet", "30041010", 80, "Store below 25°C, protect from light"},
		{"Cefixime", "Taxim-O / Suprax", "Antibiotic (Cephalosporin)", "Tablet", "30041010", 60, "Store below 30°C"},
		{"Dexamethasone", "Dexona / Decadron", "Corticosteroid", "Tablet", "30026010", 50, "Store below 30°C, protect from light"},
		{"Prednisolone", "Wysolone / Deltacortril", "Corticosteroid", "Tablet", "30026010", 50, "Store below 25°C, protect from light and moisture"},
		{"Ondansetron Hydrochloride", "Emeset / Zofran", "Antiemetic", "Tablet", "30049099", 80, "Store below 30°C"},
		{"Diclofenac Sodium", "Voveran / Voltaren", "NSAID", "Tablet", "30049099", 100, "Store below 25°C"},
		{"Furosemide", "Lasix / Frusemide", "Loop Diuretic", "Tablet", "30049099", 60, "Store below 25°C, protect from light"},
		{"Insulin Regular (Human)", "Actrapid HM / Humulin R", "Antidiabetic (Insulin)", "Vial (100 IU/mL)", "30043100", 30, "Refrigerate 2–8°C; do not freeze"},
		{"Insulin Glargine", "Lantus / Basalog", "Antidiabetic (Insulin)", "Vial (100 IU/mL)", "30043100", 30, "Refrigerate 2–8°C; do not freeze"},
		{"Lisinopril", "Listril / Zestril", "ACE Inhibitor", "Tablet", "30049099", 80, "Store below 30°C, protect from moisture"},
		{"Tramadol Hydrochloride", "Tramazac / Ultram", "Opioid Analgesic", "Tablet", "30049099", 50, "Store below 25°C, Schedule H drug"},
		{"Morphine Sulfate", "Morphine Sulfate IP", "Opioid Analgesic", "Ampoule (15 mg/mL)", "30049099", 20, "Refrigerate 2–8°C; Schedule X drug"},
		{"Salbutamol", "Asthalin / Ventolin", "Bronchodilator (β2-agonist)", "Inhaler (100 mcg/dose)", "30039090", 40, "Store below 30°C, protect from sunlight and frost"},
		{"Theophylline", "Uniphyl / Theo-SR", "Bronchodilator (Xanthine)", "Tablet", "30039090", 60, "Store below 30°C"},
		{"Digoxin", "Lanoxin / Digoxin IP", "Cardiac Glycoside", "Tablet", "30029099", 30, "Store below 25°C, protect from light"},
		{"Warfarin Sodium", "Warf / Coumadin", "Anticoagulant", "Tablet", "30049099", 40, "Store below 25°C, protect from light"},
		{"Clopidogrel", "Plavix / Clopilet", "Antiplatelet", "Tablet", "30049099", 80, "Store below 30°C"},
	}

	type batchData struct {
		BatchNumber   string
		Supplier      string
		ExpiryMonths  int // months from a fixed base date
		PurchaseDate  string
		Quantity      float64
		PurchasePrice float64
	}

	batchTemplates := [][]batchData{
		// Paracetamol
		{{"BT-PCM-2401", "Sun Pharma Ltd.", 18, "2024-01-10", 5000, 0.85}, {"BT-PCM-2402", "Cipla Ltd.", 24, "2024-06-15", 3000, 0.90}},
		// Amoxicillin
		{{"BT-AMX-2401", "Ranbaxy Labs", 14, "2024-02-20", 2000, 4.50}, {"BT-AMX-2402", "Alkem Labs", 20, "2024-08-05", 1500, 4.80}},
		// Metformin
		{{"BT-MET-2401", "USV Ltd.", 22, "2024-01-15", 3000, 2.10}, {"BT-MET-2402", "Micro Labs", 28, "2024-09-10", 2000, 2.20}},
		// Atorvastatin
		{{"BT-ATV-2401", "Pfizer India", 20, "2024-03-01", 1500, 8.50}, {"BT-ATV-2402", "Torrent Pharma", 26, "2024-10-15", 1000, 8.75}},
		// Omeprazole
		{{"BT-OMP-2401", "AstraZeneca India", 16, "2024-02-10", 2500, 3.20}, {"BT-OMP-2402", "Cipla Ltd.", 22, "2024-07-20", 2000, 3.40}},
		// Amlodipine
		{{"BT-AML-2401", "Pfizer India", 24, "2024-04-05", 1200, 5.60}, {"BT-AML-2402", "Sun Pharma", 30, "2024-11-10", 800, 5.80}},
		// Aspirin
		{{"BT-ASP-2401", "Bayer India", 20, "2024-01-20", 4000, 0.60}, {"BT-ASP-2402", "Zydus Cadila", 26, "2024-08-25", 3000, 0.65}},
		// Losartan
		{{"BT-LOS-2401", "MSD Pharmaceuticals", 18, "2024-03-15", 1000, 6.20}, {"BT-LOS-2402", "Sun Pharma", 24, "2024-09-20", 800, 6.40}},
		// Cetirizine
		{{"BT-CTZ-2401", "UCB India", 22, "2024-02-05", 2000, 1.80}, {"BT-CTZ-2402", "Cipla Ltd.", 28, "2024-08-01", 1500, 1.90}},
		// Pantoprazole
		{{"BT-PNT-2401", "Wyeth Pharma", 18, "2024-04-10", 2500, 4.10}, {"BT-PNT-2402", "Alkem Labs", 24, "2024-10-05", 2000, 4.30}},
		// Azithromycin
		{{"BT-AZT-2401", "Pfizer India", 14, "2024-03-20", 1000, 12.50}, {"BT-AZT-2402", "Cipla Ltd.", 20, "2024-09-15", 800, 13.00}},
		// Ciprofloxacin
		{{"BT-CPF-2401", "Bayer India", 16, "2024-02-28", 1500, 7.80}, {"BT-CPF-2402", "Sun Pharma", 22, "2024-08-30", 1000, 8.10}},
		// Ibuprofen
		{{"BT-IBP-2401", "Abbott India", 20, "2024-01-30", 3000, 1.50}, {"BT-IBP-2402", "Cipla Ltd.", 26, "2024-07-15", 2000, 1.60}},
		// Metronidazole
		{{"BT-MTZ-2401", "Flagyl India", 18, "2024-03-05", 2000, 1.20}, {"BT-MTZ-2402", "Sun Pharma", 24, "2024-09-01", 1500, 1.30}},
		// Cefixime
		{{"BT-CFX-2401", "Lupin Ltd.", 14, "2024-04-15", 800, 18.00}, {"BT-CFX-2402", "Alkem Labs", 20, "2024-10-20", 600, 18.50}},
		// Dexamethasone
		{{"BT-DEX-2401", "Fulford India", 22, "2024-02-15", 500, 2.50}, {"BT-DEX-2402", "Cipla Ltd.", 28, "2024-08-10", 400, 2.70}},
		// Prednisolone
		{{"BT-PRD-2401", "Sun Pharma", 20, "2024-03-10", 600, 1.80}, {"BT-PRD-2402", "Wyeth Pharma", 26, "2024-09-05", 500, 1.90}},
		// Ondansetron
		{{"BT-OND-2401", "GSK India", 18, "2024-04-01", 1500, 5.20}, {"BT-OND-2402", "Cipla Ltd.", 24, "2024-10-01", 1200, 5.50}},
		// Diclofenac
		{{"BT-DCF-2401", "Novartis India", 20, "2024-02-20", 2000, 2.80}, {"BT-DCF-2402", "Sun Pharma", 26, "2024-08-15", 1500, 2.90}},
		// Furosemide
		{{"BT-FRS-2401", "Sanofi India", 22, "2024-03-25", 800, 1.20}, {"BT-FRS-2402", "Cipla Ltd.", 28, "2024-09-25", 600, 1.30}},
		// Insulin Regular
		{{"BT-INR-2401", "Novo Nordisk India", 12, "2024-04-20", 200, 180.00}, {"BT-INR-2402", "Eli Lilly India", 18, "2024-10-10", 150, 190.00}},
		// Insulin Glargine
		{{"BT-ING-2401", "Sanofi India", 12, "2024-04-25", 150, 950.00}, {"BT-ING-2402", "Biocon Ltd.", 18, "2024-10-15", 100, 820.00}},
		// Lisinopril
		{{"BT-LSN-2401", "AstraZeneca India", 22, "2024-03-30", 1000, 4.50}, {"BT-LSN-2402", "Cipla Ltd.", 28, "2024-09-30", 800, 4.70}},
		// Tramadol
		{{"BT-TRM-2401", "Grunenthal India", 18, "2024-04-05", 500, 6.80}, {"BT-TRM-2402", "Sun Pharma", 24, "2024-10-05", 400, 7.00}},
		// Morphine Sulfate
		{{"BT-MRP-2401", "Neon Laboratories", 24, "2024-01-25", 100, 45.00}, {"BT-MRP-2402", "Rusan Pharma", 30, "2024-07-25", 80, 48.00}},
		// Salbutamol
		{{"BT-SLB-2401", "GSK India", 24, "2024-02-01", 300, 62.00}, {"BT-SLB-2402", "Cipla Ltd.", 30, "2024-08-20", 250, 65.00}},
		// Theophylline
		{{"BT-TPH-2401", "Glenmark Pharma", 22, "2024-03-15", 700, 3.10}, {"BT-TPH-2402", "Cadila Pharma", 28, "2024-09-10", 500, 3.30}},
		// Digoxin
		{{"BT-DGX-2401", "GSK India", 24, "2024-04-10", 300, 4.20}, {"BT-DGX-2402", "Sun Pharma", 30, "2024-10-10", 250, 4.40}},
		// Warfarin
		{{"BT-WRF-2401", "Abbott India", 20, "2024-02-10", 400, 3.50}, {"BT-WRF-2402", "Cipla Ltd.", 26, "2024-08-05", 300, 3.70}},
		// Clopidogrel
		{{"BT-CLP-2401", "Sanofi India", 22, "2024-03-05", 1200, 9.80}, {"BT-CLP-2402", "Sun Pharma", 28, "2024-09-20", 1000, 10.20}},
	}

	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i, m := range medicines {
		med := model.Medicine{
			GenericName:       m.GenericName,
			BrandName:         m.BrandName,
			Category:          m.Category,
			Unit:              m.Unit,
			HSNCode:           m.HSNCode,
			ReorderLevel:      m.ReorderLevel,
			CurrentStock:      0,
			StorageConditions: m.StorageConditions,
			Status:            "active",
		}
		if err := db.Create(&med).Error; err != nil {
			log.Printf("seed medicine %s: %v", m.GenericName, err)
			continue
		}

		totalStock := 0.0
		if i < len(batchTemplates) {
			for _, bt := range batchTemplates[i] {
				purchaseDate, _ := time.Parse("2006-01-02", bt.PurchaseDate)
				expiryDate := baseDate.AddDate(0, bt.ExpiryMonths, 0)
				price := bt.PurchasePrice

				batch := model.MedicineBatch{
					MedicineID:        med.MedicineID,
					BatchNumber:       bt.BatchNumber,
					ExpiryDate:        expiryDate,
					Quantity:          bt.Quantity,
					QuantityRemaining: bt.Quantity,
					Supplier:          bt.Supplier,
					PurchaseDate:      purchaseDate,
					PurchasePrice:     &price,
				}
				if err := db.Create(&batch).Error; err != nil {
					log.Printf("seed batch %s: %v", bt.BatchNumber, err)
					continue
				}
				totalStock += bt.Quantity
			}
		}

		db.Model(&model.Medicine{}).Where("medicine_id = ?", med.MedicineID).
			Update("current_stock", totalStock)
	}

	log.Println("seeded 30 medicines with batches")
}

func seedAdmin(db *gorm.DB) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), 12)

	var count int64
	db.Model(&model.User{}).Where("email = ?", "admin@apollo.health").Count(&count)

	if count > 0 {
		// Update existing admin password
		db.Model(&model.User{}).Where("email = ?", "admin@apollo.health").
			Update("password_hash", string(hash))

		// Also update admin table if exists
		db.Model(&model.Admin{}).Where("email = ?", "admin@apollo.health").
			Update("hashed_password", string(hash))

		log.Println("updated default admin password: admin@apollo.health / Admin@123")
		return
	}

	admin := model.User{
		Email:        "admin@apollo.health",
		Username:     "admin",
		PasswordHash: string(hash),
		FullName:     "System Admin",
		Phone:        "+91-9876543210",
		Role:         "admin",
		IsActive:     true,
	}
	db.Create(&admin)

	var adminCount int64
	db.Model(&model.Admin{}).Where("email = ?", "admin@apollo.health").Count(&adminCount)
	if adminCount == 0 {
		adminRow := model.Admin{
			Email:          "admin@apollo.health",
			HashedPassword: string(hash),
		}
		db.Create(&adminRow)
	}

	log.Println("seeded default admin: admin@apollo.health / Admin@123")
}

func seedDepartments(db *gorm.DB) {
	var count int64
	db.Model(&model.Department{}).Count(&count)
	if count > 0 {
		// Ensure the 3 role-based departments exist even if other departments were already seeded
		roleDepts := []model.Department{
			{Name: "Clinical Department", BedCount: 0, HasICU: false, Status: "active"},
			{Name: "Administration Department", BedCount: 0, HasICU: false, Status: "active"},
			{Name: "Support Services Department", BedCount: 0, HasICU: false, Status: "active"},
		}
		for i := range roleDepts {
			var existing model.Department
			if err := db.Where("name = ?", roleDepts[i].Name).First(&existing).Error; err != nil {
				db.Create(&roleDepts[i])
				log.Printf("seeded department: %s", roleDepts[i].Name)
			}
		}
		return
	}

	departments := []model.Department{
		{Name: "General Medicine", BedCount: 50, HasICU: false, Status: "active"},
		{Name: "Cardiology", BedCount: 30, HasICU: true, Status: "active"},
		{Name: "Neurology", BedCount: 25, HasICU: true, Status: "active"},
		{Name: "Orthopedics", BedCount: 30, HasICU: false, Status: "active"},
		{Name: "Pediatrics", BedCount: 25, HasICU: true, Status: "active"},
		{Name: "Gynecology & Obstetrics", BedCount: 30, HasICU: false, Status: "active"},
		{Name: "Dermatology", BedCount: 10, HasICU: false, Status: "active"},
		{Name: "ENT", BedCount: 15, HasICU: false, Status: "active"},
		{Name: "Ophthalmology", BedCount: 15, HasICU: false, Status: "active"},
		{Name: "Psychiatry", BedCount: 10, HasICU: false, Status: "active"},
		{Name: "Oncology", BedCount: 20, HasICU: true, Status: "active"},
		{Name: "Nephrology", BedCount: 15, HasICU: true, Status: "active"},
		{Name: "Urology", BedCount: 15, HasICU: false, Status: "active"},
		{Name: "Pulmonology", BedCount: 20, HasICU: true, Status: "active"},
		{Name: "Gastroenterology", BedCount: 15, HasICU: false, Status: "active"},
		{Name: "Endocrinology", BedCount: 10, HasICU: false, Status: "active"},
		{Name: "Radiology", BedCount: 0, OTCount: 0, HasICU: false, Status: "active"},
		{Name: "Pathology", BedCount: 0, OTCount: 0, HasICU: false, Status: "active"},
		{Name: "Emergency Medicine", BedCount: 20, HasICU: true, Status: "active"},
		{Name: "General Surgery", BedCount: 30, OTCount: 4, HasICU: true, Status: "active"},
		{Name: "Clinical Department", BedCount: 0, HasICU: false, Status: "active"},
		{Name: "Administration Department", BedCount: 0, HasICU: false, Status: "active"},
		{Name: "Support Services Department", BedCount: 0, HasICU: false, Status: "active"},
	}

	for i := range departments {
		db.Create(&departments[i])
	}
	log.Println("seeded 23 departments")
}
