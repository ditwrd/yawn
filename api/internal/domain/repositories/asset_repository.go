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

// Package repositories provides data access layer implementations for
// domain entities using GORM.
//
// This package contains repository interfaces and implementations for
// managing database operations with proper error handling and logging.
// All repositories support pagination, filtering, and soft delete patterns.
package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ditwrd/yawn/api/internal/domain/models"
)

// AssetRepository defines the interface for asset data operations.
//
// Provides methods for CRUD operations, project-based filtering, and
// asset search functionality. All operations are context-aware and
// include proper error handling.
type AssetRepository interface {
	// Create inserts a new asset into the database
	Create(ctx context.Context, asset *models.Asset) error

	// GetByID retrieves an asset by its ID
	GetByID(ctx context.Context, id string) (*models.Asset, error)

	// GetByProjectID retrieves all assets for a specific project with pagination
	GetByProjectID(
		ctx context.Context,
		projectID string,
		limit, offset int,
	) ([]*models.Asset, error)

	// GetByRepositoryID retrieves all assets for a specific repository
	GetByRepositoryID(
		ctx context.Context,
		repositoryID string,
	) ([]*models.Asset, error)

	// List retrieves all assets with pagination and optional filtering
	List(
		ctx context.Context,
		limit, offset int,
		filters AssetFilters,
	) ([]*models.Asset, error)

	// Update modifies an existing asset in the database
	Update(ctx context.Context, asset *models.Asset) error

	// Delete performs a soft delete on an asset
	Delete(ctx context.Context, id string) error

	// Search finds assets by name or description with pagination
	Search(
		ctx context.Context,
		query string,
		limit, offset int,
	) ([]*models.Asset, error)

	// GetVersionHistory retrieves all versions of an asset by name and project
	GetVersionHistory(
		ctx context.Context,
		projectID, assetName string,
	) ([]*models.Asset, error)

	// GetLatestVersion retrieves the latest version of an asset by name and
	// project
	GetLatestVersion(
		ctx context.Context,
		projectID, assetName string,
	) (*models.Asset, error)

	// Count returns the total number of assets matching the filters
	Count(ctx context.Context, filters AssetFilters) (int64, error)

	// Exists checks if an asset exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByName checks if an asset exists by name within a project
	ExistsByName(ctx context.Context, projectID, name string) (bool, error)
}

// AssetFilters defines filtering options for asset queries.
type AssetFilters struct {
	ProjectID    string
	RepositoryID string
	Name         string
	Version      string
	Search       string // General search across name and description
}

// assetRepository implements the AssetRepository interface using GORM.
type assetRepository struct {
	db     *gorm.DB
	logger interface {
		Info(msg string, fields ...any)
	}
}

// getCaseInsensitiveOperator returns the appropriate case-insensitive operator
// based on the database dialect (ILIKE for PostgreSQL, LIKE for SQLite).
func (r *assetRepository) getCaseInsensitiveOperator() string {
	if dialector, ok := r.db.Dialector.(interface{ Name() string }); ok {
		if dialector.Name() == "postgres" {
			return "ILIKE"
		}
	}

	return "LIKE"
}

// NewAssetRepository creates a new instance of AssetRepository.
//
// Parameters:
//   - db: GORM database instance
//   - logger: Logger for debugging and monitoring
//
// Returns:
//   - AssetRepository: An instance of the asset repository
func NewAssetRepository(db *gorm.DB, logger interface {
	Info(msg string, fields ...any)
},
) AssetRepository {
	return &assetRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new asset into the database.
func (r *assetRepository) Create(
	ctx context.Context,
	asset *models.Asset,
) error {
	err := r.db.WithContext(ctx).Create(asset).Error
	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Asset created successfully",
			"asset_id", asset.ID,
			"name", asset.Name,
			"project_id", asset.ProjectID,
		)
	}

	return nil
}

// GetByID retrieves an asset by its ID.
func (r *assetRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Asset, error) {
	var asset models.Asset

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Repository").
		Preload("Pipelines").
		Where("id = ?", id).
		First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("asset with id %s not found", id)
		}

		return nil, fmt.Errorf("failed to get asset by id %s: %w", id, err)
	}

	return &asset, nil
}

// GetByProjectID retrieves all assets for a specific project with pagination.
func (r *assetRepository) GetByProjectID(
	ctx context.Context,
	projectID string,
	limit, offset int,
) ([]*models.Asset, error) {
	var assets []*models.Asset

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Repository").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&assets).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get assets for project %s: %w",
			projectID,
			err,
		)
	}

	return assets, nil
}

