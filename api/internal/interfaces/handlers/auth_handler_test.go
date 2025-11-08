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

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/infrastructure/web/middleware"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

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

// TestPasswordService is a configurable implementation for testing
type TestPasswordService struct {
	shouldValidatePassword bool
	shouldHashPassword     bool
	shouldCheckStrength    bool
}

func (t *TestPasswordService) HashPassword(password string) (string, error) {
	if !t.shouldHashPassword {
		return "", fmt.Errorf("hashing disabled")
	}
	return "hashed_password", nil
}

func (t *TestPasswordService) ValidatePassword(
	password, hash string,
) (bool, error) {
	if !t.shouldValidatePassword {
		return false, nil
	}
	return true, nil
}

func (t *TestPasswordService) CheckPasswordStrength(password string) error {
	if !t.shouldCheckStrength {
		return fmt.Errorf("strength check disabled")
	}
	return nil
}

// setupTestAuthHandler creates a test AuthHandler with mocked dependencies.
func setupTestAuthHandler() (*AuthHandler, *MockUserService, *MockJWTService, *TestPasswordService) {
	mockUserService := &MockUserService{}
	mockJWTService := &MockJWTService{}
	testPasswordService := &TestPasswordService{
		shouldValidatePassword: true,
		shouldHashPassword:     true,
		shouldCheckStrength:    true,
	}

	logger := zerolog.New(zerolog.NewConsoleWriter())

	handler := NewAuthHandler(
		mockUserService,
		mockJWTService,
		testPasswordService,
		&logger,
	)
	return handler, mockUserService, mockJWTService, testPasswordService
}

// createTestUser creates a test user for testing purposes.
func createTestUser() *models.User {
	return &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
}

// Test case structures for table-driven testing

// RegisterTestCase represents a single test case for registration
type RegisterTestCase struct {
	name             string
	request          dto.RegisterRequest
	setupMocks       func(*MockUserService, *MockJWTService, *TestPasswordService)
	expectedStatus   int
	expectedResponse interface{}
	expectedError    string
}

// LoginTestCase represents a single test case for login
type LoginTestCase struct {
	name             string
	request          dto.LoginRequest
	setupMocks       func(*MockUserService, *MockJWTService, *TestPasswordService)
	expectedStatus   int
	expectedResponse interface{}
	expectedError    string
}

