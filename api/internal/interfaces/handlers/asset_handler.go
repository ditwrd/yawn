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

// Package handlers provides HTTP request handlers for asset management
// operations.
//
// This package contains handlers for asset CRUD operations with proper
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

// AssetHandler handles asset management HTTP requests.
type AssetHandler struct {
	assetService services.AssetService
	logger       *zerolog.Logger
}

// NewAssetHandler creates a new AssetHandler instance.
//
// Parameters:
//   - assetService: Asset service for asset operations
//   - logger: Logger for structured logging
//
// Returns:
//   - *AssetHandler: An instance of the asset handler
func NewAssetHandler(
	assetService services.AssetService,
	logger *zerolog.Logger,
) *AssetHandler {
	return &AssetHandler{
		assetService: assetService,
		logger:       logger,
	}
}

// ListAssets handles GET /assets endpoint.
//
// @Summary List all assets
// @Description Returns a paginated list of assets with optional filtering.
// @Tags assets
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Param project_id query string false "Filter by project ID"
// @Param repository_id query string false "Filter by repository ID"
// @Param name query string false "Filter by asset name (partial match)"
// @Param version query string false "Filter by version (exact match)"
// @Param search query string false "Search across name and description"
// @Success 200 {object} dto.AssetListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets [get].
func (h *AssetHandler) ListAssets(c echo.Context) error {
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
	filters := services.AssetListFilters{
		ProjectID:    c.QueryParam("project_id"),
		RepositoryID: c.QueryParam("repository_id"),
		Name:         c.QueryParam("name"),
		Version:      c.QueryParam("version"),
		Search:       c.QueryParam("search"),
	}

	// Get assets from service
	result, err := h.assetService.List(
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
			Msg("Failed to list assets")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to retrieve assets",
			Code:    "LIST_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	assetResponses := make([]dto.AssetResponse, len(result.Assets))
	for i, asset := range result.Assets {
		assetResponses[i] = h.assetToResponse(asset)
	}

	h.logger.Info().
		Int("count", len(assetResponses)).
		Msg("Assets listed successfully")

	return c.JSON(http.StatusOK, dto.AssetListResponse{
		Assets: assetResponses,
		Total:  result.Total,
		Page:   result.Page,
		Limit:  result.Limit,
	})
}

// GetAsset handles GET /assets/{id} endpoint.
//
// @Summary Get asset by ID
// @Description Returns asset details with full relationships.
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} dto.AssetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id} [get].
func (h *AssetHandler) GetAsset(c echo.Context) error {
	assetIDStr := c.Param("id")
	if assetIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Asset ID is required",
			Code:    "MISSING_ASSET_ID",
			Details: "Please provide a valid asset ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(assetIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid asset ID format",
			Code:    "INVALID_ASSET_ID",
			Details: "Asset ID must be a valid UUID",
		})
	}

	// Get asset from service
	asset, err := h.assetService.GetByID(c.Request().Context(), assetIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("asset_id", assetIDStr).
			Msg("Failed to get asset")

		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Asset not found",
			Code:    "ASSET_NOT_FOUND",
			Details: "The requested asset does not exist",
		})
	}

	// Convert to response DTO
	assetResponse := h.assetToResponse(asset)

	h.logger.Info().
		Str("asset_id", assetIDStr).
		Msg("Asset retrieved successfully")

	return c.JSON(http.StatusOK, assetResponse)
}

// CreateAsset handles POST /assets endpoint.
//
// @Summary Create a new asset
// @Description Creates a new asset with the provided details.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body dto.CreateAssetRequest true "Asset creation data"
// @Success 201 {object} dto.AssetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets [post].
func (h *AssetHandler) CreateAsset(c echo.Context) error {
	// Parse request body
	var req dto.CreateAssetRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Msg("Failed to bind create asset request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid asset data",
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

	// Check if user can create assets in the project
	if err := h.assetService.CanCreate(c.Request().Context(), req.ProjectID, currentUserID); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("project_id", req.ProjectID).
			Msg("Access denied: user cannot create assets in project")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "CREATE_ACCESS_DENIED",
			Details: "You don't have permission to create assets in this project",
		})
	}

	// Create service request
	serviceReq := &services.CreateAssetRequest{
		Name:         req.Name,
		Description:  req.Description,
		Version:      req.Version,
		ProjectID:    req.ProjectID,
		RepositoryID: req.RepositoryID,
	}

	// Create asset
	asset, err := h.assetService.Create(c.Request().Context(), serviceReq)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("name", req.Name).
			Str("project_id", req.ProjectID).
			Msg("Failed to create asset")

		// Handle specific validation errors
		if strings.Contains(err.Error(), "already exists") {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Asset already exists",
				Code:    "ASSET_EXISTS",
				Details: err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create asset",
			Code:    "CREATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	assetResponse := h.assetToResponse(asset)

	h.logger.Info().
		Str("asset_id", asset.ID.String()).
		Str("name", asset.Name).
		Msg("Asset created successfully")

	return c.JSON(http.StatusCreated, assetResponse)
}

