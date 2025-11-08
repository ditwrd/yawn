/*
Copyright © 2025 Aditya Wardianto <hi@ditwrd.dev>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

// Package database provides database connection and configuration management
// for the YAWN application using GORM ORM.
//
// This package handles the setup and configuration of database connections
// with support for multiple database types and connection pooling. It uses
// GORM as the ORM (Object-Relational Mapping) layer to provide a clean
// interface for database operations.
//
// Supported databases:
//   - PostgreSQL (recommended for production)
//   - SQLite (recommended for development and testing)
//
// Features:
//   - Automatic database connection based on configuration
//   - Connection pooling optimization
//   - Configurable logging levels
//   - UTC timestamp handling
//   - Prepared statement caching
//   - Connection lifetime management
//   - Proper error handling and translation
//
// Configuration:
// Database settings are configured through the config package and include:
// - Database type (postgres, sqlite)
// - Connection parameters (host, port, name, user, password)
// - SSL mode for PostgreSQL
// - File path for SQLite
//
// Example usage:
//
//	cfg, err := config.LoadConfig("")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	db, err := database.NewDatabase(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use db for GORM operations
//	var users []models.User
//	result := db.Find(&users)
//
// Connection pooling:
// The package configures optimal connection pool settings:
// - Max idle connections: 10
// - Max open connections: 100
// - Connection max lifetime: 1 hour
// - Connection max idle time: 5 minutes
//
// Error handling:
// All database errors are wrapped with descriptive context to help
// with debugging and monitoring. Connection errors are clearly distinguished
// from query errors.
package database

import (
	"fmt"
	"time"

	"github.com/ditwrd/yawn/api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection based on the provided configuration.
//
// This function initializes a GORM database connection with the specified database type
// and configuration settings. It automatically configures connection pooling,
// logging, and other optimization settings based on best practices.
//
// Supported database types:
//   - "postgres" or "postgresql": PostgreSQL database
//   - "sqlite": SQLite database (file-based)
//
// Parameters:
//   - cfg: Application configuration containing database settings
//
// Returns:
//   - *gorm.DB: Configured GORM database instance ready for use
//   - error: Any error encountered during connection setup
//
// Database configuration:
// The function uses the following configuration fields:
//   - cfg.Database.Type: Database type ("postgres", "postgresql", "sqlite")
//   - cfg.Database.GetDSN(): Database connection string generated from config
//   - cfg.Logger.Level: Logging level for GORM operations
//
// GORM configuration:
// The database is configured with the following settings:
//   - Logger: Configurable log level based on application logger settings
//   - NowFunc: UTC time function for consistent timestamps
//   - SkipDefaultTransaction: false (transactions enabled)
//   - PrepareStmt: true (prepared statement caching enabled)
//   - PrepareStmtMaxSize: 100 (maximum cached prepared statements)
//   - PrepareStmtTTL: 1 hour (prepared statement cache TTL)
//   - TranslateError: true (error translation enabled)
//
// Connection pool settings:
//   - MaxIdleConns: 10 (maximum idle connections)
//   - MaxOpenConns: 100 (maximum open connections)
//   - ConnMaxLifetime: 1 hour (maximum connection lifetime)
//   - ConnMaxIdleTime: 5 minutes (maximum idle time for connections)
//
// Example usage:
//
//	cfg, err := config.LoadConfig("")
//	if err != nil {
//		return nil, fmt.Errorf("failed to load config: %w", err)
//	}
//
//	db, err := NewDatabase(cfg)
//	if err != nil {
//		return nil, fmt.Errorf("failed to connect to database: %w", err)
//	}
//
//	// Test the connection
//	sqlDB, err := db.DB()
//	if err != nil {
//		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
//	}
//
//	if err := sqlDB.Ping(); err != nil {
//		return nil, fmt.Errorf("failed to ping database: %w", err)
//	}
//
//	fmt.Println("Database connection established successfully")
//
// Error handling:
//   - Returns descriptive errors for connection failures
//   - Validates database type before attempting connection
//   - Wraps GORM errors with additional context
//   - Returns errors for unsupported database types
//
// Migration note:
// After creating the database connection, you should typically run
// database migrations to ensure all tables are created:
//
//	err = db.AutoMigrate(&models.User{}, &models.Project{}, ...)
//	if err != nil {
//		return nil, fmt.Errorf("failed to run migrations: %w", err)
//	}
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "postgres", "postgresql":
		dsn := cfg.Database.GetDSN()
		dialector = postgres.Open(dsn)
	case "sqlite":
		dsn := cfg.Database.GetDSN()
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel(cfg.Logger.Level)),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		SkipDefaultTransaction: false,
		PrepareStmt:            true,
		PrepareStmtMaxSize:     100,
		PrepareStmtTTL:         time.Hour,
		TranslateError:         true,
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

// getLogLevel converts a string log level to the corresponding GORM logger level.
//
// This helper function maps application logger configuration to GORM's
// internal logger levels, enabling consistent logging behavior across
// the application and the ORM layer.
//
// Parameters:
//   - level: String representation of the desired log level
//
// Returns:
//   - logger.LogLevel: Corresponding GORM log level constant
//
// Supported log levels:
//   - "silent": logger.Silent - No database logs
//   - "error": logger.Error - Only log database errors
//   - "warn": logger.Warn - Log warnings and errors
//   - "info": logger.Info - Log informational messages, warnings, and errors
//   - Any other value: Defaults to logger.Info
//
// Log level hierarchy (from most to least verbose):
//   1. Info: All database operations and SQL queries
//   2. Warn: Warnings and errors only
//   3. Error: Errors only
//   4. Silent: No logging
//
// Example usage:
//
//	level := getLogLevel("info")
//	// Returns: logger.Info
//
//	level := getLogLevel("debug")
//	// Returns: logger.Info (fallback for unsupported levels)
//
//	level := getLogLevel("error")
//	// Returns: logger.Error
//
// Configuration:
// This function is typically called with cfg.Logger.Level from the
// application configuration to ensure consistent logging levels.
//
// Performance considerations:
// - "silent" level provides best performance (no logging overhead)
// - "error" level is recommended for production
// - "info" level is useful for development and debugging
func getLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

