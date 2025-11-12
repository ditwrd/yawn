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

// RepositoryRepository defines the interface for repository data operations.
//
// Provides methods for CRUD operations, project-based filtering, and
// Git synchronization management. All operations are context-aware and
// include proper error handling.
type RepositoryRepository interface {
	// Create inserts a new repository into the database
	Create(ctx context.Context, repository *models.Repository) error

	// GetByID retrieves a repository by its ID
	GetByID(ctx context.Context, id string) (*models.Repository, error)

	// GetByProjectID retrieves all repositories for a specific project with
	// pagination
	GetByProjectID(
		ctx context.Context,
		projectID string,
		limit, offset int,
	) ([]*models.Repository, error)

	// List retrieves all repositories with pagination and optional filtering
	List(
		ctx context.Context,
		limit, offset int,
		filters RepositoryFilters,
	) ([]*models.Repository, error)

	// ListByURL retrieves all repositories matching a specific URL
	ListByURL(ctx context.Context, url string) ([]*models.Repository, error)

	// Update modifies an existing repository in the database
	Update(ctx context.Context, repository *models.Repository) error

	// Delete performs a soft delete on a repository
	Delete(ctx context.Context, id string) error

	// UpdateStatus updates the sync status of a repository
	UpdateStatus(
		ctx context.Context,
		id string,
		status models.RepositoryStatus,
	) error

	// UpdateLatestCommit updates the latest commit hash of a repository
	UpdateLatestCommit(ctx context.Context, id, commitHash string) error

	// GetPendingSync retrieves repositories that need synchronization
	GetPendingSync(ctx context.Context, limit int) ([]*models.Repository, error)

	// Search finds repositories by URL or branch with pagination
	Search(
		ctx context.Context,
		query string,
		limit, offset int,
	) ([]*models.Repository, error)

	// Count returns the total number of repositories matching the filters
	Count(ctx context.Context, filters RepositoryFilters) (int64, error)

	// Exists checks if a repository exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByURL checks if a repository exists by URL within a project
	ExistsByURL(ctx context.Context, projectID, url string) (bool, error)
}

// RepositoryFilters defines filtering options for repository queries.
type RepositoryFilters struct {
	ProjectID  string
	URL        string
	Branch     string
	SyncStatus models.RepositoryStatus
	Search     string // General search across URL and branch
}

// repositoryRepository implements the RepositoryRepository interface using
// GORM.
type repositoryRepository struct {
	db     *gorm.DB
	logger interface {
		Info(msg string, fields ...any)
	}
}

// NewRepositoryRepository creates a new instance of RepositoryRepository.
//
// Parameters:
//   - db: GORM database instance
//   - logger: Logger for debugging and monitoring
//
// Returns:
//   - RepositoryRepository: An instance of the repository
func NewRepositoryRepository(db *gorm.DB, logger interface {
	Info(msg string, fields ...any)
},
) RepositoryRepository {
	return &repositoryRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new repository into the database.
func (r *repositoryRepository) Create(
	ctx context.Context,
	repository *models.Repository,
) error {
	err := r.db.WithContext(ctx).Create(repository).Error
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Repository created successfully",
			"repository_id", repository.ID,
			"url", repository.URL,
			"project_id", repository.ProjectID,
		)
	}

	return nil
}

// GetByID retrieves a repository by its ID.
func (r *repositoryRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Repository, error) {
	var repository models.Repository

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Assets").
		Where("id = ?", id).
		First(&repository).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository with id %s not found", id)
		}

		return nil, fmt.Errorf("failed to get repository by id %s: %w", id, err)
	}

	return &repository, nil
}

// GetByProjectID retrieves all repositories for a specific project with
// pagination.
func (r *repositoryRepository) GetByProjectID(
	ctx context.Context,
	projectID string,
	limit, offset int,
) ([]*models.Repository, error) {
	var repositories []*models.Repository

	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&repositories).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get repositories for project %s: %w",
			projectID,
			err,
		)
	}

	return repositories, nil
}

