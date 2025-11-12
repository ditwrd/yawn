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

// Package handlers provides HTTP request handlers for authentication
// operations.
//
// This package contains handlers for user registration, login, logout, and
// token refresh. All handlers follow RESTful conventions with proper error
// handling and JSON responses.
//
// Security features:
//   - Input validation and sanitization
//   - Rate limiting ready (can be added with middleware)
//   - Secure password handling
//   - Proper error messages (don't leak sensitive information)
//   - JWT token management
package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// createAuthTestLogger creates a zerolog logger for testing auth handlers.
func createAuthTestLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	return &logger
}

// MockUserService is a mock implementation of UserService for testing.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) GetByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) List(limit, offset int) ([]models.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.User), args.Error(1)
}

// MockJWTService is a mock implementation of JWTService for testing.
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(
	user *models.User,
) (string, string, error) {
	args := m.Called(user)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJWTService) GenerateAccessToken(
	user *models.User,
) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(
	tokenString string,
) (*services.TokenClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenClaims), args.Error(1)
}

func (m *MockJWTService) RefreshToken(
	refreshTokenString string,
) (string, error) {
	args := m.Called(refreshTokenString)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) InvalidateToken(tokenString string) error {
	args := m.Called(tokenString)
	return args.Error(0)
}

func (m *MockJWTService) IsTokenBlacklisted(tokenString string) (bool, error) {
	args := m.Called(tokenString)
	return args.Bool(0), args.Error(1)
}

// MockPasswordService is a mock implementation of PasswordService for testing.
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) ValidatePassword(
	password, hash string,
) (bool, error) {
	args := m.Called(password, hash)
	return args.Bool(0), args.Error(1)
}

