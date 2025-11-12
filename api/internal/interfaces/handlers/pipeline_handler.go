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

// Package handlers provides HTTP request handlers for pipeline management
// operations.
//
// This package contains handlers for pipeline CRUD operations with proper
// authorization and validation. All handlers follow RESTful conventions
// with proper error handling and JSON responses.
package handlers

import (
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

// PipelineHandler handles pipeline management HTTP requests.
type PipelineHandler struct {
	pipelineService services.PipelineService
	logger          *zerolog.Logger
}

// NewPipelineHandler creates a new PipelineHandler instance.
//
// Parameters:
//   - pipelineService: Pipeline service for pipeline operations
//   - logger: Logger for structured logging
//
// Returns:
//   - *PipelineHandler: An instance of the pipeline handler
func NewPipelineHandler(
	pipelineService services.PipelineService,
	logger *zerolog.Logger,
) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
		logger:          logger,
	}
}

// ListPipelines handles GET /pipelines endpoint.
//
// @Summary List all pipelines
// @Description Returns a paginated list of pipelines with optional filtering.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Param project_id query string false "Filter by project ID"
// @Param status query string false "Filter by status"
// @Param name query string false "Filter by pipeline name (partial match)"
// @Param search query string false "Search across name and description"
// @Param is_enabled query bool false "Filter by enabled status"
// @Param schedule query string false "Filter by schedule"
// @Success 200 {object} dto.PipelineListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines [get].
func (h *PipelineHandler) ListPipelines(c echo.Context) error {
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

	// Parse filters
	filters := services.PipelineListFilters{
		ProjectID: c.QueryParam("project_id"),
		Status:    models.PipelineStatus(c.QueryParam("status")),
		Name:      c.QueryParam("name"),
		Search:    c.QueryParam("search"),
		Schedule:  c.QueryParam("schedule"),
	}

	// Parse is_enabled filter
	if isEnabledStr := c.QueryParam("is_enabled"); isEnabledStr != "" {
		isEnabled, err := strconv.ParseBool(isEnabledStr)
		if err == nil {
			filters.IsEnabled = &isEnabled
		}
	}

	// Get pipelines from service
	result, err := h.pipelineService.List(
		c.Request().Context(),
		page,
		limit,
		filters,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int("page", page).
			Int("limit", limit).
			Interface("filters", filters).
			Msg("Failed to list pipelines")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve pipelines",
			Code:    "LIST_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	pipelineResponses := make([]dto.PipelineResponse, len(result.Pipelines))
	for i, pipeline := range result.Pipelines {
		pipelineResponses[i] = h.pipelineToResponse(pipeline)
	}

	h.logger.Info().
		Int("count", len(pipelineResponses)).
		Msg("Pipelines listed successfully")

	return c.JSON(http.StatusOK, dto.PipelineListResponse{
		Pipelines: pipelineResponses,
		Total:     result.Total,
		Page:      result.Page,
		Limit:     result.Limit,
	})
}

// GetPipeline handles GET /pipelines/{id} endpoint.
//
// @Summary Get pipeline by ID
// @Description Returns pipeline details with full relationships.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Success 200 {object} dto.PipelineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id} [get].
func (h *PipelineHandler) GetPipeline(c echo.Context) error {
	pipelineIDStr := c.Param("id")
	if pipelineIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Pipeline ID is required",
			Code:    "MISSING_PIPELINE_ID",
			Details: "Please provide a valid pipeline ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(pipelineIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid pipeline ID format",
			Code:    "INVALID_PIPELINE_ID",
			Details: "Pipeline ID must be a valid UUID",
		})
	}

	// Get pipeline from service
	pipeline, err := h.pipelineService.GetByID(
		c.Request().Context(),
		pipelineIDStr,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Msg("Failed to get pipeline")

		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Pipeline not found",
			Code:    "PIPELINE_NOT_FOUND",
			Details: "The requested pipeline does not exist",
		})
	}

	// Convert to response DTO
	pipelineResponse := h.pipelineToResponse(pipeline)

	h.logger.Info().
		Str("pipeline_id", pipelineIDStr).
		Msg("Pipeline retrieved successfully")

	return c.JSON(http.StatusOK, pipelineResponse)
}

// CreatePipeline handles POST /pipelines endpoint.
//
// @Summary Create a new pipeline
// @Description Creates a new pipeline with the provided details.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param request body dto.CreatePipelineRequest true "Pipeline creation data"
// @Success 201 {object} dto.PipelineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines [post].
func (h *PipelineHandler) CreatePipeline(c echo.Context) error {
	// Parse request body
	var req dto.CreatePipelineRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind create pipeline request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid pipeline data",
		})
	}

	// Get current user info from context
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "User authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Check if user can create pipelines in the project
	if err := h.pipelineService.CanCreate(c.Request().Context(), req.ProjectID, currentUserID); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("project_id", req.ProjectID).
			Msg("Access denied: user cannot create pipelines in project")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "CREATE_ACCESS_DENIED",
			Details: "You don't have permission to create pipelines in this project",
		})
	}

	// Create service request
	serviceReq := &services.CreatePipelineRequest{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		Schedule:    req.Schedule,
		ProjectID:   req.ProjectID,
		AssetIDs:    req.AssetIDs,
	}

	// Create pipeline
	pipeline, err := h.pipelineService.Create(c.Request().Context(), serviceReq)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("name", req.Name).
			Str("project_id", req.ProjectID).
			Msg("Failed to create pipeline")

		// Handle specific validation errors
		if strings.Contains(err.Error(), "already exists") {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Pipeline already exists",
				Code:    "PIPELINE_EXISTS",
				Details: err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create pipeline",
			Code:    "CREATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	pipelineResponse := h.pipelineToResponse(pipeline)

	h.logger.Info().
		Str("pipeline_id", pipeline.ID.String()).
		Str("name", pipeline.Name).
		Msg("Pipeline created successfully")

	return c.JSON(http.StatusCreated, pipelineResponse)
}

// UpdatePipeline handles PUT /pipelines/{id} endpoint.
//
// @Summary Update pipeline information
// @Description Updates pipeline information with the provided details.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Param request body dto.UpdatePipelineRequest true "Pipeline update data"
// @Success 200 {object} dto.PipelineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id} [put].
func (h *PipelineHandler) UpdatePipeline(c echo.Context) error {
	pipelineIDStr := c.Param("id")
	if pipelineIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Pipeline ID is required",
			Code:    "MISSING_PIPELINE_ID",
			Details: "Please provide a valid pipeline ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(pipelineIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid pipeline ID format",
			Code:    "INVALID_PIPELINE_ID",
			Details: "Pipeline ID must be a valid UUID",
		})
	}

	// Parse request body
	var req dto.UpdatePipelineRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Msg("Failed to bind update pipeline request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid pipeline data",
		})
	}

	// Get current user info from context
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "User authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Check if user can update pipelines in the project (requires maintainer
	// role)
	if err := h.pipelineService.ValidateAccess(c.Request().Context(), pipelineIDStr, currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("pipeline_id", pipelineIDStr).
			Msg("Access denied: user cannot update pipeline")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "UPDATE_ACCESS_DENIED",
			Details: "You don't have permission to update this pipeline",
		})
	}

	// Convert DTO to service request
	serviceReq := &services.UpdatePipelineRequest{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		Schedule:    req.Schedule,
		IsEnabled:   req.IsEnabled,
		AssetIDs:    req.AssetIDs,
	}

	// Convert status string to PipelineStatus pointer if provided
	if req.Status != nil {
		status := models.PipelineStatus(*req.Status)
		serviceReq.Status = &status
	}

	// Update pipeline
	pipeline, err := h.pipelineService.Update(
		c.Request().Context(),
		pipelineIDStr,
		serviceReq,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Msg("Failed to update pipeline")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update pipeline",
			Code:    "UPDATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	pipelineResponse := h.pipelineToResponse(pipeline)

	h.logger.Info().
		Str("pipeline_id", pipelineIDStr).
		Msg("Pipeline updated successfully")

	return c.JSON(http.StatusOK, pipelineResponse)
}

