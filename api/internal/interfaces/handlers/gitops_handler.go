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

// Package handlers provides HTTP request handlers for GitOps operations
//
// This package contains handlers for Git repository synchronization, webhook
// processing, and pipeline discovery from Git repositories. All handlers
// follow RESTful conventions with proper error handling and JSON responses.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

// GitOpsHandler handles GitOps management HTTP requests.
type GitOpsHandler struct {
	gitOpsService services.GitOpsService
	logger        *zerolog.Logger
}

// NewGitOpsHandler creates a new GitOpsHandler instance.
//
// Parameters:
//   - gitOpsService: GitOps service for Git operations
//   - logger: Logger for structured logging
//
// Returns:
//   - *GitOpsHandler: An instance of the GitOps handler
func NewGitOpsHandler(
	gitOpsService services.GitOpsService,
	logger *zerolog.Logger,
) *GitOpsHandler {
	return &GitOpsHandler{
		gitOpsService: gitOpsService,
		logger:        logger,
	}
}

// SyncRepository handles POST /gitops/repositories/{id}/sync endpoint.
//
// @Summary Synchronize a Git repository
// @Description Triggers synchronization of a Git repository to discover and
// update pipelines.
// @Tags gitops
// @Accept json
// @Produce json
// @Param id path string true "Repository ID"
// @Success 200 {object} dto.SyncRepositoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /gitops/repositories/{id}/sync [post].
func (h *GitOpsHandler) SyncRepository(c echo.Context) error {
	repositoryIDStr := c.Param("id")
	if repositoryIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Repository ID is required",
			Code:    "MISSING_REPOSITORY_ID",
			Details: "Please provide a valid repository ID",
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

	h.logger.Info().
		Str("repository_id", repositoryIDStr).
		Str("triggered_by", currentUserID).
		Msg("Starting manual repository synchronization")

	// Trigger synchronization
	result, err := h.gitOpsService.SyncRepository(
		c.Request().Context(),
		repositoryIDStr,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("repository_id", repositoryIDStr).
			Str("triggered_by", currentUserID).
			Msg("Failed to synchronize repository")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to synchronize repository",
			Code:    "SYNC_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	response := dto.SyncRepositoryResponse{
		Success:    result.Success,
		CommitHash: result.CommitHash,
		Message:    result.Message,
		Pipelines:  result.Pipelines,
		Duration:   result.Duration.String(),
		SyncedAt:   result.SyncedAt.Format("2006-01-02T15:04:05Z"),
		Changes: dto.SyncChangesResponse{
			Added:    result.Changes.Added,
			Modified: result.Changes.Modified,
			Deleted:  result.Changes.Deleted,
		},
	}

	if len(result.Errors) > 0 {
		response.Errors = result.Errors
	}

	h.logger.Info().
		Str("repository_id", repositoryIDStr).
		Str("triggered_by", currentUserID).
		Bool("success", result.Success).
		Int("pipelines_found", len(result.Pipelines)).
		Dur("duration", result.Duration).
		Msg("Repository synchronization completed")

	return c.JSON(http.StatusOK, response)
}

// HandleWebhook handles POST /gitops/webhooks endpoint.
//
// @Summary Handle Git webhook
// @Description Processes Git webhook events to trigger automatic repository
// synchronization.
// @Tags gitops
// @Accept json
// @Produce json
// @Param webhook body dto.WebhookRequest true "Webhook payload"
// @Success 200 {object} dto.WebhookResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /gitops/webhooks [post].
func (h *GitOpsHandler) HandleWebhook(c echo.Context) error {
	// Parse webhook payload
	var req dto.WebhookRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind webhook request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid webhook data",
		})
	}

	// Convert to service payload
	payload := &services.WebhookPayload{
		EventType: req.EventType,
		RepoURL:   req.RepoURL,
		Branch:    req.Branch,
		Commit:    req.Commit,
		Timestamp: req.Timestamp,
		Payload:   req.Payload,
	}

	h.logger.Info().
		Str("event_type", req.EventType).
		Str("repo_url", req.RepoURL).
		Str("branch", req.Branch).
		Str("commit", req.Commit).
		Msg("Processing Git webhook")

	// Handle webhook
	err := h.gitOpsService.HandleWebhook(c.Request().Context(), payload)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("event_type", req.EventType).
			Str("repo_url", req.RepoURL).
			Msg("Failed to handle webhook")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to handle webhook",
			Code:    "WEBHOOK_FAILED",
			Details: "Please try again later",
		})
	}

	response := dto.WebhookResponse{
		Success:   true,
		Message:   "Webhook processed successfully",
		Timestamp: payload.Timestamp,
	}

	h.logger.Info().
		Str("event_type", req.EventType).
		Str("repo_url", req.RepoURL).
		Msg("Webhook processed successfully")

	return c.JSON(http.StatusOK, response)
}

