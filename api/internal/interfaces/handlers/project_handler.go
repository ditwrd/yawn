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

// Package handlers provides HTTP request handlers for project management
// operations.
//
// This package contains handlers for project CRUD operations with proper
// authorization. All handlers follow RESTful conventions with proper error
// handling and JSON responses.
//
// Security features:
//   - Role-based access control (owner/maintainer/viewer access)
//   - Input validation and sanitization
//   - Proper error messages (don't leak sensitive information)
//   - Pagination support for list endpoints
//   - Member management with authorization checks
package handlers

import (
	"errors"
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

// ProjectHandler handles project management HTTP requests.
type ProjectHandler struct {
	projectService services.ProjectService
	userService    services.UserService
	logger         *zerolog.Logger
}

// NewProjectHandler creates a new ProjectHandler instance.
//
// Parameters:
//   - projectService: Project service for project operations
//   - userService: User service for user operations
//   - logger: Logger for structured logging
//
// Returns:
//   - *ProjectHandler: An instance of the project handler
func NewProjectHandler(
	projectService services.ProjectService,
	userService services.UserService,
	logger *zerolog.Logger,
) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		userService:    userService,
		logger:         logger,
	}
}

// getCurrentUserID extracts the current user ID from the Echo context.
func (h *ProjectHandler) getCurrentUserID(c echo.Context) (string, error) {
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return "", errors.New("missing user_id in context")
	}

	return currentUserID, nil
}

// CreateProject handles POST /projects endpoint.
//
// @Summary Create a new project
// @Description Creates a new project with the current user as owner.
// @Tags projects
// @Accept json
// @Produce json
// @Param request body dto.CreateProjectRequest true "Project creation data"
// @Success 201 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects [post].
func (h *ProjectHandler) CreateProject(c echo.Context) error {
	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Parse request body
	var req dto.CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind create project request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid project data",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project name is required",
			Code:    "MISSING_NAME",
			Details: "Please provide a project name",
		})
	}

	// Set default visibility to private if not provided
	if req.Visibility == "" {
		req.Visibility = "private"
	}

	// Create project model
	project := &models.Project{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Repository:  strings.TrimSpace(req.Repository),
		Visibility:  req.Visibility,
	}

	// Create project
	err = h.projectService.Create(project, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", currentUserID).
			Str("project_name", req.Name).
			Msg("Failed to create project")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create project",
			Code:    "CREATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Get the created project with relations
	createdProject, err := h.projectService.GetByID(
		project.ID.String(),
		currentUserID,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", project.ID.String()).
			Msg("Failed to retrieve created project")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve project",
			Code:    "RETRIEVE_FAILED",
			Details: "Project was created but could not be retrieved",
		})
	}

	// Convert to response DTO
	response := h.projectToDTO(createdProject)

	h.logger.Info().
		Str("project_id", project.ID.String()).
		Str("user_id", currentUserID).
		Msg("Project created successfully")

	return c.JSON(http.StatusCreated, response)
}

// ListProjects handles GET /projects endpoint.
//
// @Summary List projects
// @Description Returns a paginated list of projects accessible to the current
// user.
// @Tags projects
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Param search query string false "Search query for project names and
// descriptions"
// @Success 200 {object} dto.ProjectListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects [get].
func (h *ProjectHandler) ListProjects(c echo.Context) error {
	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

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

	// Get search query
	searchQuery := c.QueryParam("search")

	var projects []models.Project
	if searchQuery != "" {
		// Search projects
		projects, err = h.projectService.Search(
			currentUserID,
			searchQuery,
			limit,
			offset,
		)
		if err != nil {
			h.logger.Error().
				Err(err).
				Str("user_id", currentUserID).
				Str("search_query", searchQuery).
				Msg("Failed to search projects")

			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Failed to search projects",
				Code:    "SEARCH_FAILED",
				Details: "Please try again later",
			})
		}
	} else {
		// List projects
		projects, err = h.projectService.List(currentUserID, limit, offset)
		if err != nil {
			h.logger.Error().
				Err(err).
				Str("user_id", currentUserID).
				Msg("Failed to list projects")

			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Failed to retrieve projects",
				Code:    "LIST_FAILED",
				Details: "Please try again later",
			})
		}
	}

	// Convert to response DTO
	projectResponses := make([]dto.ProjectResponse, len(projects))
	for i, project := range projects {
		projectResponses[i] = h.projectToDTO(&project)
	}

	h.logger.Info().
		Str("user_id", currentUserID).
		Int("count", len(projects)).
		Msg("Projects listed successfully")

	return c.JSON(http.StatusOK, dto.ProjectListResponse{
		Projects: projectResponses,
		Total: len(
			projectResponses,
		), // In a real implementation, get total count from service
		Page:  page,
		Limit: limit,
	})
}