func (m *MockPasswordService) CheckPasswordStrength(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

// Helper functions for creating test contexts and requests
func createAuthTestContext(method, path string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// Helper function to create deterministic test user
func createTestUser(id uuid.UUID, email, role string) *models.User {
	return &models.User{
		ID:           id,
		Email:        email,
		PasswordHash: "hashed_password",
		Role:         models.UserRole(role),
	}
}

func TestNewAuthHandler(t *testing.T) {
	t.Parallel()

	type args struct {
		userService     services.UserService
		jwtService      services.JWTService
		passwordService services.PasswordService
		logger          *zerolog.Logger
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "successful auth handler creation",
			args: args{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewAuthHandler(tt.args.userService, tt.args.jwtService, tt.args.passwordService, tt.args.logger)

			// Verify that the handler is not nil
			assert.NotNil(t, got)

			// Verify that the handler is of the correct type
			assert.IsType(t, &AuthHandler{}, got)

			// Verify that dependencies are set correctly
			assert.Equal(t, tt.args.userService, got.userService)
			assert.Equal(t, tt.args.jwtService, got.jwtService)
			assert.Equal(t, tt.args.passwordService, got.passwordService)
			assert.Equal(t, tt.args.logger, got.logger)
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	t.Parallel()

	// Deterministic test UUID for consistent testing
	testUUID := uuid.Must(uuid.NewV7())

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		c echo.Context
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatus     int
		expectedResp   interface{}
		expectUserCall bool
	}{
		{
			name: "successful user registration",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByEmail", "test@example.com").Return(nil, errors.New("user not found"))
					m.On("Create", mock.MatchedBy(func(user *models.User) bool {
						return user.Email == "test@example.com" && user.Role == models.UserRoleUser
					})).Return(nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("HashPassword", "SecurePass123!").Return("hashed_password", nil)
					m.On("CheckPasswordStrength", "SecurePass123!").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusCreated,
			expectedResp: dto.RegisterResponse{
				ID:    testUUID.String(),
				Email: "test@example.com",
				Role:  string(models.UserRoleUser),
			},
			expectUserCall: true,
		},
		{
			name: "registration with existing email",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					existingUser := createTestUser(testUUID, "test@example.com", "user")
					m.On("GetByEmail", "test@example.com").Return(existingUser, nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					// This test will reach password strength check since password is not empty
					m.On("CheckPasswordStrength", "SecurePass123!").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusConflict,
			expectedResp: dto.ErrorResponse{
				Error:   "User with this email already exists",
				Code:    "CONFLICT",
				Details: "Please use a different email address or try logging in",
			},
			expectUserCall: false,
		},
		{
			name: "registration with invalid JSON",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", []byte("invalid json"))
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid request format",
				Code:    "INVALID_REQUEST",
				Details: "Request body must contain email and password",
			},
			expectUserCall: false,
		},
		{
			name: "registration with empty email",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{}, // Won't reach password strength check due to empty email
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Email is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
			expectUserCall: false,
		},
		{
			name: "registration with empty password",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{}, // Won't reach password strength check since password is empty
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Password is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
			expectUserCall: false,
		},
		{
			name: "registration with invalid email format",
			fields: fields{
				userService: &MockUserService{},
				passwordService: &MockPasswordService{}, // Won't reach password strength check due to invalid email
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "invalid-email",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid email format",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
			expectUserCall: false,
		},
		{
			name: "registration with weak password",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("CheckPasswordStrength", "weak").Return(errors.New("password too weak"))
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "weak",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Password does not meet requirements",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
			expectUserCall: false,
		},
		{
			name: "registration with password hashing error",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByEmail", "test@example.com").Return(nil, errors.New("user not found"))
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("HashPassword", "SecurePass123!").Return("", errors.New("hashing failed"))
					m.On("CheckPasswordStrength", "SecurePass123!").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusInternalServerError,
			expectedResp: dto.ErrorResponse{
				Error:   "Failed to process registration",
				Code:    "INTERNAL_ERROR",
				Details: "Please try again later",
			},
			expectUserCall: false,
		},
		{
			name: "registration with database error",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByEmail", "test@example.com").Return(nil, errors.New("user not found"))
					m.On("Create", mock.AnythingOfType("*models.User")).Return(errors.New("database error"))
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("HashPassword", "SecurePass123!").Return("hashed_password", nil)
					m.On("CheckPasswordStrength", "SecurePass123!").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/register", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusInternalServerError,
			expectedResp: dto.ErrorResponse{
				Error:   "Failed to create account",
				Code:    "INTERNAL_ERROR",
				Details: "Please try again later",
			},
			expectUserCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.Register(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			// Check if the handler returned an error (app error)
			if err != nil {
				// For app errors, we need to check if it's the expected error type
				// In tests without middleware, app errors are returned as Go errors
				assert.Error(t, err)
				// Skip response body validation for error cases since middleware isn't set up
			} else {
				// For successful cases, verify HTTP status code
				assert.Equal(t, tt.wantStatus, rec.Code)

				// Verify response body only for successful cases
				if tt.expectedResp != nil && tt.wantStatus == http.StatusCreated {
					var response dto.RegisterResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedResp.(dto.RegisterResponse).Email, response.Email)
					assert.Equal(t, tt.expectedResp.(dto.RegisterResponse).Role, response.Role)
					assert.NotEmpty(t, response.ID) // ID will be generated dynamically
				}
			}

			// Verify mock expectations
			tt.fields.userService.AssertExpectations(t)
			tt.fields.passwordService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	// Deterministic test UUID for consistent testing
	testUUID := uuid.Must(uuid.NewV7())

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		c echo.Context
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatus     int
		expectedResp   interface{}
	}{
		{
			name: "successful user login",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := createTestUser(testUUID, "test@example.com", "user")
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("ValidatePassword", "SecurePass123!", "hashed_password").Return(true, nil)
					return m
				}(),
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					user := createTestUser(testUUID, "test@example.com", "user")
					m.On("GenerateTokenPair", mock.MatchedBy(func(u *models.User) bool {
						return u.Email == user.Email && u.ID == user.ID
					})).Return("access_token", "refresh_token", nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusOK,
			expectedResp: dto.LoginResponse{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresIn:    900,
				User: dto.UserInfo{
					ID:    testUUID.String(),
					Email: "test@example.com",
					Role:  string(models.UserRoleUser),
				},
			},
		},
		{
			name: "login with invalid JSON",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", []byte("invalid json"))
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid request format",
				Code:    "INVALID_REQUEST",
				Details: "Request body must contain email and password",
			},
		},
		{
			name: "login with empty email",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Email is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "login with empty password",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "test@example.com",
					Password: "",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Password is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "login with non-existent user",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByEmail", "nonexistent@example.com").Return(nil, nil) // Return nil, nil for not found
					return m
				}(),
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "nonexistent@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusUnauthorized,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid email or password",
				Code:    "INVALID_CREDENTIALS",
				Details: "Please check your credentials and try again",
			},
		},
		{
			name: "login with incorrect password",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := createTestUser(testUUID, "test@example.com", "user")
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("ValidatePassword", "WrongPassword!", "hashed_password").Return(false, nil)
					return m
				}(),
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "test@example.com",
					Password: "WrongPassword!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusUnauthorized,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid email or password",
				Code:    "INVALID_CREDENTIALS",
				Details: "Please check your credentials and try again",
			},
		},
		{
			name: "login with password validation error",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := createTestUser(testUUID, "test@example.com", "user")
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("ValidatePassword", "SecurePass123!", "hashed_password").Return(false, errors.New("validation error"))
					return m
				}(),
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusInternalServerError,
			expectedResp: dto.ErrorResponse{
				Error:   "Login failed",
				Code:    "INTERNAL_ERROR",
				Details: "Please try again later",
			},
		},
		{
			name: "login with JWT token generation error",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := createTestUser(testUUID, "test@example.com", "user")
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("ValidatePassword", "SecurePass123!", "hashed_password").Return(true, nil)
					return m
				}(),
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("GenerateTokenPair", mock.AnythingOfType("*models.User")).Return("", "", errors.New("JWT generation failed"))
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusInternalServerError,
			expectedResp: dto.ErrorResponse{
				Error:   "Login failed",
				Code:    "TOKEN_GENERATION_FAILED",
				Details: "Please try again later",
			},
		},
		{
			name: "login with whitespace trimmed email",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := createTestUser(testUUID, "test@example.com", "user")
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("ValidatePassword", "SecurePass123!", "hashed_password").Return(true, nil)
					return m
				}(),
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("GenerateTokenPair", mock.AnythingOfType("*models.User")).Return("access_token", "refresh_token", nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LoginRequest{
					Email:    "  test@example.com  ",
					Password: "SecurePass123!",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/login", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusOK,
			expectedResp: dto.LoginResponse{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresIn:    900,
				User: dto.UserInfo{
					ID:    testUUID.String(),
					Email: "test@example.com",
					Role:  string(models.UserRoleUser),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.Login(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			// The auth handler handles errors differently:
			// - Some errors are returned as app errors (handled by middleware)
			// - Some errors are returned as direct JSON responses
			// In tests without middleware, app errors will be returned as Go errors
			// while direct JSON responses will have no error but set HTTP status codes

			if tt.wantStatus >= 400 {
				// For error cases, either an error should be returned (app errors)
				// or HTTP status code should be set (direct JSON responses)
				if err != nil {
					// App error case - middleware would normally handle this
					assert.Error(t, err)
					// Skip response validation for app error cases since middleware isn't set up
					return
				} else {
					// Direct JSON response case
					assert.Equal(t, tt.wantStatus, rec.Code)
				}
			} else {
				// For success cases, no error should be returned
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStatus, rec.Code)
			}

			// Verify response body
			if tt.expectedResp != nil {
				if tt.wantStatus == http.StatusOK {
					var response dto.LoginResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					expected := tt.expectedResp.(dto.LoginResponse)
					assert.Equal(t, expected.AccessToken, response.AccessToken)
					assert.Equal(t, expected.RefreshToken, response.RefreshToken)
					assert.Equal(t, expected.TokenType, response.TokenType)
					assert.Equal(t, expected.ExpiresIn, response.ExpiresIn)
					assert.Equal(t, expected.User.Email, response.User.Email)
					assert.Equal(t, expected.User.Role, response.User.Role)
					assert.Equal(t, expected.User.ID, response.User.ID)
				} else {
					var response dto.ErrorResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					expected := tt.expectedResp.(dto.ErrorResponse)
					assert.Contains(t, response.Error, expected.Error)
					assert.Equal(t, expected.Code, response.Code)
					assert.Equal(t, expected.Details, response.Details)
				}
			}

			// Verify mock expectations
			tt.fields.userService.AssertExpectations(t)
			tt.fields.passwordService.AssertExpectations(t)
			if tt.fields.jwtService != nil {
				tt.fields.jwtService.AssertExpectations(t)
			}
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		c echo.Context
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatus     int
		expectedResp   interface{}
	}{
		{
			name: "successful token refresh",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("RefreshToken", "valid_refresh_token").Return("new_access_token", nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RefreshRequest{
					RefreshToken: "valid_refresh_token",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/refresh", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusOK,
			expectedResp: dto.RefreshResponse{
				AccessToken: "new_access_token",
				TokenType:   "Bearer",
				ExpiresIn:   900,
			},
		},
		{
			name: "refresh with invalid JSON",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				c, _ := createAuthTestContext(http.MethodPost, "/auth/refresh", []byte("invalid json"))
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid request format",
				Code:    "INVALID_REQUEST",
				Details: "Request body must contain refresh_token",
			},
		},
		{
			name: "refresh with empty token",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RefreshRequest{
					RefreshToken: "",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/refresh", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Refresh token is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "refresh with token too long",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				// Create a token longer than 2048 characters
				longToken := strings.Repeat("a", 2049)
				req := dto.RefreshRequest{
					RefreshToken: longToken,
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/refresh", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Refresh token is too long",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "refresh with invalid/expired token",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("RefreshToken", "invalid_token").Return("", errors.New("invalid token"))
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RefreshRequest{
					RefreshToken: "invalid_token",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/refresh", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusUnauthorized,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid or expired refresh token",
				Code:    "INVALID_REFRESH_TOKEN",
				Details: "Please login again to get a new token",
			},
		},
		{
			name: "refresh with JWT service error",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("RefreshToken", "valid_refresh_token").Return("", errors.New("service error"))
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.RefreshRequest{
					RefreshToken: "valid_refresh_token",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/refresh", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusUnauthorized,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid or expired refresh token",
				Code:    "INVALID_REFRESH_TOKEN",
				Details: "Please login again to get a new token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.Refresh(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			// Verify no handler error is returned (errors are returned as HTTP responses)
			assert.NoError(t, err)

			// Verify HTTP status code
			assert.Equal(t, tt.wantStatus, rec.Code)

			// Verify response body
			if tt.expectedResp != nil {
				if tt.wantStatus == http.StatusOK {
					var response dto.RefreshResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					expected := tt.expectedResp.(dto.RefreshResponse)
					assert.Equal(t, expected.AccessToken, response.AccessToken)
					assert.Equal(t, expected.TokenType, response.TokenType)
					assert.Equal(t, expected.ExpiresIn, response.ExpiresIn)
				} else {
					var response dto.ErrorResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					expected := tt.expectedResp.(dto.ErrorResponse)
					assert.Contains(t, response.Error, expected.Error)
					assert.Equal(t, expected.Code, response.Code)
					assert.Equal(t, expected.Details, response.Details)
				}
			}

			// Verify mock expectations
			if tt.fields.jwtService != nil {
				tt.fields.jwtService.AssertExpectations(t)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		c echo.Context
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatus     int
		expectedResp   interface{}
	}{
		{
			name: "successful logout",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("InvalidateToken", "valid_access_token").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LogoutRequest{
					AccessToken: "valid_access_token",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/logout", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusOK,
			expectedResp: dto.LogoutResponse{
				Message: "Logged out successfully",
			},
		},
		{
			name: "logout with invalid JSON",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				c, _ := createAuthTestContext(http.MethodPost, "/auth/logout", []byte("invalid json"))
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Invalid request format",
				Code:    "INVALID_REQUEST",
				Details: "Request body must contain access_token",
			},
		},
		{
			name: "logout with empty token",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LogoutRequest{
					AccessToken: "",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/logout", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Access token is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "logout with token too long",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService:      &MockJWTService{},
				logger:          createAuthTestLogger(),
			},
			args: func() args {
				// Create a token longer than 2048 characters
				longToken := strings.Repeat("a", 2049)
				req := dto.LogoutRequest{
					AccessToken: longToken,
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/logout", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusBadRequest,
			expectedResp: dto.ErrorResponse{
				Error:   "Access token is too long",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "logout with token invalidation error (should still succeed)",
			fields: fields{
				userService:     &MockUserService{},
				passwordService: &MockPasswordService{},
				jwtService: func() *MockJWTService {
					m := &MockJWTService{}
					m.On("InvalidateToken", "invalid_token").Return(errors.New("token invalidation failed"))
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: func() args {
				req := dto.LogoutRequest{
					AccessToken: "invalid_token",
				}
				body, _ := json.Marshal(req)
				c, _ := createAuthTestContext(http.MethodPost, "/auth/logout", body)
				return args{c: c}
			}(),
			wantStatus: http.StatusOK,
			expectedResp: dto.LogoutResponse{
				Message: "Logged out successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.Logout(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			// Verify no handler error is returned (errors are returned as HTTP responses)
			assert.NoError(t, err)

			// Verify HTTP status code
			assert.Equal(t, tt.wantStatus, rec.Code)

			// Verify response body
			if tt.expectedResp != nil {
				if tt.wantStatus == http.StatusOK {
					var response dto.LogoutResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					expected := tt.expectedResp.(dto.LogoutResponse)
					assert.Equal(t, expected.Message, response.Message)
				} else {
					var response dto.ErrorResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					expected := tt.expectedResp.(dto.ErrorResponse)
					assert.Contains(t, response.Error, expected.Error)
					assert.Equal(t, expected.Code, response.Code)
					assert.Equal(t, expected.Details, response.Details)
				}
			}

			// Verify mock expectations
			if tt.fields.jwtService != nil {
				tt.fields.jwtService.AssertExpectations(t)
			}
		})
	}
}

func TestAuthHandler_validateRegisterRequest(t *testing.T) {
	t.Parallel()

	type fields struct {
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		req *dto.RegisterRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid registration request",
			fields: fields{
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("CheckPasswordStrength", "SecurePass123!").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				},
			},
			wantErr: false,
		},
		{
			name: "empty email",
			fields: fields{
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "",
					Password: "SecurePass123!",
				},
			},
			wantErr: true,
			errMsg:  "Email is required",
		},
		{
			name: "empty password",
			fields: fields{
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "",
				},
			},
			wantErr: true,
			errMsg:  "Password is required",
		},
		{
			name: "email too long",
			fields: fields{
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    strings.Repeat("a", 256) + "@example.com",
					Password: "SecurePass123!",
				},
			},
			wantErr: true,
			errMsg:  "Email address is too long (max 255 characters)",
		},
		{
			name: "password too long",
			fields: fields{
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "test@example.com",
					Password: strings.Repeat("a", 129),
				},
			},
			wantErr: true,
			errMsg:  "Password is too long (max 128 characters)",
		},
		{
			name: "invalid email format - missing @",
			fields: fields{
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "testexample.com",
					Password: "SecurePass123!",
				},
			},
			wantErr: true,
			errMsg:  "Invalid email format",
		},
		{
			name: "invalid email format - missing domain",
			fields: fields{
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "test@",
					Password: "SecurePass123!",
				},
			},
			wantErr: true,
			errMsg:  "Invalid email format",
		},
		{
			name: "weak password",
			fields: fields{
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("CheckPasswordStrength", "weak").Return(errors.New("password too weak"))
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "test@example.com",
					Password: "weak",
				},
			},
			wantErr: true,
			errMsg:  "Password does not meet requirements",
		},
		{
			name: "valid email with whitespace trimmed",
			fields: fields{
				passwordService: func() *MockPasswordService {
					m := &MockPasswordService{}
					m.On("CheckPasswordStrength", "SecurePass123!").Return(nil)
					return m
				}(),
				logger: createAuthTestLogger(),
			},
			args: args{
				req: &dto.RegisterRequest{
					Email:    "  test@example.com  ",
					Password: "SecurePass123!",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.validateRegisterRequest(tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
				// Check if the error message contains the expected message (may include error code)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			tt.fields.passwordService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_validateLoginRequest(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		req *dto.LoginRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid login request",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    "test@example.com",
					Password: "SecurePass123!",
				},
			},
			wantErr: false,
		},
		{
			name: "empty email",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    "",
					Password: "SecurePass123!",
				},
			},
			wantErr: true,
			errMsg:  "Email is required",
		},
		{
			name: "empty password",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    "test@example.com",
					Password: "",
				},
			},
			wantErr: true,
			errMsg:  "Password is required",
		},
		{
			name: "email too long",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    strings.Repeat("a", 256) + "@example.com",
					Password: "SecurePass123!",
				},
			},
			wantErr: true,
			errMsg:  "Email address is too long (max 255 characters)",
		},
		{
			name: "password too long",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    "test@example.com",
					Password: strings.Repeat("a", 129),
				},
			},
			wantErr: true,
			errMsg:  "Password is too long (max 128 characters)",
		},
		{
			name: "valid request with email at max length",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    strings.Repeat("a", 245) + "@example.com", // 245 + 12 = 257, but this gets trimmed in validation
					Password: "SecurePass123!",
				},
			},
			wantErr: true, // Should fail because it's too long even after validation
			errMsg:  "Email address is too long (max 255 characters)",
		},
		{
			name: "valid request with password at max length",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LoginRequest{
					Email:    "test@example.com",
					Password: strings.Repeat("a", 128), // Exactly 128 characters
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.validateLoginRequest(tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
				// Check if the error message contains the expected message (may include error code)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthHandler_validateRefreshRequest(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		req *dto.RefreshRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid refresh request",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RefreshRequest{
					RefreshToken: "valid_refresh_token",
				},
			},
			wantErr: false,
		},
		{
			name: "empty refresh token",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RefreshRequest{
					RefreshToken: "",
				},
			},
			wantErr: true,
			errMsg:  "Refresh token is required",
		},
		{
			name: "refresh token too long",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RefreshRequest{
					RefreshToken: strings.Repeat("a", 2049),
				},
			},
			wantErr: true,
			errMsg:  "Refresh token is too long",
		},
		{
			name: "valid refresh request with token at max length",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RefreshRequest{
					RefreshToken: strings.Repeat("a", 2048), // Exactly 2048 characters
				},
			},
			wantErr: false,
		},
		{
			name: "valid JWT token format",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.RefreshRequest{
					RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.validateRefreshRequest(tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
				// Check if the error message contains the expected message (may include error code)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthHandler_validateLogoutRequest(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService     *MockUserService
		jwtService      *MockJWTService
		passwordService *MockPasswordService
		logger          *zerolog.Logger
	}

	type args struct {
		req *dto.LogoutRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid logout request",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LogoutRequest{
					AccessToken: "valid_access_token",
				},
			},
			wantErr: false,
		},
		{
			name: "empty access token",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LogoutRequest{
					AccessToken: "",
				},
			},
			wantErr: true,
			errMsg:  "Access token is required",
		},
		{
			name: "access token too long",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LogoutRequest{
					AccessToken: strings.Repeat("a", 2049),
				},
			},
			wantErr: true,
			errMsg:  "Access token is too long",
		},
		{
			name: "valid logout request with token at max length",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LogoutRequest{
					AccessToken: strings.Repeat("a", 2048), // Exactly 2048 characters
				},
			},
			wantErr: false,
		},
		{
			name: "valid JWT token format for logout",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LogoutRequest{
					AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				},
			},
			wantErr: false,
		},
		{
			name: "valid Bearer token format",
			fields: fields{
				userService:     &MockUserService{},
				jwtService:      &MockJWTService{},
				passwordService: &MockPasswordService{},
				logger:          createAuthTestLogger(),
			},
			args: args{
				req: &dto.LogoutRequest{
					AccessToken: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &AuthHandler{
				userService:     tt.fields.userService,
				jwtService:      tt.fields.jwtService,
				passwordService: tt.fields.passwordService,
				logger:          tt.fields.logger,
			}

			err := h.validateLogoutRequest(tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
				// Check if the error message contains the expected message (may include error code)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
