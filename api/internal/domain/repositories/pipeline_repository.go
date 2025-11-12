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
	"time"

	"gorm.io/gorm"

	"github.com/ditwrd/yawn/api/internal/domain/models"
)

// PipelineRepository defines the interface for pipeline data operations.
//
// Provides methods for CRUD operations, execution management, dependency
// resolution, and project-based filtering. All operations are context-aware
// and include proper error handling.
type PipelineRepository interface {
	// Create inserts a new pipeline into the database
	Create(ctx context.Context, pipeline *models.Pipeline) error

	// GetByID retrieves a pipeline by its ID with relationships
	GetByID(ctx context.Context, id string) (*models.Pipeline, error)

	// GetByProjectID retrieves all pipelines for a specific project with
	// pagination
	GetByProjectID(
		ctx context.Context,
		projectID string,
		limit, offset int,
	) ([]*models.Pipeline, error)

	// List retrieves all pipelines with pagination and optional filtering
	List(
		ctx context.Context,
		limit, offset int,
		filters PipelineFilters,
	) ([]*models.Pipeline, error)

	// Update modifies an existing pipeline in the database
	Update(ctx context.Context, pipeline *models.Pipeline) error

	// Delete performs a soft delete on a pipeline
	Delete(ctx context.Context, id string) error

	// Search finds pipelines by name or description with pagination
	Search(
		ctx context.Context,
		query string,
		limit, offset int,
	) ([]*models.Pipeline, error)

	// GetActivePipelines retrieves all active pipelines that can be executed
	GetActivePipelines(ctx context.Context) ([]*models.Pipeline, error)

	// UpdateStatus updates the status of a pipeline
	UpdateStatus(
		ctx context.Context,
		id string,
		status models.PipelineStatus,
	) error

	// Count returns the total number of pipelines matching the filters
	Count(ctx context.Context, filters PipelineFilters) (int64, error)

	// Exists checks if a pipeline exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByName checks if a pipeline exists by name within a project
	ExistsByName(ctx context.Context, projectID, name string) (bool, error)

	// === Pipeline Execution Operations ===

	// CreateExecution creates a new pipeline execution
	CreateExecution(
		ctx context.Context,
		execution *models.PipelineExecution,
	) error

	// GetExecutionByID retrieves an execution by its ID
	GetExecutionByID(
		ctx context.Context,
		id string,
	) (*models.PipelineExecution, error)

	// GetExecutionsByPipelineID retrieves executions for a pipeline with
	// pagination
	GetExecutionsByPipelineID(
		ctx context.Context,
		pipelineID string,
		limit, offset int,
	) ([]*models.PipelineExecution, error)

	// UpdateExecution updates an existing execution
	UpdateExecution(
		ctx context.Context,
		execution *models.PipelineExecution,
	) error

	// UpdateExecutionStatus updates the status of an execution
	UpdateExecutionStatus(
		ctx context.Context,
		id string,
		status models.ExecutionStatus,
	) error

	// GetRunningExecutions retrieves all currently running executions
	GetRunningExecutions(ctx context.Context) ([]*models.PipelineExecution, error)

	// === Execution Step Operations ===

	// CreateStep creates a new execution step
	CreateStep(ctx context.Context, step *models.ExecutionStep) error

	// GetStepsByExecutionID retrieves steps for an execution ordered by execution
	// order
	GetStepsByExecutionID(
		ctx context.Context,
		executionID string,
	) ([]*models.ExecutionStep, error)

	// UpdateStep updates an existing execution step
	UpdateStep(ctx context.Context, step *models.ExecutionStep) error

	// === Pipeline Dependency Operations ===

	// CreateDependency creates a new pipeline dependency
	CreateDependency(
		ctx context.Context,
		dependency *models.PipelineDependency,
	) error

	// GetDependenciesByPipelineID retrieves dependencies for a pipeline
	GetDependenciesByPipelineID(
		ctx context.Context,
		pipelineID string,
	) ([]*models.PipelineDependency, error)

	// GetDependentsByPipelineID retrieves pipelines that depend on a specific
	// pipeline
	GetDependentsByPipelineID(
		ctx context.Context,
		pipelineID string,
	) ([]*models.PipelineDependency, error)

	// DeleteDependency removes a pipeline dependency
	DeleteDependency(ctx context.Context, pipelineID, dependsOnID string) error

	// GetDependencyGraph builds a dependency graph for pipelines in a project
	GetDependencyGraph(
		ctx context.Context,
		projectID string,
	) (map[string][]string, error)
}

