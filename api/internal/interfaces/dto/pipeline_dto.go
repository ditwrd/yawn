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

// Package dto defines data transfer objects for API request/response handling.
//
// This package contains structured types for API contracts, providing
// validation, serialization, and documentation support for all endpoints.
// All DTOs include JSON tags and validation annotations.
package dto

import (
	"time"

	"github.com/gofrs/uuid"
)

// PipelineResponse represents a pipeline in API responses.
type PipelineResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ProjectID   string `json:"project_id"`
	Status      string `json:"status"`
	Config      string `json:"config,omitempty"`
	Schedule    string `json:"schedule,omitempty"`
	IsEnabled   bool   `json:"is_enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`

	// Optional nested relationships
	Project      *ProjectSummaryResponse      `json:"project,omitempty"`
	Assets       []AssetSummaryResponse       `json:"assets,omitempty"`
	Executions   []PipelineExecutionResponse  `json:"executions,omitempty"`
	Dependencies []PipelineDependencyResponse `json:"dependencies,omitempty"`
	Dependents   []PipelineDependencyResponse `json:"dependents,omitempty"`
}

// PipelineSummaryResponse represents a simplified pipeline for nested
// responses.
type PipelineSummaryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	IsEnabled bool   `json:"is_enabled"`
}

// PipelineRequestsWithID represents a request that needs Pipeline ID as path
// parameter.
type PipelineRequestsWithID struct {
	PipelineID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ListPipelinesRequest represents the request to list pipelines with filtering
// and pagination.
type ListPipelinesRequest struct {
	Page      int    `query:"page"       example:"1"`
	Limit     int    `query:"limit"      example:"20"`
	ProjectID string `query:"project_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status    string `query:"status"     example:"active"`
	Name      string `query:"name"       example:"build-pipeline"`
}

// CreatePipelineRequest represents the request to create a new pipeline.
type CreatePipelineRequest struct {
	Name        string   `json:"name"                validate:"required,min=1,max=255"`
	Description string   `json:"description"         validate:"max=1000"`
	Config      string   `json:"config"              validate:"max=10000"`
	Schedule    string   `json:"schedule"            validate:"max=255"`
	ProjectID   string   `json:"project_id"          validate:"required,uuid"`
	AssetIDs    []string `json:"asset_ids,omitempty" validate:"omitempty,dive,uuid"`
}

// UpdatePipelineRequest represents the request to update a pipeline.
type UpdatePipelineRequest struct {
	PipelineID  uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        *string   `                                                          json:"name,omitempty"        validate:"omitempty,min=1,max=255"`
	Description *string   `                                                          json:"description,omitempty" validate:"omitempty,max=1000"`
	Config      *string   `                                                          json:"config,omitempty"      validate:"omitempty,max=10000"`
	Schedule    *string   `                                                          json:"schedule,omitempty"    validate:"omitempty,max=255"`
	Status      *string   `                                                          json:"status,omitempty"      validate:"omitempty,oneof=draft active paused running completed failed cancelled"`
	IsEnabled   *bool     `                                                          json:"is_enabled,omitempty"`
	AssetIDs    []string  `                                                          json:"asset_ids,omitempty"   validate:"omitempty,dive,uuid"`
}

// PipelineListResponse represents a paginated list of pipelines.
type PipelineListResponse struct {
	Pipelines []PipelineResponse `json:"pipelines"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
}

// PipelineSearchResponse represents search results for pipelines.
type PipelineSearchResponse struct {
	Pipelines []PipelineResponse `json:"pipelines"`
	Total     int64              `json:"total"`
	Query     string             `json:"query"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
}

// PipelineDeleteResponse represents the response when a pipeline is deleted.
type PipelineDeleteResponse struct {
	Message string `json:"message"`
}

// PipelineExecutionResponse represents a pipeline execution in API responses.
type PipelineExecutionResponse struct {
	ID          string  `json:"id"`
	PipelineID  string  `json:"pipeline_id"`
	Status      string  `json:"status"`
	TriggerType string  `json:"trigger_type"`
	TriggeredBy *string `json:"triggered_by,omitempty"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	Duration    int     `json:"duration"`
	Config      string  `json:"config,omitempty"`
	Logs        string  `json:"logs,omitempty"`
	ErrorMsg    string  `json:"error_msg,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`

	// Optional nested relationships
	Pipeline    *PipelineSummaryResponse `json:"pipeline,omitempty"`
	TriggerUser *UserSummaryResponse     `json:"trigger_user,omitempty"`
	Steps       []ExecutionStepResponse  `json:"steps,omitempty"`
}

// ExecutionStepResponse represents an execution step in API responses.
type ExecutionStepResponse struct {
	ID          string  `json:"id"`
	ExecutionID string  `json:"execution_id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Config      string  `json:"config,omitempty"`
	Result      string  `json:"result,omitempty"`
	ErrorMsg    string  `json:"error_msg,omitempty"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	Duration    int     `json:"duration"`
	Order       int     `json:"order"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// PipelineExecutionListResponse represents a paginated list of pipeline
// executions.
type PipelineExecutionListResponse struct {
	Executions []PipelineExecutionResponse `json:"executions"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
}

// TriggerExecutionRequest represents the request to trigger a pipeline
// execution.
type TriggerExecutionRequest struct {
	Config *string `json:"config,omitempty" validate:"omitempty,max=10000"`
}

// TriggerExecutionResponse represents the response when triggering an
// execution.
type TriggerExecutionResponse struct {
	Execution *PipelineExecutionResponse `json:"execution"`
	Message   string                     `json:"message"`
}

// CancelExecutionResponse represents the response when cancelling an execution.
type CancelExecutionResponse struct {
	Message string `json:"message"`
}

// PipelineDependencyResponse represents a pipeline dependency in API responses.
type PipelineDependencyResponse struct {
	ID                  string `json:"id"`
	PipelineID          string `json:"pipeline_id"`
	DependsOnPipelineID string `json:"depends_on_pipeline_id"`
	Condition           string `json:"condition,omitempty"`
	CreatedAt           string `json:"created_at"`

	// Optional nested relationships
	Pipeline  *PipelineSummaryResponse `json:"pipeline,omitempty"`
	DependsOn *PipelineSummaryResponse `json:"depends_on,omitempty"`
}

// CreateDependencyRequest represents the request to create a pipeline
// dependency.
type CreateDependencyRequest struct {
	DependsOnPipelineID string `json:"depends_on_pipeline_id" validate:"required,uuid"`
	Condition           string `json:"condition,omitempty"    validate:"max=500"`
}

// DependencyGraphResponse represents a dependency graph for pipelines.
type DependencyGraphResponse struct {
	Nodes []DependencyNodeResponse `json:"nodes"`
	Edges []DependencyEdgeResponse `json:"edges"`
}

// DependencyNodeResponse represents a node in the dependency graph.
type DependencyNodeResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// DependencyEdgeResponse represents an edge in the dependency graph.
type DependencyEdgeResponse struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

// UserSummaryResponse represents a simplified user for nested responses.
type UserSummaryResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// PipelineMetricsResponse represents metrics for pipelines.
type PipelineMetricsResponse struct {
	TotalPipelines     int64                      `json:"total_pipelines"`
	PipelinesByStatus  map[string]int64           `json:"pipelines_by_status"`
	PipelinesByProject []ProjectPipelineCount     `json:"pipelines_by_project"`
	RecentExecutions   []PipelineExecutionSummary `json:"recent_executions"`
	ExecutionStats     PipelineExecutionStats     `json:"execution_stats"`
}

// ProjectPipelineCount represents pipeline count per project.
type ProjectPipelineCount struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Count       int64  `json:"count"`
}

