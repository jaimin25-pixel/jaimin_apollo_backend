package repository

import (
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

// ─── StaffRepo ────────────────────────────────────────────────────────────────
// Used by AuthService to authenticate staff (nurses, receptionists, etc.)

type StaffRepo struct{ DB *gorm.DB }

func NewStaffRepo(db *gorm.DB) *StaffRepo { return &StaffRepo{DB: db} }

func (r *StaffRepo) FindByEmail(email string) (*model.Staff, error) {
	var s model.Staff
	err := r.DB.Where("email = ? AND status = 'active'", email).First(&s).Error
	return &s, err
}

func (r *StaffRepo) FindByID(id uint) (*model.Staff, error) {
	var s model.Staff
	err := r.DB.Where("staff_id = ? AND status = 'active'", id).First(&s).Error
	return &s, err
}

// ─── QueueStats ───────────────────────────────────────────────────────────────

type QueueStats struct {
	TotalToday     int64          `json:"total_today"`
	CheckedInCount int64          `json:"checked_in_count"`
	WaitingByDept  []DeptQueueStat `json:"waiting_by_dept"`
}

type DeptQueueStat struct {
	DeptID   uint   `json:"dept_id"`
	DeptName string `json:"dept_name"`
	Waiting  int64  `json:"waiting"`
}

// ─── ReceptionistRepo ─────────────────────────────────────────────────────────

type ReceptionistRepo struct{ DB *gorm.DB }

func NewReceptionistRepo(db *gorm.DB) *ReceptionistRepo { return &ReceptionistRepo{DB: db} }

// ── Dashboard ─────────────────────────────────────────────────────────────────

func (r *ReceptionistRepo) TodayQueueStats() (*QueueStats, error) {
	stats := &QueueStats{}

	today := time.Now().Format("2006-01-02")

	// Total appointments today
	if err := r.DB.Model(&model.Appointment{}).
		Where("DATE(scheduled_at AT TIME ZONE 'Asia/Kolkata') = ?", today).
		Count(&stats.TotalToday).Error; err != nil {
		return nil, err
	}

	// Checked-in count
	if err := r.DB.Model(&model.Appointment{}).
		Where("DATE(scheduled_at AT TIME ZONE 'Asia/Kolkata') = ? AND status = 'checked_in'", today).
		Count(&stats.CheckedInCount).Error; err != nil {
		return nil, err
	}

	// Waiting (checked_in) grouped by department
	rows, err := r.DB.Raw(`
		SELECT a.dept_id, d.name AS dept_name, COUNT(*) AS waiting
		FROM appointments a
		JOIN departments d ON d.dept_id = a.dept_id
		WHERE DATE(a.scheduled_at AT TIME ZONE 'Asia/Kolkata') = ?
		  AND a.status = 'checked_in'
		GROUP BY a.dept_id, d.name
		ORDER BY waiting DESC
	`, today).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row DeptQueueStat
		if err := rows.Scan(&row.DeptID, &row.DeptName, &row.Waiting); err != nil {
			return nil, err
		}
		stats.WaitingByDept = append(stats.WaitingByDept, row)
	}
	if stats.WaitingByDept == nil {
		stats.WaitingByDept = []DeptQueueStat{}
	}
	return stats, nil
}

// ── Patient ───────────────────────────────────────────────────────────────────

func (r *ReceptionistRepo) CreatePatient(p *model.Patient) error {
	return r.DB.Create(p).Error
}

func (r *ReceptionistRepo) PatCodeExists(code string) bool {
	var count int64
	r.DB.Model(&model.Patient{}).Where("pat_code = ?", code).Count(&count)
	return count > 0
}

func (r *ReceptionistRepo) SearchPatients(q, bloodGroup, insurance string) ([]model.Patient, error) {
	var patients []model.Patient
	tx := r.DB.Model(&model.Patient{})
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("full_name ILIKE ? OR contact_number ILIKE ? OR pat_code ILIKE ?", like, like, like)
	}
	if bloodGroup != "" {
		tx = tx.Where("blood_group = ?", bloodGroup)
	}
	if insurance != "" {
		tx = tx.Where("insurance_provider ILIKE ?", "%"+insurance+"%")
	}
	err := tx.Order("created_at DESC").Limit(50).Find(&patients).Error
	return patients, err
}

