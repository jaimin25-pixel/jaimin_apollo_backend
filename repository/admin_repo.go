package repository

import (
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

type AdminRepo struct{ DB *gorm.DB }

func NewAdminRepo(db *gorm.DB) *AdminRepo { return &AdminRepo{DB: db} }

// ── Dashboard ───────────────────────────────────────────────────────

type BedSummary struct {
	Total     int64 `json:"total"`
	Occupied  int64 `json:"occupied"`
	Available int64 `json:"available"`
	ICU       int64 `json:"icu"`
	Emergency int64 `json:"emergency"`
}

type DeptDoctorStat struct {
	DeptID       uint   `json:"dept_id"`
	DeptName     string `json:"dept_name"`
	DoctorCount  int64  `json:"doctor_count"`
	ActiveCount  int64  `json:"active_count"`
}

func (r *AdminRepo) BedStats() BedSummary {
	var s BedSummary
	r.DB.Model(&model.Bed{}).Count(&s.Total)
	r.DB.Model(&model.Bed{}).Where("status = 'occupied'").Count(&s.Occupied)
	r.DB.Model(&model.Bed{}).Where("status = 'available'").Count(&s.Available)
	r.DB.Model(&model.Bed{}).
		Joins("JOIN wards ON wards.ward_id = beds.ward_id").
		Where("wards.ward_type = 'icu'").Count(&s.ICU)
	r.DB.Model(&model.Bed{}).
		Joins("JOIN wards ON wards.ward_id = beds.ward_id").
		Where("wards.ward_type = 'emergency'").Count(&s.Emergency)
	return s
}

func (r *AdminRepo) DeptDoctorStats() []DeptDoctorStat {
	var stats []DeptDoctorStat
	r.DB.Raw(`
		SELECT d.dept_id, d.name AS dept_name,
			COUNT(doc.doctor_id) AS doctor_count,
			COUNT(doc.doctor_id) FILTER (WHERE doc.status = 'active') AS active_count
		FROM departments d
		LEFT JOIN doctors doc ON doc.dept_id = d.dept_id
		WHERE d.status = 'active'
		GROUP BY d.dept_id, d.name
		ORDER BY d.name
	`).Scan(&stats)
	return stats
}

func (r *AdminRepo) OPDCountToday() int64 {
	var c int64
	today := time.Now().Truncate(24 * time.Hour)
	r.DB.Model(&model.Appointment{}).
		Where("scheduled_at >= ? AND scheduled_at < ?", today, today.Add(24*time.Hour)).
		Count(&c)
	return c
}

func (r *AdminRepo) IPDAdmissionsToday() int64 {
	var c int64
	today := time.Now().Truncate(24 * time.Hour)
	r.DB.Model(&model.Admission{}).
		Where("admitted_at >= ? AND admitted_at < ?", today, today.Add(24*time.Hour)).
		Count(&c)
	return c
}

func (r *AdminRepo) DischargeCountToday() int64 {
	var c int64
	today := time.Now().Truncate(24 * time.Hour)
	r.DB.Model(&model.Admission{}).
		Where("discharged_at >= ? AND discharged_at < ? AND status = 'discharged'", today, today.Add(24*time.Hour)).
		Count(&c)
	return c
}

func (r *AdminRepo) StockAlerts() []model.Medicine {
	var meds []model.Medicine
	r.DB.Where("current_stock <= reorder_level AND status = 'active'").Find(&meds)
	return meds
}

func (r *AdminRepo) EmergencyAdmissions24h() []model.Admission {
	var admissions []model.Admission
	since := time.Now().Add(-24 * time.Hour)
	r.DB.Preload("Patient").Preload("AdmittingDoctor").Preload("Department").
		Joins("JOIN wards ON wards.ward_id = admissions.ward_id").
		Where("wards.ward_type = 'emergency' AND admissions.admitted_at >= ?", since).
		Order("admissions.admitted_at DESC").
		Find(&admissions)
	return admissions
}

// ── Doctors CRUD ────────────────────────────────────────────────────

func (r *AdminRepo) ListDoctors(deptID uint, status, search string) ([]model.Doctor, error) {
	q := r.DB.Preload("Department")
	if deptID > 0 {
		q = q.Where("dept_id = ?", deptID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("full_name ILIKE ? OR email ILIKE ? OR doc_code ILIKE ?", like, like, like)
	}
	var docs []model.Doctor
	err := q.Order("full_name").Find(&docs).Error
	return docs, err
}

func (r *AdminRepo) GetDoctor(id uint) (*model.Doctor, error) {
	var doc model.Doctor
	err := r.DB.Preload("Department").First(&doc, "doctor_id = ?", id).Error
	return &doc, err
}

func (r *AdminRepo) CreateDoctor(doc *model.Doctor) error {
	return r.DB.Create(doc).Error
}

func (r *AdminRepo) UpdateDoctor(doc *model.Doctor) error {
	return r.DB.Save(doc).Error
}

func (r *AdminRepo) DeleteDoctor(id uint) error {
	return r.DB.Delete(&model.Doctor{}, "doctor_id = ?", id).Error
}

func (r *AdminRepo) NextDocCode() string {
	var count int64
	r.DB.Model(&model.Doctor{}).Count(&count)
	return "DOC-" + padNumber(int(count)+1, 4)
}

// ── Departments CRUD ────────────────────────────────────────────────

func (r *AdminRepo) DeleteDepartment(id uint) error {
	return r.DB.Delete(&model.Department{}, "dept_id = ?", id).Error
}

func (r *AdminRepo) ListDepartments() ([]model.Department, error) {
	var depts []model.Department
	err := r.DB.Preload("HODDoctor").Order("name").Find(&depts).Error
	return depts, err
}

func (r *AdminRepo) GetDepartment(id uint) (*model.Department, error) {
	var dept model.Department
	err := r.DB.Preload("HODDoctor").First(&dept, "dept_id = ?", id).Error
	return &dept, err
}

func (r *AdminRepo) CreateDepartment(dept *model.Department) error {
	return r.DB.Create(dept).Error
}

func (r *AdminRepo) UpdateDepartment(dept *model.Department) error {
	return r.DB.Save(dept).Error
}

// ── Staff CRUD ──────────────────────────────────────────────────────

func (r *AdminRepo) ListStaff(role string, deptID uint, status string) ([]model.Staff, error) {
	q := r.DB.Preload("Department")
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if deptID > 0 {
		q = q.Where("dept_id = ?", deptID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var staff []model.Staff
	err := q.Order("full_name").Find(&staff).Error
	return staff, err
}

func (r *AdminRepo) GetStaff(id uint) (*model.Staff, error) {
	var s model.Staff
	err := r.DB.Preload("Department").First(&s, "staff_id = ?", id).Error
	return &s, err
}

func (r *AdminRepo) CreateStaff(s *model.Staff) error {
	return r.DB.Create(s).Error
}

func (r *AdminRepo) UpdateStaff(s *model.Staff) error {
	return r.DB.Save(s).Error
}

func (r *AdminRepo) DeleteStaff(id uint) error {
	return r.DB.Delete(&model.Staff{}, "staff_id = ?", id).Error
}

// ── Pharmacists CRUD ────────────────────────────────────────────────

func (r *AdminRepo) ListPharmacists(status string) ([]model.Pharmacist, error) {
	q := r.DB.Model(&model.Pharmacist{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var list []model.Pharmacist
	err := q.Order("full_name").Find(&list).Error
	return list, err
}

func (r *AdminRepo) GetPharmacist(id uint) (*model.Pharmacist, error) {
	var p model.Pharmacist
	err := r.DB.First(&p, "pharmacist_id = ?", id).Error
	return &p, err
}

func (r *AdminRepo) CreatePharmacist(p *model.Pharmacist) error {
	return r.DB.Create(p).Error
}

func (r *AdminRepo) UpdatePharmacist(p *model.Pharmacist) error {
	return r.DB.Save(p).Error
}

// ── Partner Pharmacies CRUD ─────────────────────────────────────────

func (r *AdminRepo) ListPartnerPharmacies(status string) ([]model.PartnerPharmacy, error) {
	q := r.DB.Model(&model.PartnerPharmacy{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var list []model.PartnerPharmacy
	err := q.Order("name").Find(&list).Error
	return list, err
}

func (r *AdminRepo) GetPartnerPharmacy(id uint) (*model.PartnerPharmacy, error) {
	var p model.PartnerPharmacy
	err := r.DB.First(&p, "partner_id = ?", id).Error
	return &p, err
}

func (r *AdminRepo) CreatePartnerPharmacy(p *model.PartnerPharmacy) error {
	return r.DB.Create(p).Error
}

func (r *AdminRepo) UpdatePartnerPharmacy(p *model.PartnerPharmacy) error {
	return r.DB.Save(p).Error
}

// ── Hospital Config ─────────────────────────────────────────────────

func (r *AdminRepo) GetConfig() (*model.HospitalConfig, error) {
	var cfg model.HospitalConfig
	err := r.DB.First(&cfg).Error
	if err == gorm.ErrRecordNotFound {
		cfg = model.HospitalConfig{HospitalName: "Apollo Hospital"}
		r.DB.Create(&cfg)
		return &cfg, nil
	}
	return &cfg, err
}

func (r *AdminRepo) UpdateConfig(cfg *model.HospitalConfig) error {
	return r.DB.Save(cfg).Error
}

// ── Reports ─────────────────────────────────────────────────────────

type FinancialSummary struct {
	TotalInvoices    int64   `json:"total_invoices"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalPaid        float64 `json:"total_paid"`
	TotalOutstanding float64 `json:"total_outstanding"`
}

func (r *AdminRepo) FinancialReport(from, to time.Time) FinancialSummary {
	var s FinancialSummary
	r.DB.Model(&model.Invoice{}).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Count(&s.TotalInvoices)
	r.DB.Model(&model.Invoice{}).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Select("COALESCE(SUM(total_amount),0)").Scan(&s.TotalRevenue)
	r.DB.Model(&model.Invoice{}).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Select("COALESCE(SUM(amount_paid),0)").Scan(&s.TotalPaid)
	s.TotalOutstanding = s.TotalRevenue - s.TotalPaid
	return s
}

type OccupancyReport struct {
	DeptID    uint   `json:"dept_id"`
	DeptName  string `json:"dept_name"`
	TotalBeds int64  `json:"total_beds"`
	Occupied  int64  `json:"occupied"`
}

func (r *AdminRepo) OccupancyReport(from, to time.Time, deptID uint) []OccupancyReport {
	var reps []OccupancyReport
	q := `
		SELECT d.dept_id, d.name AS dept_name,
			COUNT(b.bed_id) AS total_beds,
			COUNT(b.bed_id) FILTER (WHERE b.status = 'occupied') AS occupied
		FROM departments d
		LEFT JOIN wards w ON w.dept_id = d.dept_id
		LEFT JOIN beds b ON b.ward_id = w.ward_id
		WHERE d.status = 'active'
	`
	args := []interface{}{}
	if deptID > 0 {
		q += " AND d.dept_id = ?"
		args = append(args, deptID)
	}
	q += " GROUP BY d.dept_id, d.name ORDER BY d.name"
	r.DB.Raw(q, args...).Scan(&reps)
	return reps
}

type PrescriptionReport struct {
	DoctorID   uint   `json:"doctor_id"`
	DoctorName string `json:"doctor_name"`
	DeptName   string `json:"dept_name"`
	RxCount    int64  `json:"rx_count"`
}

func (r *AdminRepo) PrescriptionReport(from, to time.Time) []PrescriptionReport {
	var reps []PrescriptionReport
	r.DB.Raw(`
		SELECT p.doctor_id, doc.full_name AS doctor_name, d.name AS dept_name,
			COUNT(p.rx_id) AS rx_count
		FROM prescriptions p
		JOIN doctors doc ON doc.doctor_id = p.doctor_id
		LEFT JOIN departments d ON d.dept_id = doc.dept_id
		WHERE p.created_at >= ? AND p.created_at <= ?
		GROUP BY p.doctor_id, doc.full_name, d.name
		ORDER BY rx_count DESC
	`, from, to).Scan(&reps)
	return reps
}

// ── Doctor History ──────────────────────────────────────────────────

func (r *AdminRepo) DoctorAppointments(doctorID uint) ([]model.Appointment, error) {
	var appts []model.Appointment
	err := r.DB.Preload("Patient").Where("doctor_id = ?", doctorID).
		Order("scheduled_at DESC").Limit(50).Find(&appts).Error
	return appts, err
}

func (r *AdminRepo) DoctorPrescriptions(doctorID uint) ([]model.Prescription, error) {
	var rxs []model.Prescription
	err := r.DB.Preload("Patient").Where("doctor_id = ?", doctorID).
		Order("created_at DESC").Limit(50).Find(&rxs).Error
	return rxs, err
}

// ── Helpers ─────────────────────────────────────────────────────────

func padNumber(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s += "0"
	}
	ns := s + intToStr(n)
	return ns[len(ns)-width:]
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
