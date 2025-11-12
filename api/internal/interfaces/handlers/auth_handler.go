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
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	apperrors "github.com/ditwrd/yawn/api/internal/domain/errors"
	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userService     services.UserService
	jwtService      services.JWTService
	passwordService services.PasswordService
	logger          *zerolog.Logger
}

// NewAuthHandler creates a new AuthHandler instance.
//
// Parameters:
//   - userService: User service for user operations
//   - jwtService: JWT service for token management
//   - passwordService: Password service for password operations
//   - logger: Logger for structured logging
//
// Returns:
//   - *AuthHandler: An instance of the auth handler
func NewAuthHandler(
	userService services.UserService,
	jwtService services.JWTService,
	passwordService services.PasswordService,
	logger *zerolog.Logger,
) *AuthHandler {
	return &AuthHandler{
		userService:     userService,
		jwtService:      jwtService,
		passwordService: passwordService,
		logger:          logger,
	}
}

// Register handles user registration requests.
//
// @Summary Register a new user
// @Description Creates a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post].
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind registration request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain email and password",
		})
	}

	// Validate request
	if err := h.validateRegisterRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
			Details: "Please check your input and try again",
		})
	}

	// Check if user already exists
	existingUser, err := h.userService.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		h.logger.Info().
			Str("email", req.Email).
			Msg("Registration attempt with existing email")

		return apperrors.NewConflictError(
			"User with this email already exists",
			apperrors.OriginWebHandler,
		).WithDetails("Please use a different email address or try logging in")
	}

	// Hash password
	passwordHash, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to hash password during registration")

		return apperrors.NewSystemError(
			apperrors.CodeInternalError,
			"Failed to process registration",
			apperrors.OriginWebHandler,
		).WithCause(err).WithDetails("Please try again later")
	}

	// Create user
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: passwordHash,
		Role:         models.UserRoleUser, // Default role
	}

	err = h.userService.Create(user)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("email", user.Email).
			Msg("Failed to create user during registration")

		return apperrors.NewSystemError(
			apperrors.CodeInternalError,
			"Failed to create account",
			apperrors.OriginWebHandler,
		).WithCause(err).WithDetails("Please try again later")
	}

	h.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("User registered successfully")

	// Return success response
	return c.JSON(http.StatusCreated, dto.RegisterResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  string(user.Role),
	})
}

// Login handles user login requests.
//
// @Summary User login
// @Description Authenticates a user and returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post].
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind login request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain email and password",
		})
	}

	// Validate request
	if err := h.validateLoginRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
			Details: "Please check your input and try again",
		})
	}

	// Find user by email
	user, err := h.userService.GetByEmail(
		strings.ToLower(strings.TrimSpace(req.Email)),
	)
	if err != nil || user == nil {
		h.logger.Info().
			Str("email", req.Email).
			Msg("Login attempt with non-existent email")

		return apperrors.NewAuthError(
			apperrors.CodeInvalidCredentials,
			"Invalid email or password",
			apperrors.OriginWebHandler,
		).WithDetails("Please check your credentials and try again")
	}

	// Validate password
	valid, err := h.passwordService.ValidatePassword(
		req.Password,
		user.PasswordHash,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to validate password during login")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Login failed",
			Code:    "INTERNAL_ERROR",
			Details: "Please try again later",
		})
	}

	if !valid {
		h.logger.Info().
			Str("user_id", user.ID.String()).
			Str("email", user.Email).
			Msg("Login attempt with invalid password")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Invalid email or password",
			Code:    "INVALID_CREDENTIALS",
			Details: "Please check your credentials and try again",
		})
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to generate JWT tokens during login")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Login failed",
			Code:    "TOKEN_GENERATION_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("User logged in successfully")

	// Return tokens and user info
	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes in seconds (should match JWT config)
		User: dto.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Role:  string(user.Role),
		},
	})
}