// List retrieves all repositories with pagination and optional filtering.
func (r *repositoryRepository) List(
	ctx context.Context,
	limit, offset int,
	filters RepositoryFilters,
) ([]*models.Repository, error) {
	var repositories []*models.Repository

	query := r.db.WithContext(ctx).
		Preload("Project")

	// Apply filters
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}

	if filters.URL != "" {
		query = query.Where("url ILIKE ?", "%"+filters.URL+"%")
	}

	if filters.Branch != "" {
		query = query.Where("branch = ?", filters.Branch)
	}

	if filters.SyncStatus != "" {
		query = query.Where("sync_status = ?", filters.SyncStatus)
	}

	if filters.Search != "" {
		query = query.Where(
			"url ILIKE ? OR branch ILIKE ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&repositories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repositories, nil
}

// ListByURL retrieves all repositories matching a specific URL.
func (r *repositoryRepository) ListByURL(
	ctx context.Context,
	url string,
) ([]*models.Repository, error) {
	var repositories []*models.Repository

	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("url = ?", url).
		Order("created_at DESC").
		Find(&repositories).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list repositories by URL %s: %w",
			url,
			err,
		)
	}

	return repositories, nil
}

// Update modifies an existing repository in the database.
func (r *repositoryRepository) Update(
	ctx context.Context,
	repository *models.Repository,
) error {
	result := r.db.WithContext(ctx).Save(repository)
	if result.Error != nil {
		return fmt.Errorf("failed to update repository: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"repository with id %s not found or no changes made",
			repository.ID,
		)
	}

	if r.logger != nil {
		r.logger.Info("Repository updated successfully",
			"repository_id", repository.ID,
			"url", repository.URL,
		)
	}

	return nil
}

// Delete performs a soft delete on a repository.
func (r *repositoryRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Repository{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete repository: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("repository with id %s not found", id)
	}

	if r.logger != nil {
		r.logger.Info("Repository deleted successfully", "repository_id", id)
	}

	return nil
}

// UpdateStatus updates the sync status of a repository.
func (r *repositoryRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status models.RepositoryStatus,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Repository{}).
		Where("id = ?", id).
		Update("sync_status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update repository status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("repository with id %s not found", id)
	}

	if r.logger != nil {
		r.logger.Info("Repository status updated successfully",
			"repository_id", id,
			"status", status,
		)
	}

	return nil
}

// UpdateLatestCommit updates the latest commit hash of a repository.
func (r *repositoryRepository) UpdateLatestCommit(
	ctx context.Context,
	id, commitHash string,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Repository{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"latest_commit": commitHash,
			"sync_status":   models.RepositoryStatusSuccess,
		})

	if result.Error != nil {
		return fmt.Errorf(
			"failed to update repository latest commit: %w",
			result.Error,
		)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("repository with id %s not found", id)
	}

	return nil
}

// GetPendingSync retrieves repositories that need synchronization.
func (r *repositoryRepository) GetPendingSync(
	ctx context.Context,
	limit int,
) ([]*models.Repository, error) {
	var repositories []*models.Repository

	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("sync_status = ? OR sync_status = ?",
			models.RepositoryStatusPending,
			models.RepositoryStatusError).
		Or("latest_commit IS NULL").
		Order("updated_at ASC").
		Limit(limit).
		Find(&repositories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sync repositories: %w", err)
	}

	return repositories, nil
}

// Search finds repositories by URL or branch with pagination.
func (r *repositoryRepository) Search(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]*models.Repository, error) {
	var repositories []*models.Repository

	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("url ILIKE ? OR branch ILIKE ?", "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&repositories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}

	return repositories, nil
}

// Count returns the total number of repositories matching the filters.
func (r *repositoryRepository) Count(
	ctx context.Context,
	filters RepositoryFilters,
) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.Repository{})

	// Apply filters
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}

	if filters.URL != "" {
		query = query.Where("url ILIKE ?", "%"+filters.URL+"%")
	}

	if filters.Branch != "" {
		query = query.Where("branch = ?", filters.Branch)
	}

	if filters.SyncStatus != "" {
		query = query.Where("sync_status = ?", filters.SyncStatus)
	}

	if filters.Search != "" {
		query = query.Where(
			"url ILIKE ? OR branch ILIKE ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count repositories: %w", err)
	}

	return count, nil
}

// Exists checks if a repository exists by ID.
func (r *repositoryRepository) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Repository{}).
		Where("id = ?", id).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to check repository existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByURL checks if a repository exists by URL within a project.
func (r *repositoryRepository) ExistsByURL(
	ctx context.Context,
	projectID, url string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&models.Repository{}).
		Where("project_id = ? AND url = ?", projectID, url).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf(
			"failed to check repository existence by URL: %w",
			err,
		)
	}

	return count > 0, nil
}
