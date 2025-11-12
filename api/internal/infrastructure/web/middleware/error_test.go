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

package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/ditwrd/yawn/api/internal/domain/errors"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

func TestErrorHandler_HandleAppError(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	appErr := apperrors.NewValidationError(
		apperrors.CodeMissingField,
		"Email is required",
		apperrors.OriginWebHandler,
	).WithDetails("Please provide a valid email address")

	handler.HandleError(appErr, c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(
		t,
		echo.MIMEApplicationJSONCharsetUTF8,
		rec.Header().Get(echo.HeaderContentType),
	)

	var response dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Email is required", response.Error)
	assert.Equal(t, string(apperrors.CodeMissingField), response.Code)
	assert.Equal(t, "Please provide a valid email address", response.Details)
}

func TestErrorHandler_HandleValidationError(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	// Create mock validation errors
	validationErrors := []validator.FieldError{
		&mockFieldError{field: "Email", tag: "email", param: ""},
		&mockFieldError{field: "Password", tag: "min", param: "8"},
	}

	// Create a validator.ValidationErrors from our mock errors
	validatorErrors := validator.ValidationErrors(validationErrors)
	handler.HandleError(validatorErrors, c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response dto.ValidationErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Validation failed", response.Error)
	assert.Equal(t, "VALIDATION_FAILED", response.Code)
	assert.Contains(t, response.Fields, "Email")
	assert.Contains(t, response.Fields, "Password")
}

func TestErrorHandler_HandleHTTPError(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	httpErr := echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	handler.HandleError(httpErr, c)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Resource not found", response.Error)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

func TestErrorHandler_HandleGORMErrors(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	// Test RecordNotFound error
	handler.HandleError(gorm.ErrRecordNotFound, c)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Resource not found", response.Error)
	assert.Equal(t, "NOT_FOUND", response.Code)

	// Reset recorder for next test
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	// Test DuplicateKey error
	handler.HandleError(gorm.ErrDuplicatedKey, c)
	assert.Equal(t, http.StatusConflict, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Resource already exists", response.Error)
	assert.Equal(t, "ALREADY_EXISTS", response.Code)
}

func TestErrorHandler_HandleGenericError(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	genericErr := errors.New("something went wrong")
	handler.HandleError(genericErr, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Internal server error", response.Error)
	assert.Equal(t, "INTERNAL_ERROR", response.Code)
}

func TestErrorHandler_getValidationErrorMessage(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	tests := []struct {
		name     string
		fieldErr validator.FieldError
		expected string
	}{
		{
			name: "required field",
			fieldErr: &mockFieldError{
				field: "Email",
				tag:   "required",
				param: "",
			},
			expected: "Email is required",
		},
		{
			name: "email format",
			fieldErr: &mockFieldError{
				field: "Email",
				tag:   "email",
				param: "",
			},
			expected: "Please enter a valid email address",
		},
		{
			name: "min length 1",
			fieldErr: &mockFieldError{
				field: "Name",
				tag:   "min",
				param: "1",
			},
			expected: "Name must not be empty",
		},
		{
			name: "min length",
			fieldErr: &mockFieldError{
				field: "Password",
				tag:   "min",
				param: "8",
			},
			expected: "Password must be at least 8 characters",
		},
		{
			name: "max length",
			fieldErr: &mockFieldError{
				field: "Username",
				tag:   "max",
				param: "20",
			},
			expected: "Username must be at most 20 characters",
		},
		{
			name: "exact length",
			fieldErr: &mockFieldError{
				field: "Code",
				tag:   "len",
				param: "10",
			},
			expected: "Code must be exactly 10 characters",
		},
		{
			name: "one of",
			fieldErr: &mockFieldError{
				field: "Role",
				tag:   "oneof",
				param: "admin user",
			},
			expected: "Role must be one of: admin user",
		},
		{
			name: "uuid",
			fieldErr: &mockFieldError{
				field: "UserID",
				tag:   "uuid",
				param: "",
			},
			expected: "UserID must be a valid UUID",
		},
		{
			name: "url",
			fieldErr: &mockFieldError{
				field: "Website",
				tag:   "url",
				param: "",
			},
			expected: "Website must be a valid URL",
		},
		{
			name: "unknown tag",
			fieldErr: &mockFieldError{
				field: "Field",
				tag:   "unknown",
				param: "",
			},
			expected: "Field is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.getValidationErrorMessage(tt.fieldErr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandler_getErrorCodeFromStatus(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	tests := []struct {
		status   int
		expected string
	}{
		{http.StatusBadRequest, "INVALID_REQUEST"},
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusForbidden, "FORBIDDEN"},
		{http.StatusNotFound, "NOT_FOUND"},
		{http.StatusConflict, "CONFLICT"},
		{http.StatusTooManyRequests, "TOO_MANY_REQUESTS"},
		{http.StatusInternalServerError, "INTERNAL_ERROR"},
		{http.StatusBadGateway, "BAD_GATEWAY"},
		{http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE"},
		{http.StatusCreated, "UNKNOWN_ERROR"}, // Test unknown status
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			result := handler.getErrorCodeFromStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorMiddleware(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	middleware := ErrorMiddleware(&logger)

	e := echo.New()
	e.Use(middleware)

	// Handler that returns an AppError
	e.GET("/test", func(c echo.Context) error {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Test error",
			apperrors.OriginWebHandler,
		)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Test error", response.Error)
	assert.Equal(t, string(apperrors.CodeMissingField), response.Code)
}

// Mock field error for testing
type mockFieldError struct {
	field string
	tag   string
	param string
}

func (m *mockFieldError) Field() string {
	return m.field
}

func (m *mockFieldError) Tag() string {
	return m.tag
}

func (m *mockFieldError) Param() string {
	return m.param
}

func (m *mockFieldError) Value() interface{} {
	return ""
}

func (m *mockFieldError) ActualTag() string {
	return m.tag
}

func (m *mockFieldError) Namespace() string {
	return m.field
}

func (m *mockFieldError) StructNamespace() string {
	return m.field
}

func (m *mockFieldError) StructField() string {
	return m.field
}

func (m *mockFieldError) Translate(_ ut.Translator) string {
	return ""
}

func (m *mockFieldError) Error() string {
	return "validation error"
}

func (m *mockFieldError) Kind() reflect.Kind {
	return reflect.String
}

func (m *mockFieldError) Type() reflect.Type {
	return reflect.TypeOf("")
}

func TestErrorHandler_CommittedResponse(t *testing.T) {
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewErrorHandler(&logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "test-request-id")

	// Manually commit the response
	c.Response().WriteHeader(http.StatusOK)

	appErr := apperrors.NewValidationError(
		apperrors.CodeMissingField,
		"Test error",
		apperrors.OriginWebHandler,
	)

	// Should not try to write to committed response
	handler.HandleError(appErr, c)

	// Status should remain as originally set
	assert.Equal(t, http.StatusOK, rec.Code)
}