// DeletePipeline handles DELETE /pipelines/{id} endpoint.
//
// @Summary Delete pipeline
// @Description Soft deletes a pipeline. Requires maintainer or owner role in
// the project.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Success 200 {object} dto.PipelineDeleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id} [delete].
func (h *PipelineHandler) DeletePipeline(c echo.Context) error {
	pipelineIDStr := c.Param("id")
	if pipelineIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Pipeline ID is required",
			Code:    "MISSING_PIPELINE_ID",
			Details: "Please provide a valid pipeline ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(pipelineIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid pipeline ID format",
			Code:    "INVALID_PIPELINE_ID",
			Details: "Pipeline ID must be a valid UUID",
		})
	}

	// Get current user info from context
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "User authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Check if user can delete pipelines in the project (requires maintainer
	// role)
	if err := h.pipelineService.ValidateAccess(c.Request().Context(), pipelineIDStr, currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("pipeline_id", pipelineIDStr).
			Msg("Access denied: user cannot delete pipeline")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "DELETE_ACCESS_DENIED",
			Details: "You don't have permission to delete this pipeline",
		})
	}

	// Delete pipeline
	err = h.pipelineService.Delete(c.Request().Context(), pipelineIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Msg("Failed to delete pipeline")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete pipeline",
			Code:    "DELETE_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("pipeline_id", pipelineIDStr).
		Msg("Pipeline deleted successfully")

	return c.JSON(http.StatusOK, dto.PipelineDeleteResponse{
		Message: "Pipeline deleted successfully",
	})
}

// SearchPipelines handles GET /pipelines/search endpoint.
//
// @Summary Search pipelines
// @Description Searches pipelines by name or description with pagination.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} dto.PipelineSearchResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/search [get].
func (h *PipelineHandler) SearchPipelines(c echo.Context) error {
	query := c.QueryParam("q")
	if strings.TrimSpace(query) == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Search query is required",
			Code:    "MISSING_QUERY",
			Details: "Please provide a search query using the 'q' parameter",
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

	// Search pipelines
	result, err := h.pipelineService.Search(
		c.Request().Context(),
		query,
		page,
		limit,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("query", query).
			Msg("Failed to search pipelines")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to search pipelines",
			Code:    "SEARCH_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	pipelineResponses := make([]dto.PipelineResponse, len(result.Pipelines))
	for i, pipeline := range result.Pipelines {
		pipelineResponses[i] = h.pipelineToResponse(pipeline)
	}

	h.logger.Info().
		Str("query", query).
		Int("count", len(pipelineResponses)).
		Msg("Pipeline search completed successfully")

	return c.JSON(http.StatusOK, dto.PipelineSearchResponse{
		Pipelines: pipelineResponses,
		Total:     result.Total,
		Query:     query,
		Page:      result.Page,
		Limit:     result.Limit,
	})
}

// UpdatePipelineStatus handles PUT /pipelines/{id}/status endpoint.
//
// @Summary Update pipeline status
// @Description Updates the status of a pipeline.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Param status body string true "New status" Enums(draft, active, paused,
// running, completed, failed, cancelled)
// @Success 200 {object} dto.PipelineResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id}/status [put].
func (h *PipelineHandler) UpdatePipelineStatus(c echo.Context) error {
	pipelineIDStr := c.Param("id")
	if pipelineIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Pipeline ID is required",
			Code:    "MISSING_PIPELINE_ID",
			Details: "Please provide a valid pipeline ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(pipelineIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid pipeline ID format",
			Code:    "INVALID_PIPELINE_ID",
			Details: "Pipeline ID must be a valid UUID",
		})
	}

	// Parse status from request body
	var statusReq struct {
		Status string `json:"status" validate:"required,oneof=draft active paused running completed failed cancelled"`
	}
	if err := c.Bind(&statusReq); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain a valid status",
		})
	}

	// Get current user info from context
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "User authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Check if user can update pipeline status (requires maintainer role)
	if err := h.pipelineService.ValidateAccess(c.Request().Context(), pipelineIDStr, currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("pipeline_id", pipelineIDStr).
			Msg("Access denied: user cannot update pipeline status")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "UPDATE_ACCESS_DENIED",
			Details: "You don't have permission to update this pipeline's status",
		})
	}

	// Update status
	newStatus := models.PipelineStatus(statusReq.Status)
	if err := h.pipelineService.UpdateStatus(c.Request().Context(), pipelineIDStr, newStatus); err != nil {
		h.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Str("status", string(newStatus)).
			Msg("Failed to update pipeline status")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update pipeline status",
			Code:    "STATUS_UPDATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Get updated pipeline
	pipeline, err := h.pipelineService.GetByID(
		c.Request().Context(),
		pipelineIDStr,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve updated pipeline",
			Code:    "RETRIEVE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	pipelineResponse := h.pipelineToResponse(pipeline)

	h.logger.Info().
		Str("pipeline_id", pipelineIDStr).
		Str("status", string(newStatus)).
		Msg("Pipeline status updated successfully")

	return c.JSON(http.StatusOK, pipelineResponse)
}

// === Pipeline Execution Handlers ===

// TriggerExecution handles POST /pipelines/{id}/execute endpoint.
//
// @Summary Trigger pipeline execution
// @Description Manually triggers a pipeline execution.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Param request body dto.TriggerExecutionRequest false "Execution
// configuration"
// @Success 201 {object} dto.TriggerExecutionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id}/execute [post].
func (h *PipelineHandler) TriggerExecution(c echo.Context) error {
	pipelineIDStr := c.Param("id")
	if pipelineIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Pipeline ID is required",
			Code:    "MISSING_PIPELINE_ID",
			Details: "Please provide a valid pipeline ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(pipelineIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid pipeline ID format",
			Code:    "INVALID_PIPELINE_ID",
			Details: "Pipeline ID must be a valid UUID",
		})
	}

	// Parse request body (optional)
	var req dto.TriggerExecutionRequest
	c.Bind(&req) // Ignore errors as body is optional

	// Get current user info from context
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "User authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Check if user can trigger pipeline execution (requires maintainer role)
	if err := h.pipelineService.ValidateAccess(c.Request().Context(), pipelineIDStr, currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("pipeline_id", pipelineIDStr).
			Msg("Access denied: user cannot trigger pipeline execution")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "EXECUTION_ACCESS_DENIED",
			Details: "You don't have permission to execute this pipeline",
		})
	}

	// Trigger execution
	execution, err := h.pipelineService.TriggerExecution(
		c.Request().Context(),
		pipelineIDStr,
		currentUserID,
		req.Config,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Str("user_id", currentUserID).
			Msg("Failed to trigger pipeline execution")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to trigger pipeline execution",
			Code:    "TRIGGER_FAILED",
			Details: err.Error(),
		})
	}

	// Convert to response DTO
	executionResponse := h.executionToResponse(execution)

	h.logger.Info().
		Str("execution_id", execution.ID.String()).
		Str("pipeline_id", pipelineIDStr).
		Str("triggered_by", currentUserID).
		Msg("Pipeline execution triggered successfully")

	return c.JSON(http.StatusCreated, dto.TriggerExecutionResponse{
		Execution: &executionResponse,
		Message:   "Pipeline execution triggered successfully",
	})
}

