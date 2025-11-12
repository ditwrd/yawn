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

// AssetResponse represents an asset in API responses.
type AssetResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Version      string  `json:"version"`
	ProjectID    string  `json:"project_id"`
	RepositoryID *string `json:"repository_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`

	// Optional nested relationships
	Project     *ProjectSummaryResponse `json:"project,omitempty"`
	Repository  *RepositoryResponse     `json:"repository,omitempty"`
	PipelineIDs []string                `json:"pipeline_ids,omitempty"`
}

// AssetSummaryResponse represents a simplified asset for nested responses.
type AssetSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// AssetRequestsWithID represents a request that needs Asset ID as path
// parameter.
type AssetRequestsWithID struct {
	AssetID uuid.UUID `example:"123e4567-e89b-12d3-a456-426614174000" param:"id"`
}

// CreateAssetRequest represents the request to create a new asset.
type CreateAssetRequest struct {
	Name         string  `json:"name"                    validate:"required,min=1,max=255"`
	Description  string  `json:"description"             validate:"max=1000"`
	Version      string  `json:"version"                 validate:"required,min=1,max=50"`
	ProjectID    string  `json:"project_id"              validate:"required,uuid"`
	RepositoryID *string `json:"repository_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateAssetRequest represents the request to update an asset.
type UpdateAssetRequest struct {
	AssetID      uuid.UUID `example:"123e4567-e89b-12d3-a456-426614174000" param:"id"`
	Name         *string   `                                                          json:"name,omitempty"          validate:"omitempty,min=1,max=255"`
	Description  *string   `                                                          json:"description,omitempty"   validate:"omitempty,max=1000"`
	Version      *string   `                                                          json:"version,omitempty"       validate:"omitempty,min=1,max=50"`
	RepositoryID *string   `                                                          json:"repository_id,omitempty" validate:"omitempty,uuid"`
}

// AssetListResponse represents a paginated list of assets.
type AssetListResponse struct {
	Assets []AssetResponse `json:"assets"`
	Total  int64           `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// AssetSearchResponse represents search results for assets.
type AssetSearchResponse struct {
	Assets []AssetResponse `json:"assets"`
	Total  int64           `json:"total"`
	Query  string          `json:"query"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// AssetVersionHistoryRequest represents a request to get asset version history.
type AssetVersionHistoryRequest struct {
	AssetID uuid.UUID `example:"123e4567-e89b-12d3-a456-426614174000" param:"id"`
}

// AssetVersionHistoryResponse represents the version history of an asset.
type AssetVersionHistoryResponse struct {
	AssetName string         `json:"asset_name"`
	Versions  []AssetVersion `json:"versions"`
}

// AssetVersion represents a single version of an asset.
type AssetVersion struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AssetSearchRequest represents a request to search for assets.
type AssetSearchRequest struct {
	Query     string `example:"web-app" form:"query"      json:"query,omitempty"`
	ProjectID string `                  form:"project_id" json:"project_id,omitempty" validate:"omitempty,uuid"`
	Page      int    `example:"1"       form:"page"       json:"page,omitempty"       validate:"omitempty,min=1"`
	Limit     int    `example:"10"      form:"limit"      json:"limit,omitempty"      validate:"omitempty,min=1,max=100"`
}

// AssetDeleteResponse represents the response when an asset is deleted.
type AssetDeleteResponse struct {
	Message string `json:"message"`
}

// RepositoryResponse represents a repository in API responses.
type RepositoryResponse struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	Branch       string `json:"branch"`
	LatestCommit string `json:"latest_commit"`
	SyncStatus   string `json:"sync_status"`
	ProjectID    string `json:"project_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// RepositorySummaryResponse represents a simplified repository for nested
// responses.
type RepositorySummaryResponse struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
}

// ProjectSummaryResponse represents a simplified project for nested responses.
type ProjectSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// AssetPipelineResponse represents a pipeline in API responses for asset
// context.
type AssetPipelineResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ProjectID   string                 `json:"project_id"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Assets      []AssetSummaryResponse `json:"assets,omitempty"`
}

// AssetValidationErrorResponse represents validation errors for asset
// operations.
type AssetValidationErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string][]string    `json:"details,omitempty"`
	Fields  []AssetValidationError `json:"fields,omitempty"`
}

// AssetValidationError represents a single field validation error.
type AssetValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// AssetMetricsResponse represents metrics for assets.
type AssetMetricsResponse struct {
	TotalAssets     int64               `json:"total_assets"`
	AssetsByProject []ProjectAssetCount `json:"assets_by_project"`
	AssetsByVersion map[string]int64    `json:"assets_by_version"`
	RecentActivity  []AssetActivityItem `json:"recent_activity"`
}

// ProjectAssetCount represents asset count per project.
type ProjectAssetCount struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Count       int64  `json:"count"`
}

// AssetActivityItem represents recent asset activity.
type AssetActivityItem struct {
	AssetID   string    `json:"asset_id"`
	AssetName string    `json:"asset_name"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
}
