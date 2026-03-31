package service

import (
	"apollo-backend/model"
	"apollo-backend/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type HRService interface {
	GetDashboard() (map[string]interface{}, error)
	ListStaff(deptID *uint, role string) ([]model.Staff, error)
	CreateStaff(req *StaffCreateRequest) error
	UpdateStaff(id uint, updates map[string]interface{}) error

	ListShifts(staffID *uint, date string) ([]model.StaffShift, error)
	ClockInOut(staffID uint, action string) error

	ListLeaves(staffID *uint, status string) ([]model.LeaveRequest, error)
	ApplyLeave(req *model.LeaveRequest) error
	ProcessLeave(id uint, status, notes string, approverID *uint) error

	ListPayroll(month, year int, staffID *uint) ([]model.Payroll, error)
	GeneratePayroll(month, year int, generatedBy uint) error
}

type hrService struct {
	repo repository.HRRepo
}

func NewHRService(repo repository.HRRepo) HRService {
	return &hrService{repo}
}

func (s *hrService) GetDashboard() (map[string]interface{}, error) {
	return s.repo.GetDashboardStats()
}

func (s *hrService) ListStaff(deptID *uint, role string) ([]model.Staff, error) {
	return s.repo.ListStaff(deptID, role)
}

type StaffCreateRequest struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Role     string  `json:"role"`
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	DeptID   uint    `json:"dept_id"`
	BaseSalary float64 `json:"base_salary"`
}

func (s *hrService) CreateStaff(req *StaffCreateRequest) error {
	if req.Username == "" || req.Password == "" || req.Role == "" {
		return errors.New("username, password, and role are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		Username:       req.Username,
		Email:          req.Email,
		FullName:       req.FullName,
		Phone:          req.Phone,
		PasswordHash:   string(hash),
		Role:           req.Role,
	}

	var deptPtr *uint
	if req.DeptID > 0 {
		deptPtr = &req.DeptID
	}

	staff := model.Staff{
		Role:           req.Role, // doctor, nurse, receptionist, pharmacist, admin, hr, etc.
		DeptID:         deptPtr,
		Status:         "active",
		FullName:       req.FullName,
		Email:          req.Email,
		HashedPassword: string(hash),
	}

	return s.repo.CreateStaff(&user, &staff)
}

func (s *hrService) UpdateStaff(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateStaff(id, updates)
}

func (s *hrService) ListShifts(staffID *uint, date string) ([]model.StaffShift, error) {
	return s.repo.ListShifts(staffID, date)
}

func (s *hrService) ClockInOut(staffID uint, action string) error {
	if action == "in" {
		return s.repo.ClockIn(staffID)
	} else if action == "out" {
		return s.repo.ClockOut(staffID)
	}
	return errors.New("invalid action")
}

func (s *hrService) ListLeaves(staffID *uint, status string) ([]model.LeaveRequest, error) {
	return s.repo.ListLeaves(staffID, status)
}

func (s *hrService) ApplyLeave(req *model.LeaveRequest) error {
	req.Status = "pending"
	return s.repo.ApplyLeave(req)
}

func (s *hrService) ProcessLeave(id uint, status, notes string, approverID *uint) error {
	valid := map[string]bool{"approved": true, "rejected": true}
	if !valid[status] {
		return errors.New("status must be approved or rejected")
	}
	return s.repo.UpdateLeaveStatus(id, status, notes, approverID)
}

func (s *hrService) ListPayroll(month, year int, staffID *uint) ([]model.Payroll, error) {
	return s.repo.ListPayroll(month, year, staffID)
}

func (s *hrService) GeneratePayroll(month, year int, generatedBy uint) error {
	// Simple mock payroll generation
	staffList, err := s.repo.ListStaff(nil, "")
	if err != nil {
		return err
	}

	var payrolls []model.Payroll
	for _, st := range staffList {
		if st.Status != "active" {
			continue
		}
		
		var mockSalary float64 = 5000.0 // Could pull from staff record BaseSalary if added
		mockDeductions := 500.0
		net := mockSalary - mockDeductions

		p := model.Payroll{
			StaffID:     st.StaffID,
			Month:       month,
			Year:        year,
			BasicSalary: mockSalary,
			Deductions:  mockDeductions,
			NetSalary:   net,
			Status:      "generated",
			GeneratedBy: &generatedBy,
		}
		payrolls = append(payrolls, p)
	}

	return s.repo.GeneratePayroll(payrolls)
}
