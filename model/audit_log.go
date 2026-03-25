package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog is immutable — no DeletedAt, no gorm.Model
type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_audit_user" json:"user_id"`
	Action    string    `gorm:"not null;size:50;index:idx_audit_action" json:"action"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	Detail    string    `gorm:"type:text" json:"detail,omitempty"`
	CreatedAt time.Time `gorm:"not null;index:idx_audit_created" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }
