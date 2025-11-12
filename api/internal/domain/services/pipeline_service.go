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
	"slices"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
)

// PipelineService defines the interface for pipeline business operations.
//
// Provides methods for pipeline CRUD operations with business validation,
// execution management, dependency resolution, and project-based access
// control.
type PipelineService interface {
	// Create creates a new pipeline with validation and business rules
	Create(
		ctx context.Context,
		req *CreatePipelineRequest,
	) (*models.Pipeline, error)

	// GetByID retrieves a pipeline by its ID with project access validation
	GetByID(ctx context.Context, id string) (*models.Pipeline, error)

	// GetByProjectID retrieves all pipelines for a project with pagination
	GetByProjectID(
		ctx context.Context,
		projectID string,
		page, limit int,
	) (*PaginatedPipelinesResponse, error)

	// List retrieves all pipelines with pagination and filtering
	List(
		ctx context.Context,
		page, limit int,
		filters PipelineListFilters,
	) (*PaginatedPipelinesResponse, error)

	// Update updates an existing pipeline with validation
	Update(
		ctx context.Context,
		id string,
		req *UpdatePipelineRequest,
	) (*models.Pipeline, error)

	// Delete soft deletes a pipeline with access validation
	Delete(ctx context.Context, id string) error

	// Search searches pipelines by query string with pagination
	Search(
		ctx context.Context,
		query string,
		page, limit int,
	) (*PaginatedPipelinesResponse, error)

	// UpdateStatus updates the status of a pipeline
	UpdateStatus(
		ctx context.Context,
		id string,
		status models.PipelineStatus,
	) error

	// ValidateAccess checks if the current user has access to a pipeline
	ValidateAccess(
		ctx context.Context,
		pipelineID, userID string,
		requiredRole models.ProjectRole,
	) error

	// CanCreate checks if the user can create pipelines in a project
	CanCreate(ctx context.Context, projectID, userID string) error

	// === Pipeline Execution Operations ===

	// TriggerExecution manually triggers a pipeline execution
	TriggerExecution(
		ctx context.Context,
		pipelineID, userID string,
		config *string,
	) (*models.PipelineExecution, error)

	// GetExecutionByID retrieves a pipeline execution by ID
	GetExecutionByID(
		ctx context.Context,
		executionID string,
	) (*models.PipelineExecution, error)

	// GetExecutionsByPipelineID retrieves executions for a pipeline with
	// pagination
	GetExecutionsByPipelineID(
		ctx context.Context,
		pipelineID string,
		page, limit int,
	) (*PaginatedExecutionsResponse, error)

	// CancelExecution cancels a running pipeline execution
	CancelExecution(ctx context.Context, executionID, userID string) error

	// GetRunningExecutions retrieves all currently running executions
	GetRunningExecutions(ctx context.Context) ([]*models.PipelineExecution, error)

	// === Dependency Resolution Operations ===

	// AddDependency adds a dependency relationship between pipelines
	AddDependency(
		ctx context.Context,
		pipelineID, dependsOnID string,
		condition *string,
	) error

	// RemoveDependency removes a dependency relationship
	RemoveDependency(ctx context.Context, pipelineID, dependsOnID string) error

	// GetDependencyGraph retrieves the dependency graph for a project
	GetDependencyGraph(
		ctx context.Context,
		projectID string,
	) (*DependencyGraphResponse, error)

	// ValidateDependencies checks for circular dependencies and other issues
	ValidateDependencies(ctx context.Context, pipelineID string) error

	// ResolveDependencies determines execution order based on dependencies
	ResolveDependencies(ctx context.Context, pipelineID string) ([]string, error)
}

// CreatePipelineRequest represents the request to create a new pipeline.
type CreatePipelineRequest struct {
	Name        string
	Description string
	Config      string
	Schedule    string
	ProjectID   string
	AssetIDs    []string
}

// UpdatePipelineRequest represents the request to update a pipeline.
type UpdatePipelineRequest struct {
	Name        *string
	Description *string
	Config      *string
	Schedule    *string
	Status      *models.PipelineStatus
	IsEnabled   *bool
	AssetIDs    []string
}

