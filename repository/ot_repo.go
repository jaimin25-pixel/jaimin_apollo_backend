package repository

import (
	"apollo-backend/model"
	"time"

	"gorm.io/gorm"
)

type OTRepo interface {
	GetDashboardStats(date string) (map[string]interface{}, error)
	ListSchedules(date string, roomID *uint, status string) ([]model.OTSchedule, error)
	CreateSchedule(s *model.OTSchedule) error
	GetSchedule(id uint) (*model.OTSchedule, error)
	UpdateSchedule(id uint, updates map[string]interface{}) error
	UpdateScheduleStatus(id uint, status string) error
	DeleteSchedule(id uint, reason string) error
	UpdateNotes(id uint, surgicalNotes, anesthesiaRecord string) error

	ListRooms(date string) ([]model.Ward, error)
	SterilizeRoom(id uint) error
}

type otRepo struct {
	db *gorm.DB
}

func NewOTRepo(db *gorm.DB) OTRepo {
	return &otRepo{db}
}

func (r *otRepo) GetDashboardStats(date string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	targetDate := time.Now().Truncate(24 * time.Hour)
	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			targetDate = t
		}
	}

	var scheduledTotal, inProgress, completed int64
	q := r.db.Model(&model.OTSchedule{}).Where("scheduled_at >= ? AND scheduled_at < ?", targetDate, targetDate.Add(24*time.Hour))
	q.Count(&scheduledTotal)
	q.Where("status = ?", "in_progress").Count(&inProgress)
	q.Where("status = ?", "completed").Count(&completed)

	stats["scheduled"] = scheduledTotal
	stats["in_progress"] = inProgress
	stats["completed"] = completed

	return stats, nil
}

func (r *otRepo) ListSchedules(date string, roomID *uint, status string) ([]model.OTSchedule, error) {
	var schedules []model.OTSchedule
	q := r.db.Preload("Room").Preload("Patient").Preload("Surgeon").Preload("Anesthesiologist")
	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			q = q.Where("scheduled_at >= ? AND scheduled_at < ?", t, t.Add(24*time.Hour))
		}
	}
	if roomID != nil && *roomID > 0 {
		q = q.Where("room_id = ?", *roomID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("scheduled_at asc").Find(&schedules).Error
	return schedules, err
}

func (r *otRepo) CreateSchedule(s *model.OTSchedule) error {
	return r.db.Create(s).Error
}

func (r *otRepo) GetSchedule(id uint) (*model.OTSchedule, error) {
	var s model.OTSchedule
	err := r.db.Preload("Room").Preload("Patient").Preload("Surgeon").Preload("Anesthesiologist").First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *otRepo) UpdateSchedule(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.OTSchedule{}).Where("ot_id = ?", id).Updates(updates).Error
}

func (r *otRepo) UpdateScheduleStatus(id uint, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == "in_progress" {
		now := time.Now()
		updates["actual_start_at"] = &now
	} else if status == "completed" {
		now := time.Now()
		updates["actual_end_at"] = &now
	}
	return r.db.Model(&model.OTSchedule{}).Where("ot_id = ?", id).Updates(updates).Error
}

func (r *otRepo) DeleteSchedule(id uint, reason string) error {
	return r.db.Model(&model.OTSchedule{}).Where("ot_id = ?", id).Update("status", "cancelled").Error
}

func (r *otRepo) UpdateNotes(id uint, surgicalNotes, anesthesiaRecord string) error {
	return r.db.Model(&model.OTSchedule{}).Where("ot_id = ?", id).Updates(map[string]interface{}{
		"surgical_notes":    surgicalNotes,
		"anesthesia_record": anesthesiaRecord,
	}).Error
}

func (r *otRepo) ListRooms(date string) ([]model.Ward, error) {
	var rooms []model.Ward
	// Wards of type 'ot'
	err := r.db.Where("ward_type = ?", "ot").Find(&rooms).Error
	return rooms, err
}

func (r *otRepo) SterilizeRoom(id uint) error {
	// A room's "sterilization format" is kept on its active schedule or via Ward status. We can reuse Ward status="active" assuming maintenance="needs sterilizing"
	return r.db.Model(&model.Ward{}).Where("ward_id = ?", id).Update("status", "active").Error
}