// PipelineFilters defines filtering options for pipeline queries.
type PipelineFilters struct {
	ProjectID string
	Status    models.PipelineStatus
	Name      string
	Search    string // General search across name and description
	IsEnabled *bool
	Schedule  string
}

// pipelineRepository implements the PipelineRepository interface using GORM.
type pipelineRepository struct {
	db     *gorm.DB
	logger interface {
		Info(msg string, fields ...any)
	}
}

// getCaseInsensitiveOperator returns the appropriate case-insensitive operator
// based on the database dialect (ILIKE for PostgreSQL, LIKE for SQLite).
func (r *pipelineRepository) getCaseInsensitiveOperator() string {
	if dialector, ok := r.db.Dialector.(interface{ Name() string }); ok {
		if dialector.Name() == "postgres" {
			return "ILIKE"
		}
	}

	return "LIKE"
}

// NewPipelineRepository creates a new instance of PipelineRepository.
//
// Parameters:
//   - db: GORM database instance
//   - logger: Logger for debugging and monitoring
//
// Returns:
//   - PipelineRepository: An instance of the pipeline repository
func NewPipelineRepository(db *gorm.DB, logger interface {
	Info(msg string, fields ...any)
},
) PipelineRepository {
	return &pipelineRepository{
		db:     db,
		logger: logger,
	}
}

// === Pipeline CRUD Operations ===

// Create inserts a new pipeline into the database.
func (r *pipelineRepository) Create(
	ctx context.Context,
	pipeline *models.Pipeline,
) error {
	err := r.db.WithContext(ctx).Create(pipeline).Error
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Pipeline created successfully",
			"pipeline_id", pipeline.ID,
			"name", pipeline.Name,
			"project_id", pipeline.ProjectID,
		)
	}

	return nil
}

// GetByID retrieves a pipeline by its ID with relationships.
func (r *pipelineRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Pipeline, error) {
	var pipeline models.Pipeline

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Assets").
		Preload("Dependencies.DependsOn").
		Preload("Dependents.Pipeline").
		Where("id = ?", id).
		First(&pipeline).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("pipeline with id %s not found", id)
		}

		return nil, fmt.Errorf("failed to get pipeline by id %s: %w", id, err)
	}

	return &pipeline, nil
}

// GetByProjectID retrieves all pipelines for a specific project with
// pagination.
func (r *pipelineRepository) GetByProjectID(
	ctx context.Context,
	projectID string,
	limit, offset int,
) ([]*models.Pipeline, error) {
	var pipelines []*models.Pipeline

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Assets").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&pipelines).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get pipelines for project %s: %w",
			projectID,
			err,
		)
	}

	return pipelines, nil
}

// List retrieves all pipelines with pagination and optional filtering.
func (r *pipelineRepository) List(
	ctx context.Context,
	limit, offset int,
	filters PipelineFilters,
) ([]*models.Pipeline, error) {
	var pipelines []*models.Pipeline

	query := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Assets")

	// Apply filters
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.Name != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where("name "+operator+" ?", "%"+filters.Name+"%")
	}

	if filters.Search != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where(
			"name "+operator+" ? OR description "+operator+" ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	if filters.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filters.IsEnabled)
	}

	if filters.Schedule != "" {
		query = query.Where("schedule = ?", filters.Schedule)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&pipelines).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}

	return pipelines, nil
}

