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

// ─── PatientService ───────────────────────────────────────────────────────────

type PatientService struct {
	repo      *repository.PatientRepo
	auditRepo *repository.AuditRepo
}

func NewPatientService(r *repository.PatientRepo, ar *repository.AuditRepo) *PatientService {
	return &PatientService{repo: r, auditRepo: ar}
}

// ─── Input Structs ────────────────────────────────────────────────────────────

type PatientRegisterInput struct {
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

type PatientUpdateInput struct {
	ContactNumber         string `json:"contact_number"`
	Address               string `json:"address"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	InsuranceID           string `json:"insurance_id"`
}

type PatientBookAppointmentInput struct {
	DoctorID       uint      `json:"doctor_id" binding:"required"`
	DeptID         uint      `json:"dept_id" binding:"required"`
	ScheduledAt    time.Time `json:"scheduled_at" binding:"required"`
	ChiefComplaint string    `json:"chief_complaint"`
}

type PatientRescheduleInput struct {
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// ─── PAT Code & Queue Token helpers ──────────────────────────────────────────

func (s *PatientService) generatePatCode() (string, error) {
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

func patientDeptAbbr(name string) string {
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

func (s *PatientService) generateQueueToken(deptID uint) (string, error) {
	dept, err := s.repo.GetDepartment(deptID)
	if err != nil {
		return "", errors.New("department not found")
	}
	count, err := s.repo.CountTodayCheckedInByDept(deptID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Q%s-%03d", patientDeptAbbr(dept.Name), count+1), nil
}

// ─── Registration ─────────────────────────────────────────────────────────────

func (s *PatientService) RegisterPatient(input PatientRegisterInput) (*model.Patient, []model.Patient, error) {
	// Validate required fields
	if strings.TrimSpace(input.FullName) == "" {
		return nil, nil, errors.New("full_name is required")
	}
	if strings.TrimSpace(input.ContactNumber) == "" {
		return nil, nil, errors.New("contact_number is required")
	}
	if strings.TrimSpace(input.Gender) == "" {
		return nil, nil, errors.New("gender is required")
	}
	validGenders := map[string]bool{"male": true, "female": true, "other": true}
	if !validGenders[strings.ToLower(input.Gender)] {
		return nil, nil, errors.New("gender must be male, female, or other")
	}

	// REQ-PAT-004: Duplicate detection
	duplicates, _ := s.repo.DuplicateCheck(input.FullName, input.DateOfBirth, input.ContactNumber)

	code, err := s.generatePatCode()
	if err != nil {
		return nil, nil, err
	}

	p := &model.Patient{
		PatCode:               code,
		FullName:              input.FullName,
		DateOfBirth:           input.DateOfBirth,
		Gender:                strings.ToLower(input.Gender),
		BloodGroup:            input.BloodGroup,
		ContactNumber:         input.ContactNumber,
		Address:               input.Address,
		EmergencyContactName:  input.EmergencyContactName,
		EmergencyContactPhone: input.EmergencyContactPhone,
		InsuranceID:           input.InsuranceID,
		InsuranceProvider:     input.InsuranceProvider,
	}

	if err := s.repo.CreatePatient(p); err != nil {
		return nil, nil, err
	}

	return p, duplicates, nil
}

// ─── Search ───────────────────────────────────────────────────────────────────

func (s *PatientService) SearchPatients(q string) ([]model.Patient, error) {
	return s.repo.SearchPatients(q)
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (s *PatientService) GetPatient(id uint) (*model.Patient, error) {
	p, err := s.repo.GetPatientByID(id)
	if err != nil {
		return nil, errors.New("patient not found")
	}
	return p, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *PatientService) UpdatePatient(id uint, input PatientUpdateInput) error {
	_, err := s.repo.GetPatientByID(id)
	if err != nil {
		return errors.New("patient not found")
	}
	fields := map[string]interface{}{}
	if input.ContactNumber != "" {
		fields["contact_number"] = input.ContactNumber
	}
	if input.Address != "" {
		fields["address"] = input.Address
	}
	if input.EmergencyContactName != "" {
		fields["emergency_contact_name"] = input.EmergencyContactName
	}
	if input.EmergencyContactPhone != "" {
		fields["emergency_contact_phone"] = input.EmergencyContactPhone
	}
	if input.InsuranceID != "" {
		fields["insurance_id"] = input.InsuranceID
	}
	if len(fields) == 0 {
		return errors.New("no fields to update")
	}
	return s.repo.UpdatePatient(id, fields)
}

// ─── Appointments ─────────────────────────────────────────────────────────────

func (s *PatientService) ListAppointments(patientID uint, status, fromDate string) ([]model.Appointment, error) {
	_, err := s.repo.GetPatientByID(patientID)
	if err != nil {
		return nil, errors.New("patient not found")
	}
	return s.repo.ListAppointments(patientID, status, fromDate)
}

func (s *PatientService) GetAppointment(apptID uint) (*model.Appointment, error) {
	a, err := s.repo.GetAppointment(apptID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}
	return a, nil
}

func (s *PatientService) BookAppointment(patientID uint, input PatientBookAppointmentInput) (*model.Appointment, error) {
	_, err := s.repo.GetPatientByID(patientID)
	if err != nil {
		return nil, errors.New("patient not found")
	}

	// REQ-PAT-011: Double-booking prevention
	conflict, err := s.repo.CheckDoubleBooking(input.DoctorID, input.ScheduledAt)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, errors.New("doctor already has an appointment at this time slot (REQ-PAT-011)")
	}

	a := &model.Appointment{
		PatientID:      patientID,
		DoctorID:       input.DoctorID,
		DeptID:         input.DeptID,
		ScheduledAt:    input.ScheduledAt,
		Status:         "scheduled",
		ChiefComplaint: input.ChiefComplaint,
	}
	if err := s.repo.CreateAppointment(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *PatientService) UpdateAppointment(apptID uint, input PatientRescheduleInput) error {
	a, err := s.repo.GetAppointment(apptID)
	if err != nil {
		return errors.New("appointment not found")
	}

	fields := map[string]interface{}{}

	if input.Status != "" {
		// REQ-PAT-012: Validate status transitions
		allowed := map[string][]string{
			"scheduled":       {"checked_in", "cancelled"},
			"checked_in":      {"in_consultation", "cancelled"},
			"in_consultation": {"completed"},
		}
		validTransitions, ok := allowed[a.Status]
		if !ok {
			return errors.New("appointment is in a terminal status and cannot be updated")
		}
		valid := false
		for _, s := range validTransitions {
			if s == input.Status {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("cannot transition from '%s' to '%s'", a.Status, input.Status)
		}
		fields["status"] = input.Status
	}

	if input.ScheduledAt != nil {
		if a.Status != "scheduled" {
			return errors.New("only scheduled appointments can be rescheduled")
		}
		fields["scheduled_at"] = *input.ScheduledAt
	}

	if len(fields) == 0 {
		return errors.New("no fields to update")
	}

	return s.repo.UpdateAppointment(apptID, fields)
}

// ─── Check-in ─────────────────────────────────────────────────────────────────

func (s *PatientService) CheckInAppointment(apptID uint) (*model.Appointment, error) {
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

// ─── Admissions ───────────────────────────────────────────────────────────────

func (s *PatientService) ListAdmissions(patientID uint, status string) ([]model.Admission, error) {
	_, err := s.repo.GetPatientByID(patientID)
	if err != nil {
		return nil, errors.New("patient not found")
	}
	return s.repo.ListAdmissions(patientID, status)
}

func (s *PatientService) GetAdmission(admissionID uint) (*model.Admission, error) {
	a, err := s.repo.GetAdmission(admissionID)
	if err != nil {
		return nil, errors.New("admission not found")
	}
	return a, nil
}

// ─── EHR ──────────────────────────────────────────────────────────────────────

func (s *PatientService) GetFullEHR(patientID uint) (*repository.EHRData, error) {
	_, err := s.repo.GetPatientByID(patientID)
	if err != nil {
		return nil, errors.New("patient not found")
	}
	return s.repo.GetFullEHR(patientID)
}

// ─── Invoices ─────────────────────────────────────────────────────────────────

func (s *PatientService) ListInvoices(patientID uint, status string) ([]model.Invoice, error) {
	_, err := s.repo.GetPatientByID(patientID)
	if err != nil {
		return nil, errors.New("patient not found")
	}
	return s.repo.ListInvoices(patientID, status)
}