// PipelineExecutionSummary represents a summary of a pipeline execution.
type PipelineExecutionSummary struct {
	ExecutionID  string     `json:"execution_id"`
	PipelineID   string     `json:"pipeline_id"`
	PipelineName string     `json:"pipeline_name"`
	Status       string     `json:"status"`
	TriggerType  string     `json:"trigger_type"`
	Duration     int        `json:"duration"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// PipelineExecutionStats represents execution statistics.
type PipelineExecutionStats struct {
	TotalExecutions    int64            `json:"total_executions"`
	SuccessfulRuns     int64            `json:"successful_runs"`
	FailedRuns         int64            `json:"failed_runs"`
	AverageDuration    float64          `json:"average_duration"`
	ExecutionsByStatus map[string]int64 `json:"executions_by_status"`
	ExecutionsByType   map[string]int64 `json:"executions_by_type"`
}

// PipelineValidationErrorResponse represents validation errors for pipeline
// operations.
type PipelineValidationErrorResponse struct {
	Error   string                    `json:"error"`
	Code    string                    `json:"code"`
	Details map[string][]string       `json:"details,omitempty"`
	Fields  []PipelineValidationError `json:"fields,omitempty"`
}

// PipelineValidationError represents a single field validation error.
type PipelineValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// PipelineConfig represents the configuration structure for a pipeline.
type PipelineConfig struct {
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Steps       []PipelineStepConfig `json:"steps"`
	Variables   map[string]any       `json:"variables,omitempty"`
	Timeout     *int                 `json:"timeout,omitempty"` // Timeout in seconds
	Retries     *int                 `json:"retries,omitempty"`
}

// PipelineStepConfig represents a step configuration within a pipeline.
type PipelineStepConfig struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Description     string         `json:"description,omitempty"`
	Config          map[string]any `json:"config,omitempty"`
	DependsOn       []string       `json:"depends_on,omitempty"`
	Condition       *string        `json:"condition,omitempty"`
	Timeout         *int           `json:"timeout,omitempty"` // Timeout in seconds
	ContinueOnError *bool          `json:"continue_on_error,omitempty"`
}

// PipelineSearchRequest represents a request to search for pipelines.
type PipelineSearchRequest struct {
	Query     string `json:"query,omitempty"      form:"query"      example:"deployment"`
	ProjectID string `json:"project_id,omitempty" form:"project_id"                      validate:"omitempty,uuid"`
	Status    string `json:"status,omitempty"     form:"status"                          validate:"omitempty,oneof=draft active paused running completed failed cancelled"`
	Page      int    `json:"page,omitempty"       form:"page"       example:"1"          validate:"omitempty,min=1"`
	Limit     int    `json:"limit,omitempty"      form:"limit"      example:"10"         validate:"omitempty,min=1,max=100"`
}

// PipelineTriggerRequest represents a request to trigger a pipeline execution.
type PipelineTriggerRequest struct {
	PipelineID uuid.UUID              `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Branch     string                 `           example:"main"                                 json:"branch,omitempty"`
	Commit     string                 `           example:"abc123"                               json:"commit,omitempty"`
	Params     map[string]interface{} `                                                          json:"params,omitempty"`
	DryRun     bool                   `                                                          json:"dry_run,omitempty"`
	Config     *string                `                                                          json:"config,omitempty"  validate:"omitempty,max=10000"`
}