// Update modifies an existing pipeline in the database.
func (r *pipelineRepository) Update(
	ctx context.Context,
	pipeline *models.Pipeline,
) error {
	result := r.db.WithContext(ctx).Save(pipeline)
	if result.Error != nil {
		return fmt.Errorf("failed to update pipeline: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"pipeline with id %s not found or no changes made",
			pipeline.ID,
		)
	}

	if r.logger != nil {
		r.logger.Info("Pipeline updated successfully",
			"pipeline_id", pipeline.ID,
			"name", pipeline.Name,
		)
	}

	return nil
}

// Delete performs a soft delete on a pipeline.
func (r *pipelineRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Pipeline{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete pipeline: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("pipeline with id %s not found", id)
	}

	if r.logger != nil {
		r.logger.Info("Pipeline deleted successfully", "pipeline_id", id)
	}

	return nil
}

// Search finds pipelines by name or description with pagination.
func (r *pipelineRepository) Search(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]*models.Pipeline, error) {
	var pipelines []*models.Pipeline

	operator := r.getCaseInsensitiveOperator()

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Assets").
		Where("name "+operator+" ? OR description "+operator+" ?", "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&pipelines).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search pipelines: %w", err)
	}

	return pipelines, nil
}

// GetActivePipelines retrieves all active pipelines that can be executed.
func (r *pipelineRepository) GetActivePipelines(
	ctx context.Context,
) ([]*models.Pipeline, error) {
	var pipelines []*models.Pipeline

	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Assets").
		Where("status = ? AND is_enabled = ?", models.PipelineStatusActive, true).
		Order("created_at DESC").
		Find(&pipelines).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active pipelines: %w", err)
	}

	return pipelines, nil
}

// UpdateStatus updates the status of a pipeline.
func (r *pipelineRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status models.PipelineStatus,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Pipeline{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update pipeline status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("pipeline with id %s not found", id)
	}

	if r.logger != nil {
		r.logger.Info("Pipeline status updated successfully",
			"pipeline_id", id,
			"status", status,
		)
	}

	return nil
}

// Count returns the total number of pipelines matching the filters.
func (r *pipelineRepository) Count(
	ctx context.Context,
	filters PipelineFilters,
) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.Pipeline{})

	// Apply filters
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.Name != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where("name "+operator+" ?", "%"+filters.Name+"%")
	}

	if filters.Search != "" {
		operator := r.getCaseInsensitiveOperator()
		query = query.Where(
			"name "+operator+" ? OR description "+operator+" ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	if filters.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filters.IsEnabled)
	}

	if filters.Schedule != "" {
		query = query.Where("schedule = ?", filters.Schedule)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count pipelines: %w", err)
	}

	return count, nil
}

// Exists checks if a pipeline exists by ID.
func (r *pipelineRepository) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Pipeline{}).
		Where("id = ?", id).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to check pipeline existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a pipeline exists by name within a project.
func (r *pipelineRepository) ExistsByName(
	ctx context.Context,
	projectID, name string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&models.Pipeline{}).
		Where("project_id = ? AND name = ?", projectID, name).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf(
			"failed to check pipeline existence by name: %w",
			err,
		)
	}

	return count > 0, nil
}

// === Pipeline Execution Operations ===

// CreateExecution creates a new pipeline execution.
func (r *pipelineRepository) CreateExecution(
	ctx context.Context,
	execution *models.PipelineExecution,
) error {
	err := r.db.WithContext(ctx).Create(execution).Error
	if err != nil {
		return fmt.Errorf("failed to create pipeline execution: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Pipeline execution created successfully",
			"execution_id", execution.ID,
			"pipeline_id", execution.PipelineID,
		)
	}

	return nil
}