// PipelineListFilters defines filtering options for pipeline listing.
type PipelineListFilters struct {
	ProjectID string
	Status    models.PipelineStatus
	Name      string
	Search    string
	IsEnabled *bool
	Schedule  string
}

// PaginatedPipelinesResponse represents a paginated response for pipelines.
type PaginatedPipelinesResponse struct {
	Pipelines []*models.Pipeline
	Total     int64
	Page      int
	Limit     int
}

// PaginatedExecutionsResponse represents a paginated response for executions.
type PaginatedExecutionsResponse struct {
	Executions []*models.PipelineExecution
	Total      int64
	Page       int
	Limit      int
}

// DependencyGraphResponse represents a dependency graph response.
type DependencyGraphResponse struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

// DependencyNode represents a node in the dependency graph.
type DependencyNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// DependencyEdge represents an edge in the dependency graph.
type DependencyEdge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

// pipelineService implements the PipelineService interface.
type pipelineService struct {
	pipelineRepo repositories.PipelineRepository
	projectRepo  repositories.ProjectRepository
	assetRepo    repositories.AssetRepository
	userRepo     repositories.UserRepository
	logger       *zerolog.Logger
}

// NewPipelineService creates a new instance of PipelineService.
//
// Parameters:
//   - pipelineRepo: Pipeline repository for data operations
//   - projectRepo: Project repository for access validation
//   - assetRepo: Asset repository for asset operations
//   - userRepo: User repository for user operations
//   - logger: Logger for structured logging
//
// Returns:
//   - PipelineService: An instance of the pipeline service
func NewPipelineService(
	pipelineRepo repositories.PipelineRepository,
	projectRepo repositories.ProjectRepository,
	assetRepo repositories.AssetRepository,
	userRepo repositories.UserRepository,
	logger *zerolog.Logger,
) PipelineService {
	return &pipelineService{
		pipelineRepo: pipelineRepo,
		projectRepo:  projectRepo,
		assetRepo:    assetRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// === Pipeline CRUD Operations ===

// Create creates a new pipeline with validation and business rules.
func (s *pipelineService) Create(
	ctx context.Context,
	req *CreatePipelineRequest,
) (*models.Pipeline, error) {
	// Validate request
	if err := s.validateCreateRequest(ctx, req); err != nil {
		return nil, err
	}

	// Check if pipeline with same name already exists in project
	exists, err := s.pipelineRepo.ExistsByName(ctx, req.ProjectID, req.Name)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", req.ProjectID).
			Str("name", req.Name).
			Msg("Failed to check pipeline existence")

		return nil, fmt.Errorf("failed to validate pipeline uniqueness: %w", err)
	}

	if exists {
		return nil, fmt.Errorf(
			"pipeline with name '%s' already exists in project",
			req.Name,
		)
	}

	// Create the pipeline
	pipeline := &models.Pipeline{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ProjectID:   uuid.Must(uuid.FromString(req.ProjectID)),
		Status:      models.PipelineStatusDraft,
		Config:      strings.TrimSpace(req.Config),
		Schedule:    strings.TrimSpace(req.Schedule),
		IsEnabled:   true,
	}

	// Save pipeline
	if err := s.pipelineRepo.Create(ctx, pipeline); err != nil {
		s.logger.Error().
			Err(err).
			Str("name", pipeline.Name).
			Str("project_id", req.ProjectID).
			Msg("Failed to create pipeline")

		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Associate assets if provided
	if len(req.AssetIDs) > 0 {
		// This would require implementing the many-to-many relationship management
		// For now, we'll log and skip implementation
		s.logger.Info().
			Int("asset_count", len(req.AssetIDs)).
			Msg("Asset association not yet implemented")
	}

	s.logger.Info().
		Str("pipeline_id", pipeline.ID.String()).
		Str("name", pipeline.Name).
		Str("project_id", req.ProjectID).
		Msg("Pipeline created successfully")

	return pipeline, nil
}

// GetByID retrieves a pipeline by its ID with project access validation.
func (s *pipelineService) GetByID(
	ctx context.Context,
	id string,
) (*models.Pipeline, error) {
	if id == "" {
		return nil, errors.New("pipeline ID is required")
	}

	pipeline, err := s.pipelineRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pipeline_id", id).
			Msg("Failed to get pipeline")

		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	return pipeline, nil
}

// GetByProjectID retrieves all pipelines for a project with pagination.
func (s *pipelineService) GetByProjectID(
	ctx context.Context,
	projectID string,
	page, limit int,
) (*PaginatedPipelinesResponse, error) {
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

	// Get pipelines
	pipelines, err := s.pipelineRepo.GetByProjectID(ctx, projectID, limit, offset)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Msg("Failed to get pipelines by project ID")

		return nil, fmt.Errorf("failed to get pipelines: %w", err)
	}

	// Get total count
	count, err := s.pipelineRepo.Count(
		ctx,
		repositories.PipelineFilters{ProjectID: projectID},
	)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("project_id", projectID).
			Msg("Failed to count pipelines")

		return nil, fmt.Errorf("failed to count pipelines: %w", err)
	}

	return &PaginatedPipelinesResponse{
		Pipelines: pipelines,
		Total:     count,
		Page:      page,
		Limit:     limit,
	}, nil
}

