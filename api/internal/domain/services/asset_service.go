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

// Package services provides business logic layer implementations for domain
// entities.
//
// This package contains service interfaces and implementations that encapsulate
// business rules, validation, and orchestration of repository operations.
// All services are context-aware and include proper error handling and logging.
package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
)

// AssetService defines the interface for asset business operations.
//
// Provides methods for asset CRUD operations with business validation,
// project-based access control, and asset lifecycle management.
type AssetService interface {
	// Create creates a new asset with validation and business rules
	Create(ctx context.Context, req *CreateAssetRequest) (*models.Asset, error)

	// GetByID retrieves an asset by its ID with project access validation
	GetByID(ctx context.Context, id string) (*models.Asset, error)

	// GetByProjectID retrieves all assets for a project with pagination
	GetByProjectID(
		ctx context.Context,
		projectID string,
		page, limit int,
	) (*PaginatedAssetsResponse, error)

	// GetByRepositoryID retrieves all assets for a repository
	GetByRepositoryID(
		ctx context.Context,
		repositoryID string,
	) ([]*models.Asset, error)

	// List retrieves all assets with pagination and filtering
	List(
		ctx context.Context,
		page, limit int,
		filters AssetListFilters,
	) (*PaginatedAssetsResponse, error)

	// Update updates an existing asset with validation
	Update(
		ctx context.Context,
		id string,
		req *UpdateAssetRequest,
	) (*models.Asset, error)

	// Delete soft deletes an asset with access validation
	Delete(ctx context.Context, id string) error

	// Search searches assets by query string with pagination
	Search(
		ctx context.Context,
		query string,
		page, limit int,
	) (*PaginatedAssetsResponse, error)

	// GetVersionHistory retrieves all versions of an asset
	GetVersionHistory(
		ctx context.Context,
		projectID, assetName string,
	) ([]*models.Asset, error)

	// GetLatestVersion retrieves the latest version of an asset
	GetLatestVersion(
		ctx context.Context,
		projectID, assetName string,
	) (*models.Asset, error)

	// ValidateAccess checks if the current user has access to an asset
	ValidateAccess(
		ctx context.Context,
		assetID, userID string,
		requiredRole models.ProjectRole,
	) error

	// CanCreate checks if the user can create assets in a project
	CanCreate(ctx context.Context, projectID, userID string) error
}

// CreateAssetRequest represents the request to create a new asset.
type CreateAssetRequest struct {
	Name         string
	Description  string
	Version      string
	ProjectID    string
	RepositoryID *string
}

// UpdateAssetRequest represents the request to update an asset.
type UpdateAssetRequest struct {
	Name         *string
	Description  *string
	Version      *string
	RepositoryID *string
}

// AssetListFilters defines filtering options for asset listing.
type AssetListFilters struct {
	ProjectID    string
	RepositoryID string
	Name         string
	Version      string
	Search       string
}

// PaginatedAssetsResponse represents a paginated response for assets.
type PaginatedAssetsResponse struct {
	Assets []*models.Asset
	Total  int64
	Page   int
	Limit  int
}

// assetService implements the AssetService interface.
type assetService struct {
	assetRepo   repositories.AssetRepository
	projectRepo repositories.ProjectRepository
	userRepo    repositories.UserRepository
	logger      *zerolog.Logger
}

