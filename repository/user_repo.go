package repository

import (
	"apollo-backend/model"

	"gorm.io/gorm"
)

type UserRepo struct{ DB *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{DB: db} }

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("email = ? AND is_active = true", email).First(&u).Error
	return &u, err
}

func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.DB.Where("id = ? AND is_active = true", id).First(&u).Error
	return &u, err
}

func (r *UserRepo) Create(u *model.User) error {
	return r.DB.Create(u).Error
}

func (r *UserRepo) UpdateLastLogin(id uint) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *UserRepo) UpdatePassword(id uint, hash string) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Update("password_hash", hash).Error
}
