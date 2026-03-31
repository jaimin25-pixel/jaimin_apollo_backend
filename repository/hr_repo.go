package repository

import (
	"apollo-backend/model"
	"time"

	"gorm.io/gorm"
)

type HRRepo interface {
	GetDashboardStats() (map[string]interface{}, error)
	ListStaff(deptID *uint, role string) ([]model.Staff, error)
	CreateStaff(user *model.User, staff *model.Staff) error
	UpdateStaff(id uint, updates map[string]interface{}) error

	ListShifts(staffID *uint, date string) ([]model.StaffShift, error)
	ClockIn(staffID uint) error
	ClockOut(staffID uint) error

	ListLeaves(staffID *uint, status string) ([]model.LeaveRequest, error)
	ApplyLeave(req *model.LeaveRequest) error
	UpdateLeaveStatus(id uint, status string, notes string, approvedBy *uint) error

	ListPayroll(month int, year int, staffID *uint) ([]model.Payroll, error)
	GeneratePayroll(payrolls []model.Payroll) error
}

type hrRepo struct {
	db *gorm.DB
}

func NewHRRepo(db *gorm.DB) HRRepo {
	return &hrRepo{db}
}

func (r *hrRepo) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	var totalStaff, onLeave, pendingLeaves int64

	r.db.Model(&model.Staff{}).Where("status = ?", "active").Count(&totalStaff)

	today := time.Now().Truncate(24 * time.Hour)
	r.db.Model(&model.LeaveRequest{}).Where("status = ? AND from_date <= ? AND to_date >= ?", "approved", today, today).Count(&onLeave)
	r.db.Model(&model.LeaveRequest{}).Where("status = ?", "pending").Count(&pendingLeaves)

	// "On Duty" simple proxy
	onDuty := totalStaff - onLeave

	stats["total_staff"] = totalStaff
	stats["on_duty"] = onDuty
	stats["on_leave"] = onLeave
	stats["pending_leave_requests"] = pendingLeaves

	return stats, nil
}

func (r *hrRepo) ListStaff(deptID *uint, role string) ([]model.Staff, error) {
	var staff []model.Staff
	q := r.db.Preload("Department").Preload("User")
	if deptID != nil && *deptID > 0 {
		q = q.Where("dept_id = ?", *deptID)
	}
	if role != "" {
		q = q.Where("role = ?", role)
	}
	err := q.Find(&staff).Error
	return staff, err
}

func (r *hrRepo) CreateStaff(user *model.User, staff *model.Staff) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(staff).Error
	})
}

func (r *hrRepo) UpdateStaff(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Staff{}).Where("staff_id = ?", id).Updates(updates).Error
}

func (r *hrRepo) ListShifts(staffID *uint, date string) ([]model.StaffShift, error) {
	var shifts []model.StaffShift
	q := r.db.Preload("Staff")
	if staffID != nil && *staffID > 0 {
		q = q.Where("staff_id = ?", *staffID)
	}
	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			q = q.Where("shift_date = ?", t)
		}
	}
	err := q.Order("shift_date desc").Find(&shifts).Error
	return shifts, err
}

func (r *hrRepo) ClockIn(staffID uint) error {
	now := time.Now()
	// create shift record
	shift := model.StaffShift{
		StaffID:   staffID,
		ShiftDate: now.Truncate(24 * time.Hour),
		CheckInAt: &now,
		ShiftType: "regular",
		StartTime: "09:00:00",
		EndTime:   "17:00:00",
	}
	// Check if already clocked in today
	var existing model.StaffShift
	if err := r.db.Where("staff_id = ? AND shift_date = ?", staffID, shift.ShiftDate).First(&existing).Error; err == nil {
		// Update early clock in
		return r.db.Model(&model.StaffShift{}).Where("shift_id = ?", existing.ShiftID).Update("check_in_at", now).Error
	}
	return r.db.Create(&shift).Error
}

func (r *hrRepo) ClockOut(staffID uint) error {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	return r.db.Model(&model.StaffShift{}).Where("staff_id = ? AND shift_date = ?", staffID, today).Update("check_out_at", now).Error
}

func (r *hrRepo) ListLeaves(staffID *uint, status string) ([]model.LeaveRequest, error) {
	var leaves []model.LeaveRequest
	q := r.db.Preload("Staff").Preload("Approver")
	if staffID != nil && *staffID > 0 {
		q = q.Where("staff_id = ?", *staffID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at desc").Find(&leaves).Error
	return leaves, err
}

func (r *hrRepo) ApplyLeave(req *model.LeaveRequest) error {
	return r.db.Create(req).Error
}

func (r *hrRepo) UpdateLeaveStatus(id uint, status string, notes string, approvedBy *uint) error {
	return r.db.Model(&model.LeaveRequest{}).Where("request_id = ?", id).Updates(map[string]interface{}{
		"status":         status,
		"approval_notes": notes,
		"approved_by":    approvedBy,
	}).Error
}

func (r *hrRepo) ListPayroll(month int, year int, staffID *uint) ([]model.Payroll, error) {
	var records []model.Payroll
	q := r.db.Preload("Staff")
	if month > 0 {
		q = q.Where("month = ?", month)
	}
	if year > 0 {
		q = q.Where("year = ?", year)
	}
	if staffID != nil && *staffID > 0 {
		q = q.Where("staff_id = ?", *staffID)
	}
	err := q.Order("year desc, month desc").Find(&records).Error
	return records, err
}

func (r *hrRepo) GeneratePayroll(payrolls []model.Payroll) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, p := range payrolls {
			// Upsert or simple create
			var existing model.Payroll
			if err := tx.Where("staff_id = ? AND month = ? AND year = ?", p.StaffID, p.Month, p.Year).First(&existing).Error; err == nil {
				tx.Model(&existing).Updates(p)
			} else {
				if err := tx.Create(&p).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
