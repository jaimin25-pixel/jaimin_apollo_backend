package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"apollo-backend/config"
	"apollo-backend/model"
	"apollo-backend/repository"
	"apollo-backend/util"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo          *repository.UserRepo
	doctorRepo        *repository.DoctorRepo
	auditRepo         *repository.AuditRepo
	passwordResetRepo *repository.PasswordResetRepo
	cfg               *config.Config
}

func NewAuthService(ur *repository.UserRepo, dr *repository.DoctorRepo, ar *repository.AuditRepo, prr *repository.PasswordResetRepo, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: ur, doctorRepo: dr, auditRepo: ar, passwordResetRepo: prr, cfg: cfg}
}

func (s *AuthService) decryptPassword(encrypted string) (string, error) {
	if len(encrypted) < 20 {
		return encrypted, nil
	}
	decrypted, err := util.DecryptAES256(encrypted, s.cfg.AESKeyBytes())
	if err != nil {
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
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type RegisterInput struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResult is returned by Login regardless of whether the caller is a
// system admin (User) or a doctor (Doctor). Exactly one of User / Doctor is set.
type LoginResult struct {
	User   *model.User   `json:"user,omitempty"`
	Doctor *model.Doctor `json:"doctor,omitempty"`
	Tokens *TokenPair    `json:"tokens"`
}

func (s *AuthService) mapDoctorToUser(doc *model.Doctor) *model.User {
	return &model.User{
		ID:        uint(doc.DoctorID),
		Email:     doc.Email,
		Username:  strings.Split(doc.Email, "@")[0],
		FullName:  doc.FullName,
		Phone:     doc.Phone,
		Role:      "doctor",
		IsActive:  strings.ToLower(doc.Status) == "active",
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

func (s *AuthService) Login(input LoginInput, ip, ua string) (*LoginResult, error) {
	password, _ := s.decryptPassword(input.Password)

	// Try the users table first (admin / legacy accounts).
	user, err := s.userRepo.FindByEmail(input.Email)
	if err == nil {
		if bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); bcryptErr != nil {
			return nil, errors.New("invalid credentials")
		}
		tokens, err := s.generateTokens(user)
		if err != nil {
			return nil, err
		}
		_ = s.userRepo.UpdateLastLogin(user.ID)
		_ = s.auditRepo.Log(&model.AuditLog{
			UserID: user.ID, UserRole: user.Role, Action: "LOGIN",
			TblName: "users", RecordID: user.ID, IPAddress: ip,
		})
		return &LoginResult{User: user, Tokens: tokens}, nil
	}

	// Fallback: try the doctors table.
	doc, err := s.doctorRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if bcryptErr := bcrypt.CompareHashAndPassword([]byte(doc.HashedPassword), []byte(password)); bcryptErr != nil {
		return nil, errors.New("invalid credentials")
	}
	tokens, err := s.generateTokensForDoctor(doc)
	if err != nil {
		return nil, err
	}
	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: doc.DoctorID, UserRole: "doctor", Action: "LOGIN",
		TblName: "doctors", RecordID: doc.DoctorID, IPAddress: ip,
	})
	doctorUser := s.mapDoctorToUser(doc)
	return &LoginResult{User: doctorUser, Doctor: doc, Tokens: tokens}, nil
}

func (s *AuthService) Register(input RegisterInput) (*model.User, *TokenPair, error) {
	existing, err := s.userRepo.FindByEmail(input.Email)
	if err == nil && existing.ID != 0 {
		return nil, nil, errors.New("email already registered")
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
		Role:         "admin",
		IsActive:     true,
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

func (s *AuthService) Logout(userID uint, ip, ua string) {
	_ = s.auditRepo.Log(&model.AuditLog{
		UserID: userID, UserRole: "admin", Action: "LOGOUT",
		TblName: "users", RecordID: userID, IPAddress: ip,
	})
}

type GetMeResponse struct {
	User   *model.User   `json:"user,omitempty"`
	Doctor *model.Doctor `json:"doctor,omitempty"`
}

func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err == nil && user != nil {
		return user, nil
	}

	doc, err := s.doctorRepo.FindByID(id)
	if err == nil && doc != nil {
		return s.mapDoctorToUser(doc), nil
	}

	return nil, errors.New("user not found")
}

// GetUserWithProfile returns both User/Doctor data with role-specific profile info
func (s *AuthService) GetUserWithProfile(id uint) (*GetMeResponse, error) {
	// Try users table first
	user, err := s.userRepo.FindByID(id)
	if err == nil && user != nil {
		return &GetMeResponse{User: user}, nil
	}

	// Try doctors table
	doc, err := s.doctorRepo.FindByID(id)
	if err == nil && doc != nil {
		return &GetMeResponse{
			User:   s.mapDoctorToUser(doc),
			Doctor: doc,
		}, nil
	}

	return nil, errors.New("user not found")
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
			Subject:   fmt.Sprintf("%d", user.ID),
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
			Subject:   fmt.Sprintf("%d", user.ID),
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

func (s *AuthService) generateTokensForDoctor(doc *model.Doctor) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(s.cfg.JWTExpiryMinutes) * time.Minute)

	accessClaims := Claims{
		UserID: doc.DoctorID, Email: doc.Email, Role: "doctor",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", doc.DoctorID),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshExp := now.Add(time.Duration(s.cfg.RefreshExpiryDays) * 24 * time.Hour)
	refreshClaims := Claims{
		UserID: doc.DoctorID, Email: doc.Email, Role: "doctor",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", doc.DoctorID),
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
		UserID: user.ID, UserRole: user.Role, Action: "UPDATE",
		TblName: "password_resets", RecordID: pr.ID, IPAddress: ip,
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
		UserID: user.ID, UserRole: user.Role, Action: "UPDATE",
		TblName: "users", RecordID: user.ID, IPAddress: ip,
	})

	return nil
}

func generateRandomCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
