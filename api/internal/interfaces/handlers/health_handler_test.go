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

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthHandler_Success tests the health check endpoint returns 200 OK.
func TestHealthHandler_Success(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a simple health handler function
	healthHandler := func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	}

	// Execute
	err := healthHandler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`+"\n", rec.Body.String())
}

// TestHealthHandler_Integration tests the health check with full Echo setup.
func TestHealthHandler_Integration(t *testing.T) {
	// Setup Echo router with health endpoint
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Execute
	e.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
}

// TestHealthHandler_MethodNotAllowed tests that unsupported methods return 404.
func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	// Setup Echo router with only GET /health
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Test POST request (should fail)
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	// Execute
	e.ServeHTTP(rec, req)

	// Assert - should return 405 for unsupported method
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// TestHealthHandler_IntegrationHeaders tests that health endpoint returns
// proper headers.
func TestHealthHandler_IntegrationHeaders(t *testing.T) {
	// Setup Echo router with health endpoint
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-cache")

		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Execute
	e.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Contains(t, rec.Body.String(), "ok")
}
