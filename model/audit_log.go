package model

import (
	"time"

	"gorm.io/datatypes"
)

// AuditLog is an immutable audit trail for all data modifications.
// Populated by database triggers. Records are never deleted.
type AuditLog struct {
	LogID     uint64         `gorm:"primaryKey;autoIncrement" json:"log_id"`
	UserID    uint           `gorm:"not null;index:idx_audit_user" json:"user_id"`
	UserRole  string         `gorm:"not null;size:50" json:"user_role"`
	Action    string         `gorm:"not null;size:20;index:idx_audit_action" json:"action"`
	TblName   string         `gorm:"column:table_name;not null;size:100" json:"table_name"`
	RecordID  uint           `gorm:"not null" json:"record_id"`
	OldVal    datatypes.JSON `gorm:"type:jsonb" json:"old_val,omitempty"`
	NewVal    datatypes.JSON `gorm:"type:jsonb" json:"new_val,omitempty"`
	IPAddress string         `gorm:"type:inet;size:45" json:"ip_address,omitempty"`
	Ts        time.Time      `gorm:"not null;default:now();index:idx_audit_ts" json:"ts"`
}

func (AuditLog) TableName() string { return "audit_logs" }
