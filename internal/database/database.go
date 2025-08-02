package database

import (
	"fmt"
	"log"
	"time"

	"portal64api/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Databases holds all database connections
type Databases struct {
	MVDSB       *gorm.DB
	Portal64BDW *gorm.DB
}

// Connect establishes connections to all databases
func Connect(cfg *config.Config) (*Databases, error) {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.Server.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Connect to MVDSB database
	mvdsb, err := connectToDatabase(cfg.Database.MVDSB, "MVDSB", gormLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MVDSB database: %w", err)
	}

	// Connect to Portal64_BDW database
	portal64BDW, err := connectToDatabase(cfg.Database.Portal64BDW, "Portal64_BDW", gormLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Portal64_BDW database: %w", err)
	}

	log.Println("Successfully connected to all databases")

	return &Databases{
		MVDSB:       mvdsb,
		Portal64BDW: portal64BDW,
	}, nil
}

// connectToDatabase establishes a connection to a single database
func connectToDatabase(dbConfig config.DatabaseConnection, dbName string, gormLogger logger.Interface) (*gorm.DB, error) {
	dsn := dbConfig.GetDSN()
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", dbName, err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB for %s: %w", dbName, err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping %s database: %w", dbName, err)
	}

	log.Printf("Successfully connected to %s database", dbName)
	return db, nil
}

// Close closes all database connections
func (dbs *Databases) Close() error {
	var errors []error

	if dbs.MVDSB != nil {
		if sqlDB, err := dbs.MVDSB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close MVDSB connection: %w", err))
			}
		}
	}

	if dbs.Portal64BDW != nil {
		if sqlDB, err := dbs.Portal64BDW.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close Portal64_BDW connection: %w", err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	log.Println("All database connections closed successfully")
	return nil
}
