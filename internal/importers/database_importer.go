package importers

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseImporter handles importing SQL dumps into MySQL databases
type DatabaseImporter struct {
	config *config.ImportDBConfig
	dbConfig *config.DatabaseConfig
	logger *log.Logger
}

// NewDatabaseImporter creates a new database importer instance
func NewDatabaseImporter(config *config.ImportDBConfig, dbConfig *config.DatabaseConfig, logger *log.Logger) *DatabaseImporter {
	return &DatabaseImporter{
		config:   config,
		dbConfig: dbConfig,
		logger:   logger,
	}
}

// ImportDatabases imports SQL dumps for all target databases
func (di *DatabaseImporter) ImportDatabases(dumpFiles map[string]string) error {
	di.logger.Printf("Starting database import for %d databases", len(dumpFiles))

	for _, targetDB := range di.config.TargetDatabases {
		dumpPath, exists := dumpFiles[targetDB.Name]
		if !exists {
			di.logger.Printf("Warning: No dump file found for database %s", targetDB.Name)
			continue
		}

		if err := di.ImportDatabase(targetDB.Name, dumpPath); err != nil {
			return fmt.Errorf("failed to import database %s: %w", targetDB.Name, err)
		}
	}

	di.logger.Printf("Successfully imported all databases")
	return nil
}

// ImportDatabase imports a single database from SQL dump
func (di *DatabaseImporter) ImportDatabase(dbName, dumpPath string) error {
	di.logger.Printf("Importing database %s from %s", dbName, dumpPath)
	
	start := time.Now()
	
	// Get database connection info
	var dbConn config.DatabaseConnection
	switch dbName {
	case "mvdsb":
		dbConn = di.dbConfig.MVDSB
	case "portal64_bdw":
		dbConn = di.dbConfig.Portal64BDW
	default:
		return fmt.Errorf("unknown database: %s", dbName)
	}

	// Connect to MySQL server (not specific database)
	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=True&loc=Local&timeout=%s",
		dbConn.Username, dbConn.Password, dbConn.Host, dbConn.Port, di.config.ImportTimeout)
	
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	// Drop existing database
	di.logger.Printf("Dropping existing database %s", dbName)
	if err := di.dropDatabase(db, dbName); err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	// Create new database
	di.logger.Printf("Creating database %s", dbName)
	if err := di.createDatabase(db, dbName, dbConn.Charset); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Import dump
	di.logger.Printf("Importing SQL dump into %s", dbName)
	if err := di.importSQLDump(dbConn, dumpPath); err != nil {
		return fmt.Errorf("failed to import SQL dump: %w", err)
	}

	// Verify import
	if err := di.verifyImport(dbConn, dbName); err != nil {
		return fmt.Errorf("failed to verify import: %w", err)
	}

	duration := time.Since(start)
	di.logger.Printf("Successfully imported database %s in %s", dbName, duration)
	
	return nil
}

// dropDatabase drops a database if it exists
func (di *DatabaseImporter) dropDatabase(db *sql.DB, dbName string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
	_, err := db.Exec(query)
	return err
}

// createDatabase creates a new database with specified charset
func (di *DatabaseImporter) createDatabase(db *sql.DB, dbName, charset string) error {
	query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s_general_ci", 
		dbName, charset, charset)
	_, err := db.Exec(query)
	return err
}

// importSQLDump imports SQL dump using command line mysql client
func (di *DatabaseImporter) importSQLDump(dbConn config.DatabaseConnection, dumpPath string) error {
	// Verify dump file exists and is readable
	if err := di.validateDumpFile(dumpPath); err != nil {
		return fmt.Errorf("dump file validation failed: %w", err)
	}

	// For now, use programmatic import instead of command line
	// This is more portable and doesn't require mysql client
	return di.importSQLDumpProgrammatically(dbConn, dumpPath)
}

// importSQLDumpProgrammatically imports SQL dump programmatically
func (di *DatabaseImporter) importSQLDumpProgrammatically(dbConn config.DatabaseConnection, dumpPath string) error {
	// Connect to the specific database
	dbDSN := dbConn.GetDSN()
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Set timeout
	db.SetConnMaxLifetime(di.config.ImportTimeout)
	
	// Open and read SQL dump file
	file, err := os.Open(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to open dump file: %w", err)
	}
	defer file.Close()

	// Process SQL commands
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB buffer for large lines
	
	var sqlCommand strings.Builder
	var commandCount int
	var errorCount int

	di.logger.Printf("Processing SQL dump file...")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Add line to current command
		sqlCommand.WriteString(line)
		sqlCommand.WriteString(" ")

		// Check if command is complete (ends with semicolon)
		if strings.HasSuffix(line, ";") {
			command := strings.TrimSpace(sqlCommand.String())
			if len(command) > 0 {
				if err := di.executeSQL(db, command); err != nil {
					errorCount++
					di.logger.Printf("Warning: SQL command failed: %v", err)
					// Continue with next command - some errors might be acceptable
				}
				commandCount++

				// Report progress
				if commandCount%1000 == 0 {
					di.logger.Printf("Processed %d SQL commands (%d errors)", commandCount, errorCount)
				}
			}
			sqlCommand.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dump file: %w", err)
	}

	di.logger.Printf("Completed SQL import: %d commands processed, %d errors", commandCount, errorCount)
	
	// If too many errors, consider it a failure
	if errorCount > commandCount/10 { // More than 10% errors
		return fmt.Errorf("too many SQL errors during import: %d out of %d", errorCount, commandCount)
	}

	return nil
}