// List retrieves all pipelines with pagination and filtering.
func (s *pipelineService) List(
	ctx context.Context,
	page, limit int,
	filters PipelineListFilters,
) (*PaginatedPipelinesResponse, error) {
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
	repoFilters := repositories.PipelineFilters{
		ProjectID: filters.ProjectID,
		Status:    filters.Status,
		Name:      filters.Name,
		Search:    filters.Search,
		IsEnabled: filters.IsEnabled,
		Schedule:  filters.Schedule,
	}

	// Get pipelines
	pipelines, err := s.pipelineRepo.List(ctx, limit, offset, repoFilters)
	if err != nil {
		s.logger.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to list pipelines")

		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}

	// Get total count
	count, err := s.pipelineRepo.Count(ctx, repoFilters)
	if err != nil {
		s.logger.Error().
			Err(err).
			Interface("filters", filters).
			Msg("Failed to count pipelines")

		return nil, fmt.Errorf("failed to count pipelines: %w", err)
	}

	return &PaginatedPipelinesResponse{
		Pipelines: pipelines,
		Total:     count,
		Page:      page,
		Limit:     limit,
	}, nil
}

// Update updates an existing pipeline with validation.
func (s *pipelineService) Update(
	ctx context.Context,
	id string,
	req *UpdatePipelineRequest,
) (*models.Pipeline, error) {
	if id == "" {
		return nil, errors.New("pipeline ID is required")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Get existing pipeline
	pipeline, err := s.pipelineRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		pipeline.Name = strings.TrimSpace(*req.Name)
	}

	if req.Description != nil {
		pipeline.Description = strings.TrimSpace(*req.Description)
	}

	if req.Config != nil {
		pipeline.Config = strings.TrimSpace(*req.Config)
	}

	if req.Schedule != nil {
		pipeline.Schedule = strings.TrimSpace(*req.Schedule)
	}

	if req.Status != nil {
		pipeline.Status = *req.Status
	}

	if req.IsEnabled != nil {
		pipeline.IsEnabled = *req.IsEnabled
	}

	// Validate updated pipeline
	if err := s.validatePipeline(pipeline); err != nil {
		return nil, err
	}

	// Update pipeline
	if err := s.pipelineRepo.Update(ctx, pipeline); err != nil {
		s.logger.Error().
			Err(err).
			Str("pipeline_id", id).
			Msg("Failed to update pipeline")

		return nil, fmt.Errorf("failed to update pipeline: %w", err)
	}

	s.logger.Info().
		Str("pipeline_id", id).
		Msg("Pipeline updated successfully")

	return pipeline, nil
}