// ValidateRepository handles POST /gitops/repositories/validate endpoint.
//
// @Summary Validate Git repository
// @Description Validates that a Git repository is accessible and can be
// synchronized.
// @Tags gitops
// @Accept json
// @Produce json
// @Param request body dto.ValidateRepositoryRequest true "Repository validation
// data"
// @Success 200 {object} dto.ValidateRepositoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /gitops/repositories/validate [post].
func (h *GitOpsHandler) ValidateRepository(c echo.Context) error {
	// Parse request body
	var req dto.ValidateRepositoryRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind repository validation request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid repository data",
		})
	}

	if req.URL == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Repository URL is required",
			Code:    "MISSING_URL",
			Details: "Please provide a valid Git repository URL",
		})
	}

	h.logger.Info().
		Str("repo_url", req.URL).
		Msg("Validating Git repository")

	// Validate repository
	err := h.gitOpsService.ValidateGitRepository(c.Request().Context(), req.URL)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("repo_url", req.URL).
			Msg("Repository validation failed")

		return c.JSON(http.StatusBadRequest, dto.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Repository validation failed: " + err.Error(),
			URL:     req.URL,
		})
	}

	// Get latest commit information
	commitInfo, err := h.gitOpsService.GetLatestCommit(
		c.Request().Context(),
		req.URL,
		req.Branch,
	)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("repo_url", req.URL).
			Msg("Failed to get latest commit information")

		// Don't fail the validation, just log the warning
		commitInfo = nil
	}

	response := dto.ValidateRepositoryResponse{
		Valid:   true,
		Message: "Repository is accessible and valid",
		URL:     req.URL,
	}

	if commitInfo != nil {
		response.LatestCommit = &dto.CommitInfoResponse{
			Hash:      commitInfo.Hash,
			Message:   commitInfo.Message,
			Author:    commitInfo.Author,
			Email:     commitInfo.Email,
			Timestamp: commitInfo.Timestamp.Format("2006-01-02T15:04:05Z"),
			Branch:    commitInfo.Branch,
		}
	}

	h.logger.Info().
		Str("repo_url", req.URL).
		Bool("valid", response.Valid).
		Msg("Repository validation completed")

	return c.JSON(http.StatusOK, response)
}

// GetPendingSync handles GET /gitops/repositories/pending-sync endpoint.
//
// @Summary Get pending synchronizations
// @Description Retrieves repositories that need to be synchronized.
// @Tags gitops
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of repositories to return"
// default(10)
// @Success 200 {object} dto.PendingSyncResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /gitops/repositories/pending-sync [get].
func (h *GitOpsHandler) GetPendingSync(c echo.Context) error {
	// Parse limit parameter
	limitStr := c.QueryParam("limit")
	limit := 10 // default limit

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil &&
			parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	h.logger.Info().
		Int("limit", limit).
		Msg("Getting pending repository synchronizations")

	// This would require additional methods in the GitOpsService
	// For now, return a basic response
	response := dto.PendingSyncResponse{
		Repositories: []dto.PendingRepositoryResponse{}, // Would be populated by service
		Total:        0,
		Limit:        limit,
		Message:      "Pending synchronization retrieval not yet implemented",
	}

	h.logger.Info().
		Int("limit", limit).
		Int("count", response.Total).
		Msg("Pending synchronizations retrieved")

	return c.JSON(http.StatusOK, response)
}

// GetSyncStatus handles GET /gitops/repositories/{id}/sync-status endpoint.
//
// @Summary Get repository sync status
// @Description Retrieves the synchronization status of a specific repository.
// @Tags gitops
// @Accept json
// @Produce json
// @Param id path string true "Repository ID"
// @Success 200 {object} dto.SyncStatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /gitops/repositories/{id}/sync-status [get].
func (h *GitOpsHandler) GetSyncStatus(c echo.Context) error {
	repositoryIDStr := c.Param("id")
	if repositoryIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Repository ID is required",
			Code:    "MISSING_REPOSITORY_ID",
			Details: "Please provide a valid repository ID",
		})
	}

	h.logger.Info().
		Str("repository_id", repositoryIDStr).
		Msg("Getting repository sync status")

	// This would require additional methods in the GitOpsService
	// For now, return a basic response
	response := dto.SyncStatusResponse{
		RepositoryID: repositoryIDStr,
		Status:       "unknown",
		Message:      "Sync status retrieval not yet implemented",
	}

	h.logger.Info().
		Str("repository_id", repositoryIDStr).
		Str("status", response.Status).
		Msg("Repository sync status retrieved")

	return c.JSON(http.StatusOK, response)
}
