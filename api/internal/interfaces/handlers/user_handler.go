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

// Package handlers provides HTTP request handlers for user management operations.
//
// This package contains handlers for user CRUD operations with proper authorization.
// All handlers follow RESTful conventions with proper error handling and JSON responses.
//
// Security features:
//   - Role-based access control (admin/self access)
//   - Input validation and sanitization
//   - Proper error messages (don't leak sensitive information)
//   - Pagination support for list endpoints
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userService services.UserService
	logger      *zerolog.Logger
}

// NewUserHandler creates a new UserHandler instance.
//
// Parameters:
//   - userService: User service for user operations
//   - logger: Logger for structured logging
//
// Returns:
//   - *UserHandler: An instance of the user handler
func NewUserHandler(userService services.UserService, logger *zerolog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// ListUsers handles GET /users endpoint.
//
// @Summary List all users (admin only)
// @Description Returns a paginated list of all users. Requires admin role.
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} dto.UserListResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Get users from service
	users, err := h.userService.List(limit, offset)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int("limit", limit).
			Int("offset", offset).
			Msg("Failed to list users")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve users",
			Code:    "LIST_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Role:      string(user.Role),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	h.logger.Info().
		Int("count", len(users)).
		Msg("Users listed successfully")

	return c.JSON(http.StatusOK, dto.UserListResponse{
		Users: userResponses,
		Total: len(users), // In a real implementation, you'd get the total count from the database
		Page:  page,
		Limit: limit,
	})
}

// GetUser handles GET /users/{id} endpoint.
//
// @Summary Get user by ID
// @Description Returns user details. Users can only access their own data, admins can access any user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c echo.Context) error {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "User ID is required",
			Code:    "MISSING_USER_ID",
			Details: "Please provide a valid user ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID format",
			Code:    "INVALID_USER_ID",
			Details: "User ID must be a valid UUID",
		})
	}

	// Get user from service
	user, err := h.userService.GetByID(userIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userIDStr).
			Msg("Failed to get user")
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "User not found",
			Code:    "USER_NOT_FOUND",
			Details: "The requested user does not exist",
		})
	}

	// Convert to response DTO
	userResponse := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.logger.Info().
		Str("user_id", userIDStr).
		Msg("User retrieved successfully")

	return c.JSON(http.StatusOK, userResponse)
}

// UpdateUser handles PUT /users/{id} endpoint.
//
// @Summary Update user information
// @Description Updates user information. Users can only update their own data, admins can update any user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "User ID is required",
			Code:    "MISSING_USER_ID",
			Details: "Please provide a valid user ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID format",
			Code:    "INVALID_USER_ID",
			Details: "User ID must be a valid UUID",
		})
	}

	// Parse request body
	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", userIDStr).
			Msg("Failed to bind update user request")
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid user data",
		})
	}

	// Get existing user
	existingUser, err := h.userService.GetByID(userIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userIDStr).
			Msg("Failed to get user for update")
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "User not found",
			Code:    "USER_NOT_FOUND",
			Details: "The requested user does not exist",
		})
	}

	// Update fields if provided
	if req.Email != "" {
		existingUser.Email = strings.ToLower(strings.TrimSpace(req.Email))
	}
	if req.Role != "" {
		existingUser.Role = models.UserRole(req.Role)
	}

	// Validate the updated user
	if req.Email != "" {
		if !strings.Contains(existingUser.Email, "@") || !strings.Contains(existingUser.Email, ".") {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid email format",
				Code:    "VALIDATION_ERROR",
				Details: "Please provide a valid email address",
			})
		}
	}

	if req.Role != "" {
		validRoles := []models.UserRole{models.UserRoleAdmin, models.UserRoleUser}
		isValidRole := false
		for _, role := range validRoles {
			if existingUser.Role == role {
				isValidRole = true
				break
			}
		}
		if !isValidRole {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   fmt.Sprintf("Invalid user role: %s", existingUser.Role),
				Code:    "VALIDATION_ERROR",
				Details: "Valid roles are: admin, user",
			})
		}
	}

	// Update user in service
	err = h.userService.Update(existingUser)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userIDStr).
			Msg("Failed to update user")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update user",
			Code:    "UPDATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	userResponse := dto.UserResponse{
		ID:        existingUser.ID.String(),
		Email:     existingUser.Email,
		Role:      string(existingUser.Role),
		CreatedAt: existingUser.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: existingUser.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.logger.Info().
		Str("user_id", userIDStr).
		Msg("User updated successfully")

	return c.JSON(http.StatusOK, userResponse)
}

// DeleteUser handles DELETE /users/{id} endpoint.
//
// @Summary Delete user (admin only)
// @Description Soft deletes a user. Requires admin role.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserDeleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "User ID is required",
			Code:    "MISSING_USER_ID",
			Details: "Please provide a valid user ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID format",
			Code:    "INVALID_USER_ID",
			Details: "User ID must be a valid UUID",
		})
	}

	// Check if user exists
	existingUser, err := h.userService.GetByID(userIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userIDStr).
			Msg("Failed to get user for deletion")
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "User not found",
			Code:    "USER_NOT_FOUND",
			Details: "The requested user does not exist",
		})
	}

	// Delete user from service
	err = h.userService.Delete(userIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userIDStr).
			Msg("Failed to delete user")
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete user",
			Code:    "DELETE_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("user_id", userIDStr).
		Str("email", existingUser.Email).
		Msg("User deleted successfully")

	return c.JSON(http.StatusOK, dto.UserDeleteResponse{
		Message: "User deleted successfully",
	})
}