func (r *ReceptionistRepo) GetPatient(id uint) (*model.Patient, error) {
	var p model.Patient
	err := r.DB.Where("patient_id = ?", id).First(&p).Error
	return &p, err
}

func (r *ReceptionistRepo) UpdatePatientContact(id uint, contact, address, emergName, emergPhone string) error {
	return r.DB.Model(&model.Patient{}).
		Where("patient_id = ?", id).
		Updates(map[string]interface{}{
			"contact_number":          contact,
			"address":                 address,
			"emergency_contact_name":  emergName,
			"emergency_contact_phone": emergPhone,
			"updated_at":              time.Now(),
		}).Error
}

// ── Appointment ───────────────────────────────────────────────────────────────

func (r *ReceptionistRepo) CreateAppointment(a *model.Appointment) error {
	return r.DB.Create(a).Error
}

func (r *ReceptionistRepo) ListAppointments(status, date string, deptID, doctorID uint) ([]model.Appointment, error) {
	var appts []model.Appointment
	tx := r.DB.Preload("Patient").Preload("Doctor").Preload("Department")
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if date != "" {
		tx = tx.Where("DATE(scheduled_at AT TIME ZONE 'Asia/Kolkata') = ?", date)
	}
	if deptID != 0 {
		tx = tx.Where("dept_id = ?", deptID)
	}
	if doctorID != 0 {
		tx = tx.Where("doctor_id = ?", doctorID)
	}
	err := tx.Order("scheduled_at ASC").Find(&appts).Error
	return appts, err
}

func (r *ReceptionistRepo) GetAppointment(id uint) (*model.Appointment, error) {
	var a model.Appointment
	err := r.DB.Preload("Patient").Preload("Doctor").Preload("Department").Preload("CreatedByStaff").
		Where("appt_id = ?", id).First(&a).Error
	return &a, err
}

func (r *ReceptionistRepo) RescheduleAppointment(id uint, newTime time.Time) error {
	return r.DB.Model(&model.Appointment{}).
		Where("appt_id = ? AND status = 'scheduled'", id).
		Updates(map[string]interface{}{
			"scheduled_at": newTime,
			"updated_at":   time.Now(),
		}).Error
}

func (r *ReceptionistRepo) CancelAppointment(id uint) error {
	return r.DB.Model(&model.Appointment{}).
		Where("appt_id = ? AND status IN ('scheduled','checked_in')", id).
		Updates(map[string]interface{}{
			"status":     "cancelled",
			"updated_at": time.Now(),
		}).Error
}

func (r *ReceptionistRepo) CheckInAppointment(id uint, token string) error {
	return r.DB.Model(&model.Appointment{}).
		Where("appt_id = ? AND status = 'scheduled'", id).
		Updates(map[string]interface{}{
			"status":      "checked_in",
			"queue_token": token,
			"updated_at":  time.Now(),
		}).Error
}

func (r *ReceptionistRepo) CountTodayCheckedInByDept(deptID uint) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.DB.Model(&model.Appointment{}).
		Where("dept_id = ? AND DATE(scheduled_at AT TIME ZONE 'Asia/Kolkata') = ? AND status IN ('checked_in','in_consultation','completed')", deptID, today).
		Count(&count).Error
	return count, err
}

func (r *ReceptionistRepo) GetDepartment(id uint) (*model.Department, error) {
	var d model.Department
	err := r.DB.Where("dept_id = ?", id).First(&d).Error
	return &d, err
}

// ── Visitor Log ───────────────────────────────────────────────────────────────

func (r *ReceptionistRepo) CreateVisitor(v *model.VisitorLog) error {
	return r.DB.Create(v).Error
}

func (r *ReceptionistRepo) ListVisitors(date string, patientID uint) ([]model.VisitorLog, error) {
	var visitors []model.VisitorLog
	tx := r.DB.Preload("Patient").Preload("LoggedByStaff")
	if date != "" {
		tx = tx.Where("DATE(time_in AT TIME ZONE 'Asia/Kolkata') = ?", date)
	}
	if patientID != 0 {
		tx = tx.Where("patient_id = ?", patientID)
	}
	err := tx.Order("time_in DESC").Find(&visitors).Error
	return visitors, err
}

func (r *ReceptionistRepo) CheckoutVisitor(id uint) error {
	now := time.Now()
	return r.DB.Model(&model.VisitorLog{}).
		Where("visitor_id = ? AND time_out IS NULL", id).
		Update("time_out", now).Error
}
