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

// Package logger provides structured logging configuration and setup for the YAWN application.
//
// This package configures zerolog for structured JSON logging with configurable
// output formats and log levels. It provides a centralized logging solution
// that can be used throughout the application for consistent log formatting.
//
// Features:
//   - Structured JSON logging with zerolog
//   - Configurable log levels (trace, debug, info, warn, error, fatal, panic)
//   - Multiple output formats (JSON, console)
//   - Automatic timestamp formatting
//   - Service context injection
//   - Graceful fallback for invalid configurations
//   - Zero-allocation logging for high performance
//
// Log levels:
//   - trace: Most detailed logging, typically for debugging
//   - debug: Debug information for developers
//   - info: General information about application flow (default)
//   - warn: Warning messages for potentially problematic situations
//   - error: Error messages for failures
//   - fatal: Fatal errors that cause the application to exit
//   - panic: Panic-level errors
//
// Output formats:
//   - JSON: Structured JSON output (default, recommended for production)
//   - Console: Human-readable console output with colors (recommended for development)
//
// Configuration:
// Logger configuration is managed through the config package:
//   - cfg.Logger.Level: Sets the minimum log level
//   - cfg.Logger.Format: Sets the output format ("json" or "console")
//
// Example usage:
//
//	cfg, err := config.LoadConfig("")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	logger := logger.NewLogger(cfg)
//
//	// Use the logger
//	logger.Info().
//		Str("user_id", "123").
//		Str("action", "login").
//		Msg("User logged in successfully")
//
//	logger.Error().
//		Err(err).
//		Str("component", "database").
//		Msg("Database connection failed")
//
// Performance:
// zerolog is designed for high-performance logging with minimal allocations:
//   - Zero-allocation API for structured fields
//   - Lazy evaluation of expensive operations
//   - Efficient JSON serialization
//   - Level-based filtering to avoid unnecessary work
//
// Integration:
// The logger can be easily integrated with other components:
//   - HTTP request logging middleware
//   - Database query logging
//   - Error tracking and monitoring
//   - Audit logging
//
// Best practices:
//   - Use structured fields with Str(), Int(), etc. instead of formatted strings
//   - Include relevant context like user_id, request_id, component
//   - Use appropriate log levels for different types of messages
//   - Avoid logging sensitive information
//   - Use JSON format in production for log aggregation
package logger

import (
	"os"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLogger creates a new zerolog logger instance based on the provided configuration.
//
// This function configures a structured logger with the specified log level and output format.
// It automatically injects service context and timestamps, and provides graceful fallback
// for invalid configuration values.
//
// Parameters:
//   - cfg: Application configuration containing logger settings
//
// Returns:
//   - *zerolog.Logger: Configured logger instance ready for use
//
// Configuration used:
//   - cfg.Logger.Level: Minimum log level to output (default: "info")
//   - cfg.Logger.Format: Output format ("json" or "console", default: "json")
//
// Log level handling:
//   - Valid levels: "trace", "debug", "info", "warn", "error", "fatal", "panic"
//   - Invalid levels fall back to "info" with a warning log
//   - Case-insensitive level parsing
//
// Output formats:
//   - "json": Structured JSON output with timestamps (recommended for production)
//   - "console": Human-readable output with colors and timestamps (recommended for development)
//   - Invalid formats fall back to JSON output
//
// Automatic features:
//   - Timestamps are added to all log entries
//   - Service context ("service": "yawn-api") is automatically added
//   - Global log level is set based on configuration
//   - Output is written to stdout
//
// Example usage:
//
//	cfg := &config.Config{
//		Logger: config.LoggerConfig{
//			Level:  "info",
//			Format: "json",
//		},
//	}
//
//	logger := NewLogger(cfg)
//
//	// Basic logging
//	logger.Info().Msg("Application started")
//
//	// Structured logging with fields
//	logger.Info().
//		Str("user_id", "12345").
//		Str("ip", "192.168.1.1").
//		Msg("User login successful")
//
//	// Error logging with error field
//	logger.Error().
//		Err(err).
//		Str("operation", "database_query").
//		Dur("duration", time.Since(start)).
//		Msg("Database operation failed")
//
//	// Warning with context
//	logger.Warn().
//		Str("rate_limit", "100/min").
//		Int("current_rate", 95).
//		Msg("Approaching rate limit")
//
// Console output example (format: "console"):
//	2006-01-02 15:04:05 INF User login successful service=yawn-api user_id=12345 ip=192.168.1.1
//
// JSON output example (format: "json"):
//	{"level":"info","service":"yawn-api","user_id":"12345","ip":"192.168.1.1","message":"User login successful","time":"2006-01-02T15:04:05Z"}
//
// Performance considerations:
//   - zerolog provides zero-allocation logging for best performance
//   - Use field methods (Str, Int, Bool, etc.) instead of formatted strings
//   - Log level filtering happens early to avoid unnecessary work
//   - JSON serialization is highly optimized
//
// Error handling:
//   - Invalid log levels fall back to "info" with warning
//   - Function never panics or returns nil
//   - All configuration errors are handled gracefully
func NewLogger(cfg *config.Config) *zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().Err(err).Msgf("Invalid log level '%s', using 'info'", cfg.Logger.Level)
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var output zerolog.Logger
	if cfg.Logger.Format == "console" {
		output = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}).With().Timestamp().Logger()
	} else {
		// JSON format (default)
		output = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Add service context
	output = output.With().Str("service", "yawn-api").Logger()

	return &output
}