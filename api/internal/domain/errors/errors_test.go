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

package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "error without details",
			appErr: &AppError{
				Code:    CodeInvalidRequest,
				Message: "Invalid request",
			},
			expected: "Invalid request (INVALID_REQUEST)",
		},
		{
			name: "error with details",
			appErr: &AppError{
				Code:    CodeInvalidRequest,
				Message: "Invalid request",
				Details: "Missing required field",
			},
			expected: "Invalid request: Missing required field (INVALID_REQUEST)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appErr.Error())
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := &AppError{
		Code:    CodeInternalError,
		Message: "System error",
		Cause:   originalErr,
	}

	assert.Equal(t, originalErr, appErr.Unwrap())
}

func TestAppError_WithCause(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewValidationError(
		CodeMissingField,
		"Field required",
		OriginWebHandler,
	)

	result := appErr.WithCause(originalErr)

	assert.Same(t, originalErr, result.Cause)
	assert.Equal(t, appErr, result) // Should return same instance
}

func TestAppError_WithContext(t *testing.T) {
	appErr := NewValidationError(
		CodeMissingField,
		"Field required",
		OriginWebHandler,
	)

	result := appErr.WithContext("field_name", "email").
		WithContext("user_id", "123")

	assert.Equal(t, "email", result.Context["field_name"])
	assert.Equal(t, "123", result.Context["user_id"])
	assert.Equal(t, appErr, result) // Should return same instance
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := NewValidationError(
		CodeMissingField,
		"Field required",
		OriginWebHandler,
	)

	result := appErr.WithDetails("Email address is required")

	assert.Equal(t, "Email address is required", result.Details)
	assert.Equal(t, appErr, result) // Should return same instance
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError(
		CodeInvalidFormat,
		"Invalid email",
		OriginWebHandler,
	)

	assert.Equal(t, CodeInvalidFormat, err.Code)
	assert.Equal(t, CategoryValidation, err.Category)
	assert.Equal(t, "Invalid email", err.Message)
	assert.Equal(t, OriginWebHandler, err.Origin)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
}

func TestNewAuthError(t *testing.T) {
	tests := []struct {
		name           string
		code           ErrorCode
		expectedStatus int
	}{
		{
			name:           "unauthorized error",
			code:           CodeUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden error",
			code:           CodeForbidden,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthError(tt.code, "Access denied", OriginAuthService)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, CategoryAuth, err.Category)
			assert.Equal(t, "Access denied", err.Message)
			assert.Equal(t, OriginAuthService, err.Origin)
			assert.Equal(t, tt.expectedStatus, err.HTTPStatus)
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("User", OriginUserService)

	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, CategoryNotFound, err.Category)
	assert.Equal(t, "User not found", err.Message)
	assert.Equal(t, "The requested User could not be found", err.Details)
	assert.Equal(t, OriginUserService, err.Origin)
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("User already exists", OriginUserService)

	assert.Equal(t, CodeConflict, err.Code)
	assert.Equal(t, CategoryConflict, err.Category)
	assert.Equal(t, "User already exists", err.Message)
	assert.Equal(t, OriginUserService, err.Origin)
	assert.Equal(t, http.StatusConflict, err.HTTPStatus)
}

func TestNewBusinessError(t *testing.T) {
	err := NewBusinessError(
		CodeBusinessRuleViolation,
		"Invalid operation",
		OriginProjectService,
	)

	assert.Equal(t, CodeBusinessRuleViolation, err.Code)
	assert.Equal(t, CategoryBusiness, err.Category)
	assert.Equal(t, "Invalid operation", err.Message)
	assert.Equal(t, OriginProjectService, err.Origin)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
}

func TestNewSystemError(t *testing.T) {
	err := NewSystemError(
		CodeDatabaseError,
		"Database connection failed",
		OriginDatabase,
	)

	assert.Equal(t, CodeDatabaseError, err.Code)
	assert.Equal(t, CategorySystem, err.Category)
	assert.Equal(t, "Database connection failed", err.Message)
	assert.Equal(t, OriginDatabase, err.Origin)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
}

