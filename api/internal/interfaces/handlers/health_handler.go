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

// Package handlers provides HTTP handlers for the YAWN application interfaces layer.
//
// This package contains the HTTP request handlers that form the presentation layer
// of the application. Handlers are responsible for processing incoming HTTP requests,
// validating input, coordinating with services and repositories, and formatting
// appropriate HTTP responses.
//
// Architecture pattern:
// The handlers follow the Controller pattern from clean architecture:
// - Handle HTTP request/response lifecycle
// - Parse and validate request data
// - Call appropriate service methods
// - Format and return HTTP responses
// - Handle error scenarios and status codes
//
// Current handlers:
//   - HealthHandler: Provides health check endpoints for monitoring
//
// Future handlers may include:
//   - UserHandler: User management endpoints
//   - ProjectHandler: Project management endpoints
//   - AssetHandler: Asset management endpoints
//   - RepositoryHandler: Repository management endpoints
//   - PipelineHandler: Pipeline management endpoints
//
// HTTP standards:
//   - Proper HTTP status codes
//   - JSON request/response format
//   - Error handling with appropriate status codes
//   - Content-Type headers
//   - Request validation
//
// Example usage:
//
//	// Create handler
//	healthHandler := handlers.NewHealthHandler()
//
//	// Register routes
//	e.GET("/health", healthHandler.Health)
//	e.GET("/health/live", healthHandler.Health) // Liveness probe
//	e.GET("/health/ready", healthHandler.Health) // Readiness probe
//
// Monitoring and observability:
// Health check endpoints are designed for:
//   - Kubernetes liveness and readiness probes
//   - Load balancer health checks
//   - Monitoring system alerts
//   - Manual health verification
//
// Response format:
// Health endpoints return consistent JSON responses:
//
//	{
//	  "status": "healthy",
//	  "service": "yawn-api",
//	  "timestamp": "2006-01-02T15:04:05Z"
//	}
//
// Error handling:
// Handlers follow consistent error response patterns:
//   - 400: Bad Request (validation errors)
//   - 401: Unauthorized (authentication required)
//   - 403: Forbidden (insufficient permissions)
//   - 404: Not Found (resource doesn't exist)
//   - 500: Internal Server Error (unexpected failures)
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health performs a health check
func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "yawn-api",
	})
}