// GetExecutions handles GET /pipelines/{id}/executions endpoint.
//
// @Summary Get pipeline executions
// @Description Returns a paginated list of executions for a pipeline.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Pipeline ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} dto.PipelineExecutionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id}/executions [get].
func (h *PipelineHandler) GetExecutions(c echo.Context) error {
	pipelineIDStr := c.Param("id")
	if pipelineIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Pipeline ID is required",
			Code:    "MISSING_PIPELINE_ID",
			Details: "Please provide a valid pipeline ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(pipelineIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid pipeline ID format",
			Code:    "INVALID_PIPELINE_ID",
			Details: "Pipeline ID must be a valid UUID",
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

	// Get executions from service
	result, err := h.pipelineService.GetExecutionsByPipelineID(
		c.Request().Context(),
		pipelineIDStr,
		page,
		limit,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineIDStr).
			Msg("Failed to get pipeline executions")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve executions",
			Code:    "LIST_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	executionResponses := make(
		[]dto.PipelineExecutionResponse,
		len(result.Executions),
	)
	for i, execution := range result.Executions {
		executionResponses[i] = h.executionToResponse(execution)
	}

	h.logger.Info().
		Str("pipeline_id", pipelineIDStr).
		Int("count", len(executionResponses)).
		Msg("Pipeline executions retrieved successfully")

	return c.JSON(http.StatusOK, dto.PipelineExecutionListResponse{
		Executions: executionResponses,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
	})
}