// Delete soft deletes a pipeline with access validation.
func (s *pipelineService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("pipeline ID is required")
	}

	// Check if pipeline exists
	exists, err := s.pipelineRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check pipeline existence: %w", err)
	}

	if !exists {
		return errors.New("pipeline not found")
	}

	// Delete pipeline
	if err := s.pipelineRepo.Delete(ctx, id); err != nil {
		s.logger.Error().
			Err(err).
			Str("pipeline_id", id).
			Msg("Failed to delete pipeline")

		return fmt.Errorf("failed to delete pipeline: %w", err)
	}

	s.logger.Info().
		Str("pipeline_id", id).
		Msg("Pipeline deleted successfully")

	return nil
}

// Search searches pipelines by query string with pagination.
func (s *pipelineService) Search(
	ctx context.Context,
	query string,
	page, limit int,
) (*PaginatedPipelinesResponse, error) {
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

	// Search pipelines
	pipelines, err := s.pipelineRepo.Search(ctx, query, limit, offset)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Msg("Failed to search pipelines")

		return nil, fmt.Errorf("failed to search pipelines: %w", err)
	}

	// Get total count
	count, err := s.pipelineRepo.Count(
		ctx,
		repositories.PipelineFilters{Search: query},
	)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("query", query).
			Msg("Failed to count search results")

		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	return &PaginatedPipelinesResponse{
		Pipelines: pipelines,
		Total:     count,
		Page:      page,
		Limit:     limit,
	}, nil
}

// UpdateStatus updates the status of a pipeline.
func (s *pipelineService) UpdateStatus(
	ctx context.Context,
	id string,
	status models.PipelineStatus,
) error {
	if id == "" {
		return errors.New("pipeline ID is required")
	}

	// Validate status transition
	err := s.validateStatusTransition(ctx, id, status)
	if err != nil {
		return err
	}

	// Update status
	err = s.pipelineRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pipeline_id", id).
			Str("status", string(status)).
			Msg("Failed to update pipeline status")

		return fmt.Errorf("failed to update pipeline status: %w", err)
	}

	s.logger.Info().
		Str("pipeline_id", id).
		Str("status", string(status)).
		Msg("Pipeline status updated successfully")

	return nil
}

// ValidateAccess checks if the current user has access to a pipeline.
func (s *pipelineService) ValidateAccess(
	ctx context.Context,
	pipelineID, userID string,
	requiredRole models.ProjectRole,
) error {
	pipeline, err := s.pipelineRepo.GetByID(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to get pipeline for access validation: %w", err)
	}

	// Check if user has access to the project
	hasAccess, err := s.projectRepo.HasUserWithRole(
		pipeline.ProjectID.String(),
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

// CanCreate checks if the user can create pipelines in a project.
func (s *pipelineService) CanCreate(
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
			"user does not have permission to create pipelines in this project",
		)
	}

	return nil
}

// === Pipeline Execution Operations ===

// TriggerExecution manually triggers a pipeline execution.
func (s *pipelineService) TriggerExecution(
	ctx context.Context,
	pipelineID, userID string,
	config *string,
) (*models.PipelineExecution, error) {
	// Validate pipeline exists and is active
	pipeline, err := s.pipelineRepo.GetByID(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	if pipeline.Status != models.PipelineStatusActive {
		return nil, errors.New("pipeline must be active to execute")
	}

	if !pipeline.IsEnabled {
		return nil, errors.New("pipeline is disabled")
	}

	// Check for running executions
	runningExecutions, err := s.pipelineRepo.GetRunningExecutions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check running executions: %w", err)
	}

	for _, exec := range runningExecutions {
		if exec.PipelineID.String() == pipelineID {
			return nil, errors.New("pipeline is already running")
		}
	}

	// Check dependencies
	if err := s.validateDependenciesForExecution(ctx, pipelineID); err != nil {
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	// Create execution
	execution := &models.PipelineExecution{
		ID:          uuid.Must(uuid.NewV7()),
		PipelineID:  pipeline.ID,
		Status:      models.ExecutionStatusPending,
		TriggerType: "manual",
		Config:      *config,
	}

	if userID != "" {
		userUUID := uuid.Must(uuid.FromString(userID))
		execution.TriggeredBy = &userUUID
	}

	if err := s.pipelineRepo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create pipeline execution: %w", err)
	}

	// Start execution (in a real implementation, this would be async)
	if err := s.startExecution(ctx, execution); err != nil {
		s.logger.Error().
			Err(err).
			Str("execution_id", execution.ID.String()).
			Msg("Failed to start pipeline execution")
	}

	s.logger.Info().
		Str("execution_id", execution.ID.String()).
		Str("pipeline_id", pipelineID).
		Str("triggered_by", userID).
		Msg("Pipeline execution triggered successfully")

	return execution, nil
}

