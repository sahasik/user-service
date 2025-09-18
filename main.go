// ================================================================
// user-service/main.go - Complete User Management with GORM
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"gitlab.com/nodiviti/user-service/config"
	"gitlab.com/nodiviti/user-service/database"
	"gitlab.com/nodiviti/user-service/handlers"
	"gitlab.com/nodiviti/user-service/middleware"
	"gitlab.com/nodiviti/user-service/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database with GORM
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto-migrations (single users table)
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed initial admin user
	if err := database.SeedData(); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.Upload.Path, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("🛑 Shutting down user service...")
		database.Close()
		os.Exit(0)
	}()

	// Initialize services
	userService := services.NewUserService()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(cfg, userService)

	// Setup routes
	router := setupRoutes(userHandler, cfg)

	// Start server
	log.Printf("🚀 User Service starting on port %s", cfg.Port)
	log.Println("📊 Database: PostgreSQL with GORM (Single users table)")
	log.Println("📋 Features: Complete user management for all roles")
	log.Println("👤 Initial Admin: admin@pesantren.com / Admin123!@#")
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(userHandler *handlers.UserHandler, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	router.Static("/files", cfg.Upload.Path)

	// Health check
	router.GET("/health", userHandler.HealthCheck)

	// API routes
	api := router.Group("/api/v1")

	// Protected routes (require authentication)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// My profile routes (all authenticated users)
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetMyProfile)
			users.PUT("/me", userHandler.UpdateMyProfile)
			users.POST("/me/photo", userHandler.UploadProfilePhoto)
		}

		// Admin/Teacher routes
		adminTeacher := protected.Group("/")
		adminTeacher.Use(middleware.TeacherOnly())
		{
			adminTeacher.GET("/users/:id", userHandler.GetUserByID)
			adminTeacher.GET("/teachers", userHandler.GetTeachers)
			adminTeacher.GET("/students", userHandler.GetStudents)
			adminTeacher.GET("/students/class/:class", userHandler.GetStudentsByClass)
			adminTeacher.GET("/classes", userHandler.GetClassList)
		}

		// Admin only routes
		admin := protected.Group("/")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/users", userHandler.GetAllUsers)
			admin.POST("/users", userHandler.CreateUser) // Admin creates teachers/students
			admin.PUT("/users/:id", userHandler.UpdateUser)
			admin.DELETE("/users/:id", userHandler.DeactivateUser)
			admin.GET("/users/stats", userHandler.GetUserStats)
			admin.GET("/search/users", userHandler.SearchUsers)
		}
	}

	return router
}
