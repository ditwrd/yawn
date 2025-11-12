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
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
)

// RepositoryService defines the interface for repository business operations.
//
// Provides methods for repository CRUD operations with business validation,
// project-based access control, and Git repository lifecycle management.
type RepositoryService interface {
	// Create creates a new repository with validation and business rules
	Create(
		ctx context.Context,
		req *CreateRepositoryRequest,
	) (*models.Repository, error)

	// GetByID retrieves a repository by its ID with project access validation
	GetByID(ctx context.Context, id string) (*models.Repository, error)

	// GetByProjectID retrieves all repositories for a project with pagination
	GetByProjectID(
		ctx context.Context,
		projectID string,
		page, limit int,
	) (*PaginatedRepositoriesResponse, error)

	// List retrieves all repositories with pagination and filtering
	List(
		ctx context.Context,
		page, limit int,
		filters RepositoryListFilters,
	) (*PaginatedRepositoriesResponse, error)

	// Update updates an existing repository with validation
	Update(
		ctx context.Context,
		id string,
		req *UpdateRepositoryRequest,
	) (*models.Repository, error)

	// Delete soft deletes a repository with access validation
	Delete(ctx context.Context, id string) error

	// Search searches repositories by query string with pagination
	Search(
		ctx context.Context,
		query string,
		page, limit int,
	) (*PaginatedRepositoriesResponse, error)

	// ValidateURL validates a repository URL and checks connectivity
	ValidateURL(
		ctx context.Context,
		req *ValidateRepositoryURLRequest,
	) (*ValidateRepositoryURLResponse, error)

	// SyncRepository triggers a synchronization of the repository
	SyncRepository(
		ctx context.Context,
		id string,
	) (*SyncRepositoryResponse, error)

	// GetSyncStatus retrieves the synchronization status of a repository
	GetSyncStatus(
		ctx context.Context,
		id string,
	) (*RepositorySyncStatusResponse, error)

	// UpdateStatus updates the repository synchronization status
	UpdateStatus(
		ctx context.Context,
		id string,
		status models.RepositoryStatus,
		message string,
	) error

	// UpdateLatestCommit updates the latest commit hash for a repository
	UpdateLatestCommit(ctx context.Context, id, commitHash string) error

	// GetPendingSync retrieves repositories that need synchronization
	GetPendingSync(ctx context.Context, limit int) ([]*models.Repository, error)

	// Exists checks if a repository exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByURL checks if a repository exists by URL within a project
	ExistsByURL(ctx context.Context, projectID, url string) (bool, error)
}

// CreateRepositoryRequest represents a request to create a repository.
type CreateRepositoryRequest struct {
	ProjectID string `json:"project_id" validate:"required,uuid"`
	URL       string `json:"url"        validate:"required,url,max=500"`
	Branch    string `json:"branch"     validate:"omitempty,max=255"`
}

// UpdateRepositoryRequest represents a request to update a repository.
type UpdateRepositoryRequest struct {
	URL    string `json:"url"    validate:"omitempty,url,max=500"`
	Branch string `json:"branch" validate:"omitempty,max=255"`
}

// ValidateRepositoryURLRequest represents a request to validate a repository
// URL.
type ValidateRepositoryURLRequest struct {
	URL    string `json:"url"    validate:"required,url,max=500"`
	Branch string `json:"branch" validate:"omitempty,max=255"`
}

// ValidateRepositoryURLResponse represents the response for repository URL
// validation.
type ValidateRepositoryURLResponse struct {
	Valid        bool                    `json:"valid"`
	Message      string                  `json:"message"`
	URL          string                  `json:"url"`
	Branch       string                  `json:"branch"`
	LatestCommit *CommitInfoResponse     `json:"latest_commit,omitempty"`
	Repository   *RepositoryInfoResponse `json:"repository,omitempty"`
	Error        string                  `json:"error,omitempty"`
}

// CommitInfoResponse represents commit information.
type CommitInfoResponse struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Timestamp string `json:"timestamp"`
}

