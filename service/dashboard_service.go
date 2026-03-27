package service

import (
	"apollo-backend/model"
	"apollo-backend/repository"
)

type DashboardService struct {
	dashRepo *repository.DashboardRepo
	userRepo *repository.UserRepo
}

func NewDashboardService(dr *repository.DashboardRepo, ur *repository.UserRepo) *DashboardService {
	return &DashboardService{dashRepo: dr, userRepo: ur}
}

type DashboardStats struct {
	TotalDoctors      int64 `json:"total_doctors"`
	TotalPatients     int64 `json:"total_patients"`
	TotalStaff        int64 `json:"total_staff"`
	AppointmentsToday int64 `json:"appointments_today"`
	AvailableBeds     int64 `json:"available_beds"`
	TotalDepartments  int64 `json:"total_departments"`
}

type DashboardData struct {
	Stats          DashboardStats   `json:"stats"`
	RecentActivity []model.AuditLog `json:"recent_activity"`
}

func (s *DashboardService) GetDashboard(userID uint) (*DashboardData, error) {
	activity, err := s.dashRepo.RecentAuditLogs(userID, 10)
	if err != nil {
		activity = []model.AuditLog{}
	}

	return &DashboardData{
		Stats: DashboardStats{
			TotalDoctors:      s.dashRepo.CountDoctors(),
			TotalPatients:     s.dashRepo.CountPatients(),
			TotalStaff:        s.dashRepo.CountStaff(),
			AppointmentsToday: s.dashRepo.CountAppointmentsToday(),
			AvailableBeds:     s.dashRepo.CountAvailableBeds(),
			TotalDepartments:  s.dashRepo.CountDepartments(),
		},
		RecentActivity: activity,
	}, nil
}