// GetByRepositoryID retrieves all assets for a specific repository.
func (r *assetRepository) GetByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*models.Asset, error) {
	var assets []*models.Asset

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Repository").
		Where("repository_id = ?", repositoryID).
		Order("created_at DESC").
		Find(&assets).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get assets for repository %s: %w",
			repositoryID,
			err,
		)
	}

	return assets, nil
}

// List retrieves all assets with pagination and optional filtering.
func (r *assetRepository) List(
	ctx context.Context,
	limit, offset int,
	filters AssetFilters,
) ([]*models.Asset, error) {
	var assets []*models.Asset

	query := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Repository")

	// Apply filters
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}

	if filters.RepositoryID != "" {
		query = query.Where("repository_id = ?", filters.RepositoryID)
	}

	if filters.Name != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where("name "+operator+" ?", "%"+filters.Name+"%")
	}

	if filters.Version != "" {
		query = query.Where("version = ?", filters.Version)
	}

	if filters.Search != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where(
			"name "+operator+" ? OR description "+operator+" ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&assets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	return assets, nil
}

// Update modifies an existing asset in the database.
func (r *assetRepository) Update(
	ctx context.Context,
	asset *models.Asset,
) error {
	result := r.db.WithContext(ctx).Save(asset)
	if result.Error != nil {
		return fmt.Errorf("failed to update asset: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("asset with id %s not found or no changes made", asset.ID)
	}

	if r.logger != nil {
		r.logger.Info("Asset updated successfully",
			"asset_id", asset.ID,
			"name", asset.Name,
		)
	}

	return nil
}

// Delete performs a soft delete on an asset.
func (r *assetRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Asset{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete asset: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("asset with id %s not found", id)
	}

	if r.logger != nil {
		r.logger.Info("Asset deleted successfully", "asset_id", id)
	}

	return nil
}

// Search finds assets by name or description with pagination.
func (r *assetRepository) Search(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]*models.Asset, error) {
	var assets []*models.Asset

	operator := r.getCaseInsensitiveOperator()

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Repository").
		Where("name "+operator+" ? OR description "+operator+" ?", "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&assets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search assets: %w", err)
	}

	return assets, nil
}

// GetVersionHistory retrieves all versions of an asset by name and project.
func (r *assetRepository) GetVersionHistory(
	ctx context.Context,
	projectID, assetName string,
) ([]*models.Asset, error) {
	var assets []*models.Asset

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Repository").
		Where("project_id = ? AND name = ?", projectID, assetName).
		Order("created_at DESC").
		Find(&assets).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get version history for asset %s: %w",
			assetName,
			err,
		)
	}

	return assets, nil
}

// GetLatestVersion retrieves the latest version of an asset by name and
// project.
func (r *assetRepository) GetLatestVersion(
	ctx context.Context,
	projectID, assetName string,
) (*models.Asset, error) {
	var asset models.Asset

	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, assetName).
		Order("created_at DESC").
		First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(
				"no versions found for asset %s in project %s",
				assetName,
				projectID,
			)
		}

		return nil, fmt.Errorf(
			"failed to get latest version for asset %s: %w",
			assetName,
			err,
		)
	}

	return &asset, nil
}

// Count returns the total number of assets matching the filters.
func (r *assetRepository) Count(
	ctx context.Context,
	filters AssetFilters,
) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.Asset{})

	// Apply filters
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}

	if filters.RepositoryID != "" {
		query = query.Where("repository_id = ?", filters.RepositoryID)
	}

	if filters.Name != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where("name "+operator+" ?", "%"+filters.Name+"%")
	}

	if filters.Version != "" {
		query = query.Where("version = ?", filters.Version)
	}

	if filters.Search != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where(
			"name "+operator+" ? OR description "+operator+" ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count assets: %w", err)
	}

	return count, nil
}

// Exists checks if an asset exists by ID.
func (r *assetRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Asset{}).
		Where("id = ?", id).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to check asset existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if an asset exists by name within a project.
func (r *assetRepository) ExistsByName(
	ctx context.Context,
	projectID, name string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&models.Asset{}).
		Where("project_id = ? AND name = ?", projectID, name).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check asset existence by name: %w", err)
	}

	return count > 0, nil
}