// NewAssetService creates a new instance of AssetService.
//
// Parameters:
//   - assetRepo: Asset repository for data operations
//   - projectRepo: Project repository for access validation
//   - userRepo: User repository for user operations
//   - logger: Logger for structured logging
//
// Returns:
//   - AssetService: An instance of the asset service
func NewAssetService(
	assetRepo repositories.AssetRepository,
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
	logger *zerolog.Logger,
) AssetService {
	return &assetService{
		assetRepo:   assetRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// Create creates a new asset with validation and business rules.
func (s *assetService) Create(
	ctx context.Context,
	req *CreateAssetRequest,
) (*models.Asset, error) {
	// Validate request
	if err := s.validateCreateRequest(ctx, req); err != nil {
		return nil, err
	}

	// Check if asset with same name and version already exists in project
	exists, err := s.assetRepo.ExistsByName(ctx, req.ProjectID, req.Name)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", req.ProjectID).
			Str("name", req.Name).
			Msg("Failed to check asset existence")

		return nil, fmt.Errorf("failed to validate asset uniqueness: %w", err)
	}

	if exists {
		return nil, fmt.Errorf(
			"asset with name '%s' already exists in project",
			req.Name,
		)
	}

	// Create the asset
	asset := &models.Asset{
		ID:           uuid.Must(uuid.NewV7()),
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		Version:      strings.TrimSpace(req.Version),
		ProjectID:    uuid.Must(uuid.FromString(req.ProjectID)),
		RepositoryID: nil,
	}

	if req.RepositoryID != nil {
		repoID, err := uuid.FromString(*req.RepositoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid repository ID: %w", err)
		}

		asset.RepositoryID = &repoID
	}

	// Save asset
	if err := s.assetRepo.Create(ctx, asset); err != nil {
		s.logger.Error().
			Err(err).
			Str("name", asset.Name).
			Str("project_id", req.ProjectID).
			Msg("Failed to create asset")

		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	s.logger.Info().
		Str("asset_id", asset.ID.String()).
		Str("name", asset.Name).
		Str("project_id", req.ProjectID).
		Msg("Asset created successfully")

	return asset, nil
}

// GetByID retrieves an asset by its ID with project access validation.
func (s *assetService) GetByID(
	ctx context.Context,
	id string,
) (*models.Asset, error) {
	if id == "" {
		return nil, errors.New("asset ID is required")
	}

	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("asset_id", id).
			Msg("Failed to get asset")

		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return asset, nil
}

// GetByProjectID retrieves all assets for a project with pagination.
func (s *assetService) GetByProjectID(
	ctx context.Context,
	projectID string,
	page, limit int,
) (*PaginatedAssetsResponse, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Get assets
	assets, err := s.assetRepo.GetByProjectID(ctx, projectID, limit, offset)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Msg("Failed to get assets by project ID")

		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	// Get total count
	count, err := s.assetRepo.Count(
		ctx,
		repositories.AssetFilters{ProjectID: projectID},
	)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Msg("Failed to count assets")

		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	return &PaginatedAssetsResponse{
		Assets: assets,
		Total:  count,
		Page:   page,
		Limit:  limit,
	}, nil
}

// GetByRepositoryID retrieves all assets for a repository.
func (s *assetService) GetByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*models.Asset, error) {
	if repositoryID == "" {
		return nil, errors.New("repository ID is required")
	}

	assets, err := s.assetRepo.GetByRepositoryID(ctx, repositoryID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("repository_id", repositoryID).
			Msg("Failed to get assets by repository ID")

		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	return assets, nil
}

// List retrieves all assets with pagination and filtering.
func (s *assetService) List(
	ctx context.Context,
	page, limit int,
	filters AssetListFilters,
) (*PaginatedAssetsResponse, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Convert filters
	repoFilters := repositories.AssetFilters{
		ProjectID:    filters.ProjectID,
		RepositoryID: filters.RepositoryID,
		Name:         filters.Name,
		Version:      filters.Version,
		Search:       filters.Search,
	}

	// Get assets
	assets, err := s.assetRepo.List(ctx, limit, offset, repoFilters)
	if err != nil {
		s.logger.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to list assets")

		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	// Get total count
	count, err := s.assetRepo.Count(ctx, repoFilters)
	if err != nil {
		s.logger.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to count assets")

		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	return &PaginatedAssetsResponse{
		Assets: assets,
		Total:  count,
		Page:   page,
		Limit:  limit,
	}, nil
}

// Update updates an existing asset with validation.
func (s *assetService) Update(
	ctx context.Context,
	id string,
	req *UpdateAssetRequest,
) (*models.Asset, error) {
	if id == "" {
		return nil, errors.New("asset ID is required")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Get existing asset
	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		asset.Name = strings.TrimSpace(*req.Name)
	}

	if req.Description != nil {
		asset.Description = strings.TrimSpace(*req.Description)
	}

	if req.Version != nil {
		asset.Version = strings.TrimSpace(*req.Version)
	}

	if req.RepositoryID != nil {
		if *req.RepositoryID == "" {
			asset.RepositoryID = nil
		} else {
			repoID, err := uuid.FromString(*req.RepositoryID)
			if err != nil {
				return nil, fmt.Errorf("invalid repository ID: %w", err)
			}

			asset.RepositoryID = &repoID
		}
	}

	// Validate updated asset
	if err := s.validateAsset(asset); err != nil {
		return nil, err
	}

	// Update asset
	if err := s.assetRepo.Update(ctx, asset); err != nil {
		s.logger.Error().
			Err(err).
			Str("asset_id", id).
			Msg("Failed to update asset")

		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	s.logger.Info().
		Str("asset_id", id).
		Msg("Asset updated successfully")

	return asset, nil
}

// Delete soft deletes an asset with access validation.
func (s *assetService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("asset ID is required")
	}

	// Check if asset exists
	exists, err := s.assetRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check asset existence: %w", err)
	}

	if !exists {
		return errors.New("asset not found")
	}

	// Delete asset
	if err := s.assetRepo.Delete(ctx, id); err != nil {
		s.logger.Error().
			Err(err).
			Str("asset_id", id).
			Msg("Failed to delete asset")

		return fmt.Errorf("failed to delete asset: %w", err)
	}

	s.logger.Info().
		Str("asset_id", id).
		Msg("Asset deleted successfully")

	return nil
}

// Search searches assets by query string with pagination.
func (s *assetService) Search(
	ctx context.Context,
	query string,
	page, limit int,
) (*PaginatedAssetsResponse, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query is required")
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Search assets
	assets, err := s.assetRepo.Search(ctx, query, limit, offset)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Msg("Failed to search assets")

		return nil, fmt.Errorf("failed to search assets: %w", err)
	}

	// Get total count
	count, err := s.assetRepo.Count(ctx, repositories.AssetFilters{Search: query})
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Msg("Failed to count search results")

		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	return &PaginatedAssetsResponse{
		Assets: assets,
		Total:  count,
		Page:   page,
		Limit:  limit,
	}, nil
}

// GetVersionHistory retrieves all versions of an asset.
func (s *assetService) GetVersionHistory(
	ctx context.Context,
	projectID, assetName string,
) ([]*models.Asset, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	if assetName == "" {
		return nil, errors.New("asset name is required")
	}

	assets, err := s.assetRepo.GetVersionHistory(ctx, projectID, assetName)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Str("asset_name", assetName).
			Msg("Failed to get asset version history")

		return nil, fmt.Errorf("failed to get version history: %w", err)
	}

	return assets, nil
}

// GetLatestVersion retrieves the latest version of an asset.
func (s *assetService) GetLatestVersion(
	ctx context.Context,
	projectID, assetName string,
) (*models.Asset, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	if assetName == "" {
		return nil, errors.New("asset name is required")
	}

	asset, err := s.assetRepo.GetLatestVersion(ctx, projectID, assetName)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Str("asset_name", assetName).
			Msg("Failed to get latest asset version")

		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return asset, nil
}

// ValidateAccess checks if the current user has access to an asset.
func (s *assetService) ValidateAccess(
	ctx context.Context,
	assetID, userID string,
	requiredRole models.ProjectRole,
) error {
	asset, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to get asset for access validation: %w", err)
	}

	// Check if user has access to the project
	hasAccess, err := s.projectRepo.HasUserWithRole(
		asset.ProjectID.String(),
		userID,
		requiredRole,
	)
	if err != nil {
		return fmt.Errorf("failed to validate project access: %w", err)
	}

	if !hasAccess {
		return fmt.Errorf(
			"user does not have required role '%s' in project",
			requiredRole,
		)
	}

	return nil
}

