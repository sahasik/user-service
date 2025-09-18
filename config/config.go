package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	GinMode     string
	ServiceName string
	Version     string

	Database    DatabaseConfig
	Redis       RedisConfig
	AuthService AuthServiceConfig
	Upload      UploadConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type AuthServiceConfig struct {
	URL     string
	Timeout time.Duration
}

type UploadConfig struct {
	Path             string
	MaxSize          int64
	AllowedFileTypes []string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "1"))
	maxUploadSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "5242880"), 10, 64)

	// Parse auth service timeout
	timeoutStr := getEnv("AUTH_SERVICE_TIMEOUT", "5s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 5 * time.Second
	}

	return &Config{
		Port:        getEnv("PORT", "8081"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		ServiceName: getEnv("SERVICE_NAME", "user-service"),
		Version:     getEnv("SERVICE_VERSION", "1.0.0"),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "pesantren_users"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},

		AuthService: AuthServiceConfig{
			URL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8080"),
			Timeout: timeout,
		},

		Upload: UploadConfig{
			Path:             getEnv("UPLOAD_PATH", "./uploads"),
			MaxSize:          maxUploadSize,
			AllowedFileTypes: []string{"jpg", "jpeg", "png", "pdf", "doc", "docx"},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
