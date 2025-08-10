package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Cache    CacheConfig
	Import   ImportConfig
	Logging  LoggingConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int
	Host         string
	Environment  string
	EnableHTTPS  bool
	CertFile     string
	KeyFile      string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	MVDSB      DatabaseConnection
	Portal64BDW DatabaseConnection
}
// DatabaseConnection holds individual database connection details
type DatabaseConnection struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Charset  string
}

// CacheConfig holds Redis cache configuration
type CacheConfig struct {
	Enabled          bool
	Address          string           
	Password         string          
	Database         int          
	MaxRetries       int       
	PoolSize         int         
	MinIdleConns     int    
	DialTimeout      time.Duration      
	ReadTimeout      time.Duration      
	WriteTimeout     time.Duration     
	PoolTimeout      time.Duration      
	IdleTimeout      time.Duration      
	MaxConnAge       time.Duration      
	
	// Background refresh settings
	RefreshThreshold float64 
	RefreshWorkers   int           
}

// ImportConfig holds SCP import configuration
type ImportConfig struct {
	Enabled   bool
	Schedule  string   // cron format
	LoadCheck LoadCheckConfig
	SCP       SCPConfig
	ZIP       ZIPConfig
	Storage   StorageConfig
	Freshness FreshnessConfig
	Database  ImportDBConfig
	Retry     RetryConfig
}

// LoadCheckConfig holds load-based delay configuration
type LoadCheckConfig struct {
	Enabled       bool
	DelayDuration time.Duration
	MaxDelays     int
	LoadThreshold int
}

// SCPConfig holds SCP connection configuration
type SCPConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	RemotePath   string
	FilePatterns []string
	Timeout      time.Duration
}

// ZIPConfig holds ZIP extraction configuration
type ZIPConfig struct {
	PasswordMVDSB     string
	PasswordPortal64  string
	ExtractTimeout    time.Duration
}

// StorageConfig holds local storage configuration
type StorageConfig struct {
	TempDir          string
	MetadataFile     string
	CleanupOnSuccess bool
	KeepFailedFiles  bool
}

// FreshnessConfig holds file freshness checking configuration
type FreshnessConfig struct {
	Enabled         bool
	CompareTimestamp bool
	CompareSize     bool
	CompareChecksum bool
	SkipIfNotNewer  bool
}

// ImportDBConfig holds database import configuration
type ImportDBConfig struct {
	ImportTimeout    time.Duration
	TargetDatabases  []TargetDatabase
}

// TargetDatabase holds individual target database configuration
type TargetDatabase struct {
	Name        string
	FilePattern string
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	Enabled     bool
	MaxAttempts int
	RetryDelay  time.Duration
	FailFast    bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Enabled        bool
	Level          string
	MainLogFile    string
	ImportLogFile  string
	MaxSizeMB      int
	MaxBackups     int
	MaxAgeDays     int
	Compress       bool
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file with multiple approaches
	loadEnvFile()
	
	return loadConfig(), nil
}

// loadEnvFile attempts to load .env file with fallback methods
func loadEnvFile() {
	envFiles := []string{
		".env",
		"/opt/portal64api/.env",
	}
	
	var loaded bool
	
	// Try godotenv first
	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			if err := godotenv.Load(envFile); err == nil {
				log.Printf("Successfully loaded .env file using godotenv: %s", envFile)
				loaded = true
				break
			} else {
				log.Printf("godotenv failed to load %s: %v", envFile, err)
			}
		}
	}
	
	// If godotenv failed, try manual parsing
	if !loaded {
		for _, envFile := range envFiles {
			if manualLoadEnvFile(envFile) {
				log.Printf("Successfully loaded .env file manually: %s", envFile)
				loaded = true
				break
			}
		}
	}
	
	if !loaded {
		log.Println("No .env file found or could be loaded, using system environment variables only")
	}
}

// manualLoadEnvFile manually parses and loads environment variables from .env file
func manualLoadEnvFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}
		
		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	
	return scanner.Err() == nil
}

