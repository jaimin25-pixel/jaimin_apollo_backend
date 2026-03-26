# Apollo HMS - Backend

## Project Overview
Hospital Management System (HMS) backend built with **Go 1.22**, **Gin**, **GORM**, and **PostgreSQL**.

## Tech Stack
- **Language:** Go 1.22
- **Router:** Gin (github.com/gin-gonic/gin)
- **ORM:** GORM (gorm.io/gorm) with PostgreSQL driver
- **Auth:** JWT (HS256) with access + refresh tokens
- **Encryption:** AES-256-GCM for password transit, bcrypt for storage
- **Config:** Viper (.env file)
- **Database:** PostgreSQL (UUID primary keys via `gen_random_uuid()`)

## Architecture (Clean Architecture)
```
apollo-backend/
├── main.go                  # Entry point, routes, DI wiring, CORS
├── config/
│   ├── config.go            # Viper config loader (DB, JWT, AES settings)
│   └── database.go          # GORM connect, AutoMigrate, admin seed
├── model/
│   ├── user.go              # User struct (UUID PK, 4 role relations)
│   ├── doctor.go            # DoctorProfile (license, specialization, fees)
│   ├── patient.go           # PatientProfile (DOB, gender, blood group, insurance)
│   ├── pharmacist.go        # PharmacistProfile (license, branch, shift)
│   ├── admin.go             # AdminProfile (employee_id, department, access_level)
│   ├── audit_log.go         # Immutable audit trail (no DeletedAt)
│   └── password_reset.go    # OTP code with expiry (immutable)
├── repository/
│   ├── user_repo.go         # FindByEmail, FindByID (Preload all profiles), Create, UpdateLastLogin, UpdatePassword
│   ├── password_reset_repo.go # Create, FindValidCode, MarkUsed, InvalidateAll
│   └── audit_repo.go        # Log audit entries
├── service/
│   └── auth_service.go      # Login, Register, ForgotPassword, VerifyResetCode, ResetPassword, JWT generation
├── handler/
│   └── auth_handler.go      # HTTP handlers: Login, Register, ForgotPassword, VerifyCode, ResetPassword, Me
├── middleware/
│   ├── auth.go              # JWT Bearer token extraction & claims validation
│   └── crypto.go            # (Legacy) crypto middleware reference
├── util/
│   └── crypto.go            # AES-256-GCM DecryptAES256 function
└── swagger-apollo-api.json  # OpenAPI 3.0.3 API documentation
```

## Database Tables
| Table | Description | Key Indexes |
|-------|-------------|-------------|
| `users` | Core user accounts | `idx_users_role`, `idx_users_active`, unique email & username |
| `doctor_profiles` | Doctor-specific data | `idx_doctor_user` (unique), `idx_doctor_license` (unique), `idx_doctor_specialization` |
| `patient_profiles` | Patient-specific data | `idx_patient_user` (unique), `idx_patient_gender`, `idx_patient_blood`, `idx_patient_insurance` |
| `pharmacist_profiles` | Pharmacist-specific data | `idx_pharmacist_user` (unique), `idx_pharmacist_license` (unique), `idx_pharmacist_branch` |
| `admin_profiles` | Admin-specific data | `idx_admin_user` (unique), `idx_admin_employee` (unique), `idx_admin_department` |
| `audit_logs` | Login/password reset events | Immutable (no soft delete) |
| `password_resets` | 6-digit OTP codes | 10 min expiry, `used` flag |

## API Endpoints
| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `POST` | `/api/auth/register` | No | Register user (doctor/patient/pharmacist/admin) |
| `POST` | `/api/auth/login` | No | Login with AES-256 encrypted password |
| `GET` | `/api/auth/me` | JWT | Get current user profile with role-specific data |
| `POST` | `/api/auth/forgot-password` | No | Request 6-digit OTP reset code |
| `POST` | `/api/auth/verify-code` | No | Verify OTP code validity |
| `POST` | `/api/auth/reset-password` | No | Reset password with OTP + new AES-encrypted password |
| `GET` | `/api/auth/encryption-key` | No | Get AES-256 key (dev mode) |
| `GET` | `/api/health` | No | Health check |

## User Roles
- `doctor` — License number, specialization, consultation fee, availability
- `patient` — DOB, gender, blood group, insurance, emergency contact, allergies
- `pharmacist` — License number, branch location, shift, duty status
- `admin` — Employee ID, department, access level (super_admin/manager/staff), requires access key `APOLLO-ADMIN-2024`

## Environment Variables (.env)
```
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hms
DB_USER=postgres
DB_PASSWORD=jaimin123
JWT_SECRET=apollo-hospital-jwt-secret-2024
JWT_EXPIRY_MINUTES=60
REFRESH_EXPIRY_DAYS=7
AES_ENCRYPTION_KEY=apollo-aes256-encryption-key-32
```

## Commands
```bash
# Run server
go run main.go

# Build
go build -o main.exe .

# Run built binary
./main.exe
```

Server runs on `http://localhost:8080`

## Security
- Passwords encrypted with **AES-256-GCM** in transit (frontend → backend)
- Passwords hashed with **bcrypt** (DefaultCost) at rest
- **JWT HS256** tokens with configurable expiry
- CORS restricted to `localhost:3000` and `localhost:5173`
- Admin registration requires secret access key
- Audit logging for login and password reset events
- Plaintext password fallback for backwards compatibility

---

## Task Log — 2026-03-26

### AP-01: Authentication System (Full Build)

| # | Task | Status |
|---|------|--------|
| 1 | Project scaffolding — Go module, Gin router, Viper config, GORM PostgreSQL | Done |
| 2 | User model with UUID PK, role enum (doctor/patient/pharmacist/admin) | Done |
| 3 | Role-specific profile models — DoctorProfile, PatientProfile, PharmacistProfile, AdminProfile | Done |
| 4 | AuditLog model (immutable, no soft delete) | Done |
| 5 | PasswordReset model (6-digit OTP, 10 min expiry, used flag) | Done |
| 6 | Database AutoMigrate all 7 tables with proper indexes and foreign keys (ON DELETE CASCADE) | Done |
| 7 | Admin user seed on startup | Done |
| 8 | UserRepo — FindByEmail, FindByID (Preload 4 profiles), Create, UpdateLastLogin, UpdatePassword | Done |
| 9 | PasswordResetRepo — Create, FindValidCode (JOIN users), MarkUsed, InvalidateAll | Done |
| 10 | AuditRepo — Log method | Done |
| 11 | AuthService — Login with credential validation and role check | Done |
| 12 | AuthService — Register with role-specific profile creation | Done |
| 13 | AuthService — JWT token generation (access + refresh, HS256) | Done |
| 14 | AuthService — ForgotPassword (crypto/rand 6-digit code, 10 min expiry) | Done |
| 15 | AuthService — VerifyResetCode and ResetPassword | Done |
| 16 | AuthHandler — Login, Register, ForgotPassword, VerifyCode, ResetPassword, Me endpoints | Done |
| 17 | JWT middleware — Bearer token extraction, claims validation | Done |
| 18 | CORS configuration for frontend origins | Done |
| 19 | AES-256-GCM password decryption in service layer (`util/crypto.go`) | Done |
| 20 | AES encryption key endpoint (`/api/auth/encryption-key`) | Done |
| 21 | Plaintext password fallback for backwards compatibility | Done |
| 22 | Swagger API documentation (`swagger-apollo-api.json`) — OpenAPI 3.0.3 | Done |
