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

// UserResponse represents user information returned by API endpoints.
type UserResponse struct {
	ID        string `json:"id"         example:"123e4567-e89b-12d3-a456-426614174000"`
	Email     string `json:"email"      example:"user@example.com"`
	Role      string `json:"role"       example:"user"`
	CreatedAt string `json:"created_at" example:"2025-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2025-01-01T00:00:00Z"`
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
	Email string `json:"email,omitempty" example:"newemail@example.com"`
	Role  string `json:"role,omitempty"  example:"admin"`
}

// UserDeleteResponse represents a successful user deletion response.
type UserDeleteResponse struct {
	Message string `json:"message" example:"User deleted successfully"`
}