// GetProject handles GET /projects/{id} endpoint.
//
// @Summary Get project by ID
// @Description Returns project details. Users can only access projects they're
// members of.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id} [get].
func (h *ProjectHandler) GetProject(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Get project from service
	project, err := h.projectService.GetByID(projectIDStr, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to get project")

		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Project not found",
			Code:    "PROJECT_NOT_FOUND",
			Details: "The requested project does not exist or you don't have access",
		})
	}

	// Convert to response DTO
	response := h.projectToDTO(project)

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("user_id", currentUserID).
		Msg("Project retrieved successfully")

	return c.JSON(http.StatusOK, response)
}

// UpdateProject handles PUT /projects/{id} endpoint.
//
// @Summary Update project
// @Description Updates project information. Only project owners and maintainers
// can update projects.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body dto.UpdateProjectRequest true "Project update data"
// @Success 200 {object} dto.ProjectResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id} [put].
func (h *ProjectHandler) UpdateProject(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Parse request body
	var req dto.UpdateProjectRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind update project request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid project data",
		})
	}

	// Get existing project to get ID
	existingProject, err := h.projectService.GetByID(projectIDStr, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to get project for update")

		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Project not found",
			Code:    "PROJECT_NOT_FOUND",
			Details: "The requested project does not exist or you don't have access",
		})
	}

	// Create update model with ID
	project := &models.Project{
		ID:          existingProject.ID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Repository:  strings.TrimSpace(req.Repository),
		Visibility:  req.Visibility,
	}

	// Update project
	err = h.projectService.Update(project, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to update project")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update project",
			Code:    "UPDATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Get updated project
	updatedProject, err := h.projectService.GetByID(projectIDStr, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Msg("Failed to retrieve updated project")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve project",
			Code:    "RETRIEVE_FAILED",
			Details: "Project was updated but could not be retrieved",
		})
	}

	// Convert to response DTO
	response := h.projectToDTO(updatedProject)

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("user_id", currentUserID).
		Msg("Project updated successfully")

	return c.JSON(http.StatusOK, response)
}

// DeleteProject handles DELETE /projects/{id} endpoint.
//
// @Summary Delete project
// @Description Deletes a project. Only project owners can delete projects.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectDeleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id} [delete].
func (h *ProjectHandler) DeleteProject(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Get project details for logging
	existingProject, err := h.projectService.GetByID(projectIDStr, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to get project for deletion")

		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Project not found",
			Code:    "PROJECT_NOT_FOUND",
			Details: "The requested project does not exist or you don't have access",
		})
	}

	// Delete project
	err = h.projectService.Delete(projectIDStr, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to delete project")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete project",
			Code:    "DELETE_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("project_name", existingProject.Name).
		Str("user_id", currentUserID).
		Msg("Project deleted successfully")

	return c.JSON(http.StatusOK, dto.ProjectDeleteResponse{
		Message: "Project deleted successfully",
	})
}