// PipelineExecutionsRequest represents a request to get pipeline executions.
type PipelineExecutionsRequest struct {
	PipelineID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Page       int       `           example:"1"                                    json:"page,omitempty"   form:"page"   validate:"omitempty,min=1"`
	Limit      int       `           example:"10"                                   json:"limit,omitempty"  form:"limit"  validate:"omitempty,min=1,max=100"`
	Status     string    `                                                          json:"status,omitempty" form:"status" validate:"omitempty,oneof=pending running completed failed cancelled"`
}

// PipelineExecutionCancelRequest represents a request to cancel a pipeline
// execution.
type PipelineExecutionCancelRequest struct {
	PipelineID  uuid.UUID `param:"id"          example:"123e4567-e89b-12d3-a456-426614174000"`
	ExecutionID uuid.UUID `param:"executionId" example:"456e7890-f12c-34d5-a678-426614174111"`
}

// PipelineDependenciesRequest represents a request to get pipeline dependency
// graph.
type PipelineDependenciesRequest struct {
	PipelineID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// TriggerPipelineRequest represents a request to trigger a pipeline execution.
type TriggerPipelineRequest struct {
	Branch string                 `json:"branch,omitempty"`
	Commit string                 `json:"commit,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
	DryRun bool                   `json:"dry_run,omitempty"`
}

// UpdatePipelineStatusRequest represents a request to update pipeline status.
type UpdatePipelineStatusRequest struct {
	PipelineID uuid.UUID              `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status     string                 `                                                          json:"status"            validate:"required,oneof=active inactive paused"`
	Message    string                 `                                                          json:"message,omitempty"`
	Config     map[string]interface{} `                                                          json:"config,omitempty"`
}
