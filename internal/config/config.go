package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	Environment  string `yaml:"environment"`
	EnableHTTPS  bool   `yaml:"enable_https"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	MVDSB      DatabaseConnection `yaml:"mvdsb"`
	Portal64BDW DatabaseConnection `yaml:"portal64_bdw"`
	Portal64SVW DatabaseConnection `yaml:"portal64_svw"`
}

// DatabaseConnection holds individual database connection details
type DatabaseConnection struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getIntEnv("SERVER_PORT", 8080),
			Host:         getStringEnv("SERVER_HOST", "0.0.0.0"),
			Environment:  getStringEnv("ENVIRONMENT", "development"),
			EnableHTTPS:  getBoolEnv("ENABLE_HTTPS", false),
			CertFile:     getStringEnv("CERT_FILE", ""),
			KeyFile:      getStringEnv("KEY_FILE", ""),
			ReadTimeout:  getIntEnv("READ_TIMEOUT", 10),
			WriteTimeout: getIntEnv("WRITE_TIMEOUT", 10),
		},
		Database: DatabaseConfig{
			MVDSB: DatabaseConnection{
				Host:     getStringEnv("MVDSB_HOST", "localhost"),
				Port:     getIntEnv("MVDSB_PORT", 3306),
				Username: getStringEnv("MVDSB_USERNAME", "root"),
				Password: getStringEnv("MVDSB_PASSWORD", ""),
				Database: getStringEnv("MVDSB_DATABASE", "mvdsb"),
				Charset:  getStringEnv("MVDSB_CHARSET", "utf8mb4"),
			},
			Portal64BDW: DatabaseConnection{
				Host:     getStringEnv("PORTAL64_BDW_HOST", "localhost"),
				Port:     getIntEnv("PORTAL64_BDW_PORT", 3306),
				Username: getStringEnv("PORTAL64_BDW_USERNAME", "root"),
				Password: getStringEnv("PORTAL64_BDW_PASSWORD", ""),
				Database: getStringEnv("PORTAL64_BDW_DATABASE", "portal64_bdw"),
				Charset:  getStringEnv("PORTAL64_BDW_CHARSET", "utf8mb4"),
			},
			Portal64SVW: DatabaseConnection{
				Host:     getStringEnv("PORTAL64_SVW_HOST", "localhost"),
				Port:     getIntEnv("PORTAL64_SVW_PORT", 3306),
				Username: getStringEnv("PORTAL64_SVW_USERNAME", "root"),
				Password: getStringEnv("PORTAL64_SVW_PASSWORD", ""),
				Database: getStringEnv("PORTAL64_SVW_DATABASE", "portal64_svw"),
				Charset:  getStringEnv("PORTAL64_SVW_CHARSET", "utf8mb4"),
			},
		},
	}

	return config, nil
}

// Helper functions to get environment variables with defaults
func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDSN returns a MySQL DSN for the given database connection
func (dc DatabaseConnection) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dc.Username, dc.Password, dc.Host, dc.Port, dc.Database, dc.Charset)
}
