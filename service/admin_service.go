package service

import (
	"errors"
	"fmt"
	"time"

	"apollo-backend/config"
	"apollo-backend/model"
	"apollo-backend/repository"
	"apollo-backend/util"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	repo      *repository.AdminRepo
	auditRepo *repository.AuditRepo
	cfg       *config.Config
}

func NewAdminService(repo *repository.AdminRepo, ar *repository.AuditRepo, cfg *config.Config) *AdminService {
	return &AdminService{repo: repo, auditRepo: ar, cfg: cfg}
}

func (s *AdminService) decryptPassword(encrypted string) (string, error) {
	if len(encrypted) < 20 {
		return encrypted, nil
	}
	decrypted, err := util.DecryptAES256(encrypted, s.cfg.AESKeyBytes())
	if err != nil {
		return encrypted, nil
	}
	return decrypted, nil
}

func (s *AdminService) hashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	return string(hash), nil
}

// ── Dashboard ───────────────────────────────────────────────────────

type AdminDashboardData struct {
	Beds                 repository.BedSummary        `json:"beds"`
	DeptDoctors          []repository.DeptDoctorStat  `json:"dept_doctors"`
	OPDAppointments      int64                        `json:"opd_appointments"`
	IPDAdmissions        int64                        `json:"ipd_admissions"`
	Discharges           int64                        `json:"discharges"`
	StockAlerts          []model.Medicine             `json:"stock_alerts"`
	EmergencyAdmissions  []model.Admission            `json:"emergency_admissions"`
}

func (s *AdminService) GetDashboard() (*AdminDashboardData, error) {
	return &AdminDashboardData{
		Beds:                s.repo.BedStats(),
		DeptDoctors:         s.repo.DeptDoctorStats(),
		OPDAppointments:     s.repo.OPDCountToday(),
		IPDAdmissions:       s.repo.IPDAdmissionsToday(),
		Discharges:          s.repo.DischargeCountToday(),
		StockAlerts:         s.repo.StockAlerts(),
		EmergencyAdmissions: s.repo.EmergencyAdmissions24h(),
	}, nil
}

// ── Doctors ─────────────────────────────────────────────────────────

type CreateDoctorInput struct {
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	DeptID         uint   `json:"dept_id" binding:"required"`
	Specialization string `json:"specialization" binding:"required"`
	Qualification  string `json:"qualification" binding:"required"`
	Phone          string `json:"phone"`
	JoiningDate    string `json:"joining_date" binding:"required"`
}

type UpdateDoctorInput struct {
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	DeptID         uint   `json:"dept_id"`
	Specialization string `json:"specialization"`
	Qualification  string `json:"qualification"`
	Phone          string `json:"phone"`
	JoiningDate    string `json:"joining_date"`
}

type StatusInput struct {
	Status string `json:"status" binding:"required"`
}

func (s *AdminService) ListDoctors(deptID uint, status, search string) ([]model.Doctor, error) {
	return s.repo.ListDoctors(deptID, status, search)
}

func (s *AdminService) GetDoctor(id uint) (*model.Doctor, []model.Appointment, []model.Prescription, error) {
	doc, err := s.repo.GetDoctor(id)
	if err != nil {
		return nil, nil, nil, errors.New("doctor not found")
	}
	appts, _ := s.repo.DoctorAppointments(id)
	rxs, _ := s.repo.DoctorPrescriptions(id)
	return doc, appts, rxs, nil
}

func (s *AdminService) CreateDoctor(input CreateDoctorInput) (*model.Doctor, error) {
	password, _ := s.decryptPassword(input.Password)
	hash, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}

	jd, err := time.Parse("2006-01-02", input.JoiningDate)
	if err != nil {
		return nil, errors.New("invalid joining_date format, use YYYY-MM-DD")
	}

	doc := &model.Doctor{
		DocCode:        s.repo.NextDocCode(),
		FullName:       input.FullName,
		Email:          input.Email,
		HashedPassword: hash,
		DeptID:         input.DeptID,
		Specialization: input.Specialization,
		Qualification:  input.Qualification,
		Phone:          input.Phone,
		JoiningDate:    jd,
		Status:         "active",
	}

	if err := s.repo.CreateDoctor(doc); err != nil {
		return nil, fmt.Errorf("failed to create doctor: %w", err)
	}
	return doc, nil
}

