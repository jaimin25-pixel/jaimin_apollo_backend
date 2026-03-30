package repository

import (
	"apollo-backend/model"

	"gorm.io/gorm"
)

type PharmacistRepo struct{ DB *gorm.DB }

func NewPharmacistRepo(db *gorm.DB) *PharmacistRepo { return &PharmacistRepo{DB: db} }

func (r *PharmacistRepo) FindByEmail(email string) (*model.Pharmacist, error) {
	var p model.Pharmacist
	err := r.DB.Where("email = ? AND status = 'active'", email).First(&p).Error
	return &p, err
}

func (r *PharmacistRepo) FindByID(id uint) (*model.Pharmacist, error) {
	var p model.Pharmacist
	err := r.DB.Where("pharmacist_id = ? AND status = 'active'", id).First(&p).Error
	return &p, err
}