// GetExecutionByID retrieves an execution by its ID.
func (r *pipelineRepository) GetExecutionByID(
	ctx context.Context,
	id string,
) (*models.PipelineExecution, error) {
	var execution models.PipelineExecution

	err := r.db.WithContext(ctx).
		Preload("Pipeline").
		Preload("TriggerUser").
		Preload("Steps").
		Where("id = ?", id).
		First(&execution).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("pipeline execution with id %s not found", id)
		}

		return nil, fmt.Errorf(
			"failed to get pipeline execution by id %s: %w",
			id,
			err,
		)
	}

	return &execution, nil
}

// GetExecutionsByPipelineID retrieves executions for a pipeline with
// pagination.
func (r *pipelineRepository) GetExecutionsByPipelineID(
	ctx context.Context,
	pipelineID string,
	limit, offset int,
) ([]*models.PipelineExecution, error) {
	var executions []*models.PipelineExecution

	err := r.db.WithContext(ctx).
		Preload("Pipeline").
		Preload("TriggerUser").
		Where("pipeline_id = ?", pipelineID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get executions for pipeline %s: %w",
			pipelineID,
			err,
		)
	}

	return executions, nil
}

// UpdateExecution updates an existing execution.
func (r *pipelineRepository) UpdateExecution(
	ctx context.Context,
	execution *models.PipelineExecution,
) error {
	result := r.db.WithContext(ctx).Save(execution)
	if result.Error != nil {
		return fmt.Errorf("failed to update pipeline execution: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"pipeline execution with id %s not found or no changes made",
			execution.ID,
		)
	}

	return nil
}

// UpdateExecutionStatus updates the status of an execution.
func (r *pipelineRepository) UpdateExecutionStatus(
	ctx context.Context,
	id string,
	status models.ExecutionStatus,
) error {
	now := time.Now()
	updates := map[string]any{
		"status":     status,
		"updated_at": now,
	}

	// Update completion fields based on status
	switch status {
	case models.ExecutionStatusRunning:
		updates["started_at"] = &now
	case models.ExecutionStatusCompleted,
		models.ExecutionStatusFailed,
		models.ExecutionStatusCancelled,
		models.ExecutionStatusTimeout:
		updates["completed_at"] = &now
		// Calculate duration if started_at is set
		var execution models.PipelineExecution
		err := r.db.WithContext(ctx).
			Select("started_at").
			Where("id = ?", id).
			First(&execution).
			Error
		if err == nil &&
			execution.StartedAt != nil {
			durationSeconds := now.Sub(*execution.StartedAt).Seconds()

			duration := int(durationSeconds)
			if duration == 0 && durationSeconds > 0 {
				duration = 1 // Ensure at least 1 second if there was any duration
			}

			updates["duration"] = duration
		}
	}

	result := r.db.WithContext(ctx).
		Model(&models.PipelineExecution{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update execution status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("pipeline execution with id %s not found", id)
	}

	return nil
}

// GetRunningExecutions retrieves all currently running executions.
func (r *pipelineRepository) GetRunningExecutions(
	ctx context.Context,
) ([]*models.PipelineExecution, error) {
	var executions []*models.PipelineExecution

	err := r.db.WithContext(ctx).
		Preload("Pipeline").
		Where("status = ?", models.ExecutionStatusRunning).
		Order("started_at ASC").
		Find(&executions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get running executions: %w", err)
	}

	return executions, nil
}

// === Execution Step Operations ===

// CreateStep creates a new execution step.
func (r *pipelineRepository) CreateStep(
	ctx context.Context,
	step *models.ExecutionStep,
) error {
	err := r.db.WithContext(ctx).Create(step).Error
	if err != nil {
		return fmt.Errorf("failed to create execution step: %w", err)
	}

	return nil
}

// GetStepsByExecutionID retrieves steps for an execution ordered by execution
// order.
func (r *pipelineRepository) GetStepsByExecutionID(
	ctx context.Context,
	executionID string,
) ([]*models.ExecutionStep, error) {
	var steps []*models.ExecutionStep

	err := r.db.WithContext(ctx).
		Where("execution_id = ?", executionID).
		Order("order ASC").
		Find(&steps).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get steps for execution %s: %w",
			executionID,
			err,
		)
	}

	return steps, nil
}

