package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"apollo-backend/config"
	"apollo-backend/util"
	"apollo-backend/model"
	"apollo-backend/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo          *repository.UserRepo
	auditRepo         *repository.AuditRepo
	passwordResetRepo *repository.PasswordResetRepo
	cfg               *config.Config
}

func NewAuthService(ur *repository.UserRepo, ar *repository.AuditRepo, prr *repository.PasswordResetRepo, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: ur, auditRepo: ar, passwordResetRepo: prr, cfg: cfg}
}

func (s *AuthService) decryptPassword(encrypted string) (string, error) {
	// If the password doesn't look like base64 (contains @ or spaces or is short), treat as plaintext for backwards compat
	if len(encrypted) < 20 {
		return encrypted, nil
	}
	decrypted, err := util.DecryptAES256(encrypted, s.cfg.AESKeyBytes())
	if err != nil {
		// Fallback: treat as plaintext (for non-encrypted clients)
		return encrypted, nil
	}
	return decrypted, nil
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type Claims struct {
	UserID uuid.UUID  `json:"user_id"`
	Email  string     `json:"email"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

type RegisterInput struct {
	FullName       string     `json:"full_name" binding:"required"`
	Email          string     `json:"email" binding:"required,email"`
	Password       string     `json:"password" binding:"required,min=6"`
	Role           model.Role `json:"role" binding:"required"`
	Phone          string     `json:"phone"`
	// Doctor
	LicenseNumber  string `json:"license_number"`
	Specialization string `json:"specialization"`
	// Pharmacist
	BranchLocation string `json:"branch_location"`
	// Admin
	EmployeeID string `json:"employee_id"`
	Department string `json:"department"`
	AccessKey  string `json:"access_key"`
	// Patient
	DateOfBirth       string `json:"date_of_birth"`
	InsuranceID       string `json:"insurance_id"`
	Gender            string `json:"gender"`
}

type LoginInput struct {
	Email    string     `json:"email" binding:"required,email"`
	Password string     `json:"password" binding:"required,min=6"`
	Role     model.Role `json:"role" binding:"required"`
}

func (s *AuthService) Login(input LoginInput, ip, ua string) (*model.User, *TokenPair, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	password, _ := s.decryptPassword(input.Password)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	if user.Role != input.Role {
		return nil, nil, errors.New("role mismatch")
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	_ = s.userRepo.UpdateLastLogin(user.ID)
	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: user.ID, Action: "login", IPAddress: ip, UserAgent: ua,
	})

	return user, tokens, nil
}

func (s *AuthService) Register(input RegisterInput) (*model.User, *TokenPair, error) {
	existing, err := s.userRepo.FindByEmail(input.Email)
	if err == nil && existing.ID != uuid.Nil {
		return nil, nil, errors.New("email already registered")
	}

	if input.Role == model.RoleAdmin && input.AccessKey != "APOLLO-ADMIN-2024" {
		return nil, nil, errors.New("invalid admin access key")
	}

	password, _ := s.decryptPassword(input.Password)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, errors.New("failed to hash password")
	}

	username := strings.Split(input.Email, "@")[0]

	user := &model.User{
		Email:        input.Email,
		Username:     username,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		Phone:        input.Phone,
		Role:         input.Role,
		IsActive:     true,
	}

	// Attach role-specific profile
	switch input.Role {
	case model.RoleDoctor:
		user.DoctorProfile = &model.DoctorProfile{
			LicenseNumber:  input.LicenseNumber,
			Specialization: input.Specialization,
		}
	case model.RolePatient:
		profile := &model.PatientProfile{
			InsuranceID: input.InsuranceID,
			Gender:      model.Gender(input.Gender),
		}
		if input.DateOfBirth != "" {
			dob, err := time.Parse("2006-01-02", input.DateOfBirth)
			if err != nil {
				return nil, nil, errors.New("invalid date_of_birth format, use YYYY-MM-DD")
			}
			profile.DateOfBirth = &dob
		}
		user.PatientProfile = profile
	case model.RolePharmacist:
		user.PharmacistProfile = &model.PharmacistProfile{
			LicenseNumber:  input.LicenseNumber,
			BranchLocation: input.BranchLocation,
		}
	case model.RoleAdmin:
		user.AdminProfile = &model.AdminProfile{
			EmployeeID:  input.EmployeeID,
			Department:  input.Department,
			AccessLevel: model.AccessStaff,
		}
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, errors.New("failed to create user")
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) GetUserByID(id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) generateTokens(user *model.User) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(s.cfg.JWTExpiryMinutes) * time.Minute)

	accessClaims := Claims{
		UserID: user.ID, Email: user.Email, Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshExp := now.Add(time.Duration(s.cfg.RefreshExpiryDays) * 24 * time.Hour)
	refreshClaims := Claims{
		UserID: user.ID, Email: user.Email, Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken: accessToken, RefreshToken: refreshToken, ExpiresAt: accessExp.Unix(),
	}, nil
}

// ---- Password Reset ----

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyCodeInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type ResetPasswordInput struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=12"`
}

func (s *AuthService) ForgotPassword(input ForgotPasswordInput, ip, ua string) (string, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return "", errors.New("user not found")
	}

	code, err := generateRandomCode()
	if err != nil {
		return "", errors.New("failed to generate reset code")
	}

	pr := &model.PasswordReset{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.passwordResetRepo.Create(pr); err != nil {
		return "", errors.New("failed to create reset code")
	}

	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: user.ID, Action: "password_reset_requested", IPAddress: ip, UserAgent: ua,
	})

	return code, nil
}

func (s *AuthService) VerifyResetCode(input VerifyCodeInput) error {
	_, err := s.passwordResetRepo.FindValidCode(input.Email, input.Code)
	if err != nil {
		return errors.New("invalid or expired code")
	}
	return nil
}

func (s *AuthService) ResetPassword(input ResetPasswordInput, ip, ua string) error {
	pr, err := s.passwordResetRepo.FindValidCode(input.Email, input.Code)
	if err != nil {
		return errors.New("invalid or expired code")
	}

	newPassword, _ := s.decryptPassword(input.NewPassword)
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.userRepo.UpdatePassword(user.ID, string(hash)); err != nil {
		return errors.New("failed to update password")
	}

	_ = s.passwordResetRepo.MarkUsed(pr.ID)
	_ = s.passwordResetRepo.InvalidateAll(user.ID)

	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: user.ID, Action: "password_reset_completed", IPAddress: ip, UserAgent: ua,
	})

	return nil
}

// generateRandomCode returns a cryptographically random 6-digit string.
func generateRandomCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