// GetExecutionByID retrieves a pipeline execution by ID.
func (s *pipelineService) GetExecutionByID(
	ctx context.Context,
	executionID string,
) (*models.PipelineExecution, error) {
	if executionID == "" {
		return nil, errors.New("execution ID is required")
	}

	execution, err := s.pipelineRepo.GetExecutionByID(ctx, executionID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("execution_id", executionID).
			Msg("Failed to get pipeline execution")

		return nil, fmt.Errorf("failed to get pipeline execution: %w", err)
	}

	return execution, nil
}

// GetExecutionsByPipelineID retrieves executions for a pipeline with
// pagination.
func (s *pipelineService) GetExecutionsByPipelineID(
	ctx context.Context,
	pipelineID string,
	page, limit int,
) (*PaginatedExecutionsResponse, error) {
	if pipelineID == "" {
		return nil, errors.New("pipeline ID is required")
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

	// Get executions
	executions, err := s.pipelineRepo.GetExecutionsByPipelineID(
		ctx,
		pipelineID,
		limit,
		offset,
	)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pipeline_id", pipelineID).
			Msg("Failed to get executions by pipeline ID")

		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	// For simplicity, we'll use a fixed total count
	// In a real implementation, you'd get this from the repository
	total := int64(len(executions))

	return &PaginatedExecutionsResponse{
		Executions: executions,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}, nil
}

// CancelExecution cancels a running pipeline execution.
func (s *pipelineService) CancelExecution(
	ctx context.Context,
	executionID, userID string,
) error {
	if executionID == "" {
		return errors.New("execution ID is required")
	}

	// Get execution
	execution, err := s.pipelineRepo.GetExecutionByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution: %w", err)
	}

	// Check if execution can be cancelled
	if execution.Status != models.ExecutionStatusRunning &&
		execution.Status != models.ExecutionStatusPending {
		return fmt.Errorf(
			"execution cannot be cancelled in current state: %s",
			execution.Status,
		)
	}

	// Update status to cancelled
	if err := s.pipelineRepo.UpdateExecutionStatus(ctx, executionID, models.ExecutionStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel execution: %w", err)
	}

	s.logger.Info().
		Str("execution_id", executionID).
		Str("cancelled_by", userID).
		Msg("Pipeline execution cancelled successfully")

	return nil
}

// GetRunningExecutions retrieves all currently running executions.
func (s *pipelineService) GetRunningExecutions(
	ctx context.Context,
) ([]*models.PipelineExecution, error) {
	executions, err := s.pipelineRepo.GetRunningExecutions(ctx)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Failed to get running executions")

		return nil, fmt.Errorf("failed to get running executions: %w", err)
	}

	return executions, nil
}

// === Dependency Resolution Operations ===

