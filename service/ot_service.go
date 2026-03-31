package service

import (
	"apollo-backend/model"
	"apollo-backend/repository"
	"errors"
)

type OTService interface {
	GetDashboard(date string) (map[string]interface{}, error)
	ListSchedules(date string, roomID *uint, status string) ([]model.OTSchedule, error)
	CreateSchedule(s *model.OTSchedule) error
	GetSchedule(id uint) (*model.OTSchedule, error)
	UpdateSchedule(id uint, updates map[string]interface{}) error
	AdvanceStatus(id uint, status string) error
	CancelSchedule(id uint, reason string) error
	AddNotes(id uint, surgicalNotes, anesthesiaRecord string) error

	ListRooms(date string) ([]model.Ward, error)
	SterilizeRoom(id uint) error
}

type otService struct {
	repo repository.OTRepo
}

func NewOTService(repo repository.OTRepo) OTService {
	return &otService{repo}
}

func (s *otService) GetDashboard(date string) (map[string]interface{}, error) {
	return s.repo.GetDashboardStats(date)
}

func (s *otService) ListSchedules(date string, roomID *uint, status string) ([]model.OTSchedule, error) {
	return s.repo.ListSchedules(date, roomID, status)
}

func (s *otService) CreateSchedule(sched *model.OTSchedule) error {
	if sched.PatientID == 0 || sched.RoomID == 0 || sched.SurgeonID == 0 {
		return errors.New("patient, room, and surgeon are required")
	}
	return s.repo.CreateSchedule(sched)
}

func (s *otService) GetSchedule(id uint) (*model.OTSchedule, error) {
	return s.repo.GetSchedule(id)
}

func (s *otService) UpdateSchedule(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateSchedule(id, updates)
}

func (s *otService) AdvanceStatus(id uint, status string) error {
	valid := map[string]bool{"prepared": true, "in_progress": true, "completed": true, "closed": true}
	if !valid[status] {
		return errors.New("invalid status advancement")
	}
	return s.repo.UpdateScheduleStatus(id, status)
}

func (s *otService) CancelSchedule(id uint, reason string) error {
	return s.repo.DeleteSchedule(id, reason)
}

func (s *otService) AddNotes(id uint, surgicalNotes, anesthesiaRecord string) error {
	return s.repo.UpdateNotes(id, surgicalNotes, anesthesiaRecord)
}

func (s *otService) ListRooms(date string) ([]model.Ward, error) {
	return s.repo.ListRooms(date)
}

func (s *otService) SterilizeRoom(id uint) error {
	return s.repo.SterilizeRoom(id)
}
