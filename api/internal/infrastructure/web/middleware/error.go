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

// Package middleware provides HTTP middleware for the Echo framework.
//
// This package contains middleware for error handling, authentication,
// authorization, logging, and other cross-cutting concerns.
//
// Error handling middleware provides:
// - Consistent error response format across all endpoints
// - Proper HTTP status code mapping
// - Structured logging with error context
// - Security-conscious error messages (no sensitive data leakage)
package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	apperrors "github.com/ditwrd/yawn/api/internal/domain/errors"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

// ErrorHandler provides centralized error handling for Echo applications.
type ErrorHandler struct {
	logger *zerolog.Logger
}

// NewErrorHandler creates a new error handler instance.
func NewErrorHandler(logger *zerolog.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles errors and returns consistent JSON responses.
func (h *ErrorHandler) HandleError(err error, c echo.Context) {
	// Set the content type
	c.Response().
		Header().
		Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

	// Handle different error types
	var appErr *apperrors.AppError
	var validationErrors validator.ValidationErrors
	var httpErr *echo.HTTPError

	switch {
	case errors.As(err, &appErr):
		h.handleAppError(appErr, c)
	case errors.As(err, &validationErrors):
		h.handleValidationError(validationErrors, c)
	case errors.As(err, &httpErr):
		h.handleHTTPError(httpErr, c)
	case errors.Is(err, gorm.ErrRecordNotFound):
		h.handleNotFoundError(err, c)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		h.handleDuplicateKeyError(err, c)
	case errors.Is(err, gorm.ErrInvalidTransaction):
		h.handleDatabaseError(err, c)
	default:
		h.handleGenericError(err, c)
	}
}

// handleAppError handles custom application errors.
func (h *ErrorHandler) handleAppError(
	appErr *apperrors.AppError,
	c echo.Context,
) {
	// Log the error with context
	event := h.logger.Error().
		Str("error_code", string(appErr.Code)).
		Str("error_category", string(appErr.Category)).
		Str("origin", string(appErr.Origin)).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	if appErr.Cause != nil {
		event = event.Err(appErr.Cause)
	}

	if len(appErr.Context) > 0 {
		event = event.Fields(appErr.Context)
	}

	if appErr.Details != "" {
		event = event.Str("details", appErr.Details)
	}

	event.Msg(appErr.Message)

	// Return error response
	response := dto.ErrorResponse{
		Error:   appErr.Message,
		Code:    string(appErr.Code),
		Details: appErr.Details,
	}

	if !c.Response().Committed {
		c.JSON(appErr.HTTPStatus, response)
	}
}

// handleValidationError handles validation errors from request binding.
func (h *ErrorHandler) handleValidationError(
	validationErrors validator.ValidationErrors,
	c echo.Context,
) {
	h.logger.Warn().
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Interface("validation_errors", validationErrors).
		Msg("Request validation failed")

	// Convert validation errors to field-level messages
	fieldErrors := make(map[string]string)
	for _, e := range validationErrors {
		fieldErrors[e.Field()] = h.getValidationErrorMessage(e)
	}

	response := dto.ValidationErrorResponse{
		Error:   "Validation failed",
		Code:    "VALIDATION_FAILED",
		Details: "Please check your input and try again",
		Fields:  fieldErrors,
	}

	if !c.Response().Committed {
		c.JSON(http.StatusBadRequest, response)
	}
}

// handleHTTPError handles Echo HTTP errors.
func (h *ErrorHandler) handleHTTPError(
	httpErr *echo.HTTPError,
	c echo.Context,
) {
	h.logger.Warn().
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Int("status", httpErr.Code).
		Msg("HTTP error occurred")

	var message string
	var details string

	if httpErr.Message != nil {
		message = fmt.Sprintf("%v", httpErr.Message)
	} else {
		message = http.StatusText(httpErr.Code)
	}

	response := dto.ErrorResponse{
		Error:   message,
		Code:    h.getErrorCodeFromStatus(httpErr.Code),
		Details: details,
	}

	if !c.Response().Committed {
		c.JSON(httpErr.Code, response)
	}
}

// handleNotFoundError handles GORM record not found errors.
func (h *ErrorHandler) handleNotFoundError(err error, c echo.Context) {
	h.logger.Info().
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Err(err).
		Msg("Resource not found")

	response := dto.ErrorResponse{
		Error:   "Resource not found",
		Code:    "NOT_FOUND",
		Details: "The requested resource could not be found",
	}

	if !c.Response().Committed {
		c.JSON(http.StatusNotFound, response)
	}
}

// handleDuplicateKeyError handles GORM duplicate key errors.
func (h *ErrorHandler) handleDuplicateKeyError(err error, c echo.Context) {
	h.logger.Info().
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Err(err).
		Msg("Resource already exists")

	response := dto.ErrorResponse{
		Error:   "Resource already exists",
		Code:    "ALREADY_EXISTS",
		Details: "A resource with these values already exists",
	}

	if !c.Response().Committed {
		c.JSON(http.StatusConflict, response)
	}
}

// handleDatabaseError handles GORM database errors.
func (h *ErrorHandler) handleDatabaseError(err error, c echo.Context) {
	h.logger.Error().
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Err(err).
		Msg("Database error occurred")

	response := dto.ErrorResponse{
		Error:   "Database operation failed",
		Code:    "DATABASE_ERROR",
		Details: "An error occurred while processing your request",
	}

	if !c.Response().Committed {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// handleGenericError handles any other type of error.
func (h *ErrorHandler) handleGenericError(err error, c echo.Context) {
	h.logger.Error().
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Err(err).
		Msg("Unexpected error occurred")

	response := dto.ErrorResponse{
		Error:   "Internal server error",
		Code:    "INTERNAL_ERROR",
		Details: "An unexpected error occurred while processing your request",
	}

	if !c.Response().Committed {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// getValidationErrorMessage converts validator.FieldError to user-friendly
// message.
func (h *ErrorHandler) getValidationErrorMessage(
	e validator.FieldError,
) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return "Please enter a valid email address"
	case "min":
		if e.Param() == "1" {
			return fmt.Sprintf("%s must not be empty", e.Field())
		}
		return fmt.Sprintf(
			"%s must be at least %s characters",
			e.Field(),
			e.Param(),
		)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", e.Field(), e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", e.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", e.Field())
	default:
		return fmt.Sprintf("%s is invalid", e.Field())
	}
}

// getErrorCodeFromStatus converts HTTP status to error code.
func (h *ErrorHandler) getErrorCodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	default:
		return "UNKNOWN_ERROR"
	}
}

// ErrorMiddleware returns an Echo middleware function for error handling.
func ErrorMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	handler := NewErrorHandler(logger)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				handler.HandleError(err, c)
				return nil // Don't return the error to avoid double handling
			}
			return nil
		}
	}
}
