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

// Package web provides HTTP routing configuration for the Echo framework.
//
// Configures all API routes with appropriate middleware for authentication,
// authorization, and request handling. Supports modular route organization.
package web

import (
	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/infrastructure/web/middleware"
	"github.com/ditwrd/yawn/api/internal/interfaces/handlers"
	"github.com/labstack/echo/v4"
)

// RouterConfig contains dependencies for route setup.
type RouterConfig struct {
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	AuthMiddleware  *middleware.AuthMiddleware
	AuthzMiddleware *middleware.AuthorizationMiddleware
}

// SetupRoutes configures all application routes with appropriate middleware.
//
// Sets up authentication and user management endpoints with proper
// security middleware. Organizes routes by feature with versioning support.
func SetupRoutes(e *echo.Echo, cfg *RouterConfig) {
	// Health check endpoint (no auth required)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Authentication routes (no auth required)
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", cfg.AuthHandler.Register)
	authGroup.POST("/login", cfg.AuthHandler.Login)
	authGroup.POST("/refresh", cfg.AuthHandler.Refresh)
	authGroup.POST("/logout", cfg.AuthHandler.Logout)

	// User management routes (auth required)
	userGroup := v1.Group("/users")
	userGroup.Use(
		cfg.AuthMiddleware.RequireAuth(),
	) // All user routes require authentication

	// Admin-only routes
	userGroup.Use(
		cfg.AuthzMiddleware.RequireAdmin(),
	) // GET /users and DELETE /users require admin
	userGroup.GET("", cfg.UserHandler.ListUsers)
	userGroup.DELETE("/:id", cfg.UserHandler.DeleteUser)

	// Self or admin routes
	// For these routes, we need to reset the admin requirement and add role or
	// ownership
	// So we create separate groups with different middleware
	userSelfOrAdmin := v1.Group("/users")
	userSelfOrAdmin.Use(
		cfg.AuthMiddleware.RequireAuth(),
	) // Require authentication
	userSelfOrAdmin.Use(
		cfg.AuthzMiddleware.RequireRoleOrOwnership(models.UserRoleAdmin),
	) // Admin or owner
	userSelfOrAdmin.GET("/:id", cfg.UserHandler.GetUser)
	userSelfOrAdmin.PUT("/:id", cfg.UserHandler.UpdateUser)
}
