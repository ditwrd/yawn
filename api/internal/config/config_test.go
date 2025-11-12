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
package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Ensure no config file exists and no env vars are set
	config, err := LoadConfig("")

	require.NoError(t, err)
	assert.NotNil(t, config)

	// Test default values
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 30, config.Server.ReadTimeout)
	assert.Equal(t, 30, config.Server.WriteTimeout)

	assert.Equal(t, "sqlite", config.Database.Type)
	assert.Equal(t, "./yawn.db", config.Database.Path)
	assert.Equal(t, "change-me-in-production", config.JWT.Secret)
	assert.Equal(t, 3600, config.JWT.TTL)
	assert.Equal(t, "info", config.Logger.Level)
	assert.Equal(t, "json", config.Logger.Format)

	// Test default CORS values
	assert.Equal(t, []string{"http://localhost:3000"}, config.CORS.AllowedOrigins)
	assert.True(t, config.CORS.AllowCredentials)
	assert.Equal(t, []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"}, config.CORS.AllowedMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Accept", "Authorization"}, config.CORS.AllowedHeaders)
	assert.Equal(t, 3600, config.CORS.MaxAge)
	assert.True(t, config.CORS.EnableWildcardPort)
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("YAWN_SERVER_PORT", "9000")
	os.Setenv("YAWN_DATABASE_TYPE", "postgres")
	os.Setenv("YAWN_DATABASE_HOST", "localhost")
	os.Setenv("YAWN_DATABASE_USER", "testuser")
	os.Setenv("YAWN_DATABASE_PASSWORD", "testpass")
	os.Setenv("YAWN_DATABASE_NAME", "testdb")
	os.Setenv("YAWN_DATABASE_PORT", "5432")
	os.Setenv("YAWN_JWT_SECRET", "test-secret")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("YAWN_SERVER_PORT")
		os.Unsetenv("YAWN_DATABASE_TYPE")
		os.Unsetenv("YAWN_DATABASE_HOST")
		os.Unsetenv("YAWN_DATABASE_USER")
		os.Unsetenv("YAWN_DATABASE_PASSWORD")
		os.Unsetenv("YAWN_DATABASE_NAME")
		os.Unsetenv("YAWN_DATABASE_PORT")
		os.Unsetenv("YAWN_JWT_SECRET")
	}()

	config, err := LoadConfig("")
	require.NoError(t, err)

	// Test that environment variables override defaults
	assert.Equal(t, "9000", config.Server.Port)
	assert.Equal(t, "postgres", config.Database.Type)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, "testuser", config.Database.User)
	assert.Equal(t, "testpass", config.Database.Password)
	assert.Equal(t, "testdb", config.Database.Name)
	assert.Equal(t, "5432", config.Database.Port)
	assert.Equal(t, "test-secret", config.JWT.Secret)
}

func TestLoadConfig_CORS_EnvironmentVariables(t *testing.T) {
	// Set CORS environment variables
	os.Setenv("YAWN_CORS_ALLOWED_ORIGINS", "https://example.com,https://app.example.com")
	os.Setenv("YAWN_CORS_ALLOW_CREDENTIALS", "false")
	os.Setenv("YAWN_CORS_ALLOWED_METHODS", "GET,POST")
	os.Setenv("YAWN_CORS_ALLOWED_HEADERS", "Content-Type,Authorization")
	os.Setenv("YAWN_CORS_MAX_AGE", "7200")
	os.Setenv("YAWN_CORS_ENABLE_WILDCARD_PORT", "false")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("YAWN_CORS_ALLOWED_ORIGINS")
		os.Unsetenv("YAWN_CORS_ALLOW_CREDENTIALS")
		os.Unsetenv("YAWN_CORS_ALLOWED_METHODS")
		os.Unsetenv("YAWN_CORS_ALLOWED_HEADERS")
		os.Unsetenv("YAWN_CORS_MAX_AGE")
		os.Unsetenv("YAWN_CORS_ENABLE_WILDCARD_PORT")
	}()

	config, err := LoadConfig("")
	require.NoError(t, err)

	// Test that CORS environment variables override defaults
	assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, config.CORS.AllowedOrigins)
	assert.False(t, config.CORS.AllowCredentials)
	assert.Equal(t, []string{"GET", "POST"}, config.CORS.AllowedMethods)
	assert.Equal(t, []string{"Content-Type", "Authorization"}, config.CORS.AllowedHeaders)
	assert.Equal(t, 7200, config.CORS.MaxAge)
	assert.False(t, config.CORS.EnableWildcardPort)
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		dbConfig := DatabaseConfig{
			Type: "sqlite",
			Path: "/tmp/test.db",
		}
		dsn := dbConfig.GetDSN()
		assert.Equal(t, "/tmp/test.db", dsn)
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		dbConfig := DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     "5432",
			Name:     "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		}
		dsn := dbConfig.GetDSN()
		expected := "host=localhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable"
		assert.Equal(t, expected, dsn)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	type args struct {
		configPath string
	}

	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := LoadConfig(tt.args.configPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"LoadConfig() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_setDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setDefaults()
		})
	}
}
