package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gitlab.com/nodiviti/user-service/config"
	"gitlab.com/nodiviti/user-service/models"
	"gitlab.com/nodiviti/user-service/utils"
)

var (
	DB *gorm.DB
)

func InitDatabase(cfg *config.Config) error {
	// PostgreSQL connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// GORM config
	var gormLogger logger.Interface
	if cfg.GinMode == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info) // Show SQL queries in debug mode
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().In(time.FixedZone("WIB", 7*3600)) // UTC+7 for Indonesia
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL with GORM (User Service)")
	return nil
}

// AutoMigrate runs database migrations for single users table
func AutoMigrate() error {
	log.Println("🔄 Running GORM auto-migrations for users table...")

	// Auto migrate the single users table
	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate users table: %v", err)
	}

	// Create additional indexes for performance
	err = createAdditionalIndexes()
	if err != nil {
		return fmt.Errorf("failed to create additional indexes: %v", err)
	}

	log.Println("✅ GORM auto-migration completed successfully (single users table)")
	return nil
}

// createAdditionalIndexes creates additional database indexes for performance
func createAdditionalIndexes() error {
	// Composite indexes for common queries on single users table
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_role_active ON users(role, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_role_status ON users(role, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_class_year ON users(class_level, academic_year) WHERE role = 'student' AND deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_specialization ON users(specialization) WHERE role = 'teacher' AND deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_employee_id ON users(employee_id) WHERE employee_id IS NOT NULL AND deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_student_id ON users(student_id) WHERE student_id IS NOT NULL AND deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_full_name ON users(full_name) WHERE full_name IS NOT NULL AND deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at) WHERE deleted_at IS NULL",
	}

	for _, indexSQL := range indexes {
		result := DB.Exec(indexSQL)
		if result.Error != nil {
			log.Printf("Warning: Failed to create index: %v", result.Error)
			// Continue with other indexes
		}
	}

	log.Println("✅ Additional indexes created for single users table")
	return nil
}

// SeedData creates initial admin user
func SeedData() error {
	log.Println("🌱 Checking for seed data...")

	// Check if we have any users
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)

	if userCount == 0 {
		log.Println("🌱 Creating initial admin user...")

		// Hash default password
		hashedPassword, err := utils.HashPassword("Admin123!@#")
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %v", err)
		}

		// Create default admin user
		adminUser := models.User{
			Username:     "admin",
			Email:        "admin@pesantren.com",
			PasswordHash: hashedPassword,
			Role:         "admin",
			IsActive:     true,
			FullName:     stringPtr("System Administrator"),
		}

		result := DB.Create(&adminUser)
		if result.Error != nil {
			return fmt.Errorf("failed to create admin user: %v", result.Error)
		}

		log.Printf("✅ Admin user created with ID: %d", adminUser.ID)
		log.Println("   📧 Email: admin@pesantren.com")
		log.Println("   👤 Username: admin")
		log.Println("   🔑 Password: Admin123!@# (please change this!)")
	} else {
		log.Printf("⏭️  Found %d existing users, skipping seed data", userCount)
	}

	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// GetDB returns the GORM database instance
func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}

		err = sqlDB.Close()
		if err != nil {
			return err
		}

		log.Println("📦 Database connection closed (User Service)")
	}
	return nil
}

// HealthCheck checks database connectivity
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	return sqlDB.Ping()
}