// AddProjectMember handles POST /projects/{id}/members endpoint.
//
// @Summary Add member to project
// @Description Adds a new member to a project. Only project owners can add
// members.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body dto.AddProjectMemberRequest true "Member addition data"
// @Success 201 {object} dto.ProjectMemberResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id}/members [post].
func (h *ProjectHandler) AddProjectMember(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Parse request body
	var req dto.AddProjectMemberRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind add member request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid member data",
		})
	}

	// Validate required fields
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Member email is required",
			Code:    "MISSING_EMAIL",
			Details: "Please provide a valid email address",
		})
	}

	if req.Role == "" {
		req.Role = "viewer" // Default role
	}

	// Add member
	member, err := h.projectService.AddMember(
		projectIDStr,
		currentUserID,
		req.Email,
		req.Role,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Str("member_email", req.Email).
			Msg("Failed to add project member")

		if strings.Contains(err.Error(), "access denied") {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Access denied",
				Code:    "ACCESS_DENIED",
				Details: "Only project owners can add members",
			})
		}

		if strings.Contains(err.Error(), "user not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "User not found",
				Code:    "USER_NOT_FOUND",
				Details: "No user found with the provided email address",
			})
		}

		if strings.Contains(err.Error(), "already a member") {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "User already member",
				Code:    "ALREADY_MEMBER",
				Details: "The user is already a member of this project",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to add member",
			Code:    "ADD_MEMBER_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	response := dto.ProjectMemberResponse{
		UserID:    member.User.ID.String(),
		Email:     member.User.Email,
		Role:      string(member.Role),
		CreatedAt: member.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: member.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("user_id", currentUserID).
		Str("member_email", req.Email).
		Str("role", req.Role).
		Msg("Project member added successfully")

	return c.JSON(http.StatusCreated, response)
}

// ListProjectMembers handles GET /projects/{id}/members endpoint.
//
// @Summary List project members
// @Description Returns a list of all members in a project. All project members
// can view this.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectMemberListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id}/members [get].
func (h *ProjectHandler) ListProjectMembers(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// List members
	members, err := h.projectService.ListMembers(projectIDStr, currentUserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to list project members")

		if strings.Contains(err.Error(), "access denied") {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Access denied",
				Code:    "ACCESS_DENIED",
				Details: "You don't have permission to view project members",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list members",
			Code:    "LIST_MEMBERS_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	memberResponses := make([]dto.ProjectMemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = dto.ProjectMemberResponse{
			UserID:    member.User.ID.String(),
			Email:     member.User.Email,
			Role:      string(member.Role),
			CreatedAt: member.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: member.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("user_id", currentUserID).
		Int("count", len(members)).
		Msg("Project members listed successfully")

	return c.JSON(http.StatusOK, dto.ProjectMemberListResponse{
		Members: memberResponses,
		Total:   len(memberResponses),
	})
}

// RemoveProjectMember handles DELETE /projects/{id}/members/{memberId}
// endpoint.
//
// @Summary Remove member from project
// @Description Removes a member from a project. Only project owners can remove
// members.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param memberId path string true "Member User ID"
// @Success 200 {object} dto.ProjectDeleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id}/members/{memberId} [delete].
func (h *ProjectHandler) RemoveProjectMember(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	memberIDStr := c.Param("memberId")
	if memberIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Member ID is required",
			Code:    "MISSING_MEMBER_ID",
			Details: "Please provide a valid member ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	_, err = uuid.FromString(memberIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid member ID format",
			Code:    "INVALID_MEMBER_ID",
			Details: "Member ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Remove member
	err = h.projectService.RemoveMember(projectIDStr, currentUserID, memberIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Str("member_id", memberIDStr).
			Msg("Failed to remove project member")

		if strings.Contains(err.Error(), "access denied") {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Access denied",
				Code:    "ACCESS_DENIED",
				Details: "Only project owners can remove members",
			})
		}

		if strings.Contains(err.Error(), "member not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Member not found",
				Code:    "MEMBER_NOT_FOUND",
				Details: "The specified member is not part of this project",
			})
		}

		if strings.Contains(err.Error(), "cannot remove yourself") {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Cannot remove yourself",
				Code:    "CANNOT_REMOVE_SELF",
				Details: "You cannot remove yourself from the project",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to remove member",
			Code:    "REMOVE_MEMBER_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("user_id", currentUserID).
		Str("member_id", memberIDStr).
		Msg("Project member removed successfully")

	return c.JSON(http.StatusOK, dto.ProjectDeleteResponse{
		Message: "Member removed successfully",
	})
}