// UpdateAsset handles PUT /assets/{id} endpoint.
//
// @Summary Update asset information
// @Description Updates asset information with the provided details.
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Param request body dto.UpdateAssetRequest true "Asset update data"
// @Success 200 {object} dto.AssetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id} [put].
func (h *AssetHandler) UpdateAsset(c echo.Context) error {
	assetIDStr := c.Param("id")
	if assetIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Asset ID is required",
			Code:    "MISSING_ASSET_ID",
			Details: "Please provide a valid asset ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(assetIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid asset ID format",
			Code:    "INVALID_ASSET_ID",
			Details: "Asset ID must be a valid UUID",
		})
	}

	// Parse request body
	var req dto.UpdateAssetRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn().
			Err(err).
			Str("asset_id", assetIDStr).
			Msg("Failed to bind update asset request")

		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: "Request body must contain valid asset data",
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

	// Check if user can update assets in the project (requires maintainer role)
	if err := h.assetService.ValidateAccess(c.Request().Context(), assetIDStr, currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("asset_id", assetIDStr).
			Msg("Access denied: user cannot update asset")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "UPDATE_ACCESS_DENIED",
			Details: "You don't have permission to update this asset",
		})
	}

	// Convert DTO to service request
	serviceReq := &services.UpdateAssetRequest{
		Name:         req.Name,
		Description:  req.Description,
		Version:      req.Version,
		RepositoryID: req.RepositoryID,
	}

	// Update asset
	asset, err := h.assetService.Update(
		c.Request().Context(),
		assetIDStr,
		serviceReq,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("asset_id", assetIDStr).
			Msg("Failed to update asset")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update asset",
			Code:    "UPDATE_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	assetResponse := h.assetToResponse(asset)

	h.logger.Info().
		Str("asset_id", assetIDStr).
		Msg("Asset updated successfully")

	return c.JSON(http.StatusOK, assetResponse)
}

// DeleteAsset handles DELETE /assets/{id} endpoint.
//
// @Summary Delete asset
// @Description Soft deletes an asset. Requires maintainer or owner role in the
// project.
// @Tags assets
// @Accept json
// @Produce json
// @Param id path string true "Asset ID"
// @Success 200 {object} dto.AssetDeleteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{id} [delete].
func (h *AssetHandler) DeleteAsset(c echo.Context) error {
	assetIDStr := c.Param("id")
	if assetIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Asset ID is required",
			Code:    "MISSING_ASSET_ID",
			Details: "Please provide a valid asset ID",
		})
	}

	// Validate UUID format
	_, err := uuid.FromString(assetIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid asset ID format",
			Code:    "INVALID_ASSET_ID",
			Details: "Asset ID must be a valid UUID",
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

	// Check if user can delete assets in the project (requires maintainer role)
	if err := h.assetService.ValidateAccess(c.Request().Context(), assetIDStr, currentUserID, models.ProjectRoleMaintainer); err != nil {
		h.logger.Warn().
			Err(err).
			Str("user_id", currentUserID).
			Str("asset_id", assetIDStr).
			Msg("Access denied: user cannot delete asset")

		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Access denied",
			Code:    "DELETE_ACCESS_DENIED",
			Details: "You don't have permission to delete this asset",
		})
	}

	// Delete asset
	err = h.assetService.Delete(c.Request().Context(), assetIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("asset_id", assetIDStr).
			Msg("Failed to delete asset")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete asset",
			Code:    "DELETE_FAILED",
			Details: "Please try again later",
		})
	}

	h.logger.Info().
		Str("asset_id", assetIDStr).
		Msg("Asset deleted successfully")

	return c.JSON(http.StatusOK, dto.AssetDeleteResponse{
		Message: "Asset deleted successfully",
	})
}

