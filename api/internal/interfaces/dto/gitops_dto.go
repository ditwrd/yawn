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

// SyncRepositoryResponse represents the response for repository
// synchronization.
type SyncRepositoryResponse struct {
	Success    bool                `json:"success"`
	CommitHash string              `json:"commit_hash"`
	Message    string              `json:"message"`
	Pipelines  []string            `json:"pipelines"`
	Duration   string              `json:"duration"`
	SyncedAt   string              `json:"synced_at"`
	Changes    SyncChangesResponse `json:"changes"`
	Errors     []string            `json:"errors,omitempty"`
}

// SyncChangesResponse represents synchronization changes.
type SyncChangesResponse struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Deleted  []string `json:"deleted"`
}

// WebhookRequest represents a Git webhook request.
type WebhookRequest struct {
	EventType string         `json:"event_type" validate:"required"`
	RepoURL   string         `json:"repo_url"   validate:"required,url"`
	Branch    string         `json:"branch"     validate:"required"`
	Commit    string         `json:"commit"     validate:"required"`
	Timestamp time.Time      `json:"timestamp"  validate:"required"`
	Payload   map[string]any `json:"payload"`
}

// WebhookResponse represents the response for webhook processing.
type WebhookResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// GitOpsSyncResponse represents a GitOps synchronization response.
type GitOpsSyncResponse struct {
	Success    bool                `json:"success"`
	Message    string              `json:"message"`
	CommitHash string              `json:"commit_hash,omitempty"`
	Duration   string              `json:"duration,omitempty"`
	Changes    SyncChangesResponse `json:"changes,omitempty"`
	Errors     []string            `json:"errors,omitempty"`
}

// GitOpsStatusResponse represents the GitOps status response.
type GitOpsStatusResponse struct {
	RepositoryID   string            `json:"repository_id"`
	Status         string            `json:"status"`
	LastSync       string            `json:"last_sync"`
	NextSync       string            `json:"next_sync,omitempty"`
	PendingChanges int               `json:"pending_changes"`
	SyncHistory    []SyncStatusEntry `json:"sync_history"`
	Errors         []string          `json:"errors,omitempty"`
}

// SyncStatusEntry represents a single sync status entry.
type SyncStatusEntry struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Commit    string `json:"commit,omitempty"`
	Message   string `json:"message"`
}

// ValidationResponse represents a validation response.
type ValidationResponse struct {
	Valid   bool     `json:"valid"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

// GitOpsRepositorySyncRequest represents a request to sync a GitOps repository.
type GitOpsRepositorySyncRequest struct {
	RepositoryID uuid.UUID `param:"repoId" example:"123e4567-e89b-12d3-a456-426614174000"`
	Branch       string    `               example:"main"                                 json:"branch,omitempty"  form:"branch"`
	Force        bool      `                                                              json:"force,omitempty"   form:"force"`
	DryRun       bool      `                                                              json:"dry_run,omitempty" form:"dry_run"`
}

// SyncRepositoryRequest represents a request to sync a repository.
type SyncRepositoryRequest struct {
	Branch string `json:"branch,omitempty"`
	Force  bool   `json:"force,omitempty"`
	DryRun bool   `json:"dry_run,omitempty"`
}

// GitOpsRepositoryStatusRequest represents a request to get repository sync
// status.
type GitOpsRepositoryStatusRequest struct {
	RepositoryID uuid.UUID `param:"repoId" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// GitOpsPendingSyncRequest represents a request to get pending synchronizations
// for a repository.
type GitOpsPendingSyncRequest struct {
	RepositoryID uuid.UUID `param:"repoId" example:"123e4567-e89b-12d3-a456-426614174000"`
	Limit        int       `               example:"10"                                   json:"limit,omitempty" form:"limit" validate:"omitempty,min=1,max=100"`
}

// GitOpsRepositoryValidateRequest represents a request to validate a GitOps
// repository.
type GitOpsRepositoryValidateRequest struct {
	RepositoryID uuid.UUID `param:"repoId" example:"123e4567-e89b-12d3-a456-426614174000"`
	URL          string    `               example:"https://github.com/user/repo.git"     json:"url"              validate:"required,url"`
	Branch       string    `               example:"main"                                 json:"branch,omitempty"`
}

// ValidateRepositoryRequest represents the request to validate a repository.
type ValidateRepositoryRequest struct {
	URL    string `json:"url"    validate:"required,url"`
	Branch string `json:"branch"`
}

// ValidateRepositoryResponse represents the response for repository validation.
type ValidateRepositoryResponse struct {
	Valid        bool                `json:"valid"`
	Message      string              `json:"message"`
	URL          string              `json:"url"`
	LatestCommit *CommitInfoResponse `json:"latest_commit,omitempty"`
}

// CommitInfoResponse represents Git commit information.
type CommitInfoResponse struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Email     string `json:"email"`
	Timestamp string `json:"timestamp"`
	Branch    string `json:"branch"`
}