// UpdateProjectMemberRole handles PUT /projects/{id}/members/{memberId}
// endpoint.
//
// @Summary Update member role
// @Description Updates a member's role in a project. Only project owners can
// update member roles.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param memberId path string true "Member User ID"
// @Param request body dto.UpdateProjectMemberRequest true "Role update data"
// @Success 200 {object} dto.ProjectMemberResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /projects/{id}/members/{memberId} [put].
func (h *ProjectHandler) UpdateProjectMemberRole(c echo.Context) error {
	projectIDStr := c.Param("id")
	if projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	memberIDStr := c.Param("memberId")
	if memberIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Member ID is required",
			Code:    "MISSING_MEMBER_ID",
			Details: "Please provide a valid member ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid project ID format",
			Code:    "INVALID_PROJECT_ID",
			Details: "Project ID must be a valid UUID",
		})
	}

	_, err = uuid.FromString(memberIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid member ID format",
			Code:    "INVALID_MEMBER_ID",
			Details: "Member ID must be a valid UUID",
		})
	}

	// Get current user ID
	currentUserID, err := h.getCurrentUserID(c)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to get current user ID")

		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Parse request body
	var req dto.UpdateProjectMemberRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind update member role request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid role data",
		})
	}

	// Validate role
	if req.Role == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Role is required",
			Code:    "MISSING_ROLE",
			Details: "Please provide a valid role",
		})
	}

	// Update member role
	member, err := h.projectService.UpdateMemberRole(
		projectIDStr,
		currentUserID,
		memberIDStr,
		req.Role,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Str("user_id", currentUserID).
			Str("member_id", memberIDStr).
			Str("role", req.Role).
			Msg("Failed to update project member role")

		if strings.Contains(err.Error(), "access denied") {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Access denied",
				Code:    "ACCESS_DENIED",
				Details: "Only project owners can update member roles",
			})
		}

		if strings.Contains(err.Error(), "member not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Member not found",
				Code:    "MEMBER_NOT_FOUND",
				Details: "The specified member is not part of this project",
			})
		}

		if strings.Contains(err.Error(), "invalid project role") {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid role",
				Code:    "INVALID_ROLE",
				Details: "Valid roles are: owner, maintainer, viewer",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update member role",
			Code:    "UPDATE_ROLE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	response := dto.ProjectMemberResponse{
		UserID:    member.User.ID.String(),
		Email:     member.User.Email,
		Role:      string(member.Role),
		CreatedAt: member.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: member.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.logger.Info().
		Str("project_id", projectIDStr).
		Str("user_id", currentUserID).
		Str("member_id", memberIDStr).
		Str("role", req.Role).
		Msg("Project member role updated successfully")

	return c.JSON(http.StatusOK, response)
}

// projectToDTO converts a project model to a project response DTO.
func (h *ProjectHandler) projectToDTO(
	project *models.Project,
) dto.ProjectResponse {
	response := dto.ProjectResponse{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		Visibility:  project.Visibility,
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add repository URL if present
	if project.Repository != "" {
		response.Repository = &project.Repository
	}

	// Add owner information
	if project.Owner.ID != uuid.Nil {
		response.Owner = dto.UserInfo{
			ID:    project.Owner.ID.String(),
			Email: project.Owner.Email,
			Role:  string(project.Owner.Role),
		}
	}

	// Add members information
	if len(project.Users) > 0 {
		response.Members = make([]dto.ProjectMemberInfo, len(project.Users))
		for i, member := range project.Users {
			response.Members[i] = dto.ProjectMemberInfo{
				UserID:    member.User.ID.String(),
				Email:     member.User.Email,
				Role:      string(member.Role),
				CreatedAt: member.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
		}
	}

	return response
}
