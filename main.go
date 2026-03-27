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

	// services
	authSvc := service.NewAuthService(userRepo, auditRepo, passwordResetRepo, cfg)
	dashSvc := service.NewDashboardService(dashRepo, userRepo)

	// handlers
	authH := handler.NewAuthHandler(authSvc)
	dashH := handler.NewDashboardHandler(dashSvc)

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
	}

	log.Println("server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server: %v", err)
	}
}
