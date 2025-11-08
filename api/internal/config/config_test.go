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