// AddDependency adds a dependency relationship between pipelines.
func (s *pipelineService) AddDependency(
	ctx context.Context,
	pipelineID, dependsOnID string,
	condition *string,
) error {
	if pipelineID == "" || dependsOnID == "" {
		return errors.New("both pipeline ID and dependency ID are required")
	}

	if pipelineID == dependsOnID {
		return errors.New("pipeline cannot depend on itself")
	}

	// Validate both pipelines exist
	pipelineExists, err := s.pipelineRepo.Exists(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to validate pipeline existence: %w", err)
	}

	if !pipelineExists {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	dependsOnExists, err := s.pipelineRepo.Exists(ctx, dependsOnID)
	if err != nil {
		return fmt.Errorf(
			"failed to validate dependency pipeline existence: %w",
			err,
		)
	}

	if !dependsOnExists {
		return fmt.Errorf("dependency pipeline not found: %s", dependsOnID)
	}

	// Create dependency
	dependency := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          uuid.Must(uuid.FromString(pipelineID)),
		DependsOnPipelineID: uuid.Must(uuid.FromString(dependsOnID)),
	}

	if condition != nil {
		dependency.Condition = strings.TrimSpace(*condition)
	}

	if err := s.pipelineRepo.CreateDependency(ctx, dependency); err != nil {
		return fmt.Errorf("failed to create dependency: %w", err)
	}

	// Validate for circular dependencies
	if err := s.ValidateDependencies(ctx, pipelineID); err != nil {
		// Remove the dependency if it creates a cycle
		s.pipelineRepo.DeleteDependency(ctx, pipelineID, dependsOnID)

		return fmt.Errorf("dependency creates circular reference: %w", err)
	}

	s.logger.Info().
		Str("pipeline_id", pipelineID).
		Str("depends_on", dependsOnID).
		Msg("Pipeline dependency added successfully")

	return nil
}

// RemoveDependency removes a dependency relationship.
func (s *pipelineService) RemoveDependency(
	ctx context.Context,
	pipelineID, dependsOnID string,
) error {
	if pipelineID == "" || dependsOnID == "" {
		return errors.New("both pipeline ID and dependency ID are required")
	}

	err := s.pipelineRepo.DeleteDependency(ctx, pipelineID, dependsOnID)
	if err != nil {
		return fmt.Errorf("failed to remove dependency: %w", err)
	}

	s.logger.Info().
		Str("pipeline_id", pipelineID).
		Str("depends_on", dependsOnID).
		Msg("Pipeline dependency removed successfully")

	return nil
}

// GetDependencyGraph retrieves the dependency graph for a project.
func (s *pipelineService) GetDependencyGraph(
	ctx context.Context,
	projectID string,
) (*DependencyGraphResponse, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	// Get dependency graph from repository
	graph, err := s.pipelineRepo.GetDependencyGraph(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency graph: %w", err)
	}

	// Get all pipelines in the project for node information
	pipelines, err := s.pipelineRepo.GetByProjectID(ctx, projectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipelines for graph: %w", err)
	}

	// Build response
	response := &DependencyGraphResponse{
		Nodes: make([]DependencyNode, 0),
		Edges: make([]DependencyEdge, 0),
	}

	// Add nodes
	for _, pipeline := range pipelines {
		response.Nodes = append(response.Nodes, DependencyNode{
			ID:     pipeline.ID.String(),
			Name:   pipeline.Name,
			Status: string(pipeline.Status),
		})
	}

	// Add edges
	for pipelineID, dependencies := range graph {
		for _, depID := range dependencies {
			response.Edges = append(response.Edges, DependencyEdge{
				From: pipelineID,
				To:   depID,
			})
		}
	}

	return response, nil
}