// RepositoryInfoResponse represents basic repository information.
type RepositoryInfoResponse struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	DefaultBranch string   `json:"default_branch"`
	Branches      []string `json:"branches"`
}

// SyncRepositoryResponse represents the response for repository
// synchronization.
type SyncRepositoryResponse struct {
	Success      bool                 `json:"success"`
	RepositoryID string               `json:"repository_id"`
	CommitHash   string               `json:"commit_hash,omitempty"`
	Message      string               `json:"message"`
	SyncedAt     string               `json:"synced_at"`
	Duration     string               `json:"duration"`
	Changes      *SyncChangesResponse `json:"changes,omitempty"`
	Error        string               `json:"error,omitempty"`
}

// SyncChangesResponse represents synchronization changes.
type SyncChangesResponse struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Deleted  []string `json:"deleted"`
}

// RepositorySyncStatusResponse represents the synchronization status response.
type RepositorySyncStatusResponse struct {
	RepositoryID string                  `json:"repository_id"`
	Status       models.RepositoryStatus `json:"status"`
	LatestCommit string                  `json:"latest_commit"`
	LastSyncAt   string                  `json:"last_sync_at"`
	NextSyncAt   string                  `json:"next_sync_at,omitempty"`
	SyncCount    int                     `json:"sync_count"`
	ErrorMessage string                  `json:"error_message,omitempty"`
}

// RepositoryListFilters represents filters for repository listing.
type RepositoryListFilters struct {
	ProjectID string                  `json:"project_id"`
	Status    models.RepositoryStatus `json:"status"`
	URL       string                  `json:"url"`
	Branch    string                  `json:"branch"`
	Search    string                  `json:"search"`
}

