package repository

import (
	"apollo-backend/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo struct{ DB *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{DB: db} }

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.DB.
		Preload("DoctorProfile").
		Preload("PatientProfile").
		Preload("PharmacistProfile").
		Preload("AdminProfile").
		Where("email = ? AND is_active = true", email).First(&u).Error
	return &u, err
}

func (r *UserRepo) FindByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.DB.
		Preload("DoctorProfile").
		Preload("PatientProfile").
		Preload("PharmacistProfile").
		Preload("AdminProfile").
		Where("id = ? AND is_active = true", id).First(&u).Error
	return &u, err
}

func (r *UserRepo) Create(u *model.User) error {
	return r.DB.Create(u).Error
}

func (r *UserRepo) UpdateLastLogin(id uuid.UUID) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}
