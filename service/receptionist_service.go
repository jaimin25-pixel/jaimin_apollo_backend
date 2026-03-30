package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"apollo-backend/model"
	"apollo-backend/repository"
)

type ReceptionistService struct {
	repo      *repository.ReceptionistRepo
	auditRepo *repository.AuditRepo
}

func NewReceptionistService(r *repository.ReceptionistRepo, ar *repository.AuditRepo) *ReceptionistService {
	return &ReceptionistService{repo: r, auditRepo: ar}
}

// ─── Input Structs ─────────────────────────────────────────────────────────────

type RegisterPatientInput struct {
	FullName              string    `json:"full_name" binding:"required"`
	DateOfBirth           time.Time `json:"date_of_birth" binding:"required"`
	Gender                string    `json:"gender" binding:"required"`
	BloodGroup            string    `json:"blood_group"`
	ContactNumber         string    `json:"contact_number" binding:"required"`
	Address               string    `json:"address"`
	EmergencyContactName  string    `json:"emergency_contact_name"`
	EmergencyContactPhone string    `json:"emergency_contact_phone"`
	InsuranceID           string    `json:"insurance_id"`
	InsuranceProvider     string    `json:"insurance_provider"`
}

type UpdateContactInput struct {
	ContactNumber         string `json:"contact_number" binding:"required"`
	Address               string `json:"address"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
}

type BookAppointmentInput struct {
	PatientID      uint      `json:"patient_id" binding:"required"`
	DoctorID       uint      `json:"doctor_id" binding:"required"`
	DeptID         uint      `json:"dept_id" binding:"required"`
	ScheduledAt    time.Time `json:"scheduled_at" binding:"required"`
	ChiefComplaint string    `json:"chief_complaint"`
}

type RescheduleInput struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

type WalkInInput struct {
	PatientID      uint   `json:"patient_id" binding:"required"`
	DoctorID       uint   `json:"doctor_id" binding:"required"`
	DeptID         uint   `json:"dept_id" binding:"required"`
	ChiefComplaint string `json:"chief_complaint"`
}

type LogVisitorInput struct {
	PatientID   uint   `json:"patient_id" binding:"required"`
	VisitorName string `json:"visitor_name" binding:"required"`
	Relation    string `json:"relation" binding:"required"`
}

// ─── PAT Code & Queue Token helpers ──────────────────────────────────────────

func (s *ReceptionistService) generatePatCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	for attempt := 0; attempt < 10; attempt++ {
		b := make([]byte, length)
		for i := range b {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			b[i] = charset[n.Int64()]
		}
		code := "PAT-" + string(b)
		if !s.repo.PatCodeExists(code) {
			return code, nil
		}
	}
	return "", errors.New("failed to generate unique PAT code")
}

func deptAbbr(name string) string {
	words := strings.Fields(name)
	abbr := ""
	for _, w := range words {
		for _, r := range w {
			if unicode.IsLetter(r) {
				abbr += strings.ToUpper(string(r))
				break
			}
		}
		if len(abbr) >= 3 {
			break
		}
	}
	if abbr == "" {
		abbr = "XX"
	}
	return abbr
}

func (s *ReceptionistService) generateQueueToken(deptID uint) (string, error) {
	dept, err := s.repo.GetDepartment(deptID)
	if err != nil {
		return "", errors.New("department not found")
	}
	count, err := s.repo.CountTodayCheckedInByDept(deptID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Q%s-%03d", deptAbbr(dept.Name), count+1), nil
}

// ─── Dashboard ─────────────────────────────────────────────────────────────────

func (s *ReceptionistService) GetDashboard() (*repository.QueueStats, error) {
	return s.repo.TodayQueueStats()
}

// ─── Patient ───────────────────────────────────────────────────────────────────

func (s *ReceptionistService) RegisterPatient(input RegisterPatientInput) (*model.Patient, error) {
	code, err := s.generatePatCode()
	if err != nil {
		return nil, err
	}
	p := &model.Patient{
		PatCode:               code,
		FullName:              input.FullName,
		DateOfBirth:           input.DateOfBirth,
		Gender:                input.Gender,
		BloodGroup:            input.BloodGroup,
		ContactNumber:         input.ContactNumber,
		Address:               input.Address,
		EmergencyContactName:  input.EmergencyContactName,
		EmergencyContactPhone: input.EmergencyContactPhone,
		InsuranceID:           input.InsuranceID,
		InsuranceProvider:     input.InsuranceProvider,
	}
	if err := s.repo.CreatePatient(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ReceptionistService) SearchPatients(q, bloodGroup, insurance string) ([]model.Patient, error) {
	return s.repo.SearchPatients(q, bloodGroup, insurance)
}

func (s *ReceptionistService) GetPatient(id uint) (*model.Patient, error) {
	p, err := s.repo.GetPatient(id)
	if err != nil {
		return nil, errors.New("patient not found")
	}
	return p, nil
}

func (s *ReceptionistService) UpdatePatientContact(id uint, input UpdateContactInput) error {
	return s.repo.UpdatePatientContact(id, input.ContactNumber, input.Address, input.EmergencyContactName, input.EmergencyContactPhone)
}

// ─── Appointment ───────────────────────────────────────────────────────────────

func (s *ReceptionistService) BookAppointment(staffID uint, input BookAppointmentInput) (*model.Appointment, error) {
	a := &model.Appointment{
		PatientID:        input.PatientID,
		DoctorID:         input.DoctorID,
		DeptID:           input.DeptID,
		ScheduledAt:      input.ScheduledAt,
		Status:           "scheduled",
		ChiefComplaint:   input.ChiefComplaint,
		CreatedByStaffID: &staffID,
	}
	if err := s.repo.CreateAppointment(a); err != nil {
		return nil, err
	}
	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: staffID, UserRole: "receptionist", Action: "CREATE",
		TblName: "appointments", RecordID: a.ApptID,
	})
	return a, nil
}

func (s *ReceptionistService) ListAppointments(status, date string, deptID, doctorID uint) ([]model.Appointment, error) {
	return s.repo.ListAppointments(status, date, deptID, doctorID)
}

func (s *ReceptionistService) GetAppointment(id uint) (*model.Appointment, error) {
	a, err := s.repo.GetAppointment(id)
	if err != nil {
		return nil, errors.New("appointment not found")
	}
	return a, nil
}

func (s *ReceptionistService) RescheduleAppointment(id uint, input RescheduleInput) error {
	a, err := s.repo.GetAppointment(id)
	if err != nil {
		return errors.New("appointment not found")
	}
	if a.Status != "scheduled" {
		return errors.New("only scheduled appointments can be rescheduled")
	}
	return s.repo.RescheduleAppointment(id, input.ScheduledAt)
}

func (s *ReceptionistService) CancelAppointment(id uint) error {
	a, err := s.repo.GetAppointment(id)
	if err != nil {
		return errors.New("appointment not found")
	}
	if a.Status != "scheduled" && a.Status != "checked_in" {
		return errors.New("appointment cannot be cancelled in its current status")
	}
	return s.repo.CancelAppointment(id)
}

func (s *ReceptionistService) CheckInAppointment(apptID uint) (*model.Appointment, error) {
	a, err := s.repo.GetAppointment(apptID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}
	if a.Status != "scheduled" {
		return nil, errors.New("appointment is not in scheduled status")
	}
	token, err := s.generateQueueToken(a.DeptID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CheckInAppointment(apptID, token); err != nil {
		return nil, err
	}
	a.Status = "checked_in"
	a.QueueToken = token
	return a, nil
}

func (s *ReceptionistService) WalkIn(staffID uint, input WalkInInput) (*model.Appointment, error) {
	a := &model.Appointment{
		PatientID:        input.PatientID,
		DoctorID:         input.DoctorID,
		DeptID:           input.DeptID,
		ScheduledAt:      time.Now(),
		Status:           "scheduled",
		ChiefComplaint:   input.ChiefComplaint,
		CreatedByStaffID: &staffID,
	}
	if err := s.repo.CreateAppointment(a); err != nil {
		return nil, err
	}
	token, err := s.generateQueueToken(a.DeptID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CheckInAppointment(a.ApptID, token); err != nil {
		return nil, err
	}
	a.Status = "checked_in"
	a.QueueToken = token
	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: staffID, UserRole: "receptionist", Action: "CREATE",
		TblName: "appointments", RecordID: a.ApptID,
	})
	return a, nil
}

// ─── Visitor Log ───────────────────────────────────────────────────────────────

func (s *ReceptionistService) LogVisitor(staffID uint, input LogVisitorInput) (*model.VisitorLog, error) {
	v := &model.VisitorLog{
		PatientID:       input.PatientID,
		VisitorName:     input.VisitorName,
		Relation:        input.Relation,
		TimeIn:          time.Now(),
		LoggedByStaffID: staffID,
	}
	if err := s.repo.CreateVisitor(v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *ReceptionistService) ListVisitors(date string, patientID uint) ([]model.VisitorLog, error) {
	return s.repo.ListVisitors(date, patientID)
}

func (s *ReceptionistService) CheckoutVisitor(id uint) error {
	return s.repo.CheckoutVisitor(id)
}
