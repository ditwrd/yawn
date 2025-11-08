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

// Package handlers provides HTTP request handlers for the YAWN application.
//
// Follows Controller pattern with JSON responses, proper status codes, and
// health check
// endpoints for Kubernetes liveness/readiness probes and monitoring systems.
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler provides health check endpoints for monitoring systems.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns service health status for monitoring systems.
func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "yawn-api",
	})
}
