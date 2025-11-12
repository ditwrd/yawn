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

// Package dto defines data transfer objects for API requests and responses.
//
// This package contains the DTOs used for project management HTTP requests
// and responses in the YAWN API. All DTOs include proper JSON tags and
// validation tags for request binding and response formatting.
//
// Design principles:
//   - Separate request and response DTOs for clarity
//   - Use appropriate JSON tags for field names
//   - Include validation tags where applicable
//   - Provide clear documentation for each field
//   - Follow Go naming conventions
package dto

import "github.com/gofrs/uuid"

// CreateProjectRequest represents a project creation request.
type CreateProjectRequest struct {
	Name        string `example:"My Web Application"                     json:"name"                  validate:"required,min=1,max=255"`
	Description string `example:"A modern web application built with Go" json:"description,omitempty" validate:"max=1000"`
	Repository  string `example:"https://github.com/user/repo.git"       json:"repository,omitempty"  validate:"omitempty,url"`
	Visibility  string `example:"private"                                json:"visibility,omitempty"  validate:"omitempty,oneof=public private"`
}

// UpdateProjectRequest represents a project update request.
type UpdateProjectRequest struct {
	ProjectID   uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `           example:"Updated Project Name"                 json:"name,omitempty"        validate:"omitempty,min=1,max=255"`
	Description string    `           example:"Updated project description"          json:"description,omitempty" validate:"omitempty,max=1000"`
	Repository  string    `           example:"https://github.com/user/new-repo.git" json:"repository,omitempty"  validate:"omitempty,url"`
	Visibility  string    `           example:"public"                               json:"visibility,omitempty"  validate:"omitempty,oneof=public private"`
}

// ProjectResponse represents project information returned by API endpoints.
type ProjectResponse struct {
	ID          string              `example:"123e4567-e89b-12d3-a456-426614174000"   json:"id"`
	Name        string              `example:"My Web Application"                     json:"name"`
	Description string              `example:"A modern web application built with Go" json:"description,omitempty"`
	Repository  *string             `example:"https://github.com/user/repo.git"       json:"repository,omitempty"`
	Visibility  string              `example:"private"                                json:"visibility"`
	Owner       UserInfo            `                                                 json:"owner"`
	CreatedAt   string              `example:"2025-01-01T00:00:00Z"                   json:"created_at"`
	UpdatedAt   string              `example:"2025-01-01T00:00:00Z"                   json:"updated_at"`
	Members     []ProjectMemberInfo `                                                 json:"members,omitempty"`
}

// ProjectMemberInfo represents information about a project member.
type ProjectMemberInfo struct {
	UserID    string `example:"123e4567-e89b-12d3-a456-426614174000" json:"user_id"`
	Email     string `example:"user@example.com"                     json:"email"`
	Role      string `example:"maintainer"                           json:"role"`
	CreatedAt string `example:"2025-01-01T00:00:00Z"                 json:"created_at"`
}

// ProjectListResponse represents a paginated list of projects.
type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

// AddProjectMemberRequest represents a request to add a member to a project.
type AddProjectMemberRequest struct {
	ProjectID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email     string    `           example:"user@example.com"                     json:"email" validate:"required,email"`
	Role      string    `           example:"maintainer"                           json:"role"  validate:"required,oneof=owner maintainer viewer"`
}

// UpdateProjectMemberRequest represents a request to update a project member
// role.
type UpdateProjectMemberRequest struct {
	ProjectID uuid.UUID `param:"id"       example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    uuid.UUID `param:"memberId" example:"456e7890-f12c-34d5-a678-426614174111"`
	Role      string    `                 example:"viewer"                               json:"role" validate:"required,oneof=owner maintainer viewer"`
}

// ProjectMemberResponse represents a project member response.
type ProjectMemberResponse struct {
	UserID    string `example:"123e4567-e89b-12d3-a456-426614174000" json:"user_id"`
	Email     string `example:"user@example.com"                     json:"email"`
	Role      string `example:"maintainer"                           json:"role"`
	CreatedAt string `example:"2025-01-01T00:00:00Z"                 json:"created_at"`
	UpdatedAt string `example:"2025-01-01T00:00:00Z"                 json:"updated_at"`
}

// ProjectMemberListResponse represents a list of project members.
type ProjectMemberListResponse struct {
	Members []ProjectMemberResponse `json:"members"`
	Total   int                     `json:"total"`
}

// ProjectDeleteResponse represents a successful project deletion response.
type ProjectDeleteResponse struct {
	Message string `example:"Project deleted successfully" json:"message"`
}

// ProjectRequestsWithID represents a request that needs Project ID as path
// parameter.
type ProjectRequestsWithID struct {
	ProjectID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// ProjectMembersRequests represents a request that needs Project ID and User ID
// as path parameters.
type ProjectMembersRequests struct {
	ProjectID uuid.UUID `param:"id"       example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    uuid.UUID `param:"memberId" example:"456e7890-f12c-34d5-a678-426614174111"`
}
