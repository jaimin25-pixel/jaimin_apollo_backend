package repository

import (
	"time"

	"apollo-backend/model"

	"gorm.io/gorm"
)

type PasswordResetRepo struct{ DB *gorm.DB }

func NewPasswordResetRepo(db *gorm.DB) *PasswordResetRepo {
	return &PasswordResetRepo{DB: db}
}

func (r *PasswordResetRepo) Create(pr *model.PasswordReset) error {
	return r.DB.Create(pr).Error
}

// FindValidCode joins the users table to look up by email, and checks
// the code is not used and not expired.
func (r *PasswordResetRepo) FindValidCode(email string, code string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	err := r.DB.
		Joins("JOIN users ON users.id = password_resets.user_id").
		Where("users.email = ? AND password_resets.code = ? AND password_resets.used = false AND password_resets.expires_at > ?", email, code, time.Now()).
		First(&pr).Error
	return &pr, err
}

func (r *PasswordResetRepo) MarkUsed(id uint) error {
	return r.DB.Model(&model.PasswordReset{}).Where("id = ?", id).Update("used", true).Error
}

// InvalidateAll marks all unused codes for this user as used.
func (r *PasswordResetRepo) InvalidateAll(userID uint) error {
	return r.DB.Model(&model.PasswordReset{}).
		Where("user_id = ? AND used = false", userID).
		Update("used", true).Error
}
