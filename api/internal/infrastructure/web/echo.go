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

// Package web provides HTTP server configuration and middleware using Echo
// framework.
//
// Configures Echo with request logging, CORS, recovery, and request ID
// middleware.
// Supports configurable server timeouts and structured logging integration.
package web

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	appmiddleware "github.com/ditwrd/yawn/api/internal/infrastructure/web/middleware"
)

// NewEcho creates a new Echo instance with middleware and server configuration.
//
// Configures request logging, CORS, recovery, and request ID middleware.
// Sets server timeouts and address from configuration. Returns ready-to-use
// Echo instance.
func NewEcho(cfg *config.Config, logger *zerolog.Logger) *echo.Echo {
	e := echo.New()

	// Configure Echo logger to use zerolog
	e.Logger.SetOutput(zerolog.NewConsoleWriter())
	e.Use(
		echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
			LogURI:    true,
			LogStatus: true,
			LogMethod: true,
			LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
				logger.Info().
					Str("method", v.Method).
					Str("uri", v.URI).
					Int("status", v.Status).
					Dur("duration", time.Since(v.StartTime)).
					Msg("HTTP Request")

				return nil
			},
		}),
	)

	// Global error handling middleware
	e.Use(appmiddleware.ErrorMiddleware(logger))

	// Recovery middleware
	e.Use(echomiddleware.Recover())

	// CORS middleware with configurable settings
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: processAllowedOrigins(
			cfg.CORS.AllowedOrigins,
			cfg.CORS.EnableWildcardPort,
		),
		AllowCredentials: cfg.CORS.AllowCredentials,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Request ID middleware
	e.Use(echomiddleware.RequestID())

	// Remove trailing slashes
	e.Pre(echomiddleware.RemoveTrailingSlash())

	// Configure server
	e.Server.ReadTimeout = time.Duration(cfg.Server.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(cfg.Server.WriteTimeout) * time.Second
	e.Server.Addr = fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	return e
}

// processAllowedOrigins expands wildcard patterns in allowed origins for development.
// When enableWildcardPort is true, it expands origins like "http://localhost:*" to
// include common development ports. This provides flexibility during development
// while maintaining security in production.
func processAllowedOrigins(origins []string, enableWildcardPort bool) []string {
	var result []string
	commonPorts := []string{"3000", "3001", "5173", "8000", "8080", "9000"}

	for _, origin := range origins {
		if enableWildcardPort && strings.Contains(origin, "*") {
			// Handle wildcard patterns like "http://localhost:*"
			wildcardRegex := regexp.MustCompile(`^(https?://[^:]+):.*$`)
			matches := wildcardRegex.FindStringSubmatch(origin)
			if len(matches) == 2 {
				base := matches[1]
				for _, port := range commonPorts {
					result = append(result, fmt.Sprintf("%s:%s", base, port))
				}
				continue
			}
		}
		// If no wildcard pattern or wildcard processing disabled, use as-is
		result = append(result, origin)
	}

	return result
}
