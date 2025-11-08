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

// Package web provides HTTP server configuration and middleware setup for the YAWN application.
//
// This package configures the Echo web framework with essential middleware,
// request logging, CORS support, and server settings. It provides a centralized
// way to set up the HTTP server with consistent behavior across the application.
//
// Features:
//   - Echo framework configuration with sensible defaults
//   - Structured request logging with zerolog integration
//   - CORS middleware for cross-origin requests
//   - Recovery middleware for panic handling
//   - Request ID middleware for request tracing
//   - Configurable server timeouts and address binding
//   - Automatic request/response logging with duration tracking
//
// Middleware stack:
//   1. RequestID: Generates unique request IDs for tracing
//   2. Logger: Structured logging of HTTP requests and responses
//   3. Recovery: Recovers from panics and logs errors
//   4. CORS: Handles cross-origin resource sharing
//   5. RemoveTrailingSlash: Normalizes URL paths
//
// Configuration:
// Server behavior is configured through the config package:
//   - cfg.Server.Host: Server binding address (default: "0.0.0.0")
//   - cfg.Server.Port: Server port (default: "8080")
//   - cfg.Server.ReadTimeout: Maximum read duration (default: 30s)
//   - cfg.Server.WriteTimeout: Maximum write duration (default: 30s)
//
// Request logging:
// All HTTP requests are automatically logged with the following fields:
//   - method: HTTP method (GET, POST, etc.)
//   - uri: Request URI
//   - status: HTTP response status code
//   - duration: Request processing time in milliseconds
//   - Additional fields can be added by handlers
//
// CORS configuration:
//   - AllowOrigins: All origins (*) - configure for production
//   - AllowMethods: GET, PUT, POST, DELETE, OPTIONS
//   - AllowHeaders: Origin, Content-Type, Accept, Authorization
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
//	e := web.NewEcho(cfg, logger)
//
//	// Add routes
//	e.GET("/health", func(c echo.Context) error {
//		return c.JSON(200, map[string]string{"status": "ok"})
//	})
//
//	// Start server
//	e.Logger.Fatal(e.Start(e.Server.Addr))
//
// Production considerations:
//   - Configure CORS origins appropriately for your domain
//   - Use HTTPS in production (configure via reverse proxy)
//   - Set appropriate timeouts for your use case
//   - Consider rate limiting middleware
//   - Add authentication/authorization middleware
//
// Performance:
//   - Echo provides high performance HTTP routing
//   - Middleware is optimized for minimal overhead
//   - Request logging is structured and efficient
//   - Connection pooling and timeouts prevent resource leaks
package web

import (
	"fmt"
	"time"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// NewEcho creates a new Echo instance with configured middleware and settings.
//
// This function initializes a fully configured Echo web server with essential
// middleware, request logging, CORS support, and server settings based on the
// provided configuration. The returned Echo instance is ready for route registration
// and server startup.
//
// Parameters:
//   - cfg: Application configuration containing server settings
//   - logger: Configured zerolog logger for request logging
//
// Returns:
//   - *echo.Echo: Fully configured Echo instance ready for use
//
// Configuration applied:
//   - Server address: cfg.Server.Host:cfg.Server.Port
//   - Read timeout: time.Duration(cfg.Server.ReadTimeout) * time.Second
//   - Write timeout: time.Duration(cfg.Server.WriteTimeout) * time.Second
//
// Middleware stack (applied in order):
//   1. RecoveryMiddleware: Recovers from panics and logs errors
//   2. CORSMiddleware: Handles cross-origin requests with configurable origins
//   3. RequestIDMiddleware: Generates unique request IDs for tracing
//   4. RequestLoggerMiddleware: Logs all HTTP requests with structured data
//
// Request logging format:
// All HTTP requests are automatically logged with the following structure:
//
//	{
//	  "level": "info",
//	  "service": "yawn-api",
//	  "time": "2006-01-02T15:04:05Z",
//	  "method": "GET",
//	  "uri": "/api/users",
//	  "status": 200,
//	  "duration": "15.234ms"
//	}
//
// CORS configuration:
//   - AllowOrigins: ["*"] (configure for production security)
//   - AllowMethods: [echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS]
//   - AllowHeaders: [echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization]
//
// Server settings:
//   - Address: Configured from cfg.Server.Host and cfg.Server.Port
//   - ReadTimeout: Maximum duration for reading request headers and body
//   - WriteTimeout: Maximum duration for writing response
//   - RemoveTrailingSlash: Normalizes URL paths by removing trailing slashes
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
//	// Create Echo instance
//	e := web.NewEcho(cfg, logger)
//
//	// Add custom middleware if needed
//	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
//
//	// Register routes
//	api := e.Group("/api")
//	api.GET("/health", healthHandler)
//	api.POST("/users", createUserHandler)
//
//	// Start server
//	e.Logger.Infof("Starting server on %s", e.Server.Addr)
//	if err := e.Start(e.Server.Addr); err != nil {
//		e.Logger.Fatalf("Server failed to start: %v", err)
//	}
//
// Production recommendations:
//   - Configure CORS origins to specific domains instead of "*"
//   - Add authentication/authorization middleware
//   - Implement rate limiting
//   - Use HTTPS behind a reverse proxy
//   - Add request validation middleware
//   - Configure proper health check endpoints
//   - Set up monitoring and metrics collection
//
// Performance notes:
//   - Echo provides high-performance routing with radix tree
//   - Middleware stack is optimized for minimal overhead
//   - Request logging uses structured logging for efficient parsing
//   - Timeouts prevent slow clients from blocking server resources
func NewEcho(cfg *config.Config, logger *zerolog.Logger) *echo.Echo {
	e := echo.New()

	// Configure Echo logger to use zerolog
	e.Logger.SetOutput(zerolog.NewConsoleWriter())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("duration", time.Since(v.StartTime)).
				Msg("HTTP Request")
			return nil
		},
	}))

	// Recovery middleware
	e.Use(middleware.Recover())

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Configure appropriately for production
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Request ID middleware
	e.Use(middleware.RequestID())

	// Remove trailing slashes
	e.Pre(middleware.RemoveTrailingSlash())

	// Configure server
	e.Server.ReadTimeout = time.Duration(cfg.Server.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(cfg.Server.WriteTimeout) * time.Second
	e.Server.Addr = fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	return e
}