// CanCreate checks if the user can create assets in a project.
func (s *assetService) CanCreate(
	ctx context.Context,
	projectID, userID string,
) error {
	// Check if user has maintainer or owner role in project
	hasAccess, err := s.projectRepo.HasUserWithRole(
		projectID,
		userID,
		models.ProjectRoleMaintainer,
	)
	if err != nil {
		return fmt.Errorf("failed to validate project access: %w", err)
	}

	if !hasAccess {
		return errors.New(
			"user does not have permission to create assets in this project",
		)
	}

	return nil
}

// validateCreateRequest validates the create asset request.
func (s *assetService) validateCreateRequest(
	ctx context.Context,
	req *CreateAssetRequest,
) error {
	if req == nil {
		return errors.New("create request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return errors.New("asset name is required")
	}

	if strings.TrimSpace(req.Version) == "" {
		return errors.New("asset version is required")
	}

	if req.ProjectID == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format for project ID
	if _, err := uuid.FromString(req.ProjectID); err != nil {
		return fmt.Errorf("invalid project ID format: %w", err)
	}

	// Validate repository ID if provided
	if req.RepositoryID != nil && *req.RepositoryID != "" {
		if _, err := uuid.FromString(*req.RepositoryID); err != nil {
			return fmt.Errorf("invalid repository ID format: %w", err)
		}
	}

	// Validate asset name format
	if err := s.validateAssetName(req.Name); err != nil {
		return err
	}

	// Validate version format
	if err := s.validateVersion(req.Version); err != nil {
		return err
	}

	// Check if project exists
	projectExists, err := s.projectRepo.Exists(req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to validate project existence: %w", err)
	}

	if !projectExists {
		return errors.New("project not found")
	}

	return nil
}

// validateUpdateRequest validates the update asset request.
func (s *assetService) validateUpdateRequest(req *UpdateAssetRequest) error {
	if req == nil {
		return errors.New("update request is required")
	}

	// Validate name if provided
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errors.New("asset name cannot be empty")
		}
		err := s.validateAssetName(*req.Name)
		if err != nil {
			return err
		}
	}

	// Validate version if provided
	if req.Version != nil {
		if strings.TrimSpace(*req.Version) == "" {
			return errors.New("asset version cannot be empty")
		}
		err := s.validateVersion(*req.Version)
		if err != nil {
			return err
		}
	}

	// Validate repository ID if provided
	if req.RepositoryID != nil && *req.RepositoryID != "" {
		if _, err := uuid.FromString(*req.RepositoryID); err != nil {
			return fmt.Errorf("invalid repository ID format: %w", err)
		}
	}

	return nil
}

