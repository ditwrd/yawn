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
// This package contains the DTOs used for user management HTTP requests
// and responses in the YAWN API. All DTOs include proper JSON tags and
// validation
// tags for request binding and response formatting.
package dto

import "github.com/gofrs/uuid"

// UserResponse represents user information returned by API endpoints.
type UserResponse struct {
	ID        string `example:"123e4567-e89b-12d3-a456-426614174000" json:"id"`
	Email     string `example:"user@example.com"                     json:"email"`
	Role      string `example:"user"                                 json:"role"`
	CreatedAt string `example:"2025-01-01T00:00:00Z"                 json:"created_at"`
	UpdatedAt string `example:"2025-01-01T00:00:00Z"                 json:"updated_at"`
}

// UserListResponse represents a paginated list of users.
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	UserID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email  string    `           example:"newemail@example.com"                 json:"email,omitempty"`
	Role   string    `           example:"admin"                                json:"role,omitempty"`
}

// UserRequestsWithID represents a request that needs User ID as path parameter.
type UserRequestsWithID struct {
	UserID uuid.UUID `param:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UserDeleteResponse represents a successful user deletion response.
type UserDeleteResponse struct {
	Message string `example:"User deleted successfully" json:"message"`
}
