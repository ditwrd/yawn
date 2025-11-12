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

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestCORSIntegration(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		originHeader   string
		expectedStatus int
		expectAllowed  bool // Whether the origin should be allowed
	}{
		{
			name: "Development wildcard port enabled",
			config: &config.Config{
				CORS: config.CORSConfig{
					AllowedOrigins:     []string{"http://localhost:*"},
					AllowCredentials:   true,
					AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
					MaxAge:             3600,
					EnableWildcardPort: true,
				},
			},
			originHeader:   "http://localhost:3000",
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name: "Development wildcard port disabled",
			config: &config.Config{
				CORS: config.CORSConfig{
					AllowedOrigins:     []string{"http://localhost:3000"},
					AllowCredentials:   true,
					AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
					MaxAge:             3600,
					EnableWildcardPort: false,
				},
			},
			originHeader:   "http://localhost:3001",
			expectedStatus: http.StatusOK,
			expectAllowed:  false, // Should not allow origin 3001 since only 3000 is explicitly allowed
		},
		{
			name: "Production specific origins",
			config: &config.Config{
				CORS: config.CORSConfig{
					AllowedOrigins:     []string{"https://example.com", "https://app.example.com"},
					AllowCredentials:   true,
					AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
					MaxAge:             3600,
					EnableWildcardPort: false,
				},
			},
			originHeader:   "https://example.com",
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
		{
			name: "Production unauthorized origin",
			config: &config.Config{
				CORS: config.CORSConfig{
					AllowedOrigins:     []string{"https://example.com"},
					AllowCredentials:   true,
					AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
					MaxAge:             3600,
					EnableWildcardPort: false,
				},
			},
			originHeader:   "https://unauthorized.com",
			expectedStatus: http.StatusOK,
			expectAllowed:  false,
		},
		{
			name: "Multiple origins with wildcard expansion",
			config: &config.Config{
				CORS: config.CORSConfig{
					AllowedOrigins:     []string{"http://localhost:*", "https://api.example.com"},
					AllowCredentials:   true,
					AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
					MaxAge:             3600,
					EnableWildcardPort: true,
				},
			},
			originHeader:   "http://localhost:5173",
			expectedStatus: http.StatusOK,
			expectAllowed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger for test
			logger := zerolog.New(zerolog.NewConsoleWriter())

			// Create Echo instance with test config
			e := NewEcho(tt.config, &logger)

			// Add a test route
			e.GET("/test", func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"message": "test"})
			})

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.originHeader != "" {
				req.Header.Set("Origin", tt.originHeader)
			}

			// Create response recorder
			rec := httptest.NewRecorder()

			// Serve the request
			e.ServeHTTP(rec, req)

			// Check status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check Access-Control-Allow-Origin header
			allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.expectAllowed {
				assert.Equal(t, tt.originHeader, allowOrigin, "Allowed origin should match request origin")
				assert.Equal(t, "Origin", rec.Header().Get("Vary"), "Vary header should be 'Origin' for CORS")
			} else {
				assert.Empty(t, allowOrigin, "Unauthorized origin should not be allowed")
			}

			// Check for credentials header only if origin is allowed (Echo only adds it for allowed origins)
			if tt.expectAllowed && tt.config.CORS.AllowCredentials {
				assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
			} else if !tt.expectAllowed {
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	// Test preflight OPTIONS request
	logger := zerolog.New(zerolog.NewConsoleWriter())
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins:     []string{"http://localhost:*"},
			AllowCredentials:   true,
			AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			MaxAge:             3600,
			EnableWildcardPort: true,
		},
	}

	e := NewEcho(cfg, &logger)
	e.POST("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "test"})
	})

	// Create preflight request
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Preflight should succeed
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET,POST,PUT,DELETE,OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin,Content-Type,Accept,Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "3600", rec.Header().Get("Access-Control-Max-Age"))
}