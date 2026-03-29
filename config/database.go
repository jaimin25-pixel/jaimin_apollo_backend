package config

import (
	"fmt"
	"log"

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

		// Finance
		&model.Invoice{},
		&model.InsuranceClaim{},

		// HR
		&model.StaffShift{},
		&model.LeaveRequest{},

		// Audit & Partners
		&model.AuditLog{},
		&model.PartnerPharmacy{},

		// Hospital Config
		&model.HospitalConfig{},
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

		// invoices
		{"invoices", "patient_id", "patients", "patient_id"},
		{"invoices", "admission_id", "admissions", "admission_id"},
		{"invoices", "appt_id", "appointments", "appt_id"},
		{"invoices", "insurance_claim_id", "insurance_claims", "claim_id"},
		{"invoices", "created_by", "staff", "staff_id"},

		// insurance_claims
		{"insurance_claims", "invoice_id", "invoices", "invoice_id"},
		{"insurance_claims", "patient_id", "patients", "patient_id"},

		// staff_shifts
		{"staff_shifts", "staff_id", "staff", "staff_id"},

		// leave_requests
		{"leave_requests", "staff_id", "staff", "staff_id"},
		{"leave_requests", "approved_by", "staff", "staff_id"},

		// password_resets
		{"password_resets", "user_id", "users", "id"},
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

		// invoices
		`CREATE INDEX IF NOT EXISTS idx_invoice_patient ON invoices(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_status ON invoices(status)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_created ON invoices(created_at)`,

		// insurance_claims
		`CREATE INDEX IF NOT EXISTS idx_claim_invoice ON insurance_claims(invoice_id)`,
		`CREATE INDEX IF NOT EXISTS idx_claim_patient ON insurance_claims(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_claim_status ON insurance_claims(status)`,

		// staff_shifts
		`CREATE INDEX IF NOT EXISTS idx_shift_staff ON staff_shifts(staff_id)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_date ON staff_shifts(shift_date)`,
		`CREATE INDEX IF NOT EXISTS idx_shift_type ON staff_shifts(shift_type)`,

		// leave_requests
		`CREATE INDEX IF NOT EXISTS idx_leave_staff ON leave_requests(staff_id)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_status ON leave_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_dates ON leave_requests(from_date, to_date)`,

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