func (s *AdminService) UpdateDoctor(id uint, input UpdateDoctorInput) (*model.Doctor, error) {
	doc, err := s.repo.GetDoctor(id)
	if err != nil {
		return nil, errors.New("doctor not found")
	}

	if input.FullName != "" {
		doc.FullName = input.FullName
	}
	if input.Email != "" {
		doc.Email = input.Email
	}
	if input.DeptID > 0 {
		doc.DeptID = input.DeptID
	}
	if input.Specialization != "" {
		doc.Specialization = input.Specialization
	}
	if input.Qualification != "" {
		doc.Qualification = input.Qualification
	}
	if input.Phone != "" {
		doc.Phone = input.Phone
	}
	if input.JoiningDate != "" {
		jd, err := time.Parse("2006-01-02", input.JoiningDate)
		if err == nil {
			doc.JoiningDate = jd
		}
	}

	if err := s.repo.UpdateDoctor(doc); err != nil {
		return nil, fmt.Errorf("failed to update doctor: %w", err)
	}
	return doc, nil
}

func (s *AdminService) UpdateDoctorStatus(id uint, status string) error {
	doc, err := s.repo.GetDoctor(id)
	if err != nil {
		return errors.New("doctor not found")
	}
	doc.Status = status
	return s.repo.UpdateDoctor(doc)
}

func (s *AdminService) DeleteDoctor(id uint) error {
	return s.repo.DeleteDoctor(id)
}

// ── Departments ─────────────────────────────────────────────────────

type CreateDepartmentInput struct {
	Name        string `json:"name" binding:"required"`
	HODDoctorID *uint  `json:"hod_doctor_id"`
	BedCount    int    `json:"bed_count"`
	OTCount     int    `json:"ot_count"`
}

type UpdateDepartmentInput struct {
	Name        string `json:"name"`
	HODDoctorID *uint  `json:"hod_doctor_id"`
	BedCount    *int   `json:"bed_count"`
	OTCount     *int   `json:"ot_count"`
	Status      string `json:"status"`
}

func (s *AdminService) ListDepartments() ([]model.Department, error) {
	return s.repo.ListDepartments()
}

func (s *AdminService) CreateDepartment(input CreateDepartmentInput) (*model.Department, error) {
	dept := &model.Department{
		Name:        input.Name,
		HODDoctorID: input.HODDoctorID,
		BedCount:    input.BedCount,
		OTCount:     input.OTCount,
		Status:      "active",
	}
	if err := s.repo.CreateDepartment(dept); err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}
	return dept, nil
}

func (s *AdminService) UpdateDepartment(id uint, input UpdateDepartmentInput) (*model.Department, error) {
	dept, err := s.repo.GetDepartment(id)
	if err != nil {
		return nil, errors.New("department not found")
	}

	if input.Name != "" {
		dept.Name = input.Name
	}
	if input.HODDoctorID != nil {
		dept.HODDoctorID = input.HODDoctorID
	}
	if input.BedCount != nil {
		dept.BedCount = *input.BedCount
	}
	if input.OTCount != nil {
		dept.OTCount = *input.OTCount
	}
	if input.Status != "" {
		dept.Status = input.Status
	}

	if err := s.repo.UpdateDepartment(dept); err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}
	return dept, nil
}

func (s *AdminService) UpdateDepartmentStatus(id uint, status string) error {
	dept, err := s.repo.GetDepartment(id)
	if err != nil {
		return errors.New("department not found")
	}
	dept.Status = status
	return s.repo.UpdateDepartment(dept)
}

func (s *AdminService) DeleteDepartment(id uint) error {
	return s.repo.DeleteDepartment(id)
}

// ── Staff ───────────────────────────────────────────────────────────

type CreateStaffInput struct {
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	Role           string `json:"role" binding:"required"`
	DeptID         *uint  `json:"dept_id"`
	Qualification  string `json:"qualification"`
	EmploymentType string `json:"employment_type"`
	JoiningDate    string `json:"joining_date" binding:"required"`
}

type UpdateStaffInput struct {
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	DeptID         *uint  `json:"dept_id"`
	Qualification  string `json:"qualification"`
	EmploymentType string `json:"employment_type"`
	JoiningDate    string `json:"joining_date"`
}

func (s *AdminService) ListStaff(role string, deptID uint, status string) ([]model.Staff, error) {
	return s.repo.ListStaff(role, deptID, status)
}

func (s *AdminService) GetStaff(id uint) (*model.Staff, error) {
	return s.repo.GetStaff(id)
}

func (s *AdminService) CreateStaff(input CreateStaffInput) (*model.Staff, error) {
	password, _ := s.decryptPassword(input.Password)
	hash, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}

	jd, err := time.Parse("2006-01-02", input.JoiningDate)
	if err != nil {
		return nil, errors.New("invalid joining_date format, use YYYY-MM-DD")
	}

	empType := input.EmploymentType
	if empType == "" {
		empType = "full_time"
	}

	staff := &model.Staff{
		FullName:       input.FullName,
		Email:          input.Email,
		HashedPassword: hash,
		Role:           input.Role,
		DeptID:         input.DeptID,
		Qualification:  input.Qualification,
		EmploymentType: empType,
		JoiningDate:    jd,
		Status:         "active",
	}

	if err := s.repo.CreateStaff(staff); err != nil {
		return nil, fmt.Errorf("failed to create staff: %w", err)
	}
	return staff, nil
}

