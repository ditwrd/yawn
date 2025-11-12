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

// Package config provides configuration management with YAML files, environment
// variables, and defaults. Supports PostgreSQL/SQLite, JWT, HTTP server, and
// logging settings.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

// ServerConfig holds HTTP server configuration settings.
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database connection configuration settings.
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	Path     string `mapstructure:"path"`
}

// JWTConfig holds JWT configuration settings for authentication.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	TTL    int    `mapstructure:"ttl"`
}

// LoggerConfig holds logger configuration settings.
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// CORSConfig holds CORS configuration settings for cross-origin requests.
type CORSConfig struct {
	AllowedOrigins     []string `mapstructure:"allowed_origins"`
	AllowCredentials   bool     `mapstructure:"allow_credentials"`
	AllowedMethods     []string `mapstructure:"allowed_methods"`
	AllowedHeaders     []string `mapstructure:"allowed_headers"`
	MaxAge             int      `mapstructure:"max_age"`
	EnableWildcardPort bool     `mapstructure:"enable_wildcard_port"`
}

// LoadConfig loads configuration from YAML files and environment variables.
// Priority: defaults < YAML file < YAWN_ prefixed env vars.
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

		err := viper.ReadInConfig()
		if err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// Config file not found, just use defaults and env vars
				fmt.Println(
					"Config file not found, using defaults and environment variables",
				)
			} else {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	} else {
		// Try to read config, but don't fail if it doesn't exist
		err := viper.ReadInConfig()
		if err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// Config file not found, just use defaults and env vars
				fmt.Println("Config file not found, using defaults and environment variables")
			} else {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	var config Config

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values.
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

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.allowed_methods", []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	viper.SetDefault("cors.max_age", 3600) // 1 hour
	viper.SetDefault("cors.enable_wildcard_port", true) // Enable wildcard port for development
}

// GetDSN returns the database connection string for PostgreSQL/SQLite.
func (d *DatabaseConfig) GetDSN() string {
	switch strings.ToLower(d.Type) {
	case "postgres", "postgresql":
		return fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			d.Host,
			d.User,
			d.Password,
			d.Name,
			d.Port,
			d.SSLMode,
		)
	case "sqlite":
		return d.Path
	default:
		return ""
	}
}
