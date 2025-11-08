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
// This package contains the DTOs used for authentication-related HTTP requests
// and responses in the YAWN API. All DTOs include proper JSON tags and
// validation
// tags for request binding and response formatting.
//
// Design principles:
//   - Separate request and response DTOs for clarity
//   - Use appropriate JSON tags for field names
//   - Include validation tags where applicable
//   - Provide clear documentation for each field
//   - Follow Go naming conventions
package dto

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `example:"user@example.com" json:"email"`
	Password string `example:"SecurePass123!"   json:"password"`
}

// RegisterResponse represents a user registration response.
type RegisterResponse struct {
	ID    string `example:"123e4567-e89b-12d3-a456-426614174000" json:"id"`
	Email string `example:"user@example.com"                     json:"email"`
	Role  string `example:"user"                                 json:"role"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email    string `example:"user@example.com" json:"email"`
	Password string `example:"SecurePass123!"   json:"password"`
}

// UserInfo represents basic user information returned in login response.
type UserInfo struct {
	ID    string `example:"123e4567-e89b-12d3-a456-426614174000" json:"id"`
	Email string `example:"user@example.com"                     json:"email"`
	Role  string `example:"user"                                 json:"role"`
}

// LoginResponse represents a user login response with JWT tokens.
type LoginResponse struct {
	AccessToken  string   `example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." json:"access_token"`
	RefreshToken string   `example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." json:"refresh_token"`
	TokenType    string   `example:"Bearer"                                  json:"token_type"`
	ExpiresIn    int      `example:"900"                                     json:"expires_in"`
	User         UserInfo `                                                  json:"user"`
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." json:"refresh_token"`
}

// RefreshResponse represents a token refresh response.
type RefreshResponse struct {
	AccessToken string `example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." json:"access_token"`
	TokenType   string `example:"Bearer"                                  json:"token_type"`
	ExpiresIn   int    `example:"900"                                     json:"expires_in"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	AccessToken string `example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." json:"access_token"`
}

// LogoutResponse represents a logout response.
type LogoutResponse struct {
	Message string `example:"Logged out successfully" json:"message"`
}

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error   string `example:"Validation failed"                     json:"error"`
	Code    string `example:"VALIDATION_ERROR"                      json:"code"`
	Details string `example:"Please check your input and try again" json:"details,omitempty"`
}
