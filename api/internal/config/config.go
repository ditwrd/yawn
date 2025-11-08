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

// Package config provides configuration management for the YAWN API application.
//
// This package handles loading and managing application configuration from multiple sources:
// - YAML configuration files
// - Environment variables with YAWN_ prefix
// - Default values
//
// The configuration supports different database types (PostgreSQL, SQLite), JWT settings,
// HTTP server configuration, and logging settings. It uses Viper for configuration
// management and supports configuration file discovery in multiple locations.
//
// Example usage:
//
//	cfg, err := config.LoadConfig("")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access configuration values
//	port := cfg.Server.Port
//	dbDSN := cfg.Database.GetDSN()
//
// Environment variables:
// Configuration can be overridden using environment variables with the YAWN_ prefix.
// For example, YAWN_SERVER_PORT overrides server.port, YAWN_DATABASE_HOST overrides
// database.host, etc.
//
// Configuration file locations:
// The package searches for config.yaml in the following order:
// - Current directory (.)
// - ./config directory
// - /etc/yawn directory
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
//
// This is the main configuration structure that contains all sub-configurations
// for the application. Each field represents a different aspect of the application
// configuration and is populated from the configuration file or environment variables.
//
// The configuration follows a hierarchical structure where each sub-configuration
// is embedded as a separate struct type.
type Config struct {
	// Server contains HTTP server configuration settings
	Server   ServerConfig   `mapstructure:"server"`
	// Database contains database connection and configuration settings
	Database DatabaseConfig `mapstructure:"database"`
	// JWT contains JWT token configuration for authentication
	JWT      JWTConfig      `mapstructure:"jwt"`
	// Logger contains logging configuration settings
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// ServerConfig holds HTTP server configuration settings.
//
// This configuration controls how the HTTP server behaves, including the
// binding address, timeouts, and other server-specific settings.
//
// Environment variables:
//   - YAWN_SERVER_PORT: Server port (default: "8080")
//   - YAWN_SERVER_HOST: Server host (default: "0.0.0.0")
//   - YAWN_SERVER_READ_TIMEOUT: Read timeout in seconds (default: 30)
//   - YAWN_SERVER_WRITE_TIMEOUT: Write timeout in seconds (default: 30)
type ServerConfig struct {
	// Port specifies the TCP port for the server to listen on
	Port         string `mapstructure:"port"`
	// Host specifies the host address for the server to bind to
	Host         string `mapstructure:"host"`
	// ReadTimeout specifies the maximum duration for reading the entire request, including the body
	ReadTimeout  int    `mapstructure:"read_timeout"`
	// WriteTimeout specifies the maximum duration before timing out writes of the response
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database connection configuration settings.
//
// This configuration supports multiple database types including PostgreSQL and SQLite.
// The specific fields used depend on the database type selected.
//
// Environment variables:
//   - YAWN_DATABASE_TYPE: Database type (default: "sqlite")
//   - YAWN_DATABASE_HOST: Database host (default: "localhost")
//   - YAWN_DATABASE_PORT: Database port (default: "5432")
//   - YAWN_DATABASE_NAME: Database name (default: "yawn")
//   - YAWN_DATABASE_USER: Database user (default: "yawn")
//   - YAWN_DATABASE_PASSWORD: Database password (default: "")
//   - YAWN_DATABASE_SSL_MODE: SSL mode for PostgreSQL (default: "disable")
//   - YAWN_DATABASE_PATH: File path for SQLite (default: "./yawn.db")
type DatabaseConfig struct {
	// Type specifies the database type ("postgres", "postgresql", or "sqlite")
	Type     string `mapstructure:"type"`
	// Host specifies the database host address (for PostgreSQL)
	Host     string `mapstructure:"host"`
	// Port specifies the database port (for PostgreSQL)
	Port     string `mapstructure:"port"`
	// Name specifies the database name (for PostgreSQL)
	Name     string `mapstructure:"name"`
	// User specifies the database username (for PostgreSQL)
	User     string `mapstructure:"user"`
	// Password specifies the database password (for PostgreSQL)
	Password string `mapstructure:"password"`
	// SSLMode specifies the SSL mode for PostgreSQL connections
	SSLMode  string `mapstructure:"ssl_mode"`
	// Path specifies the file path for SQLite database
	Path     string `mapstructure:"path"` // For SQLite
}

// JWTConfig holds JWT (JSON Web Token) configuration settings.
//
// This configuration controls JWT token generation and validation for
// authentication in the application.
//
// Environment variables:
//   - YAWN_JWT_SECRET: JWT secret key (default: "change-me-in-production")
//   - YAWN_JWT_TTL: JWT token time-to-live in seconds (default: 3600)
//
// Security note:
// The default secret should be changed in production environments.
// Use a strong, randomly generated secret key.
type JWTConfig struct {
	// Secret is the secret key used to sign and validate JWT tokens
	Secret string `mapstructure:"secret"`
	// TTL specifies the token time-to-live in seconds (default: 3600 = 1 hour)
	TTL    int    `mapstructure:"ttl"`
}

// LoggerConfig holds logger configuration settings.
//
// This configuration controls the logging behavior, including log levels and output format.
// The application uses zerolog for structured logging.
//
// Environment variables:
//   - YAWN_LOGGER_LEVEL: Log level (default: "info")
//   - YAWN_LOGGER_FORMAT: Log format (default: "json")
//
// Supported log levels:
//   - trace: Most detailed level, typically used for debugging
//   - debug: Debug information for developers
//   - info: General information about application flow
//   - warn: Warning messages for potentially problematic situations
//   - error: Error messages for failures
//   - fatal: Fatal errors that cause the application to exit
//   - panic: Panic-level errors
//
// Supported formats:
//   - json: Structured JSON output (default)
//   - console: Human-readable console output with colors
type LoggerConfig struct {
	// Level specifies the minimum log level to output
	Level  string `mapstructure:"level"`
	// Format specifies the output format for logs ("json" or "console")
	Format string `mapstructure:"format"`
}

// LoadConfig loads configuration from file and environment variables.
//
// This function loads application configuration from multiple sources in the following priority order:
// 1. Default values (lowest priority)
// 2. Configuration file (config.yaml)
// 3. Environment variables with YAWN_ prefix (highest priority)
//
// The function searches for configuration files in the following locations:
// - Current directory (.)
// - ./config directory
// - /etc/yawn directory
//
// If configPath is provided, it will use that specific configuration file instead
// of searching the default locations.
//
// Parameters:
//   - configPath: Optional path to a specific configuration file. If empty, searches default locations.
//
// Returns:
//   - *Config: The loaded configuration structure
//   - error: Any error encountered during configuration loading
//
// Example usage:
//
//	// Load from default locations
//	cfg, err := config.LoadConfig("")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Load from specific file
//	cfg, err := config.LoadConfig("/path/to/config.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Environment variable format:
// Environment variables should use the YAWN_ prefix and replace dots with underscores.
// For example:
//   - YAWN_SERVER_PORT overrides server.port
//   - YAWN_DATABASE_HOST overrides database.host
//   - YAWN_JWT_SECRET overrides jwt.secret
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/yawn")

	// Set environment variable prefix
	viper.SetEnvPrefix("YAWN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found, just use defaults and env vars
				fmt.Println("Config file not found, using defaults and environment variables")
			} else {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	} else {
		// Try to read config, but don't fail if it doesn't exist
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found, just use defaults and env vars
				fmt.Println("Config file not found, using defaults and environment variables")
			} else {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values for all configuration options.
//
// This function is called during configuration loading to ensure that all
// configuration fields have sensible default values. These defaults can be
// overridden by configuration files or environment variables.
//
// Default values set:
//   - Server:
//     * port: "8080"
//     * host: "0.0.0.0"
//     * read_timeout: 30 seconds
//     * write_timeout: 30 seconds
//   - Database:
//     * type: "sqlite"
//     * path: "./yawn.db"
//     * host: "localhost"
//     * port: "5432"
//     * name: "yawn"
//     * user: "yawn"
//     * password: ""
//     * ssl_mode: "disable"
//   - JWT:
//     * secret: "change-me-in-production"
//     * ttl: 3600 seconds (1 hour)
//   - Logger:
//     * level: "info"
//     * format: "json"
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./yawn.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.name", "yawn")
	viper.SetDefault("database.user", "yawn")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.ssl_mode", "disable")

	// JWT defaults
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.ttl", 3600) // 1 hour

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
}

// GetDSN returns the database connection string (Data Source Name) based on database type.
//
// This method generates the appropriate connection string for the configured database type.
// The format of the DSN varies depending on the database:
//
// For PostgreSQL:
//   host=%s user=%s password=%s dbname=%s port=%s sslmode=%s
//
// For SQLite:
//   Returns the file path directly
//
// Returns:
//   - string: The formatted database connection string
//   - empty string: If the database type is not supported
//
// Example usage:
//
//	cfg := &config.DatabaseConfig{
//		Type:     "postgres",
//		Host:     "localhost",
//		User:     "myuser",
//		Password: "mypassword",
//		Name:     "mydb",
//		Port:     "5432",
//		SSLMode:  "disable",
//	}
//	dsn := cfg.GetDSN()
//	// dsn = "host=localhost user=myuser password=mypassword dbname=mydb port=5432 sslmode=disable"
//
// SQLite example:
//
//	cfg := &config.DatabaseConfig{
//		Type: "sqlite",
//		Path: "./myapp.db",
//	}
//	dsn := cfg.GetDSN()
//	// dsn = "./myapp.db"
func (d *DatabaseConfig) GetDSN() string {
	switch strings.ToLower(d.Type) {
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode)
	case "sqlite":
		return d.Path
	default:
		return ""
	}
}