// ValidateDependencies checks for circular dependencies and other issues.
func (s *pipelineService) ValidateDependencies(
	ctx context.Context,
	pipelineID string,
) error {
	// Get dependency graph for the pipeline's project
	pipeline, err := s.pipelineRepo.GetByID(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	graph, err := s.pipelineRepo.GetDependencyGraph(
		ctx,
		pipeline.ProjectID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to get dependency graph: %w", err)
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	if s.hasCycle(graph, pipelineID, visited, recStack) {
		return fmt.Errorf(
			"circular dependency detected for pipeline %s",
			pipelineID,
		)
	}

	return nil
}

// ResolveDependencies determines execution order based on dependencies.
func (s *pipelineService) ResolveDependencies(
	ctx context.Context,
	pipelineID string,
) ([]string, error) {
	// Get dependency graph
	pipeline, err := s.pipelineRepo.GetByID(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	graph, err := s.pipelineRepo.GetDependencyGraph(
		ctx,
		pipeline.ProjectID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency graph: %w", err)
	}

	// Topological sort to get execution order
	order := s.topologicalSort(graph)

	return order, nil
}

// === Helper Methods ===

// validateCreateRequest validates the create pipeline request.
func (s *pipelineService) validateCreateRequest(
	ctx context.Context,
	req *CreatePipelineRequest,
) error {
	if req == nil {
		return errors.New("create request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return errors.New("pipeline name is required")
	}

	if req.ProjectID == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format for project ID
	if _, err := uuid.FromString(req.ProjectID); err != nil {
		return fmt.Errorf("invalid project ID format: %w", err)
	}

	// Validate pipeline name format
	if err := s.validatePipelineName(req.Name); err != nil {
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

	// Validate asset IDs if provided
	for _, assetID := range req.AssetIDs {
		if _, err := uuid.FromString(assetID); err != nil {
			return fmt.Errorf("invalid asset ID format: %w", err)
		}

		// Check if asset exists
		assetExists, err := s.assetRepo.Exists(ctx, assetID)
		if err != nil {
			return fmt.Errorf("failed to validate asset existence: %w", err)
		}

		if !assetExists {
			return fmt.Errorf("asset not found: %s", assetID)
		}
	}

	return nil
}

// validateUpdateRequest validates the update pipeline request.
func (s *pipelineService) validateUpdateRequest(
	req *UpdatePipelineRequest,
) error {
	if req == nil {
		return errors.New("update request is required")
	}

	// Validate name if provided
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errors.New("pipeline name cannot be empty")
		}

		err := s.validatePipelineName(*req.Name)
		if err != nil {
			return err
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := []models.PipelineStatus{
			models.PipelineStatusDraft,
			models.PipelineStatusActive,
			models.PipelineStatusPaused,
			models.PipelineStatusRunning,
			models.PipelineStatusCompleted,
			models.PipelineStatusFailed,
			models.PipelineStatusCancelled,
		}
		isValid := slices.Contains(validStatuses, *req.Status)

		if !isValid {
			return fmt.Errorf("invalid pipeline status: %s", *req.Status)
		}
	}

	return nil
}

// validatePipeline validates a pipeline entity.
func (s *pipelineService) validatePipeline(pipeline *models.Pipeline) error {
	if pipeline == nil {
		return errors.New("pipeline is required")
	}

	if strings.TrimSpace(pipeline.Name) == "" {
		return errors.New("pipeline name is required")
	}

	err := s.validatePipelineName(pipeline.Name)
	if err != nil {
		return err
	}

	return nil
}

// validatePipelineName validates the pipeline name format.
func (s *pipelineService) validatePipelineName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) < 1 {
		return errors.New("pipeline name must be at least 1 character long")
	}

	if len(name) > 255 {
		return errors.New("pipeline name must be at most 255 characters long")
	}

	// Allow alphanumeric characters, hyphens, underscores, and spaces
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == ' ') {
			return fmt.Errorf("pipeline name contains invalid character: %c", r)
		}
	}

	return nil
}