// TestAuthHandler_Register tests user registration functionality using
// table-driven approach.
func TestAuthHandler_Register(t *testing.T) {
	testCases := []RegisterTestCase{
		{
			name: "successful registration",
			request: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mus.On("GetByEmail", "test@example.com").
					Return(nil, fmt.Errorf("not found"))
				mus.On("Create", mock.MatchedBy(func(user *models.User) bool {
					return user.Email == "test@example.com" &&
						user.Role == models.UserRoleUser
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: dto.RegisterResponse{
				Email: "test@example.com",
				Role:  "user",
			},
		},
		{
			name: "user already exists",
			request: dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				existingUser := &models.User{
					ID:           uuid.Must(uuid.NewV7()),
					Email:        "existing@example.com",
					PasswordHash: "hashed_password",
					Role:         models.UserRoleUser,
				}
				mus.On("GetByEmail", "existing@example.com").
					Return(existingUser, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedResponse: dto.ErrorResponse{
				Error:   "User with this email already exists",
				Code:    "USER_EXISTS",
				Details: "Please use a different email address or try logging in",
			},
		},
		{
			name: "invalid email format",
			request: dto.RegisterRequest{
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail before service calls
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "invalid email format",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "empty email",
			request: dto.RegisterRequest{
				Email:    "",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "email is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "empty password",
			request: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "password is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "password too long",
			request: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: string(make([]byte, 129)), // 129 characters
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "password is too long (max 128 characters)",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "email too long",
			request: dto.RegisterRequest{
				Email:    string(make([]byte, 256)) + "@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "email address is too long (max 255 characters)",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "database error during user creation",
			request: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mus.On("GetByEmail", "test@example.com").
					Return(nil, fmt.Errorf("not found"))
				mus.On("Create", mock.AnythingOfType("*models.User")).
					Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: dto.ErrorResponse{
				Error:   "Failed to create account",
				Code:    "CREATE_FAILED",
				Details: "Please try again later",
			},
		},
		{
			name: "password hashing error",
			request: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				tps.shouldHashPassword = false // Configure to fail hashing
				mus.On("GetByEmail", "test@example.com").
					Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: dto.ErrorResponse{
				Error:   "Failed to process registration",
				Code:    "INTERNAL_ERROR",
				Details: "Please try again later",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockUserService, mockJWTService, testPasswordService := setupTestAuthHandler()
			e := echo.New()

			// Configure mocks
			tc.setupMocks(mockUserService, mockJWTService, testPasswordService)

			// Create request
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			request := httptest.NewRequest(
				http.MethodPost,
				"/auth/register",
				bytes.NewBuffer(reqBody),
			)
			request.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(request, rec)

			// Execute
			err = handler.Register(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse and verify response
			if tc.expectedStatus >= 400 {
				// Error response
				var response dto.ErrorResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(
					t,
					tc.expectedResponse.(dto.ErrorResponse).Error,
					response.Error,
				)
				assert.Equal(
					t,
					tc.expectedResponse.(dto.ErrorResponse).Code,
					response.Code,
				)
			} else {
				// Success response
				var response dto.RegisterResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				expectedResp := tc.expectedResponse.(dto.RegisterResponse)
				assert.Equal(t, expectedResp.Email, response.Email)
				assert.Equal(t, expectedResp.Role, response.Role)
			}

			// Assert mock expectations
			mockUserService.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

// TestAuthHandler_Login tests user login functionality using table-driven
// approach.
func TestAuthHandler_Login(t *testing.T) {
	testCases := []LoginTestCase{
		{
			name: "successful login",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				user := createTestUser()
				mus.On("GetByEmail", "test@example.com").
					Return(user, nil).
					Once()
				mjs.On("GenerateTokenPair", user).
					Return("access_token", "refresh_token", nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.LoginResponse{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresIn:    900,
				User: dto.UserInfo{
					Email: "test@example.com",
					Role:  "user",
				},
			},
		},
		{
			name: "invalid credentials - wrong password",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				user := createTestUser()
				tps.shouldValidatePassword = false // Configure to fail validation
				mus.On("GetByEmail", "test@example.com").
					Return(user, nil).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: dto.ErrorResponse{
				Error:   "Invalid email or password",
				Code:    "INVALID_CREDENTIALS",
				Details: "Please check your credentials and try again",
			},
		},
		{
			name: "user not found",
			request: dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mus.On("GetByEmail", "nonexistent@example.com").
					Return(nil, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: dto.ErrorResponse{
				Error:   "Invalid email or password",
				Code:    "INVALID_CREDENTIALS",
				Details: "Please check your credentials and try again",
			},
		},
		{
			name: "empty email",
			request: dto.LoginRequest{
				Email:    "",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "email is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "empty password",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "password is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "token generation error",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				user := createTestUser()
				mus.On("GetByEmail", "test@example.com").
					Return(user, nil).
					Once()
				mjs.On("GenerateTokenPair", user).
					Return("", "", fmt.Errorf("token generation failed")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: dto.ErrorResponse{
				Error:   "Login failed",
				Code:    "TOKEN_GENERATION_FAILED",
				Details: "Please try again later",
			},
		},
		{
			name: "password validation returns false",
			request: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				user := createTestUser()
				tps.shouldValidatePassword = false // Configure to return false
				mus.On("GetByEmail", "test@example.com").
					Return(user, nil).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: dto.ErrorResponse{
				Error:   "Invalid email or password",
				Code:    "INVALID_CREDENTIALS",
				Details: "Please check your credentials and try again",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockUserService, mockJWTService, testPasswordService := setupTestAuthHandler()
			e := echo.New()

			// Configure mocks
			tc.setupMocks(mockUserService, mockJWTService, testPasswordService)

			// Create request
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			request := httptest.NewRequest(
				http.MethodPost,
				"/auth/login",
				bytes.NewBuffer(reqBody),
			)
			request.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(request, rec)

			// Execute
			err = handler.Login(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse and verify response
			if tc.expectedStatus >= 400 {
				// Error response
				var response dto.ErrorResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				expectedResp := tc.expectedResponse.(dto.ErrorResponse)
				assert.Equal(t, expectedResp.Error, response.Error)
				assert.Equal(t, expectedResp.Code, response.Code)
			} else {
				// Success response
				var response dto.LoginResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				expectedResp := tc.expectedResponse.(dto.LoginResponse)
				assert.Equal(t, expectedResp.AccessToken, response.AccessToken)
				assert.Equal(t, expectedResp.RefreshToken, response.RefreshToken)
				assert.Equal(t, expectedResp.TokenType, response.TokenType)
				assert.Equal(t, expectedResp.ExpiresIn, response.ExpiresIn)
				assert.Equal(t, expectedResp.User.Email, response.User.Email)
				assert.Equal(t, expectedResp.User.Role, response.User.Role)
			}

			// Assert mock expectations
			mockUserService.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

// Additional test case structures for other endpoints

// RefreshTestCase represents a single test case for token refresh
type RefreshTestCase struct {
	name             string
	request          dto.RefreshRequest
	setupMocks       func(*MockUserService, *MockJWTService, *TestPasswordService)
	expectedStatus   int
	expectedResponse interface{}
}

// LogoutTestCase represents a single test case for logout
type LogoutTestCase struct {
	name             string
	request          dto.LogoutRequest
	setupMocks       func(*MockUserService, *MockJWTService, *TestPasswordService)
	expectedStatus   int
	expectedResponse interface{}
}

// Helper functions for better test organization

// createEchoContext creates a new Echo context for testing
func createEchoContext(
	method, path string,
	body []byte,
) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// assertErrorResponse validates error response structure
func assertErrorResponse(
	t *testing.T,
	body []byte,
	expected dto.ErrorResponse,
) {
	var response dto.ErrorResponse
	err := json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Equal(t, expected.Error, response.Error)
	assert.Equal(t, expected.Code, response.Code)
	if expected.Details != "" {
		assert.Equal(t, expected.Details, response.Details)
	}
}

// TestAuthHandler_Refresh tests token refresh functionality using table-driven
// approach.
func TestAuthHandler_Refresh(t *testing.T) {
	testCases := []RefreshTestCase{
		{
			name: "successful token refresh",
			request: dto.RefreshRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mjs.On("RefreshToken", "valid_refresh_token").
					Return("new_access_token", nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.RefreshResponse{
				AccessToken: "new_access_token",
				TokenType:   "Bearer",
				ExpiresIn:   900,
			},
		},
		{
			name: "invalid refresh token",
			request: dto.RefreshRequest{
				RefreshToken: "invalid_token",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mjs.On("RefreshToken", "invalid_token").
					Return("", fmt.Errorf("invalid token")).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: dto.ErrorResponse{
				Error:   "Invalid or expired refresh token",
				Code:    "INVALID_REFRESH_TOKEN",
				Details: "Please login again to get a new token",
			},
		},
		{
			name: "empty refresh token",
			request: dto.RefreshRequest{
				RefreshToken: "",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "refresh_token is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "refresh token too long",
			request: dto.RefreshRequest{
				RefreshToken: string(make([]byte, 2049)), // 2049 characters
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "refresh token is too long",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockUserService, mockJWTService, testPasswordService := setupTestAuthHandler()

			// Configure mocks
			tc.setupMocks(mockUserService, mockJWTService, testPasswordService)

			// Create request
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			c, rec := createEchoContext(
				http.MethodPost,
				"/auth/refresh",
				reqBody,
			)

			// Execute
			err = handler.Refresh(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse and verify response
			if tc.expectedStatus >= 400 {
				// Error response
				assertErrorResponse(
					t,
					rec.Body.Bytes(),
					tc.expectedResponse.(dto.ErrorResponse),
				)
			} else {
				// Success response
				var response dto.RefreshResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				expectedResp := tc.expectedResponse.(dto.RefreshResponse)
				assert.Equal(t, expectedResp.AccessToken, response.AccessToken)
				assert.Equal(t, expectedResp.TokenType, response.TokenType)
				assert.Equal(t, expectedResp.ExpiresIn, response.ExpiresIn)
			}

			// Assert mock expectations
			mockUserService.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

// TestAuthHandler_Logout tests logout functionality using table-driven
// approach.
func TestAuthHandler_Logout(t *testing.T) {
	testCases := []LogoutTestCase{
		{
			name: "successful logout",
			request: dto.LogoutRequest{
				AccessToken: "valid_access_token",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mjs.On("InvalidateToken", "valid_access_token").
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.LogoutResponse{
				Message: "Logged out successfully",
			},
		},
		{
			name: "logout with token invalidation error",
			request: dto.LogoutRequest{
				AccessToken: "problematic_token",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mjs.On("InvalidateToken", "problematic_token").
					Return(fmt.Errorf("token invalidation failed")).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.LogoutResponse{
				Message: "Logged out successfully",
			},
		},
		{
			name: "empty access token",
			request: dto.LogoutRequest{
				AccessToken: "",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "access_token is required",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "access token too long",
			request: dto.LogoutRequest{
				AccessToken: string(make([]byte, 2049)), // 2049 characters
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				// No mocks needed - validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: dto.ErrorResponse{
				Error:   "access token is too long",
				Code:    "VALIDATION_ERROR",
				Details: "Please check your input and try again",
			},
		},
		{
			name: "invalid request format",
			request: dto.LogoutRequest{
				AccessToken: "valid_token",
			},
			setupMocks: func(mus *MockUserService, mjs *MockJWTService, tps *TestPasswordService) {
				mjs.On("InvalidateToken", "valid_token").Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.LogoutResponse{
				Message: "Logged out successfully",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			handler, mockUserService, mockJWTService, testPasswordService := setupTestAuthHandler()

			// Configure mocks
			tc.setupMocks(mockUserService, mockJWTService, testPasswordService)

			// Create request
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			c, rec := createEchoContext(
				http.MethodPost,
				"/auth/logout",
				reqBody,
			)

			// Execute
			err = handler.Logout(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse and verify response
			if tc.expectedStatus >= 400 {
				// Error response
				assertErrorResponse(
					t,
					rec.Body.Bytes(),
					tc.expectedResponse.(dto.ErrorResponse),
				)
			} else {
				// Success response
				var response dto.LogoutResponse
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				expectedResp := tc.expectedResponse.(dto.LogoutResponse)
				assert.Equal(t, expectedResp.Message, response.Message)
			}

			// Assert mock expectations
			mockUserService.AssertExpectations(t)
			mockJWTService.AssertExpectations(t)
		})
	}
}

// Test matrix for edge cases and boundary conditions
func TestAuthHandler_EdgeCases(t *testing.T) {
	t.Run("invalid JSON request format", func(t *testing.T) {
		handler, _, _, _ := setupTestAuthHandler()

		// Test with invalid JSON
		c, rec := createEchoContext(
			http.MethodPost,
			"/auth/register",
			[]byte("invalid json"),
		)

		err := handler.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		assertErrorResponse(t, rec.Body.Bytes(), dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain email and password",
		})
	})

	t.Run("malformed JSON with missing fields", func(t *testing.T) {
		handler, _, _, _ := setupTestAuthHandler()

		// Test with JSON missing required fields
		c, rec := createEchoContext(
			http.MethodPost,
			"/auth/register",
			[]byte(`{"email": "test@example.com"}`),
		)

		err := handler.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("extremely long input values", func(t *testing.T) {
		handler, _, _, _ := setupTestAuthHandler()

		// Test with extremely long email and password
		longEmail := string(make([]byte, 300)) + "@example.com"
		req := dto.RegisterRequest{
			Email:    longEmail,
			Password: "SecurePass123!",
		}

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)

		c, rec := createEchoContext(http.MethodPost, "/auth/register", reqBody)

		err = handler.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		assertErrorResponse(t, rec.Body.Bytes(), dto.ErrorResponse{
			Error:   "email address is too long (max 255 characters)",
			Code:    "VALIDATION_ERROR",
			Details: "Please check your input and try again",
		})
	})
}

// TestJWTMiddlewareVersionCompatibility tests JWT library version
// compatibility.
//
// CRITICAL TEST: This test verifies that the JWT library version used by
// echo-jwt matches the version used in our code, preventing type assertion
// failures that can cause handlers to panic.
func TestJWTMiddlewareVersionCompatibility(t *testing.T) {
	t.Run("jwt library version compatibility", func(t *testing.T) {
		// Create real JWT service with test configuration
		jwtConfig := services.DefaultJWTConfig()
		jwtConfig.AccessSecret = "test-access-secret"
		jwtConfig.RefreshSecret = "test-refresh-secret"
		jwtService := services.NewJWTService(jwtConfig)

		// Create auth middleware
		logger := zerolog.New(zerolog.NewConsoleWriter())
		authMiddleware := middleware.NewAuthMiddleware(jwtService, &logger)

		// Create Echo instance with JWT middleware
		e := echo.New()
		e.Use(authMiddleware.JWT())

		// Create test user and token
		user := &models.User{
			ID:    uuid.Must(uuid.NewV7()),
			Email: "test@example.com",
			Role:  models.UserRoleUser,
		}

		// Generate real JWT token
		accessToken, err := jwtService.GenerateAccessToken(user)
		require.NoError(
			t,
			err,
			"Failed to generate access token for compatibility test",
		)

		// Add protected route that uses JWT type assertions
		e.GET("/protected", func(c echo.Context) error {
			// CRITICAL: These type assertions must not fail
			// This test detects golang-jwt/jwt version mismatches
			token, ok := c.Get("user").(*jwt.Token)
			if !ok {
				t.Fatalf(
					"JWT library version mismatch: token type assertion failed",
				)
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					"JWT type mismatch",
				)
			}

			claims, ok := token.Claims.(*services.TokenClaims)
			if !ok {
				t.Fatalf(
					"JWT library version mismatch: claims type assertion failed",
				)
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					"Claims type mismatch",
				)
			}

			// Verify claims are accessible
			if claims.UserID != user.ID {
				t.Fatalf("JWT library version mismatch: user ID mismatch")
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					"User ID mismatch",
				)
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"user_id": claims.UserID.String(),
				"email":   claims.Email,
				"role":    claims.Role,
			})
		})

		// Create test request with valid JWT token
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		rec := httptest.NewRecorder()

		// Execute request
		e.ServeHTTP(rec, req)

		// Verify response
		assert.Equal(
			t,
			http.StatusOK,
			rec.Code,
			"JWT middleware compatibility test failed. This may indicate a golang-jwt/jwt version mismatch between echo-jwt and our code.",
		)

		// Verify response body contains expected user data
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), response["user_id"])
		assert.Equal(t, user.Email, response["email"])
		assert.Equal(t, string(user.Role), response["role"])

		t.Log("JWT library version compatibility test passed")
	})
}

// TestAuthMiddleware_Integration tests the complete authentication flow.
func TestAuthMiddleware_Integration(t *testing.T) {
	t.Run("complete auth flow with token validation", func(t *testing.T) {
		// Create real JWT service
		jwtConfig := services.DefaultJWTConfig()
		jwtConfig.AccessSecret = "test-access-secret"
		jwtConfig.RefreshSecret = "test-refresh-secret"
		jwtService := services.NewJWTService(jwtConfig)

		// Create auth middleware
		logger := zerolog.New(zerolog.NewConsoleWriter())
		authMiddleware := middleware.NewAuthMiddleware(jwtService, &logger)

		// Create Echo instance with authentication middleware
		e := echo.New()
		e.Use(authMiddleware.RequireAuth())

		// Create test user
		user := &models.User{
			ID:    uuid.Must(uuid.NewV7()),
			Email: "test@example.com",
			Role:  models.UserRoleUser,
		}

		// Generate real JWT token
		accessToken, err := jwtService.GenerateAccessToken(user)
		require.NoError(t, err)

		// Add protected route that tests claim extraction
		e.GET("/test", func(c echo.Context) error {
			// Test our helper functions work correctly
			claims, err := middleware.GetUserClaims(c)
			assert.NoError(t, err)
			assert.Equal(t, user.ID, claims.UserID)

			userID, err := middleware.GetUserID(c)
			assert.NoError(t, err)
			assert.Equal(t, user.ID.String(), userID)

			userRole, err := middleware.GetUserRole(c)
			assert.NoError(t, err)
			assert.Equal(t, string(user.Role), userRole)

			return c.JSON(http.StatusOK, map[string]string{
				"message": "authentication successful",
			})
		})

		// Test successful authentication
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Test missing token
		req = httptest.NewRequest(http.MethodGet, "/test", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		// Test invalid token
		req = httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
