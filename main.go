package main

import (
	"log"

	"apollo-backend/config"
	"apollo-backend/handler"
	"apollo-backend/middleware"
	"apollo-backend/repository"
	"apollo-backend/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db := config.ConnectDB(cfg)

	// repositories
	userRepo := repository.NewUserRepo(db)
	auditRepo := repository.NewAuditRepo(db)
	passwordResetRepo := repository.NewPasswordResetRepo(db)
	dashRepo := repository.NewDashboardRepo(db)
	adminRepo := repository.NewAdminRepo(db)

	// services
	authSvc := service.NewAuthService(userRepo, auditRepo, passwordResetRepo, cfg)
	dashSvc := service.NewDashboardService(dashRepo, userRepo)
	adminSvc := service.NewAdminService(adminRepo, auditRepo, cfg)

	// handlers
	authH := handler.NewAuthHandler(authSvc)
	dashH := handler.NewDashboardHandler(dashSvc)
	adminH := handler.NewAdminHandler(adminSvc)

	// router
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3002", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/forgot-password", authH.ForgotPassword)
			auth.POST("/verify-code", authH.VerifyCode)
			auth.POST("/reset-password", authH.ResetPassword)
			auth.GET("/me", middleware.JWTAuth(authSvc), authH.Me)
			auth.POST("/logout", middleware.JWTAuth(authSvc), authH.Logout)
		}
		api.GET("/auth/encryption-key", func(c *gin.Context) {
			c.JSON(200, gin.H{"key": cfg.AESKey})
		})
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Dashboard (JWT protected)
		api.GET("/dashboard", middleware.JWTAuth(authSvc), dashH.GetDashboard)

		// Admin routes (JWT + admin role required)
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(authSvc), middleware.RequireRole("admin"))
		{
			// Dashboard
			admin.GET("/dashboard", adminH.GetDashboard)

			// Doctors
			admin.GET("/doctors", adminH.ListDoctors)
			admin.POST("/doctors", adminH.CreateDoctor)
			admin.GET("/doctors/:id", adminH.GetDoctor)
			admin.PUT("/doctors/:id", adminH.UpdateDoctor)
			admin.PATCH("/doctors/:id/status", adminH.UpdateDoctorStatus)
			admin.DELETE("/doctors/:id", adminH.DeleteDoctor)

			// Departments
			admin.GET("/departments", adminH.ListDepartments)
			admin.POST("/departments", adminH.CreateDepartment)
			admin.PUT("/departments/:id", adminH.UpdateDepartment)
			admin.PATCH("/departments/:id/status", adminH.UpdateDepartmentStatus)
			admin.DELETE("/departments/:id", adminH.DeleteDepartment)

			// Staff
			admin.GET("/staff", adminH.ListStaff)
			admin.POST("/staff", adminH.CreateStaff)
			admin.GET("/staff/:id", adminH.GetStaff)
			admin.PUT("/staff/:id", adminH.UpdateStaff)
			admin.PATCH("/staff/:id/status", adminH.UpdateStaffStatus)
			admin.DELETE("/staff/:id", adminH.DeleteStaff)

			// Pharmacists
			admin.GET("/pharmacists", adminH.ListPharmacists)
			admin.POST("/pharmacists", adminH.CreatePharmacist)
			admin.PUT("/pharmacists/:id", adminH.UpdatePharmacist)

			// Partner Pharmacies
			admin.GET("/partner-pharmacies", adminH.ListPartnerPharmacies)
			admin.POST("/partner-pharmacies", adminH.CreatePartnerPharmacy)
			admin.PUT("/partner-pharmacies/:id", adminH.UpdatePartnerPharmacy)
			admin.PATCH("/partner-pharmacies/:id/status", adminH.UpdatePartnerPharmacyStatus)

			// Hospital Config
			admin.GET("/config", adminH.GetConfig)
			admin.PUT("/config", adminH.UpdateConfig)

			// Reports
			admin.GET("/reports/financial", adminH.FinancialReport)
			admin.GET("/reports/occupancy", adminH.OccupancyReport)
			admin.GET("/reports/prescriptions", adminH.PrescriptionReport)

			// Export
			admin.POST("/export", adminH.Export)
		}
	}

	log.Println("server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server: %v", err)
	}
}