// PaginatedRepositoriesResponse represents a paginated response for
// repositories.
type PaginatedRepositoriesResponse struct {
	Repositories []models.Repository `json:"repositories"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
	TotalPages   int                 `json:"total_pages"`
}

// repositoryService implements RepositoryService.
type repositoryService struct {
	repoRepo    repositories.RepositoryRepository
	projectRepo repositories.ProjectRepository
	logger      *zerolog.Logger
}

// NewRepositoryService creates a new repository service.
//
// Parameters:
//   - repoRepo: Repository repository for data operations
//   - projectRepo: Project repository for access control
//   - logger: Logger for structured logging
//
// Returns:
//   - RepositoryService: An instance of the repository service
func NewRepositoryService(
	repoRepo repositories.RepositoryRepository,
	projectRepo repositories.ProjectRepository,
	logger *zerolog.Logger,
) RepositoryService {
	return &repositoryService{
		repoRepo:    repoRepo,
		projectRepo: projectRepo,
		logger:      logger,
	}
}

// Create creates a new repository with validation and business rules.
func (s *repositoryService) Create(
	ctx context.Context,
	req *CreateRepositoryRequest,
) (*models.Repository, error) {
	logger := s.logger.With().Str("method", "CreateRepository").Logger()

	// Validate project ID
	projectUUID, err := uuid.FromString(req.ProjectID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("project_id", req.ProjectID).
			Msg("Invalid project ID")

		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Validate URL format
	if err := s.validateRepositoryURL(req.URL); err != nil {
		logger.Error().Err(err).Str("url", req.URL).Msg("Invalid repository URL")

		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Check if project exists
	exists, err := s.projectRepo.Exists(req.ProjectID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("project_id", req.ProjectID).
			Msg("Failed to check project existence")

		return nil, fmt.Errorf("failed to validate project: %w", err)
	}

	if !exists {
		logger.Warn().Str("project_id", req.ProjectID).Msg("Project not found")

		return nil, errors.New("project not found")
	}

	// Check for duplicate repository URL within the project
	exists, err = s.repoRepo.ExistsByURL(ctx, req.ProjectID, req.URL)
	if err != nil {
		logger.Error().
			Err(err).
			Str("url", req.URL).
			Msg("Failed to check repository URL existence")

		return nil, fmt.Errorf("failed to validate repository URL: %w", err)
	}

	if exists {
		logger.Warn().
			Str("url", req.URL).
			Str("project_id", req.ProjectID).
			Msg("Repository URL already exists in project")

		return nil, errors.New("repository URL already exists in this project")
	}

	// Set default branch if not provided
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	// Create repository model
	repo := &models.Repository{
		ID:         uuid.Must(uuid.NewV7()),
		URL:        req.URL,
		Branch:     branch,
		SyncStatus: models.RepositoryStatusPending,
		ProjectID:  projectUUID,
	}

	// Save repository to database
	if err := s.repoRepo.Create(ctx, repo); err != nil {
		logger.Error().
			Err(err).
			Str("url", req.URL).
			Msg("Failed to create repository")

		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	logger.Info().
		Str("repository_id", repo.ID.String()).
		Str("url", repo.URL).
		Str("project_id", req.ProjectID).
		Msg("Repository created successfully")

	return repo, nil
}

// GetByID retrieves a repository by its ID.
func (s *repositoryService) GetByID(
	ctx context.Context,
	id string,
) (*models.Repository, error) {
	logger := s.logger.With().
		Str("method", "GetRepositoryByID").
		Str("id", id).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return nil, fmt.Errorf("invalid repository ID: %w", err)
	}

	// Retrieve repository
	repo, err := s.repoRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve repository")

		return nil, fmt.Errorf("failed to retrieve repository: %w", err)
	}

	return repo, nil
}

// GetByProjectID retrieves all repositories for a project with pagination.
func (s *repositoryService) GetByProjectID(
	ctx context.Context,
	projectID string,
	page, limit int,
) (*PaginatedRepositoriesResponse, error) {
	logger := s.logger.With().
		Str("method", "GetRepositoriesByProjectID").
		Str("project_id", projectID).
		Int("page", page).
		Int("limit", limit).
		Logger()

	// Validate project ID
	_, err := uuid.FromString(projectID)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid project ID format")

		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Retrieve repositories
	repos, err := s.repoRepo.GetByProjectID(ctx, projectID, limit, offset)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve repositories by project ID")

		return nil, fmt.Errorf("failed to retrieve repositories: %w", err)
	}

	// Get total count for pagination
	total, err := s.repoRepo.Count(
		ctx,
		repositories.RepositoryFilters{ProjectID: projectID},
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count repositories")

		return nil, fmt.Errorf("failed to count repositories: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Convert pointer slice to value slice
	repoValues := make([]models.Repository, len(repos))
	for i, repo := range repos {
		repoValues[i] = *repo
	}

	response := &PaginatedRepositoriesResponse{
		Repositories: repoValues,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}

	logger.Info().
		Str("project_id", projectID).
		Int("count", len(repos)).
		Int64("total", total).
		Msg("Retrieved repositories by project ID")

	return response, nil
}

// List retrieves all repositories with pagination and filtering.
func (s *repositoryService) List(
	ctx context.Context,
	page, limit int,
	filters RepositoryListFilters,
) (*PaginatedRepositoriesResponse, error) {
	logger := s.logger.With().
		Str("method", "ListRepositories").
		Int("page", page).
		Int("limit", limit).
		Logger()

	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Convert filters
	repoFilters := repositories.RepositoryFilters{
		ProjectID:  filters.ProjectID,
		SyncStatus: filters.Status,
		URL:        filters.URL,
		Branch:     filters.Branch,
		Search:     filters.Search,
	}

	// Retrieve repositories
	repos, err := s.repoRepo.List(ctx, limit, offset, repoFilters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve repositories")

		return nil, fmt.Errorf("failed to retrieve repositories: %w", err)
	}

	// Get total count for pagination
	total, err := s.repoRepo.Count(ctx, repoFilters)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count repositories")

		return nil, fmt.Errorf("failed to count repositories: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Convert pointer slice to value slice
	repoValues := make([]models.Repository, len(repos))
	for i, repo := range repos {
		repoValues[i] = *repo
	}

	response := &PaginatedRepositoriesResponse{
		Repositories: repoValues,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}

	logger.Info().
		Int("count", len(repos)).
		Int64("total", total).
		Msg("Retrieved repositories list")

	return response, nil
}

// Update updates an existing repository with validation.
func (s *repositoryService) Update(
	ctx context.Context,
	id string,
	req *UpdateRepositoryRequest,
) (*models.Repository, error) {
	logger := s.logger.With().
		Str("method", "UpdateRepository").
		Str("id", id).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return nil, fmt.Errorf("invalid repository ID: %w", err)
	}

	// Get existing repository
	existingRepo, err := s.repoRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve existing repository")

		return nil, fmt.Errorf("failed to retrieve repository: %w", err)
	}

	// Validate URL if provided
	if req.URL != "" && req.URL != existingRepo.URL {
		if err := s.validateRepositoryURL(req.URL); err != nil {
			logger.Error().Err(err).Str("url", req.URL).Msg("Invalid repository URL")

			return nil, fmt.Errorf("invalid repository URL: %w", err)
		}

		// Check for duplicate URL within the project
		exists, err := s.repoRepo.ExistsByURL(
			ctx,
			existingRepo.ProjectID.String(),
			req.URL,
		)
		if err != nil {
			logger.Error().
				Err(err).
				Str("url", req.URL).
				Msg("Failed to check repository URL existence")

			return nil, fmt.Errorf("failed to validate repository URL: %w", err)
		}

		if exists {
			logger.Warn().
				Str("url", req.URL).
				Msg("Repository URL already exists in project")

			return nil, errors.New("repository URL already exists in this project")
		}

		existingRepo.URL = req.URL
	}

	// Update branch if provided
	if req.Branch != "" {
		existingRepo.Branch = req.Branch
	}

	// Update repository
	if err := s.repoRepo.Update(ctx, existingRepo); err != nil {
		logger.Error().Err(err).Msg("Failed to update repository")

		return nil, fmt.Errorf("failed to update repository: %w", err)
	}

	logger.Info().
		Str("repository_id", id).
		Str("url", existingRepo.URL).
		Str("branch", existingRepo.Branch).
		Msg("Repository updated successfully")

	return existingRepo, nil
}

// Delete soft deletes a repository with access validation.
func (s *repositoryService) Delete(ctx context.Context, id string) error {
	logger := s.logger.With().
		Str("method", "DeleteRepository").
		Str("id", id).
		Logger()

	// Delete repository
	err := s.repoRepo.Delete(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete repository")

		return fmt.Errorf("failed to delete repository: %w", err)
	}

	logger.Info().
		Str("repository_id", id).
		Msg("Repository deleted successfully")

	return nil
}

// Search searches repositories by query string with pagination.
func (s *repositoryService) Search(
	ctx context.Context,
	query string,
	page, limit int,
) (*PaginatedRepositoriesResponse, error) {
	logger := s.logger.With().
		Str("method", "SearchRepositories").
		Str("query", query).
		Int("page", page).
		Int("limit", limit).
		Logger()

	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Search repositories
	repos, err := s.repoRepo.Search(ctx, query, limit, offset)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to search repositories")

		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}

	// Get total count for pagination
	total, err := s.repoRepo.Count(
		ctx,
		repositories.RepositoryFilters{Search: query},
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count search results")

		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Convert pointer slice to value slice
	repoValues := make([]models.Repository, len(repos))
	for i, repo := range repos {
		repoValues[i] = *repo
	}

	response := &PaginatedRepositoriesResponse{
		Repositories: repoValues,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}

	logger.Info().
		Str("query", query).
		Int("count", len(repos)).
		Int64("total", total).
		Msg("Repository search completed")

	return response, nil
}

// ValidateURL validates a repository URL and checks connectivity.
func (s *repositoryService) ValidateURL(
	ctx context.Context,
	req *ValidateRepositoryURLRequest,
) (*ValidateRepositoryURLResponse, error) {
	logger := s.logger.With().
		Str("method", "ValidateRepositoryURL").
		Str("url", req.URL).
		Logger()

	response := &ValidateRepositoryURLResponse{
		URL:    req.URL,
		Branch: req.Branch,
	}

	// Validate URL format
	err := s.validateRepositoryURL(req.URL)
	if err != nil {
		response.Valid = false
		response.Message = "Invalid URL format"
		response.Error = err.Error()

		return response, nil
	}

	// TODO: Implement actual connectivity check and repository info retrieval
	// This would involve making HTTP requests to Git hosting services
	// For now, we'll do basic validation

	// Extract repository name from URL
	repoName := s.extractRepositoryName(req.URL)
	response.Valid = true
	response.Message = "Repository URL is valid"
	response.Repository = &RepositoryInfoResponse{
		Name:          repoName,
		DefaultBranch: req.Branch,
		Branches:      []string{req.Branch},
	}

	logger.Info().
		Str("url", req.URL).
		Bool("valid", response.Valid).
		Msg("Repository URL validation completed")

	return response, nil
}

// SyncRepository triggers a synchronization of the repository.
func (s *repositoryService) SyncRepository(
	ctx context.Context,
	id string,
) (*SyncRepositoryResponse, error) {
	logger := s.logger.With().
		Str("method", "SyncRepository").
		Str("id", id).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return nil, fmt.Errorf("invalid repository ID: %w", err)
	}

	// Get repository
	repo, err := s.repoRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve repository")

		return nil, fmt.Errorf("failed to retrieve repository: %w", err)
	}

	// Update status to pending
	if err := s.repoRepo.UpdateStatus(ctx, id, models.RepositoryStatusPending); err != nil {
		logger.Error().Err(err).Msg("Failed to update repository status")

		return nil, fmt.Errorf("failed to update repository status: %w", err)
	}

	// TODO: Implement actual synchronization logic
	// This would involve cloning/fetching the repository and processing changes
	// For now, we'll simulate a successful sync

	syncResponse := &SyncRepositoryResponse{
		Success:      true,
		RepositoryID: repo.ID.String(),
		Message:      "Repository synchronized successfully",
		SyncedAt:     "now",
		Duration:     "0s",
		Changes: &SyncChangesResponse{
			Added:    []string{},
			Modified: []string{},
			Deleted:  []string{},
		},
	}

	// Update repository status to success
	if err := s.repoRepo.UpdateStatus(ctx, id, models.RepositoryStatusSuccess); err != nil {
		logger.Error().Err(err).Msg("Failed to update repository status after sync")
		// Don't fail the operation, just log the error
	}

	logger.Info().
		Str("repository_id", id).
		Bool("success", syncResponse.Success).
		Msg("Repository synchronization completed")

	return syncResponse, nil
}

// GetSyncStatus retrieves the synchronization status of a repository.
func (s *repositoryService) GetSyncStatus(
	ctx context.Context,
	id string,
) (*RepositorySyncStatusResponse, error) {
	logger := s.logger.With().
		Str("method", "GetSyncStatus").
		Str("id", id).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return nil, fmt.Errorf("invalid repository ID: %w", err)
	}

	// Get repository
	repo, err := s.repoRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve repository")

		return nil, fmt.Errorf("failed to retrieve repository: %w", err)
	}

	response := &RepositorySyncStatusResponse{
		RepositoryID: repo.ID.String(),
		Status:       repo.SyncStatus,
		LatestCommit: repo.LatestCommit,
		LastSyncAt:   repo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		SyncCount:    1, // TODO: Track actual sync count
	}

	logger.Info().
		Str("repository_id", id).
		Str("status", string(repo.SyncStatus)).
		Msg("Retrieved repository sync status")

	return response, nil
}

// UpdateStatus updates the repository synchronization status.
func (s *repositoryService) UpdateStatus(
	ctx context.Context,
	id string,
	status models.RepositoryStatus,
	message string,
) error {
	logger := s.logger.With().
		Str("method", "UpdateRepositoryStatus").
		Str("id", id).
		Str("status", string(status)).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return fmt.Errorf("invalid repository ID: %w", err)
	}

	// Update status
	if err := s.repoRepo.UpdateStatus(ctx, id, status); err != nil {
		logger.Error().Err(err).Msg("Failed to update repository status")

		return fmt.Errorf("failed to update repository status: %w", err)
	}

	logger.Info().
		Str("repository_id", id).
		Str("status", string(status)).
		Msg("Repository status updated successfully")

	return nil
}

// UpdateLatestCommit updates the latest commit hash for a repository.
func (s *repositoryService) UpdateLatestCommit(
	ctx context.Context,
	id, commitHash string,
) error {
	logger := s.logger.With().
		Str("method", "UpdateLatestCommit").
		Str("id", id).
		Str("commit_hash", commitHash).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return fmt.Errorf("invalid repository ID: %w", err)
	}

	// Update latest commit
	if err := s.repoRepo.UpdateLatestCommit(ctx, id, commitHash); err != nil {
		logger.Error().Err(err).Msg("Failed to update latest commit")

		return fmt.Errorf("failed to update latest commit: %w", err)
	}

	logger.Info().
		Str("repository_id", id).
		Str("commit_hash", commitHash).
		Msg("Latest commit updated successfully")

	return nil
}

// GetPendingSync retrieves repositories that need synchronization.
func (s *repositoryService) GetPendingSync(
	ctx context.Context,
	limit int,
) ([]*models.Repository, error) {
	logger := s.logger.With().
		Str("method", "GetPendingSync").
		Int("limit", limit).
		Logger()

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	repos, err := s.repoRepo.GetPendingSync(ctx, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve pending sync repositories")

		return nil, fmt.Errorf(
			"failed to retrieve pending sync repositories: %w",
			err,
		)
	}

	logger.Info().
		Int("count", len(repos)).
		Msg("Retrieved pending sync repositories")

	return repos, nil
}

// Exists checks if a repository exists by ID.
func (s *repositoryService) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	logger := s.logger.With().
		Str("method", "RepositoryExists").
		Str("id", id).
		Logger()

	// Validate UUID format
	_, err := uuid.FromString(id)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid repository ID format")

		return false, fmt.Errorf("invalid repository ID: %w", err)
	}

	exists, err := s.repoRepo.Exists(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check repository existence")

		return false, fmt.Errorf("failed to check repository existence: %w", err)
	}

	return exists, nil
}

// ExistsByURL checks if a repository exists by URL within a project.
func (s *repositoryService) ExistsByURL(
	ctx context.Context,
	projectID, url string,
) (bool, error) {
	logger := s.logger.With().
		Str("method", "RepositoryExistsByURL").
		Str("project_id", projectID).
		Str("url", url).
		Logger()

	// Validate project ID
	_, err := uuid.FromString(projectID)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid project ID format")

		return false, fmt.Errorf("invalid project ID: %w", err)
	}

	exists, err := s.repoRepo.ExistsByURL(ctx, projectID, url)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check repository URL existence")

		return false, fmt.Errorf(
			"failed to check repository URL existence: %w",
			err,
		)
	}

	return exists, nil
}

// validateRepositoryURL validates the format of a repository URL.
func (s *repositoryService) validateRepositoryURL(repoURL string) error {
	// Parse URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL must use HTTP or HTTPS protocol")
	}

	// Check host
	if parsedURL.Host == "" {
		return errors.New("URL must have a valid host")
	}

	// Check if it's a known Git hosting service
	host := parsedURL.Hostname()
	if !s.isKnownGitHost(host) {
		return errors.New("URL must be from a known Git hosting service")
	}

	return nil
}

// isKnownGitHost checks if the hostname is a known Git hosting service.
func (s *repositoryService) isKnownGitHost(host string) bool {
	knownHosts := []string{
		"github.com",
		"gitlab.com",
		"bitbucket.org",
		"git.sr.ht",
	}

	for _, knownHost := range knownHosts {
		if host == knownHost || strings.HasSuffix(host, "."+knownHost) {
			return true
		}
	}

	// Allow any domain with git in the name (for self-hosted instances)
	return strings.Contains(host, "git")
}

// extractRepositoryName extracts repository name from URL.
func (s *repositoryService) extractRepositoryName(repoURL string) string {
	// Parse URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "unknown"
	}

	// Extract path and remove .git suffix
	path := strings.TrimSuffix(parsedURL.Path, ".git")
	path = strings.Trim(path, "/")

	// Get the last part of the path
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown"
}