func TestNewExternalServiceError(t *testing.T) {
	err := NewExternalServiceError(
		CodeServiceTimeout,
		"External service timeout",
		OriginGitOpsService,
	)

	assert.Equal(t, CodeServiceTimeout, err.Code)
	assert.Equal(t, CategoryExternal, err.Category)
	assert.Equal(t, "External service timeout", err.Message)
	assert.Equal(t, OriginGitOpsService, err.Origin)
	assert.Equal(t, http.StatusBadGateway, err.HTTPStatus)
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "app error",
			err: NewValidationError(
				CodeMissingField,
				"Field required",
				OriginWebHandler,
			),
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr, isAppErr := IsAppError(tt.err)
			if tt.expected {
				require.True(t, isAppErr)
				assert.NotNil(t, appErr)
			} else {
				assert.False(t, isAppErr)
				assert.Nil(t, appErr)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(
		originalErr,
		CodeInternalError,
		"Wrapped error",
		OriginWebHandler,
	)

	assert.Equal(t, CodeInternalError, wrappedErr.Code)
	assert.Equal(t, CategorySystem, wrappedErr.Category)
	assert.Equal(t, "Wrapped error", wrappedErr.Message)
	assert.Equal(t, OriginWebHandler, wrappedErr.Origin)
	assert.Equal(t, http.StatusInternalServerError, wrappedErr.HTTPStatus)
	assert.Equal(t, originalErr, wrappedErr.Cause)
}

func TestErrorCodesAndCategories(t *testing.T) {
	// Test that error codes are unique
	codes := map[ErrorCode]bool{
		CodeInvalidRequest:        false,
		CodeValidationFailed:      false,
		CodeMissingField:          false,
		CodeInvalidFormat:         false,
		CodeInvalidCredentials:    false,
		CodeUnauthorized:          false,
		CodeForbidden:             false,
		CodeTokenExpired:          false,
		CodeTokenInvalid:          false,
		CodeNotFound:              false,
		CodeAlreadyExists:         false,
		CodeConflict:              false,
		CodeResourceLocked:        false,
		CodeBusinessRuleViolation: false,
		CodeQuotaExceeded:         false,
		CodeInvalidState:          false,
		CodeInternalError:         false,
		CodeDatabaseError:         false,
		CodeNetworkError:          false,
		CodeServiceTimeout:        false,
		CodeDependencyError:       false,
	}

	for code := range codes {
		assert.False(t, codes[code], "Duplicate error code detected: %s", code)
		codes[code] = true
	}

	// Test that categories are consistent
	categories := map[ErrorCategory]bool{
		CategoryValidation:    false,
		CategoryAuth:          false,
		CategoryAuthorization: false,
		CategoryNotFound:      false,
		CategoryConflict:      false,
		CategoryBusiness:      false,
		CategorySystem:        false,
		CategoryExternal:      false,
	}

	for category := range categories {
		assert.False(
			t,
			categories[category],
			"Duplicate error category detected: %s",
			category,
		)
		categories[category] = true
	}

	// Test that origins are consistent
	origins := map[ServiceOrigin]bool{
		OriginAuthService:       false,
		OriginUserService:       false,
		OriginProjectService:    false,
		OriginAssetService:      false,
		OriginPipelineService:   false,
		OriginRepositoryService: false,
		OriginJWTService:        false,
		OriginPasswordService:   false,
		OriginGitOpsService:     false,
		OriginDatabase:          false,
		OriginWebHandler:        false,
		OriginMiddleware:        false,
	}

	for origin := range origins {
		assert.False(
			t,
			origins[origin],
			"Duplicate service origin detected: %s",
			origin,
		)
		origins[origin] = true
	}
}