// PendingSyncRequest represents the request to get pending synchronizations.
type PendingSyncRequest struct {
	Limit int `json:"limit" validate:"omitempty,min=1,max=100"`
}

// PendingSyncResponse represents repositories that need synchronization.
type PendingSyncResponse struct {
	Repositories []PendingRepositoryResponse `json:"repositories"`
	Total        int                         `json:"total"`
	Limit        int                         `json:"limit"`
	Message      string                      `json:"message,omitempty"`
}

// PendingRepositoryResponse represents a repository that needs synchronization.
type PendingRepositoryResponse struct {
	ID          string                  `json:"id"`
	URL         string                  `json:"url"`
	Branch      string                  `json:"branch"`
	SyncStatus  string                  `json:"sync_status"`
	LastSync    *string                 `json:"last_sync,omitempty"`
	LastAttempt *string                 `json:"last_attempt,omitempty"`
	Project     *ProjectSummaryResponse `json:"project,omitempty"`
}

// SyncStatusRequest represents the request to get sync status.
type SyncStatusRequest struct {
	RepositoryID string `json:"repository_id" validate:"required,uuid"`
}

// SyncStatusResponse represents the synchronization status of a repository.
type SyncStatusResponse struct {
	RepositoryID  string  `json:"repository_id"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
	LastSync      *string `json:"last_sync,omitempty"`
	NextSync      *string `json:"next_sync,omitempty"`
	SyncCount     int     `json:"sync_count"`
	ErrorCount    int     `json:"error_count"`
	LatestCommit  *string `json:"latest_commit,omitempty"`
	PipelineCount int     `json:"pipeline_count"`
}

// GitRepositoryResponse represents a Git repository in API responses.
type GitRepositoryResponse struct {
	ID           string                  `json:"id"`
	URL          string                  `json:"url"`
	Branch       string                  `json:"branch"`
	LatestCommit string                  `json:"latest_commit"`
	SyncStatus   string                  `json:"sync_status"`
	ProjectID    string                  `json:"project_id"`
	CreatedAt    string                  `json:"created_at"`
	UpdatedAt    string                  `json:"updated_at"`
	Project      *ProjectSummaryResponse `json:"project,omitempty"`
	Assets       []AssetSummaryResponse  `json:"assets,omitempty"`
	SyncInfo     *RepositorySyncInfo     `json:"sync_info,omitempty"`
}

// RepositorySyncInfo represents detailed synchronization information.
type RepositorySyncInfo struct {
	LastSync      *time.Time `json:"last_sync,omitempty"`
	NextSync      *time.Time `json:"next_sync,omitempty"`
	SyncCount     int        `json:"sync_count"`
	ErrorCount    int        `json:"error_count"`
	LastError     *string    `json:"last_error,omitempty"`
	Duration      *string    `json:"duration,omitempty"`
	PipelineCount int        `json:"pipeline_count"`
}

// GitRepositoryListResponse represents a paginated list of Git repositories.
type GitRepositoryListResponse struct {
	Repositories []GitRepositoryResponse `json:"repositories"`
	Total        int64                   `json:"total"`
	Page         int                     `json:"page"`
	Limit        int                     `json:"limit"`
}

// GitRepositorySearchResponse represents search results for repositories.
type GitRepositorySearchResponse struct {
	Repositories []GitRepositoryResponse `json:"repositories"`
	Total        int64                   `json:"total"`
	Query        string                  `json:"query"`
	Page         int                     `json:"page"`
	Limit        int                     `json:"limit"`
}

// PipelineDefinitionResponse represents a pipeline definition from Git.
type PipelineDefinitionResponse struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Path        string         `json:"path"`
	Content     string         `json:"content"`
	Hash        string         `json:"hash"`
	Metadata    map[string]any `json:"metadata"`
	UpdatedAt   string         `json:"updated_at"`
}

// PipelineDiscoveryResponse represents the result of pipeline discovery.
type PipelineDiscoveryResponse struct {
	Success   bool                         `json:"success"`
	Message   string                       `json:"message"`
	Pipelines []PipelineDefinitionResponse `json:"pipelines"`
	Count     int                          `json:"count"`
	Commit    string                       `json:"commit"`
	Branch    string                       `json:"branch"`
}

// GitOpsMetricsResponse represents GitOps metrics and statistics.
type GitOpsMetricsResponse struct {
	TotalRepositories    int64                   `json:"total_repositories"`
	ActiveRepositories   int64                   `json:"active_repositories"`
	PendingSync          int64                   `json:"pending_sync"`
	FailedSync           int64                   `json:"failed_sync"`
	SyncRate             float64                 `json:"sync_rate"`
	AverageSyncDuration  string                  `json:"average_sync_duration"`
	LastSyncTime         *string                 `json:"last_sync_time,omitempty"`
	RepositoriesByStatus []RepositoryStatusCount `json:"repositories_by_status"`
	RecentSyncs          []RecentSyncActivity    `json:"recent_syncs"`
	PipelineStats        GitOpsPipelineStats     `json:"pipeline_stats"`
}

// RepositoryStatusCount represents repository count by status.
type RepositoryStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// RecentSyncActivity represents recent synchronization activity.
type RecentSyncActivity struct {
	RepositoryID  string    `json:"repository_id"`
	RepositoryURL string    `json:"repository_url"`
	Status        string    `json:"status"`
	Duration      string    `json:"duration"`
	Timestamp     time.Time `json:"timestamp"`
}

// GitOpsPipelineStats represents pipeline statistics from Git operations.
type GitOpsPipelineStats struct {
	TotalPipelines  int64   `json:"total_pipelines"`
	FromGit         int64   `json:"from_git"`
	ManuallyCreated int64   `json:"manually_created"`
	LastImportCount int64   `json:"last_import_count"`
	LastImportTime  *string `json:"last_import_time,omitempty"`
}

// GitOpsConfigResponse represents GitOps configuration.
type GitOpsConfigResponse struct {
	SyncInterval       string         `json:"sync_interval"`
	SyncTimeout        string         `json:"sync_timeout"`
	MaxConcurrentSyncs int            `json:"max_concurrent_syncs"`
	WebhookSecret      *string        `json:"webhook_secret,omitempty"`
	DefaultBranch      string         `json:"default_branch"`
	PipelinePatterns   []string       `json:"pipeline_patterns"`
	Settings           map[string]any `json:"settings"`
}

// GitOpsConfigRequest represents the request to update GitOps configuration.
type GitOpsConfigRequest struct {
	SyncInterval       *string        `json:"sync_interval,omitempty"        validate:"omitempty"`
	SyncTimeout        *string        `json:"sync_timeout,omitempty"         validate:"omitempty"`
	MaxConcurrentSyncs *int           `json:"max_concurrent_syncs,omitempty" validate:"omitempty,min=1,max=10"`
	WebhookSecret      *string        `json:"webhook_secret,omitempty"       validate:"omitempty"`
	DefaultBranch      *string        `json:"default_branch,omitempty"       validate:"omitempty"`
	PipelinePatterns   []string       `json:"pipeline_patterns,omitempty"    validate:"omitempty,dive,required"`
	Settings           map[string]any `json:"settings,omitempty"`
}

// GitValidationErrorResponse represents validation errors for Git operations.
type GitValidationErrorResponse struct {
	Error   string               `json:"error"`
	Code    string               `json:"code"`
	Details map[string][]string  `json:"details,omitempty"`
	Fields  []GitValidationError `json:"fields,omitempty"`
}

// GitValidationError represents a single Git validation error.
type GitValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}