// CancelExecution handles POST /executions/{id}/cancel endpoint.
//
// @Summary Cancel pipeline execution
// @Description Cancels a running pipeline execution.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Execution ID"
// @Success 200 {object} dto.CancelExecutionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /executions/{id}/cancel [post].
func (h *PipelineHandler) CancelExecution(c echo.Context) error {
	executionIDStr := c.Param("id")
	if executionIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Execution ID is required",
			Code:    "MISSING_EXECUTION_ID",
			Details: "Please provide a valid execution ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(executionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid execution ID format",
			Code:    "INVALID_EXECUTION_ID",
			Details: "Execution ID must be a valid UUID",
		})
	}

	// Get current user info from context
	currentUserID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "User authentication required",
			Code:    "AUTH_REQUIRED",
			Details: "Please provide valid authentication credentials",
		})
	}

	// Get execution to determine pipeline access
	execution, err := h.pipelineService.GetExecutionByID(
		c.Request().Context(),
		executionIDStr,
	)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Execution not found",
			Code:    "EXECUTION_NOT_FOUND",
			Details: "The requested execution does not exist",
		})
	}

	// Check if user can cancel executions (requires maintainer role)
	if err := h.pipelineService.ValidateAccess(c.Request().Context(), execution.PipelineID.String(), currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("execution_id", executionIDStr).
			Msg("Access denied: user cannot cancel execution")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "CANCEL_ACCESS_DENIED",
			Details: "You don't have permission to cancel this execution",
		})
	}

	// Cancel execution
	err = h.pipelineService.CancelExecution(
		c.Request().Context(),
		executionIDStr,
		currentUserID,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("execution_id", executionIDStr).
			Str("cancelled_by", currentUserID).
			Msg("Failed to cancel execution")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to cancel execution",
			Code:    "CANCEL_FAILED",
			Details: err.Error(),
		})
	}

	h.logger.Info().
		Str("execution_id", executionIDStr).
		Str("cancelled_by", currentUserID).
		Msg("Pipeline execution cancelled successfully")

	return c.JSON(http.StatusOK, dto.CancelExecutionResponse{
		Message: "Pipeline execution cancelled successfully",
	})
}

// === Dependency Management Handlers ===

// GetDependencyGraph handles GET /pipelines/{id}/dependencies endpoint.
//
// @Summary Get pipeline dependency graph
// @Description Returns the dependency graph for pipelines in a project.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.DependencyGraphResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /pipelines/{id}/dependencies [get].
func (h *PipelineHandler) GetDependencyGraph(c echo.Context) error {
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

	// Get dependency graph from service
	graph, err := h.pipelineService.GetDependencyGraph(
		c.Request().Context(),
		projectIDStr,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectIDStr).
			Msg("Failed to get dependency graph")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve dependency graph",
			Code:    "GRAPH_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("project_id", projectIDStr).
		Int("nodes", len(graph.Nodes)).
		Int("edges", len(graph.Edges)).
		Msg("Dependency graph retrieved successfully")

	return c.JSON(http.StatusOK, graph)
}

// === Helper Methods ===