// UpdateStep updates an existing execution step.
func (r *pipelineRepository) UpdateStep(
	ctx context.Context,
	step *models.ExecutionStep,
) error {
	result := r.db.WithContext(ctx).Save(step)
	if result.Error != nil {
		return fmt.Errorf("failed to update execution step: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"execution step with id %s not found or no changes made",
			step.ID,
		)
	}

	return nil
}

// === Pipeline Dependency Operations ===

// CreateDependency creates a new pipeline dependency.
func (r *pipelineRepository) CreateDependency(
	ctx context.Context,
	dependency *models.PipelineDependency,
) error {
	err := r.db.WithContext(ctx).Create(dependency).Error
	if err != nil {
		return fmt.Errorf("failed to create pipeline dependency: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Pipeline dependency created successfully",
			"dependency_id", dependency.ID,
			"pipeline_id", dependency.PipelineID,
			"depends_on", dependency.DependsOnPipelineID,
		)
	}

	return nil
}

// GetDependenciesByPipelineID retrieves dependencies for a pipeline.
func (r *pipelineRepository) GetDependenciesByPipelineID(
	ctx context.Context,
	pipelineID string,
) ([]*models.PipelineDependency, error) {
	var dependencies []*models.PipelineDependency

	err := r.db.WithContext(ctx).
		Preload("DependsOn").
		Where("pipeline_id = ?", pipelineID).
		Order("created_at ASC").
		Find(&dependencies).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get dependencies for pipeline %s: %w",
			pipelineID,
			err,
		)
	}

	return dependencies, nil
}

// GetDependentsByPipelineID retrieves pipelines that depend on a specific
// pipeline.
func (r *pipelineRepository) GetDependentsByPipelineID(
	ctx context.Context,
	pipelineID string,
) ([]*models.PipelineDependency, error) {
	var dependencies []*models.PipelineDependency

	err := r.db.WithContext(ctx).
		Preload("Pipeline").
		Where("depends_on_pipeline_id = ?", pipelineID).
		Order("created_at ASC").
		Find(&dependencies).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get dependents for pipeline %s: %w",
			pipelineID,
			err,
		)
	}

	return dependencies, nil
}

// DeleteDependency removes a pipeline dependency.
func (r *pipelineRepository) DeleteDependency(
	ctx context.Context,
	pipelineID, dependsOnID string,
) error {
	result := r.db.WithContext(ctx).
		Where("pipeline_id = ? AND depends_on_pipeline_id = ?", pipelineID, dependsOnID).
		Delete(&models.PipelineDependency{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete pipeline dependency: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("pipeline dependency not found")
	}

	if r.logger != nil {
		r.logger.Info("Pipeline dependency deleted successfully",
			"pipeline_id", pipelineID,
			"depends_on", dependsOnID,
		)
	}

	return nil
}

// GetDependencyGraph builds a dependency graph for pipelines in a project.
func (r *pipelineRepository) GetDependencyGraph(
	ctx context.Context,
	projectID string,
) (map[string][]string, error) {
	var dependencies []models.PipelineDependency

	err := r.db.WithContext(ctx).
		Table("pipeline_dependencies pd").
		Joins("JOIN pipelines p ON pd.pipeline_id = p.id").
		Where("p.project_id = ? AND p.deleted_at IS NULL", projectID).
		Find(&dependencies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Build adjacency list representation of the graph
	graph := make(map[string][]string)

	for _, dep := range dependencies {
		pipelineID := dep.PipelineID.String()
		dependsOnID := dep.DependsOnPipelineID.String()

		if graph[pipelineID] == nil {
			graph[pipelineID] = []string{}
		}

		graph[pipelineID] = append(graph[pipelineID], dependsOnID)
	}

	return graph, nil
}