// Refresh handles token refresh requests.
//
// @Summary Refresh access token
// @Description Generates a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.RefreshResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh [post].
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind refresh request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain refresh_token",
		})
	}

	// Validate request
	if err := h.validateRefreshRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
			Details: "Please check your input and try again",
		})
	}

	// Generate new access token
	accessToken, err := h.jwtService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Info().
			Err(err).
			Msg("Failed to refresh token")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Invalid or expired refresh token",
			Code:    "INVALID_REFRESH_TOKEN",
			Details: "Please login again to get a new token",
		})
	}

	h.logger.Debug().
		Msg("Token refreshed successfully")

	// Return new access token
	return c.JSON(http.StatusOK, dto.RefreshResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   900, // 15 minutes in seconds (should match JWT config)
	})
}

// Logout handles user logout requests.
//
// @Summary User logout
// @Description Invalidates the provided JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Token to invalidate"
// @Success 200 {object} dto.LogoutResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post].
func (h *AuthHandler) Logout(c echo.Context) error {
	var req dto.LogoutRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind logout request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain access_token",
		})
	}

	// Validate request
	if err := h.validateLogoutRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
			Details: "Please check your input and try again",
		})
	}

	// Invalidate token
	err := h.jwtService.InvalidateToken(req.AccessToken)
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to invalidate token during logout")
		// Don't fail the request if token invalidation fails
		// Log the error but return success to the user
	}

	h.logger.Info().
		Msg("User logged out successfully")

	// Return success response
	return c.JSON(http.StatusOK, dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}

// validateRegisterRequest validates the registration request.
func (h *AuthHandler) validateRegisterRequest(req *dto.RegisterRequest) error {
	if req.Email == "" {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Email is required",
			apperrors.OriginWebHandler,
		)
	}

	if req.Password == "" {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Password is required",
			apperrors.OriginWebHandler,
		)
	}

	if len(req.Email) > 255 {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Email address is too long (max 255 characters)",
			apperrors.OriginWebHandler,
		)
	}

	if len(req.Password) > 128 {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Password is too long (max 128 characters)",
			apperrors.OriginWebHandler,
		)
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Invalid email format",
			apperrors.OriginWebHandler,
		)
	}

	// Validate password strength
	err := h.passwordService.CheckPasswordStrength(req.Password)
	if err != nil {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Password does not meet requirements",
			apperrors.OriginWebHandler,
		).WithCause(err)
	}

	return nil
}

// validateLoginRequest validates the login request.
func (h *AuthHandler) validateLoginRequest(req *dto.LoginRequest) error {
	if req.Email == "" {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Email is required",
			apperrors.OriginWebHandler,
		)
	}

	if req.Password == "" {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Password is required",
			apperrors.OriginWebHandler,
		)
	}

	if len(req.Email) > 255 {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Email address is too long (max 255 characters)",
			apperrors.OriginWebHandler,
		)
	}

	if len(req.Password) > 128 {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Password is too long (max 128 characters)",
			apperrors.OriginWebHandler,
		)
	}

	return nil
}

// validateRefreshRequest validates the refresh request.
func (h *AuthHandler) validateRefreshRequest(req *dto.RefreshRequest) error {
	if req.RefreshToken == "" {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Refresh token is required",
			apperrors.OriginWebHandler,
		)
	}

	if len(req.RefreshToken) > 2048 {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Refresh token is too long",
			apperrors.OriginWebHandler,
		)
	}

	return nil
}

// validateLogoutRequest validates the logout request.
func (h *AuthHandler) validateLogoutRequest(req *dto.LogoutRequest) error {
	if req.AccessToken == "" {
		return apperrors.NewValidationError(
			apperrors.CodeMissingField,
			"Access token is required",
			apperrors.OriginWebHandler,
		)
	}

	if len(req.AccessToken) > 2048 {
		return apperrors.NewValidationError(
			apperrors.CodeInvalidFormat,
			"Access token is too long",
			apperrors.OriginWebHandler,
		)
	}

	return nil
}
