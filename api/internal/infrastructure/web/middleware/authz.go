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

// Package middleware provides authorization middleware for role-based access
// control.
package middleware

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
)

// AuthorizationMiddleware provides authorization middleware functionality.
type AuthorizationMiddleware struct {
	logger *zerolog.Logger
}

// NewAuthorizationMiddleware creates a new authorization middleware.
func NewAuthorizationMiddleware(
	logger *zerolog.Logger,
) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		logger: logger,
	}
}

// RequireRole creates middleware that requires specific user roles.
func (am *AuthorizationMiddleware) RequireRole(
	roles ...models.UserRole,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user role from context
			userRole, err := GetUserRole(c)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Msg("Failed to get user role for authorization")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Authorization check failed",
						"code":  "AUTHZ_ERROR",
					},
				)
			}

			// Check if user has required role
			hasRequiredRole := false
			for _, requiredRole := range roles {
				if userRole == string(requiredRole) {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				am.logger.Warn().
					Str("user_role", userRole).
					Strs("required_roles", func() []string {
						result := make([]string, len(roles))
						for i, role := range roles {
							result[i] = string(role)
						}
						return result
					}()).
					Str("path", c.Request().URL.Path).
					Msg("User role not authorized for resource")

				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error": "Insufficient permissions",
					"code":  "INSUFFICIENT_PERMISSIONS",
				})
			}

			// User has required role, continue to next handler
			return next(c)
		}
	}
}

// RequireAdmin creates middleware that requires admin role.
func (am *AuthorizationMiddleware) RequireAdmin() echo.MiddlewareFunc {
	return am.RequireRole(models.UserRoleAdmin)
}

// RequireOwnership creates middleware that requires resource ownership.
func (am *AuthorizationMiddleware) RequireOwnership(
	resourceIDParam ...string,
) echo.MiddlewareFunc {
	paramName := "id"
	if len(resourceIDParam) > 0 {
		paramName = resourceIDParam[0]
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := GetUserID(c)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Msg("Failed to get user ID for ownership check")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Authorization check failed",
						"code":  "AUTHZ_ERROR",
					},
				)
			}

			userRole, err := GetUserRole(c)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Msg("Failed to get user role for ownership check")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Authorization check failed",
						"code":  "AUTHZ_ERROR",
					},
				)
			}

			if userRole == string(models.UserRoleAdmin) {
				am.logger.Debug().
					Str("user_id", userID).
					Str("path", c.Request().URL.Path).
					Msg("Admin user bypassing ownership check")
				return next(c)
			}

			resourceIDStr := c.Param(paramName)
			if resourceIDStr == "" {
				am.logger.Warn().
					Str("param", paramName).
					Str("path", c.Request().URL.Path).
					Msg("Resource ID parameter missing for ownership check")
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error": "Resource ID parameter missing",
					"code":  "MISSING_RESOURCE_ID",
				})
			}

			resourceID, err := uuid.FromString(resourceIDStr)
			if err != nil {
				am.logger.Warn().
					Err(err).
					Str("resource_id", resourceIDStr).
					Str("path", c.Request().URL.Path).
					Msg("Invalid resource ID format for ownership check")
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"error": "Invalid resource ID format",
					"code":  "INVALID_RESOURCE_ID",
				})
			}

			userUUID, err := uuid.FromString(userID)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("user_id", userID).
					Msg("Invalid user ID format in context")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Invalid user context",
						"code":  "INVALID_USER_CONTEXT",
					},
				)
			}

			if userUUID != resourceID {
				am.logger.Warn().
					Str("user_id", userID).
					Str("resource_id", resourceID.String()).
					Str("path", c.Request().URL.Path).
					Msg("User is not owner of resource")

				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error": "You don't have permission to access this resource",
					"code":  "NOT_OWNER",
				})
			}

			return next(c)
		}
	}
}