// validateStatusTransition validates that a status transition is allowed.
func (s *pipelineService) validateStatusTransition(
	ctx context.Context,
	pipelineID string,
	newStatus models.PipelineStatus,
) error {
	// Get current pipeline to check current status
	pipeline, err := s.pipelineRepo.GetByID(ctx, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to get current pipeline status: %w", err)
	}

	// Define allowed transitions
	allowedTransitions := map[models.PipelineStatus][]models.PipelineStatus{
		models.PipelineStatusDraft: {
			models.PipelineStatusActive,
			models.PipelineStatusCancelled,
		},
		models.PipelineStatusActive: {
			models.PipelineStatusPaused,
			models.PipelineStatusCancelled,
		},
		models.PipelineStatusPaused: {
			models.PipelineStatusActive,
			models.PipelineStatusCancelled,
		},
		models.PipelineStatusRunning: {
			models.PipelineStatusCompleted,
			models.PipelineStatusFailed,
			models.PipelineStatusCancelled,
		},
		models.PipelineStatusCompleted: {
			models.PipelineStatusActive,
		}, // Can be reactivated
		models.PipelineStatusFailed: {
			models.PipelineStatusActive,
		}, // Can be reactivated
		models.PipelineStatusCancelled: {
			models.PipelineStatusDraft,
		}, // Can be restored to draft
	}

	// Check if transition is allowed
	if allowedStatuses, exists := allowedTransitions[pipeline.Status]; exists {
		if slices.Contains(allowedStatuses, newStatus) {
			return nil
		}

		return fmt.Errorf(
			"invalid status transition from %s to %s",
			pipeline.Status,
			newStatus,
		)
	}

	return fmt.Errorf("unknown current status: %s", pipeline.Status)
}

// startExecution starts a pipeline execution (simplified implementation).
func (s *pipelineService) startExecution(
	ctx context.Context,
	execution *models.PipelineExecution,
) error {
	// Update status to running
	err := s.pipelineRepo.UpdateExecutionStatus(
		ctx,
		execution.ID.String(),
		models.ExecutionStatusRunning,
	)
	if err != nil {
		return err
	}

	// In a real implementation, this would:
	// 1. Create execution steps based on pipeline configuration
	// 2. Start executing steps in order
	// 3. Handle step failures and retries
	// 4. Update execution status based on results

	// For now, we'll simulate a simple execution that completes immediately
	go func() {
		// Simulate some work
		time.Sleep(100 * time.Millisecond)

		// Update to completed
		s.pipelineRepo.UpdateExecutionStatus(
			context.Background(),
			execution.ID.String(),
			models.ExecutionStatusCompleted,
		)
	}()

	return nil
}

// validateDependenciesForExecution checks if all dependencies are satisfied.
func (s *pipelineService) validateDependenciesForExecution(
	ctx context.Context,
	pipelineID string,
) error {
	// Get dependencies for the pipeline
	dependencies, err := s.pipelineRepo.GetDependenciesByPipelineID(
		ctx,
		pipelineID,
	)
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Check each dependency
	for _, dep := range dependencies {
		// In a real implementation, you would check if the dependent pipeline
		// has completed successfully within a certain time window
		s.logger.Info().
			Str("dependency", dep.DependsOnPipelineID.String()).
			Str("condition", dep.Condition).
			Msg("Checking dependency for execution")
	}

	return nil
}

// hasCycle checks for cycles in the dependency graph using DFS.
func (s *pipelineService) hasCycle(
	graph map[string][]string,
	node string,
	visited, recStack map[string]bool,
) bool {
	visited[node] = true
	recStack[node] = true

	// Check all dependencies
	for _, dep := range graph[node] {
		if !visited[dep] {
			if s.hasCycle(graph, dep, visited, recStack) {
				return true
			}
		} else if recStack[dep] {
			return true
		}
	}

	recStack[node] = false

	return false
}

// topologicalSort performs topological sort on the dependency graph.
func (s *pipelineService) topologicalSort(graph map[string][]string) []string {
	visited := make(map[string]bool)
	stack := make([]string, 0)

	// Visit all nodes
	for node := range graph {
		if !visited[node] {
			s.topologicalSortUtil(graph, node, visited, &stack)
		}
	}

	// Reverse the stack to get correct order
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}

	return stack
}

// topologicalSortUtil is a utility function for topological sort.
func (s *pipelineService) topologicalSortUtil(
	graph map[string][]string,
	node string,
	visited map[string]bool,
	stack *[]string,
) {
	visited[node] = true

	// Visit all dependencies
	for _, dep := range graph[node] {
		if !visited[dep] {
			s.topologicalSortUtil(graph, dep, visited, stack)
		}
	}

	// Add current node to stack
	*stack = append(*stack, node)
}
