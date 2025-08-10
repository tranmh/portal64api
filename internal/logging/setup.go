package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelMap = map[string]LogLevel{
	"debug": DEBUG,
	"info":  INFO,
	"warn":  WARN,
	"error": ERROR,
	"fatal": FATAL,
}

// Logger wraps the standard logger with level filtering
type Logger struct {
	*log.Logger
	level LogLevel
}

// NewLogger creates a new logger with the specified output and level
func NewLogger(output io.Writer, level LogLevel, prefix string) *Logger {
	return &Logger{
		Logger: log.New(output, prefix, log.LstdFlags|log.Lmicroseconds),
		level:  level,
	}
}

// Debug logs debug level messages
func (l *Logger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.Print("[DEBUG] ", fmt.Sprint(v...))
	}
}

// Debugf logs debug level formatted messages
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs info level messages
func (l *Logger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.Print("[INFO] ", fmt.Sprint(v...))
	}
}

// Infof logs info level formatted messages
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning level messages
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.Print("[WARN] ", fmt.Sprint(v...))
	}
}

// Warnf logs warning level formatted messages
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.Printf("[WARN] "+format, v...)
	}
}

// Error logs error level messages
func (l *Logger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.Print("[ERROR] ", fmt.Sprint(v...))
	}
}

// Errorf logs error level formatted messages
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.Printf("[ERROR] "+format, v...)
	}
}

// Fatal logs fatal level messages and exits
func (l *Logger) Fatal(v ...interface{}) {
	if l.level <= FATAL {
		l.Print("[FATAL] ", fmt.Sprint(v...))
		os.Exit(1)
	}
}

// Fatalf logs fatal level formatted messages and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.level <= FATAL {
		l.Printf("[FATAL] "+format, v...)
		os.Exit(1)
	}
}

// SetupLogging initializes the logging system based on configuration
func SetupLogging(cfg *config.LoggingConfig) error {
	if !cfg.Enabled {
		// Disable logging by setting output to discard
		log.SetOutput(io.Discard)
		return nil
	}

	// Create logs directory if it doesn't exist
	if err := ensureLogDir(cfg.MainLogFile); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}
	if err := ensureLogDir(cfg.ImportLogFile); err != nil {
		return fmt.Errorf("failed to create import logs directory: %w", err)
	}

	// Setup main application logger
	mainWriter := createRotatingWriter(cfg.MainLogFile, cfg.MaxSizeMB, cfg.MaxBackups, cfg.MaxAgeDays, cfg.Compress)
	
	// Combine with stdout for development
	var output io.Writer
	if os.Getenv("ENVIRONMENT") == "development" {
		output = io.MultiWriter(os.Stdout, mainWriter)
	} else {
		output = mainWriter
	}
	
	// Set the standard logger output
	log.SetOutput(output)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	return nil
}

// CreateImportLogger creates a separate logger for import service
func CreateImportLogger(cfg *config.LoggingConfig) (*log.Logger, error) {
	if !cfg.Enabled {
		return log.New(io.Discard, "", 0), nil
	}

	// Create import log directory if it doesn't exist
	if err := ensureLogDir(cfg.ImportLogFile); err != nil {
		return nil, fmt.Errorf("failed to create import logs directory: %w", err)
	}

	// Setup import service logger with rotation
	importWriter := createRotatingWriter(cfg.ImportLogFile, cfg.MaxSizeMB, cfg.MaxBackups, cfg.MaxAgeDays, cfg.Compress)
	
	// Combine with stdout for development
	var output io.Writer
	if os.Getenv("ENVIRONMENT") == "development" {
		output = io.MultiWriter(os.Stdout, importWriter)
	} else {
		output = importWriter
	}

	importLogger := log.New(output, "[IMPORT] ", log.LstdFlags|log.Lmicroseconds)

	return importLogger, nil
}

// createRotatingWriter creates a lumberjack rotating file writer
func createRotatingWriter(filename string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
		LocalTime:  true,
	}
}

// ensureLogDir creates the directory for log file if it doesn't exist
func ensureLogDir(logFile string) error {
	dir := filepath.Dir(logFile)
	if dir != "." && dir != "" {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// GetLogLevel returns the log level from string
func GetLogLevel(levelStr string) LogLevel {
	if level, exists := levelMap[strings.ToLower(levelStr)]; exists {
		return level
	}
	return INFO
}