// RequireRoleOrOwnership creates middleware that requires specific roles OR
// ownership.
//
// This middleware provides flexible authorization where users can access a
// resource
// if they have one of the specified roles OR they own the resource.
// Admin users automatically pass all checks.
//
// Parameters:
//   - roles: List of allowed roles
//
// - resourceIDParam: URL parameter name containing the resource ID (default:
// "id")
//
// Returns:
// - echo.MiddlewareFunc: Echo middleware function for role or ownership
// verification
//
// Example:
//
//	// Allow admin users, OR users who own the resource
//	e.Use(RequireRoleOrOwnership(models.UserRoleAdmin))
//
//	// Allow admin/maintainer users, OR users who own the resource
//	e.Use(RequireRoleOrOwnership(models.UserRoleAdmin,
//
// models.UserRoleMaintainer, "project_id"))
func (am *AuthorizationMiddleware) RequireRoleOrOwnership(
	roles ...models.UserRole,
) echo.MiddlewareFunc {
	// Extract resource ID parameter from the last argument if it's a string
	var resourceRoles []models.UserRole
	var resourceIDParam string = "id"

	if len(roles) > 0 {
		// Check if the last argument is a string (resource ID parameter)
		if lastArg := roles[len(roles)-1]; len(lastArg) > 0 &&
			lastArg[0] >= 'a' &&
			lastArg[0] <= 'z' {
			// Last argument looks like a string parameter name
			resourceIDParam = string(lastArg)
			resourceRoles = roles[:len(roles)-1]
		} else {
			// All arguments are roles
			resourceRoles = roles
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user role from context
			userRole, err := GetUserRole(c)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Msg("Failed to get user role for authorization")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Authorization check failed",
						"code":  "AUTHZ_ERROR",
					},
				)
			}

			// Check if user has required role
			hasRequiredRole := false
			for _, requiredRole := range resourceRoles {
				if userRole == string(requiredRole) {
					hasRequiredRole = true
					break
				}
			}

			// If user has required role, allow access
			if hasRequiredRole {
				return next(c)
			}

			// If user doesn't have required role, check ownership
			ownershipMiddleware := am.RequireOwnership(resourceIDParam)
			return ownershipMiddleware(next)(c)
		}
	}
}

// CustomPermissionChecker defines a function type for custom permission
// checking.
//
// This function allows for complex authorization logic beyond simple roles
// and ownership checks.
type CustomPermissionChecker func(c echo.Context, claims *services.TokenClaims) (bool, error)

// RequireCustomPermission creates middleware with custom permission checking.
//
// This middleware allows for complex authorization logic by providing a custom
// permission checker function.
//
// Parameters:
//   - checker: Function that checks if the user has permission
//
// Returns:
//   - echo.MiddlewareFunc: Echo middleware function for custom authorization
//
// Example:
//
//	// Custom checker that allows access to user's own resources or admin users
//	checker := func(c echo.Context, claims *services.TokenClaims) (bool, error)
//
//	{
//			resourceUserID := c.Param("user_id")
//			return claims.UserID.String() == resourceUserID || claims.Role ==
//
// models.UserRoleAdmin, nil
//
//	}
//	e.Use(RequireCustomPermission(checker))
func (am *AuthorizationMiddleware) RequireCustomPermission(
	checker CustomPermissionChecker,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user claims from context
			claims, err := GetUserClaims(c)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Msg("Failed to get user claims for custom permission check")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Authorization check failed",
						"code":  "AUTHZ_ERROR",
					},
				)
			}

			// Check custom permission
			hasPermission, err := checker(c, claims)
			if err != nil {
				am.logger.Error().
					Err(err).
					Str("user_id", claims.UserID.String()).
					Str("path", c.Request().URL.Path).
					Msg("Custom permission check failed")
				return c.JSON(
					http.StatusInternalServerError,
					map[string]interface{}{
						"error": "Authorization check failed",
						"code":  "AUTHZ_ERROR",
					},
				)
			}

			if !hasPermission {
				am.logger.Warn().
					Str("user_id", claims.UserID.String()).
					Str("path", c.Request().URL.Path).
					Msg("Custom permission check failed")

				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error": "Insufficient permissions",
					"code":  "INSUFFICIENT_PERMISSIONS",
				})
			}

			// User has permission, continue to next handler
			return next(c)
		}
	}
}
