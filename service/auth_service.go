package service

import (
	"errors"
	"strings"
	"time"

	"apollo-backend/config"
	"apollo-backend/model"
	"apollo-backend/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	auditRepo *repository.AuditRepo
	cfg       *config.Config
}

func NewAuthService(ur *repository.UserRepo, ar *repository.AuditRepo, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: ur, auditRepo: ar, cfg: cfg}
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
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

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
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
