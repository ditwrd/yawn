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

// Package errors provides custom error types with service/module origin
// context.
//
// This package defines a structured error handling system that includes:
// - Error codes and categories for consistent error identification
// - Service/module origin context for debugging and monitoring
// - HTTP status code mapping for REST API responses
// - Structured error data for additional context
//
// Design principles:
//   - Error codes are constant and machine-readable
//   - Error messages are user-friendly and localized
//   - Service origins help with debugging and monitoring
//   - HTTP status mapping provides consistent API responses
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a machine-readable error code.
type ErrorCode string

// ErrorCategory represents the category of error for grouping and handling.
type ErrorCategory string

// ServiceOrigin represents the service/module where the error originated.
type ServiceOrigin string

// Common error codes
const (
	// Validation errors
	CodeInvalidRequest     ErrorCode = "INVALID_REQUEST"
	CodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	CodeMissingField       ErrorCode = "MISSING_FIELD"
	CodeInvalidFormat      ErrorCode = "INVALID_FORMAT"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// Authentication & Authorization errors
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeTokenExpired ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid ErrorCode = "TOKEN_INVALID"

	// Resource errors
	CodeNotFound       ErrorCode = "NOT_FOUND"
	CodeAlreadyExists  ErrorCode = "ALREADY_EXISTS"
	CodeConflict       ErrorCode = "CONFLICT"
	CodeResourceLocked ErrorCode = "RESOURCE_LOCKED"

	// Business logic errors
	CodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"
	CodeQuotaExceeded         ErrorCode = "QUOTA_EXCEEDED"
	CodeInvalidState          ErrorCode = "INVALID_STATE"

	// System errors
	CodeInternalError   ErrorCode = "INTERNAL_ERROR"
	CodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	CodeNetworkError    ErrorCode = "NETWORK_ERROR"
	CodeServiceTimeout  ErrorCode = "SERVICE_TIMEOUT"
	CodeDependencyError ErrorCode = "DEPENDENCY_ERROR"
)

// Error categories
const (
	CategoryValidation    ErrorCategory = "VALIDATION"
	CategoryAuth          ErrorCategory = "AUTHENTICATION"
	CategoryAuthorization ErrorCategory = "AUTHORIZATION"
	CategoryNotFound      ErrorCategory = "NOT_FOUND"
	CategoryConflict      ErrorCategory = "CONFLICT"
	CategoryBusiness      ErrorCategory = "BUSINESS"
	CategorySystem        ErrorCategory = "SYSTEM"
	CategoryExternal      ErrorCategory = "EXTERNAL"
)

// Service origins
const (
	OriginAuthService       ServiceOrigin = "auth_service"
	OriginUserService       ServiceOrigin = "user_service"
	OriginProjectService    ServiceOrigin = "project_service"
	OriginAssetService      ServiceOrigin = "asset_service"
	OriginPipelineService   ServiceOrigin = "pipeline_service"
	OriginRepositoryService ServiceOrigin = "repository_service"
	OriginJWTService        ServiceOrigin = "jwt_service"
	OriginPasswordService   ServiceOrigin = "password_service"
	OriginGitOpsService     ServiceOrigin = "gitops_service"
	OriginDatabase          ServiceOrigin = "database"
	OriginWebHandler        ServiceOrigin = "web_handler"
	OriginMiddleware        ServiceOrigin = "middleware"
)

// AppError represents a structured application error with context.
type AppError struct {
	Code       ErrorCode      `json:"code"`
	Category   ErrorCategory  `json:"category"`
	Message    string         `json:"message"`
	Details    string         `json:"details,omitempty"`
	Origin     ServiceOrigin  `json:"origin"`
	HTTPStatus int            `json:"-"`
	Cause      error          `json:"-"`
	Context    map[string]any `json:"context,omitempty"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Message, e.Details, e.Code)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

// Unwrap returns the underlying cause.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithCause adds an underlying cause to the error.
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithContext adds context information to the error.
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithDetails adds detailed information to the error.
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// NewValidationError creates a new validation error.
func NewValidationError(
	code ErrorCode,
	message string,
	origin ServiceOrigin,
) *AppError {
	return &AppError{
		Code:       code,
		Category:   CategoryValidation,
		Message:    message,
		Origin:     origin,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewAuthError creates a new authentication error.
func NewAuthError(
	code ErrorCode,
	message string,
	origin ServiceOrigin,
) *AppError {
	status := http.StatusUnauthorized
	if code == CodeForbidden {
		status = http.StatusForbidden
	}

	return &AppError{
		Code:       code,
		Category:   CategoryAuth,
		Message:    message,
		Origin:     origin,
		HTTPStatus: status,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource string, origin ServiceOrigin) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Category:   CategoryNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Details:    fmt.Sprintf("The requested %s could not be found", resource),
		Origin:     origin,
		HTTPStatus: http.StatusNotFound,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string, origin ServiceOrigin) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Category:   CategoryConflict,
		Message:    message,
		Origin:     origin,
		HTTPStatus: http.StatusConflict,
	}
}

// NewBusinessError creates a new business logic error.
func NewBusinessError(
	code ErrorCode,
	message string,
	origin ServiceOrigin,
) *AppError {
	return &AppError{
		Code:       code,
		Category:   CategoryBusiness,
		Message:    message,
		Origin:     origin,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewSystemError creates a new system error.
func NewSystemError(
	code ErrorCode,
	message string,
	origin ServiceOrigin,
) *AppError {
	return &AppError{
		Code:       code,
		Category:   CategorySystem,
		Message:    message,
		Origin:     origin,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewExternalServiceError creates a new external service error.
func NewExternalServiceError(
	code ErrorCode,
	message string,
	origin ServiceOrigin,
) *AppError {
	return &AppError{
		Code:       code,
		Category:   CategoryExternal,
		Message:    message,
		Origin:     origin,
		HTTPStatus: http.StatusBadGateway,
	}
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a standard error into an AppError with context.
func WrapError(
	err error,
	code ErrorCode,
	message string,
	origin ServiceOrigin,
) *AppError {
	return &AppError{
		Code:       code,
		Category:   CategorySystem,
		Message:    message,
		Origin:     origin,
		HTTPStatus: http.StatusInternalServerError,
		Cause:      err,
	}
}