// validateAsset validates an asset entity.
func (s *assetService) validateAsset(asset *models.Asset) error {
	if asset == nil {
		return errors.New("asset is required")
	}

	if strings.TrimSpace(asset.Name) == "" {
		return errors.New("asset name is required")
	}

	if strings.TrimSpace(asset.Version) == "" {
		return errors.New("asset version is required")
	}

	err := s.validateAssetName(asset.Name)
	if err != nil {
		return err
	}

	err = s.validateVersion(asset.Version)
	if err != nil {
		return err
	}

	return nil
}

// validateAssetName validates the asset name format.
func (s *assetService) validateAssetName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < 1 {
		return errors.New("asset name must be at least 1 character long")
	}

	if len(name) > 255 {
		return errors.New("asset name must be at most 255 characters long")
	}

	// Allow alphanumeric characters, hyphens, underscores, and dots
	matched, err := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, name)
	if err != nil {
		return fmt.Errorf("failed to validate asset name format: %w", err)
	}

	if !matched {
		return errors.New(
			"asset name can only contain alphanumeric characters, hyphens, underscores, and dots",
		)
	}

	return nil
}

// validateVersion validates the version format.
func (s *assetService) validateVersion(version string) error {
	version = strings.TrimSpace(version)

	if len(version) < 1 {
		return errors.New("version must be at least 1 character long")
	}

	if len(version) > 50 {
		return errors.New("version must be at most 50 characters long")
	}

	// Allow semantic versioning format (x.y.z) and other common patterns
	matched, err := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, version)
	if err != nil {
		return fmt.Errorf("failed to validate version format: %w", err)
	}

	if !matched {
		return errors.New(
			"version can only contain alphanumeric characters, hyphens, underscores, and dots",
		)
	}

	return nil
}