// executeSQL executes a single SQL command
func (di *DatabaseImporter) executeSQL(db *sql.DB, command string) error {
	// Skip certain commands that might cause issues
	upperCommand := strings.ToUpper(strings.TrimSpace(command))
	
	// Skip some MySQL-specific commands that might not be needed
	skipCommands := []string{
		"SET NAMES",
		"SET CHARACTER_SET_CLIENT",
		"SET CHARACTER_SET_RESULTS",
		"SET COLLATION_CONNECTION",
		"SET SQL_MODE",
		"SET FOREIGN_KEY_CHECKS",
		"SET UNIQUE_CHECKS",
		"SET AUTOCOMMIT",
		"LOCK TABLES",
		"UNLOCK TABLES",
	}

	for _, skip := range skipCommands {
		if strings.HasPrefix(upperCommand, skip) {
			return nil // Skip silently
		}
	}

	// Execute the command
	_, err := db.Exec(command)
	return err
}

// validateDumpFile validates that dump file exists and is readable
func (di *DatabaseImporter) validateDumpFile(dumpPath string) error {
	// Check if file exists
	stat, err := os.Stat(dumpPath)
	if err != nil {
		return fmt.Errorf("dump file not found: %w", err)
	}

	// Check if file is not empty
	if stat.Size() == 0 {
		return fmt.Errorf("dump file is empty")
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filepath.Ext(dumpPath)), ".sql") {
		return fmt.Errorf("dump file does not have .sql extension")
	}

	// Try to open and read first few lines
	file, err := os.Open(dumpPath)
	if err != nil {
		return fmt.Errorf("cannot open dump file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hasContent := false
	lineCount := 0

	for scanner.Scan() && lineCount < 10 {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && !strings.HasPrefix(line, "--") {
			hasContent = true
			break
		}
		lineCount++
	}

	if !hasContent {
		return fmt.Errorf("dump file appears to contain no SQL content")
	}

	di.logger.Printf("Dump file validation passed: %s (%d bytes)", dumpPath, stat.Size())
	return nil
}

// verifyImport verifies that the import was successful
func (di *DatabaseImporter) verifyImport(dbConn config.DatabaseConnection, dbName string) error {
	db, err := sql.Open("mysql", dbConn.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to connect for verification: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}

	// Count tables
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?", dbName).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("failed to count tables: %w", err)
	}

	if tableCount == 0 {
		return fmt.Errorf("no tables found in imported database")
	}

	// Get sample data from first few tables
	rows, err := db.Query(`
		SELECT table_name, table_rows 
		FROM information_schema.tables 
		WHERE table_schema = ? 
		AND table_type = 'BASE TABLE' 
		ORDER BY table_name 
		LIMIT 5`, dbName)
	if err != nil {
		return fmt.Errorf("failed to get table information: %w", err)
	}
	defer rows.Close()

	di.logger.Printf("Import verification for %s:", dbName)
	di.logger.Printf("Total tables: %d", tableCount)

	for rows.Next() {
		var tableName string
		var tableRows sql.NullInt64
		
		if err := rows.Scan(&tableName, &tableRows); err != nil {
			continue
		}

		rowCount := "unknown"
		if tableRows.Valid {
			rowCount = fmt.Sprintf("%d", tableRows.Int64)
		}

		di.logger.Printf("  Table %s: %s rows", tableName, rowCount)
	}

	di.logger.Printf("Import verification completed successfully")
	return nil
}

// GetImportStats returns statistics about the import operation
func (di *DatabaseImporter) GetImportStats(dbName string) (*ImportStats, error) {
	var dbConn config.DatabaseConnection
	switch dbName {
	case "mvdsb":
		dbConn = di.dbConfig.MVDSB
	case "portal64_bdw":
		dbConn = di.dbConfig.Portal64BDW
	default:
		return nil, fmt.Errorf("unknown database: %s", dbName)
	}

	db, err := sql.Open("mysql", dbConn.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	stats := &ImportStats{
		DatabaseName: dbName,
	}

	// Get table count
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?", dbName).Scan(&stats.TableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}

	// Get total row count (approximate)
	rows, err := db.Query(`
		SELECT SUM(table_rows) 
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_type = 'BASE TABLE'`, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get row count: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var totalRows sql.NullInt64
		if err := rows.Scan(&totalRows); err == nil && totalRows.Valid {
			stats.TotalRows = totalRows.Int64
		}
	}

	// Get database size
	err = db.QueryRow(`
		SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as size_mb
		FROM information_schema.tables 
		WHERE table_schema = ?`, dbName).Scan(&stats.SizeMB)
	if err != nil {
		// Size query might fail, but that's okay
		stats.SizeMB = 0
	}

	return stats, nil
}

// ImportStats contains statistics about an imported database
type ImportStats struct {
	DatabaseName string  `json:"database_name"`
	TableCount   int     `json:"table_count"`
	TotalRows    int64   `json:"total_rows"`
	SizeMB       float64 `json:"size_mb"`
}