func (s *AdminService) UpdateStaff(id uint, input UpdateStaffInput) (*model.Staff, error) {
	staff, err := s.repo.GetStaff(id)
	if err != nil {
		return nil, errors.New("staff not found")
	}

	if input.FullName != "" {
		staff.FullName = input.FullName
	}
	if input.Email != "" {
		staff.Email = input.Email
	}
	if input.Role != "" {
		staff.Role = input.Role
	}
	if input.DeptID != nil {
		staff.DeptID = input.DeptID
	}
	if input.Qualification != "" {
		staff.Qualification = input.Qualification
	}
	if input.EmploymentType != "" {
		staff.EmploymentType = input.EmploymentType
	}
	if input.JoiningDate != "" {
		jd, err := time.Parse("2006-01-02", input.JoiningDate)
		if err == nil {
			staff.JoiningDate = jd
		}
	}

	if err := s.repo.UpdateStaff(staff); err != nil {
		return nil, fmt.Errorf("failed to update staff: %w", err)
	}
	return staff, nil
}

func (s *AdminService) UpdateStaffStatus(id uint, status string) error {
	staff, err := s.repo.GetStaff(id)
	if err != nil {
		return errors.New("staff not found")
	}
	staff.Status = status
	return s.repo.UpdateStaff(staff)
}

func (s *AdminService) DeleteStaff(id uint) error {
	return s.repo.DeleteStaff(id)
}

// ── Pharmacists ─────────────────────────────────────────────────────

type CreatePharmacistInput struct {
	FullName      string `json:"full_name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=6"`
	LicenseNumber string `json:"license_number" binding:"required"`
	Phone         string `json:"phone"`
}

type UpdatePharmacistInput struct {
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	LicenseNumber string `json:"license_number"`
	Phone         string `json:"phone"`
}

func (s *AdminService) ListPharmacists(status string) ([]model.Pharmacist, error) {
	return s.repo.ListPharmacists(status)
}

func (s *AdminService) CreatePharmacist(input CreatePharmacistInput) (*model.Pharmacist, error) {
	password, _ := s.decryptPassword(input.Password)
	hash, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}

	p := &model.Pharmacist{
		FullName:       input.FullName,
		Email:          input.Email,
		HashedPassword: hash,
		LicenseNumber:  input.LicenseNumber,
		Phone:          input.Phone,
		Status:         "active",
	}

	if err := s.repo.CreatePharmacist(p); err != nil {
		return nil, fmt.Errorf("failed to create pharmacist: %w", err)
	}
	return p, nil
}

func (s *AdminService) UpdatePharmacist(id uint, input UpdatePharmacistInput) (*model.Pharmacist, error) {
	p, err := s.repo.GetPharmacist(id)
	if err != nil {
		return nil, errors.New("pharmacist not found")
	}

	if input.FullName != "" {
		p.FullName = input.FullName
	}
	if input.Email != "" {
		p.Email = input.Email
	}
	if input.LicenseNumber != "" {
		p.LicenseNumber = input.LicenseNumber
	}
	if input.Phone != "" {
		p.Phone = input.Phone
	}

	if err := s.repo.UpdatePharmacist(p); err != nil {
		return nil, fmt.Errorf("failed to update pharmacist: %w", err)
	}
	return p, nil
}

// ── Partner Pharmacies ──────────────────────────────────────────────

type CreatePartnerPharmacyInput struct {
	Name          string `json:"name" binding:"required"`
	LicenseNumber string `json:"license_number" binding:"required"`
	Address       string `json:"address" binding:"required"`
	ContactPhone  string `json:"contact_phone"`
	ContactEmail  string `json:"contact_email"`
}

type UpdatePartnerPharmacyInput struct {
	Name          string `json:"name"`
	LicenseNumber string `json:"license_number"`
	Address       string `json:"address"`
	ContactPhone  string `json:"contact_phone"`
	ContactEmail  string `json:"contact_email"`
}

func (s *AdminService) ListPartnerPharmacies(status string) ([]model.PartnerPharmacy, error) {
	return s.repo.ListPartnerPharmacies(status)
}

