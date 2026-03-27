package repository

import (
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

type DashboardRepo struct{ DB *gorm.DB }

func NewDashboardRepo(db *gorm.DB) *DashboardRepo { return &DashboardRepo{DB: db} }

func (r *DashboardRepo) CountDoctors() int64 {
	var c int64
	r.DB.Model(&model.Doctor{}).Where("status = 'active'").Count(&c)
	return c
}

func (r *DashboardRepo) CountPatients() int64 {
	var c int64
	r.DB.Model(&model.Patient{}).Count(&c)
	return c
}

func (r *DashboardRepo) CountStaff() int64 {
	var c int64
	r.DB.Model(&model.Staff{}).Where("status = 'active'").Count(&c)
	return c
}

func (r *DashboardRepo) CountDepartments() int64 {
	var c int64
	r.DB.Model(&model.Department{}).Where("status = 'active'").Count(&c)
	return c
}

func (r *DashboardRepo) CountAvailableBeds() int64 {
	var c int64
	r.DB.Model(&model.Bed{}).Where("status = 'available'").Count(&c)
	return c
}

func (r *DashboardRepo) CountAppointmentsToday() int64 {
	var c int64
	today := time.Now().Truncate(24 * time.Hour)
	r.DB.Model(&model.Appointment{}).Where("scheduled_at >= ? AND scheduled_at < ?", today, today.Add(24*time.Hour)).Count(&c)
	return c
}

func (r *DashboardRepo) RecentAuditLogs(userID uint, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.DB.Where("user_id = ?", userID).Order("ts DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