// loadConfig creates configuration from environment variables
func loadConfig() *Config {
	return &Config{
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
		},
		Cache: CacheConfig{
			Enabled:          getBoolEnv("CACHE_ENABLED", true),
			Address:          getStringEnv("CACHE_ADDRESS", "localhost:6379"),
			Password:         getStringEnv("CACHE_PASSWORD", ""),
			Database:         getIntEnv("CACHE_DATABASE", 0),
			MaxRetries:       getIntEnv("CACHE_MAX_RETRIES", 3),
			PoolSize:         getIntEnv("CACHE_POOL_SIZE", 10),
			MinIdleConns:     getIntEnv("CACHE_MIN_IDLE_CONNS", 5),
			DialTimeout:      getDurationEnv("CACHE_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:      getDurationEnv("CACHE_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:     getDurationEnv("CACHE_WRITE_TIMEOUT", 3*time.Second),
			PoolTimeout:      getDurationEnv("CACHE_POOL_TIMEOUT", 4*time.Second),
			IdleTimeout:      getDurationEnv("CACHE_IDLE_TIMEOUT", 5*time.Minute),
			MaxConnAge:       getDurationEnv("CACHE_MAX_CONN_AGE", 0),
			RefreshThreshold: getFloat64Env("CACHE_REFRESH_THRESHOLD", 0.8),
			RefreshWorkers:   getIntEnv("CACHE_REFRESH_WORKERS", 5),
		},
		Import: ImportConfig{
			Enabled:  getBoolEnv("IMPORT_ENABLED", false),
			Schedule: getStringEnv("IMPORT_SCHEDULE", "0 2 * * *"),
			LoadCheck: LoadCheckConfig{
				Enabled:       getBoolEnv("IMPORT_LOAD_CHECK_ENABLED", true),
				DelayDuration: getDurationEnv("IMPORT_LOAD_CHECK_DELAY", 1*time.Hour),
				MaxDelays:     getIntEnv("IMPORT_LOAD_CHECK_MAX_DELAYS", 3),
				LoadThreshold: getIntEnv("IMPORT_LOAD_CHECK_THRESHOLD", 100),
			},
			SCP: SCPConfig{
				Host:         getStringEnv("IMPORT_SCP_HOST", "portal.svw.info"),
				Port:         getIntEnv("IMPORT_SCP_PORT", 22),
				Username:     getStringEnv("IMPORT_SCP_USERNAME", "portal64user"),
				Password:     getStringEnv("IMPORT_SCP_PASSWORD", ""),
				RemotePath:   getStringEnv("IMPORT_SCP_REMOTE_PATH", "/data/exports/"),
				FilePatterns: getStringSliceEnv("IMPORT_SCP_FILE_PATTERNS", []string{"mvdsb_*.zip", "portal64_bdw_*.zip"}),
				Timeout:      getDurationEnv("IMPORT_SCP_TIMEOUT", 300*time.Second),
			},
			ZIP: ZIPConfig{
				PasswordMVDSB:    getStringEnv("IMPORT_ZIP_PASSWORD_MVDSB", ""),
				PasswordPortal64: getStringEnv("IMPORT_ZIP_PASSWORD_PORTAL64_BDW", ""),
				ExtractTimeout:   getDurationEnv("IMPORT_ZIP_EXTRACT_TIMEOUT", 60*time.Second),
			},
			Storage: StorageConfig{
				TempDir:          getStringEnv("IMPORT_TEMP_DIR", "./data/import/temp"),
				MetadataFile:     getStringEnv("IMPORT_METADATA_FILE", "./data/import/last_import.json"),
				CleanupOnSuccess: getBoolEnv("IMPORT_CLEANUP_ON_SUCCESS", true),
				KeepFailedFiles:  getBoolEnv("IMPORT_KEEP_FAILED_FILES", true),
			},
			Freshness: FreshnessConfig{
				Enabled:         getBoolEnv("IMPORT_FRESHNESS_ENABLED", true),
				CompareTimestamp: getBoolEnv("IMPORT_FRESHNESS_COMPARE_TIMESTAMP", true),
				CompareSize:     getBoolEnv("IMPORT_FRESHNESS_COMPARE_SIZE", true),
				CompareChecksum: getBoolEnv("IMPORT_FRESHNESS_COMPARE_CHECKSUM", false),
				SkipIfNotNewer:  getBoolEnv("IMPORT_FRESHNESS_SKIP_IF_NOT_NEWER", true),
			},
			Database: ImportDBConfig{
				ImportTimeout: getDurationEnv("IMPORT_DATABASE_TIMEOUT", 600*time.Second),
				TargetDatabases: []TargetDatabase{
					{Name: "mvdsb", FilePattern: "mvdsb_*"},
					{Name: "portal64_bdw", FilePattern: "portal64_bdw_*"},
				},
			},
			Retry: RetryConfig{
				Enabled:     getBoolEnv("IMPORT_RETRY_ENABLED", true),
				MaxAttempts: getIntEnv("IMPORT_RETRY_MAX_ATTEMPTS", 2),
				RetryDelay:  getDurationEnv("IMPORT_RETRY_DELAY", 5*time.Minute),
				FailFast:    getBoolEnv("IMPORT_RETRY_FAIL_FAST", true),
			},
		},
		Logging: LoggingConfig{
			Enabled:       getBoolEnv("LOG_ENABLED", true),
			Level:         getStringEnv("LOG_LEVEL", "info"),
			MainLogFile:   getStringEnv("LOG_FILE_PATH", "./logs/portal64api.log"),
			ImportLogFile: getStringEnv("LOG_IMPORT_FILE_PATH", "./logs/portal64api-import.log"),
			MaxSizeMB:     getIntEnv("LOG_MAX_SIZE_MB", 100),
			MaxBackups:    getIntEnv("LOG_MAX_BACKUPS", 5),
			MaxAgeDays:    getIntEnv("LOG_MAX_AGE_DAYS", 30),
			Compress:      getBoolEnv("LOG_COMPRESS", true),
		},
	}
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

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getFloat64Env(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
