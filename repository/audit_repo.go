package repository

import (
	"apollo-backend/model"

	"gorm.io/gorm"
)

type AuditRepo struct{ DB *gorm.DB }

func NewAuditRepo(db *gorm.DB) *AuditRepo { return &AuditRepo{DB: db} }

func (r *AuditRepo) Log(entry *model.AuditLog) error {
	return r.DB.Create(entry).Error
}