// pipelineToResponse converts a pipeline model to a response DTO.
func (h *PipelineHandler) pipelineToResponse(
	pipeline *models.Pipeline,
) dto.PipelineResponse {
	response := dto.PipelineResponse{
		ID:          pipeline.ID.String(),
		Name:        pipeline.Name,
		Description: pipeline.Description,
		ProjectID:   pipeline.ProjectID.String(),
		Status:      string(pipeline.Status),
		Config:      pipeline.Config,
		Schedule:    pipeline.Schedule,
		IsEnabled:   pipeline.IsEnabled,
		CreatedAt:   pipeline.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   pipeline.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add project summary if loaded
	if pipeline.Project.ID != uuid.Nil {
		response.Project = &dto.ProjectSummaryResponse{
			ID:          pipeline.Project.ID.String(),
			Name:        pipeline.Project.Name,
			Description: pipeline.Project.Description,
			Visibility:  pipeline.Project.Visibility,
		}
	}

	// Add assets if loaded
	if len(pipeline.Assets) > 0 {
		assetSummaries := make([]dto.AssetSummaryResponse, len(pipeline.Assets))
		for i, asset := range pipeline.Assets {
			assetSummaries[i] = dto.AssetSummaryResponse{
				ID:      asset.ID.String(),
				Name:    asset.Name,
				Version: asset.Version,
			}
		}

		response.Assets = assetSummaries
	}

	// Add executions if loaded
	if len(pipeline.Executions) > 0 {
		executionResponses := make(
			[]dto.PipelineExecutionResponse,
			len(pipeline.Executions),
		)
		for i, execution := range pipeline.Executions {
			executionResponses[i] = h.executionToResponse(&execution)
		}

		response.Executions = executionResponses
	}

	// Add dependencies if loaded
	if len(pipeline.Dependencies) > 0 {
		dependencyResponses := make(
			[]dto.PipelineDependencyResponse,
			len(pipeline.Dependencies),
		)
		for i, dep := range pipeline.Dependencies {
			dependencyResponses[i] = h.dependencyToResponse(&dep)
		}

		response.Dependencies = dependencyResponses
	}

	// Add dependents if loaded
	if len(pipeline.Dependents) > 0 {
		dependencyResponses := make(
			[]dto.PipelineDependencyResponse,
			len(pipeline.Dependents),
		)
		for i, dep := range pipeline.Dependents {
			dependencyResponses[i] = h.dependencyToResponse(&dep)
		}

		response.Dependents = dependencyResponses
	}

	return response
}

// executionToResponse converts a pipeline execution model to a response DTO.
func (h *PipelineHandler) executionToResponse(
	execution *models.PipelineExecution,
) dto.PipelineExecutionResponse {
	response := dto.PipelineExecutionResponse{
		ID:          execution.ID.String(),
		PipelineID:  execution.PipelineID.String(),
		Status:      string(execution.Status),
		TriggerType: execution.TriggerType,
		Duration:    execution.Duration,
		Config:      execution.Config,
		Logs:        execution.Logs,
		ErrorMsg:    execution.ErrorMsg,
		CreatedAt:   execution.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   execution.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add optional fields
	if execution.TriggeredBy != nil {
		triggeredBy := execution.TriggeredBy.String()
		response.TriggeredBy = &triggeredBy
	}

	if execution.StartedAt != nil {
		startedAt := execution.StartedAt.Format("2006-01-02T15:04:05Z")
		response.StartedAt = &startedAt
	}

	if execution.CompletedAt != nil {
		completedAt := execution.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	// Add pipeline summary if loaded
	if execution.Pipeline.ID != uuid.Nil {
		response.Pipeline = &dto.PipelineSummaryResponse{
			ID:        execution.Pipeline.ID.String(),
			Name:      execution.Pipeline.Name,
			Status:    string(execution.Pipeline.Status),
			IsEnabled: execution.Pipeline.IsEnabled,
		}
	}

	// Add trigger user if loaded
	if execution.TriggerUser != nil && execution.TriggerUser.ID != uuid.Nil {
		response.TriggerUser = &dto.UserSummaryResponse{
			ID:    execution.TriggerUser.ID.String(),
			Email: execution.TriggerUser.Email,
			Role:  string(execution.TriggerUser.Role),
		}
	}

	// Add steps if loaded
	if len(execution.Steps) > 0 {
		stepResponses := make([]dto.ExecutionStepResponse, len(execution.Steps))
		for i, step := range execution.Steps {
			stepResponses[i] = h.stepToResponse(&step)
		}

		response.Steps = stepResponses
	}

	return response
}

// stepToResponse converts an execution step model to a response DTO.
func (h *PipelineHandler) stepToResponse(
	step *models.ExecutionStep,
) dto.ExecutionStepResponse {
	response := dto.ExecutionStepResponse{
		ID:          step.ID.String(),
		ExecutionID: step.ExecutionID.String(),
		Name:        step.Name,
		Type:        string(step.Type),
		Status:      string(step.Status),
		Config:      step.Config,
		Result:      step.Result,
		ErrorMsg:    step.ErrorMsg,
		Duration:    step.Duration,
		Order:       step.Order,
		CreatedAt:   step.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   step.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add optional fields
	if step.StartedAt != nil {
		startedAt := step.StartedAt.Format("2006-01-02T15:04:05Z")
		response.StartedAt = &startedAt
	}

	if step.CompletedAt != nil {
		completedAt := step.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	return response
}

// dependencyToResponse converts a pipeline dependency model to a response DTO.
func (h *PipelineHandler) dependencyToResponse(
	dependency *models.PipelineDependency,
) dto.PipelineDependencyResponse {
	response := dto.PipelineDependencyResponse{
		ID:                  dependency.ID.String(),
		PipelineID:          dependency.PipelineID.String(),
		DependsOnPipelineID: dependency.DependsOnPipelineID.String(),
		Condition:           dependency.Condition,
		CreatedAt:           dependency.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add pipeline summary if loaded
	if dependency.Pipeline.ID != uuid.Nil {
		response.Pipeline = &dto.PipelineSummaryResponse{
			ID:        dependency.Pipeline.ID.String(),
			Name:      dependency.Pipeline.Name,
			Status:    string(dependency.Pipeline.Status),
			IsEnabled: dependency.Pipeline.IsEnabled,
		}
	}

	// Add depends on summary if loaded
	if dependency.DependsOn.ID != uuid.Nil {
		response.DependsOn = &dto.PipelineSummaryResponse{
			ID:        dependency.DependsOn.ID.String(),
			Name:      dependency.DependsOn.Name,
			Status:    string(dependency.DependsOn.Status),
			IsEnabled: dependency.DependsOn.IsEnabled,
		}
	}

	return response
}
