package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port         string
	DatabaseURL  string
	AllowOrigins string
}

// Load reads configuration from .env file and environment variables
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "rogerbass"),
		getEnv("DB_SSLMODE", "disable"),
		getEnv("DB_TIMEZONE", "UTC"),
	)

	return &Config{
		Port:         getEnv("PORT", "3000"),
		DatabaseURL:  dsn,
		AllowOrigins: getEnv("ALLOW_ORIGINS", "http://localhost:4200"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