// SearchAssets handles GET /assets/search endpoint.
//
// @Summary Search assets
// @Description Searches assets by name or description with pagination.
// @Tags assets
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} dto.AssetSearchResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/search [get].
func (h *AssetHandler) SearchAssets(c echo.Context) error {
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

	// Search assets
	result, err := h.assetService.Search(
		c.Request().Context(),
		query,
		page,
		limit,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("query", query).
			Msg("Failed to search assets")

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to search assets",
			Code:    "SEARCH_FAILED",
			Details: "Please try again later",
		})
	}

	// Convert to response DTO
	assetResponses := make([]dto.AssetResponse, len(result.Assets))
	for i, asset := range result.Assets {
		assetResponses[i] = h.assetToResponse(asset)
	}

	h.logger.Info().
		Str("query", query).
		Int("count", len(assetResponses)).
		Msg("Asset search completed successfully")

	return c.JSON(http.StatusOK, dto.AssetSearchResponse{
		Assets: assetResponses,
		Total:  result.Total,
		Query:  query,
		Page:   result.Page,
		Limit:  result.Limit,
	})
}

// GetAssetVersionHistory handles GET /assets/{project_id}/{asset_name}/versions
// endpoint.
//
// @Summary Get asset version history
// @Description Returns all versions of an asset within a project.
// @Tags assets
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param asset_name path string true "Asset name"
// @Success 200 {object} dto.AssetVersionHistoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /assets/{project_id}/{asset_name}/versions [get].
func (h *AssetHandler) GetAssetVersionHistory(c echo.Context) error {
	projectID := c.Param("project_id")
	assetName := c.Param("asset_name")

	if projectID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Project ID is required",
			Code:    "MISSING_PROJECT_ID",
			Details: "Please provide a valid project ID",
		})
	}

	if assetName == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Asset name is required",
			Code:    "MISSING_ASSET_NAME",
			Details: "Please provide a valid asset name",
		})
	}

	// Get version history
	assets, err := h.assetService.GetVersionHistory(
		c.Request().Context(),
		projectID,
		assetName,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Str("asset_name", assetName).
			Msg("Failed to get asset version history")

		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Asset not found",
			Code:    "ASSET_NOT_FOUND",
			Details: "The requested asset does not exist",
		})
	}

	// Convert to response DTO
	versions := make([]dto.AssetVersion, len(assets))
	for i, asset := range assets {
		versions[i] = dto.AssetVersion{
			ID:        asset.ID.String(),
			Version:   asset.Version,
			CreatedAt: asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	h.logger.Info().
		Str("project_id", projectID).
		Str("asset_name", assetName).
		Int("versions_count", len(versions)).
		Msg("Asset version history retrieved successfully")

	return c.JSON(http.StatusOK, dto.AssetVersionHistoryResponse{
		AssetName: assetName,
		Versions:  versions,
	})
}

// assetToResponse converts an asset model to a response DTO.
func (h *AssetHandler) assetToResponse(asset *models.Asset) dto.AssetResponse {
	response := dto.AssetResponse{
		ID:          asset.ID.String(),
		Name:        asset.Name,
		Description: asset.Description,
		Version:     asset.Version,
		ProjectID:   asset.ProjectID.String(),
		CreatedAt:   asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add repository ID if present
	if asset.RepositoryID != nil {
		repoID := asset.RepositoryID.String()
		response.RepositoryID = &repoID
	}

	// Add project summary if loaded
	if asset.Project.ID != uuid.Nil {
		response.Project = &dto.ProjectSummaryResponse{
			ID:          asset.Project.ID.String(),
			Name:        asset.Project.Name,
			Description: asset.Project.Description,
			Visibility:  asset.Project.Visibility,
		}
	}

	// Add repository summary if loaded
	if asset.Repository != nil {
		response.Repository = &dto.RepositoryResponse{
			ID:           asset.Repository.ID.String(),
			URL:          asset.Repository.URL,
			Branch:       asset.Repository.Branch,
			LatestCommit: asset.Repository.LatestCommit,
			SyncStatus:   string(asset.Repository.SyncStatus),
			ProjectID:    asset.Repository.ProjectID.String(),
			CreatedAt:    asset.Repository.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    asset.Repository.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Add pipeline IDs if loaded
	if len(asset.Pipelines) > 0 {
		pipelineIDs := make([]string, len(asset.Pipelines))
		for i, pipeline := range asset.Pipelines {
			pipelineIDs[i] = pipeline.ID.String()
		}

		response.PipelineIDs = pipelineIDs
	}

	return response
}
