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
	doctorRepo := repository.NewDoctorRepo(db)
	pharmacistRepo := repository.NewPharmacistRepo(db)
	pharmacyRepo := repository.NewPharmacyRepo(db)
	staffRepo := repository.NewStaffRepo(db)
	receptionistRepo := repository.NewReceptionistRepo(db)

	// services
	authSvc := service.NewAuthService(userRepo, doctorRepo, pharmacistRepo, staffRepo, auditRepo, passwordResetRepo, cfg)
	receptionistSvc := service.NewReceptionistService(receptionistRepo, auditRepo)
	dashSvc := service.NewDashboardService(dashRepo, userRepo)
	adminSvc := service.NewAdminService(adminRepo, auditRepo, cfg)
	doctorSvc := service.NewDoctorService(doctorRepo, auditRepo)
	pharmacySvc := service.NewPharmacyService(pharmacyRepo)

	// handlers
	authH := handler.NewAuthHandler(authSvc)
	dashH := handler.NewDashboardHandler(dashSvc)
	adminH := handler.NewAdminHandler(adminSvc)
	doctorH := handler.NewDoctorHandler(doctorSvc)
	pharmacyH := handler.NewPharmacyHandler(pharmacySvc)
	receptionistH := handler.NewReceptionistHandler(receptionistSvc)

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

		// Doctor routes
		doc := api.Group("/doctor")
		doc.Use(middleware.JWTAuth(authSvc))
		{
			doc.GET("/dashboard",
				middleware.RequireRole("doctor"),
				doctorH.GetDashboard)

			doc.GET("/appointments",
				middleware.RequireRole("doctor", "admin"),
				doctorH.ListAppointments)
			doc.GET("/appointments/:id",
				middleware.RequireRole("doctor", "admin"),
				doctorH.GetAppointment)
			doc.PATCH("/appointments/:id/status",
				middleware.RequireRole("doctor"),
				doctorH.UpdateAppointmentStatus)

			doc.GET("/patients",
				middleware.RequireRole("doctor", "admin"),
				doctorH.ListPatients)
			doc.GET("/patients/:id",
				middleware.RequireRole("doctor", "admin"),
				doctorH.GetPatientEHR)
			doc.POST("/patients/:id/vitals",
				middleware.RequireRole("doctor", "nurse"),
				doctorH.RecordVital)
			doc.POST("/patients/:id/clinical-notes",
				middleware.RequireRole("doctor"),
				doctorH.CreateClinicalNote)

			doc.GET("/prescriptions",
				middleware.RequireRole("doctor", "admin"),
				doctorH.ListPrescriptions)
			doc.POST("/prescriptions",
				middleware.RequireRole("doctor"),
				doctorH.CreatePrescription)
			doc.GET("/prescriptions/:id",
				middleware.RequireRole("doctor", "admin"),
				doctorH.GetPrescription)

			doc.GET("/lab-orders",
				middleware.RequireRole("doctor"),
				doctorH.ListLabOrders)
			doc.POST("/lab-orders",
				middleware.RequireRole("doctor"),
				doctorH.CreateLabOrder)
			doc.GET("/lab-orders/:id",
				middleware.RequireRole("doctor"),
				doctorH.GetLabOrder)

			doc.GET("/radiology-orders",
				middleware.RequireRole("doctor"),
				doctorH.ListRadiologyOrders)
			doc.POST("/radiology-orders",
				middleware.RequireRole("doctor"),
				doctorH.CreateRadiologyOrder)
			doc.GET("/radiology-orders/:id",
				middleware.RequireRole("doctor"),
				doctorH.GetRadiologyOrder)

			doc.POST("/admissions",
				middleware.RequireRole("doctor"),
				doctorH.CreateAdmission)
			doc.PATCH("/admissions/:id/discharge",
				middleware.RequireRole("doctor"),
				doctorH.DischargeAdmission)
		}

		// Pharmacy routes (JWT + (pharmacist or admin) role required)
		pharmacy := api.Group("/pharmacy")
		pharmacy.Use(middleware.JWTAuth(authSvc), middleware.RequireRole("pharmacist", "admin"))
		{
			// Dashboard
			pharmacy.GET("/dashboard", pharmacyH.GetDashboard)

			// Prescriptions
			pharmacy.GET("/prescriptions", pharmacyH.ListPrescriptions)
			pharmacy.GET("/prescriptions/:id", pharmacyH.GetPrescription)
			pharmacy.PATCH("/prescriptions/:id/dispense", middleware.RequireRole("pharmacist"), pharmacyH.DispensePrescription)
			pharmacy.PATCH("/prescriptions/:id/cancel", pharmacyH.CancelPrescription)

			// Medicines
			pharmacy.GET("/medicines", pharmacyH.ListMedicines)
			pharmacy.POST("/medicines", pharmacyH.CreateMedicine)
			pharmacy.GET("/medicines/:id", pharmacyH.GetMedicine)
			pharmacy.PUT("/medicines/:id", pharmacyH.UpdateMedicine)

			// Medicine Batches
			pharmacy.GET("/medicines/:id/batches", pharmacyH.ListBatches)
			pharmacy.POST("/medicines/:id/batches", middleware.RequireRole("pharmacist"), pharmacyH.CreateBatch)

			// Stock Adjustments
			pharmacy.POST("/stock/adjust", middleware.RequireRole("pharmacist"), pharmacyH.AdjustStock)

			// Alerts
			pharmacy.GET("/alerts/low-stock", pharmacyH.GetLowStockAlerts)
			pharmacy.GET("/alerts/expiring", pharmacyH.GetExpiringAlerts)

			// Partner Pharmacies
			pharmacy.GET("/partner-pharmacies", pharmacyH.ListPartnerPharmacies)

			// Transfer Requests
			pharmacy.POST("/transfer-requests", middleware.RequireRole("pharmacist"), pharmacyH.CreateTransferRequest)
			pharmacy.GET("/transfer-requests", pharmacyH.ListTransferRequests)
			pharmacy.PATCH("/transfer-requests/:id", pharmacyH.UpdateTransferRequest)
		}

		// Receptionist routes under /api/v1
		v1 := api.Group("/v1")
		v1.Use(middleware.JWTAuth(authSvc))
		{
			// Dashboard
			v1.GET("/receptionist/dashboard",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.GetDashboard)

			// Patient Management
			v1.POST("/patient/register",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.RegisterPatient)
			v1.GET("/patient/search",
				middleware.RequireRole("receptionist", "admin", "doctor"),
				receptionistH.SearchPatients)
			v1.GET("/patient/:patient_id",
				middleware.RequireRole("receptionist", "admin", "doctor"),
				receptionistH.GetPatient)
			v1.PATCH("/patient/:patient_id/contact",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.UpdatePatientContact)
			v1.GET("/patient/:patient_id/registration-card",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.GetPatientRegistrationCard)

			// Appointments — walkin MUST be before /:appt_id to avoid Gin routing conflict
			v1.POST("/appointments/walkin",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.WalkIn)
			v1.POST("/appointments",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.BookAppointment)
			v1.GET("/appointments",
				middleware.RequireRole("receptionist", "admin", "doctor"),
				receptionistH.ListAppointments)
			v1.GET("/appointments/:appt_id",
				middleware.RequireRole("receptionist", "admin", "doctor"),
				receptionistH.GetAppointment)
			v1.PATCH("/appointments/:appt_id/reschedule",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.RescheduleAppointment)
			v1.PATCH("/appointments/:appt_id/cancel",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.CancelAppointment)
			v1.POST("/appointments/:appt_id/checkin",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.CheckInAppointment)
			v1.GET("/appointments/:appt_id/slip",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.GetAppointmentSlip)

			// Visitor Log
			v1.POST("/visitors",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.LogVisitor)
			v1.GET("/visitors",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.ListVisitors)
			v1.PATCH("/visitors/:visitor_id/checkout",
				middleware.RequireRole("receptionist", "admin"),
				receptionistH.CheckoutVisitor)
		}
	}

	log.Println("server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server: %v", err)
	}
}