func (s *AdminService) CreatePartnerPharmacy(input CreatePartnerPharmacyInput) (*model.PartnerPharmacy, error) {
	p := &model.PartnerPharmacy{
		Name:          input.Name,
		LicenseNumber: input.LicenseNumber,
		Address:       input.Address,
		ContactPhone:  input.ContactPhone,
		ContactEmail:  input.ContactEmail,
		Status:        "active",
	}

	if err := s.repo.CreatePartnerPharmacy(p); err != nil {
		return nil, fmt.Errorf("failed to create partner pharmacy: %w", err)
	}
	return p, nil
}

func (s *AdminService) UpdatePartnerPharmacy(id uint, input UpdatePartnerPharmacyInput) (*model.PartnerPharmacy, error) {
	p, err := s.repo.GetPartnerPharmacy(id)
	if err != nil {
		return nil, errors.New("partner pharmacy not found")
	}

	if input.Name != "" {
		p.Name = input.Name
	}
	if input.LicenseNumber != "" {
		p.LicenseNumber = input.LicenseNumber
	}
	if input.Address != "" {
		p.Address = input.Address
	}
	if input.ContactPhone != "" {
		p.ContactPhone = input.ContactPhone
	}
	if input.ContactEmail != "" {
		p.ContactEmail = input.ContactEmail
	}

	if err := s.repo.UpdatePartnerPharmacy(p); err != nil {
		return nil, fmt.Errorf("failed to update partner pharmacy: %w", err)
	}
	return p, nil
}

func (s *AdminService) UpdatePartnerPharmacyStatus(id uint, status string) error {
	p, err := s.repo.GetPartnerPharmacy(id)
	if err != nil {
		return errors.New("partner pharmacy not found")
	}
	p.Status = status
	return s.repo.UpdatePartnerPharmacy(p)
}

// ── Hospital Config ─────────────────────────────────────────────────

type UpdateConfigInput struct {
	HospitalName string `json:"hospital_name"`
	Address      string `json:"address"`
	GSTNumber    string `json:"gst_number"`
	NABHNumber   string `json:"nabh_number"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	Website      string `json:"website"`
}

func (s *AdminService) GetConfig() (*model.HospitalConfig, error) {
	return s.repo.GetConfig()
}

func (s *AdminService) UpdateConfig(input UpdateConfigInput) (*model.HospitalConfig, error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, err
	}

	if input.HospitalName != "" {
		cfg.HospitalName = input.HospitalName
	}
	if input.Address != "" {
		cfg.Address = input.Address
	}
	if input.GSTNumber != "" {
		cfg.GSTNumber = input.GSTNumber
	}
	if input.NABHNumber != "" {
		cfg.NABHNumber = input.NABHNumber
	}
	if input.ContactPhone != "" {
		cfg.ContactPhone = input.ContactPhone
	}
	if input.ContactEmail != "" {
		cfg.ContactEmail = input.ContactEmail
	}
	if input.Website != "" {
		cfg.Website = input.Website
	}

	if err := s.repo.UpdateConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}
	return cfg, nil
}

// ── Reports ─────────────────────────────────────────────────────────

func (s *AdminService) FinancialReport(from, to time.Time) repository.FinancialSummary {
	return s.repo.FinancialReport(from, to)
}

func (s *AdminService) OccupancyReport(from, to time.Time, deptID uint) []repository.OccupancyReport {
	return s.repo.OccupancyReport(from, to, deptID)
}

func (s *AdminService) PrescriptionReport(from, to time.Time) []repository.PrescriptionReport {
	return s.repo.PrescriptionReport(from, to)
}

// ── Export ───────────────────────────────────────────────────────────

type ExportInput struct {
	Entity   string `json:"entity" binding:"required"`
	Format   string `json:"format" binding:"required"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

type ExportResult struct {
	Entity   string      `json:"entity"`
	Format   string      `json:"format"`
	Count    int         `json:"count"`
	Data     interface{} `json:"data"`
}

func (s *AdminService) Export(input ExportInput) (*ExportResult, error) {
	var data interface{}
	var count int

	switch input.Entity {
	case "doctors":
		docs, _ := s.repo.ListDoctors(0, "", "")
		data = docs
		count = len(docs)
	case "staff":
		staff, _ := s.repo.ListStaff("", 0, "")
		data = staff
		count = len(staff)
	case "departments":
		depts, _ := s.repo.ListDepartments()
		data = depts
		count = len(depts)
	case "pharmacists":
		pharms, _ := s.repo.ListPharmacists("")
		data = pharms
		count = len(pharms)
	case "partner_pharmacies":
		pp, _ := s.repo.ListPartnerPharmacies("")
		data = pp
		count = len(pp)
	default:
		return nil, errors.New("unsupported entity")
	}

	return &ExportResult{
		Entity: input.Entity,
		Format: input.Format,
		Count:  count,
		Data:   data,
	}, nil
}
