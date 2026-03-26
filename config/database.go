package config

import (
	"fmt"
	"log"

	"apollo-backend/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Kolkata",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	migrate(db)
	seed(db)
	log.Println("database connected & migrated")
	return db
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.DoctorProfile{},
		&model.PatientProfile{},
		&model.PharmacistProfile{},
		&model.AdminProfile{},
		&model.AuditLog{},
		&model.PasswordReset{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func seed(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("email = ?", "admin@apollo.health").Count(&count)
	if count > 0 {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Email:        "admin@apollo.health",
		Username:     "admin",
		PasswordHash: string(hash),
		FullName:     "System Admin",
		Phone:        "+91-9876543210",
		Role:         model.RoleAdmin,
		IsActive:     true,
		AdminProfile: &model.AdminProfile{
			EmployeeID:  "AP-00001",
			Department:  "Administration",
			AccessLevel: model.AccessSuperAdmin,
		},
	}
	db.Create(&admin)
	log.Println("seeded default admin: admin@apollo.health / admin123")
}
