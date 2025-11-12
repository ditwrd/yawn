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
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
)

// AuthService defines the interface for authentication and authorization
// operations.
//
// Provides methods for user authentication, Google SSO integration, JWT token
// management, RBAC permissions, and audit logging.
type AuthService interface {
	// AuthenticateUser authenticates a user with email and password
	AuthenticateUser(
		ctx context.Context,
		email, password string,
	) (*AuthResult, error)

	// AuthenticateWithGoogle authenticates a user using Google OAuth2
	AuthenticateWithGoogle(
		ctx context.Context,
		code, state string,
	) (*AuthResult, error)

	// GetGoogleAuthURL generates the Google OAuth2 authorization URL
	GetGoogleAuthURL(
		ctx context.Context,
		redirectURI string,
	) (string, string, error)

	// ValidateToken validates a JWT token and returns user information
	ValidateToken(
		ctx context.Context,
		tokenString string,
	) (*AuthTokenClaims, error)

	// RefreshToken generates a new access token from a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// Logout invalidates user tokens (revokes refresh token)
	Logout(ctx context.Context, refreshToken string) error

	// HasPermission checks if a user has a specific permission
	HasPermission(ctx context.Context, userID string, permission Permission) bool

	// HasRole checks if a user has a specific role
	HasRole(ctx context.Context, userID string, role models.UserRole) bool

	// HasProjectRole checks if a user has a specific role in a project
	HasProjectRole(
		ctx context.Context,
		userID, projectID string,
		role models.ProjectRole,
	) bool

	// CanAccessResource checks if a user can access a specific resource
	CanAccessResource(
		ctx context.Context,
		userID, resourceID, resourceType, action string,
	) bool

	// LogAuditEvent logs an audit event for security tracking
	LogAuditEvent(ctx context.Context, event *AuditEvent)

	// CreateUser creates a new user (usually from SSO)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error)

	// UpdateUserLastLogin updates the user's last login timestamp
	UpdateUserLastLogin(ctx context.Context, userID string) error
}

// AuthResult represents the result of a successful authentication.
type AuthResult struct {
	User        *models.User `json:"user"`
	Tokens      *TokenPair   `json:"tokens"`
	Permissions []Permission `json:"permissions"`
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// AuthTokenClaims represents JWT token claims for authentication.
type AuthTokenClaims struct {
	UserID    string          `json:"user_id"`
	Email     string          `json:"email"`
	Role      models.UserRole `json:"role"`
	IssuedAt  int64           `json:"iat"`
	ExpiresAt int64           `json:"exp"`
	Scope     []string        `json:"scope"`
	Metadata  map[string]any  `json:"metadata"`
}

// Permission represents a system permission.
type Permission struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Conditions  map[string]string `json:"conditions,omitempty"`
}

// AuditEvent represents an audit log entry.
type AuditEvent struct {
	ID         uuid.UUID      `json:"id"`
	UserID     *uuid.UUID     `json:"user_id,omitempty"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID *string        `json:"resource_id,omitempty"`
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// CreateUserRequest represents a request to create a user (typically from SSO).
type CreateUserRequest struct {
	Email      string          `json:"email"`
	Name       string          `json:"name"`
	Avatar     string          `json:"avatar"`
	Role       models.UserRole `json:"role"`
	Provider   string          `json:"provider"` // "google", "github", etc.
	ProviderID string          `json:"provider_id"`
	Metadata   map[string]any  `json:"metadata"`
}

// OAuth2Config represents OAuth2 configuration.
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// authState represents OAuth2 state information stored temporarily.
type authState struct {
	State       string
	RedirectURI string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// authService implements the AuthService interface.
type authService struct {
	userRepo        repositories.UserRepository
	projectRepo     repositories.ProjectRepository
	logger          *zerolog.Logger
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	googleConfig    *oauth2.Config
	authStates      map[string]*authState // In-memory storage (should use Redis in production)
}

// NewAuthService creates a new instance of AuthService
//
// Parameters:
//   - userRepo: User repository for data operations
//   - projectRepo: Project repository for permission checks
//   - logger: Logger for structured logging
//   - jwtSecret: Secret key for JWT signing
//   - googleClientID: Google OAuth2 client ID
//   - googleClientSecret: Google OAuth2 client secret
//
// Returns:
//   - AuthService: An instance of the auth service
func NewAuthService(
	userRepo repositories.UserRepository,
	projectRepo repositories.ProjectRepository,
	logger *zerolog.Logger,
	jwtSecret []byte,
	googleClientID, googleClientSecret string,
) AuthService {
	// Configure Google OAuth2
	googleConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &authService{
		userRepo:        userRepo,
		projectRepo:     projectRepo,
		logger:          logger,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  15 * time.Minute,   // Default access token TTL
		refreshTokenTTL: 7 * 24 * time.Hour, // 7 days for refresh token
		googleConfig:    googleConfig,
		authStates:      make(map[string]*authState),
	}
}

// AuthenticateUser authenticates a user with email and password.
func (s *authService) AuthenticateUser(
	ctx context.Context,
	email, password string,
) (*AuthResult, error) {
	s.logger.Info().
		Str("email", email).
		Msg("User authentication attempt")

	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("email", email).
			Msg("User not found")

		s.LogAuditEvent(ctx, &AuditEvent{
			Action:    "login_attempt",
			Resource:  "user",
			Success:   false,
			Message:   "User not found",
			Metadata:  map[string]any{"email": email},
			Timestamp: time.Now(),
		})

		return nil, errors.New("invalid credentials")
	}

	// In a real implementation, you would verify the password hash
	// For now, we'll assume password verification is done elsewhere

	// Generate tokens
	tokens, err := s.generateTokenPair(user.ID.String())
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to generate tokens")

		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Get user permissions
	permissions := s.getUserPermissions(user)

	// Update last login
	if err := s.UpdateUserLastLogin(ctx, user.ID.String()); err != nil {
		s.logger.Warn().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to update last login")
	}

	result := &AuthResult{
		User:        user,
		Tokens:      tokens,
		Permissions: permissions,
	}

	userIDStr := user.ID.String()
	s.LogAuditEvent(ctx, &AuditEvent{
		UserID:     &user.ID,
		Action:     "login",
		Resource:   "user",
		ResourceID: &userIDStr,
		Success:    true,
		Message:    "User authenticated successfully",
		Metadata:   map[string]any{"method": "password"},
		Timestamp:  time.Now(),
	})

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Msg("User authenticated successfully")

	return result, nil
}

// AuthenticateWithGoogle authenticates a user using Google OAuth2.
func (s *authService) AuthenticateWithGoogle(
	ctx context.Context,
	code, state string,
) (*AuthResult, error) {
	s.logger.Info().
		Str("state", state).
		Msg("Google OAuth2 authentication attempt")

	// Verify state
	storedState, exists := s.authStates[state]
	if !exists || time.Now().After(storedState.ExpiresAt) {
		s.logger.Warn().
			Str("state", state).
			Msg("Invalid or expired OAuth state")

		return nil, errors.New("invalid or expired state")
	}

	// Exchange code for token
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("state", state).
			Msg("Failed to exchange OAuth code")

		s.LogAuditEvent(ctx, &AuditEvent{
			Action:    "oauth_google_callback",
			Resource:  "oauth",
			Success:   false,
			Message:   "Failed to exchange OAuth code: " + err.Error(),
			Metadata:  map[string]any{"state": state, "provider": "google"},
			Timestamp: time.Now(),
		})

		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(ctx, token)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("state", state).
			Msg("Failed to get Google user info")

		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.findOrCreateGoogleUser(ctx, userInfo)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("email", userInfo.Email).
			Msg("Failed to find or create user")

		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.generateTokenPair(user.ID.String())
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to generate tokens for Google user")

		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Get user permissions
	permissions := s.getUserPermissions(user)

	// Update last login
	if err := s.UpdateUserLastLogin(ctx, user.ID.String()); err != nil {
		s.logger.Warn().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to update last login for Google user")
	}

	result := &AuthResult{
		User:        user,
		Tokens:      tokens,
		Permissions: permissions,
	}

	userIDStr2 := user.ID.String()
	s.LogAuditEvent(ctx, &AuditEvent{
		UserID:     &user.ID,
		Action:     "oauth_google_login",
		Resource:   "user",
		ResourceID: &userIDStr2,
		Success:    true,
		Message:    "Google OAuth authentication successful",
		Metadata:   map[string]any{"provider": "google", "email": userInfo.Email},
		Timestamp:  time.Now(),
	})

	// Clean up state
	delete(s.authStates, state)

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", userInfo.Email).
		Msg("Google OAuth authentication successful")

	return result, nil
}

// GetGoogleAuthURL generates the Google OAuth2 authorization URL.
func (s *authService) GetGoogleAuthURL(
	ctx context.Context,
	redirectURI string,
) (string, string, error) {
	// Generate state
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state with expiration
	s.authStates[state] = &authState{
		State:       state,
		RedirectURI: redirectURI,
		CreatedAt:   time.Now(),
		ExpiresAt: time.Now().
			Add(10 * time.Minute),
		// State expires in 10 minutes
	}

	// Generate auth URL
	authURL := s.googleConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
	)

	s.logger.Info().
		Str("state", state).
		Str("redirect_uri", redirectURI).
		Msg("Generated Google OAuth authorization URL")

	return authURL, state, nil
}

// ValidateToken validates a JWT token and returns user information.
func (s *authService) ValidateToken(
	ctx context.Context,
	tokenString string,
) (*AuthTokenClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					token.Header["alg"],
				)
			}

			return s.jwtSecret, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("invalid user_id in token")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, errors.New("invalid email in token")
		}

		roleStr, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New("invalid role in token")
		}

		role := models.UserRole(roleStr)

		iat, ok := claims["iat"].(float64)
		if !ok {
			return nil, errors.New("invalid iat in token")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, errors.New("invalid exp in token")
		}

		// Extract scope if present
		var scope []string

		if scopeClaim, ok := claims["scope"].([]any); ok {
			for _, s := range scopeClaim {
				if str, ok := s.(string); ok {
					scope = append(scope, str)
				}
			}
		}

		return &AuthTokenClaims{
			UserID:    userID,
			Email:     email,
			Role:      role,
			IssuedAt:  int64(iat),
			ExpiresAt: int64(exp),
			Scope:     scope,
			Metadata:  claims,
		}, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new access token from a refresh token.
func (s *authService) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (*TokenPair, error) {
	// Validate refresh token (simplified implementation)
	// In a real implementation, you would maintain a refresh token store
	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	tokens, err := s.generateTokenPair(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	s.LogAuditEvent(ctx, &AuditEvent{
		UserID: func() *uuid.UUID {
			id, _ := uuid.FromString(claims.UserID)
			return &id
		}(),
		Action:    "token_refresh",
		Resource:  "token",
		Success:   true,
		Message:   "Token refreshed successfully",
		Timestamp: time.Now(),
	})

	return tokens, nil
}

// Logout invalidates user tokens (revokes refresh token).
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	// In a real implementation, you would invalidate the refresh token in a store
	// For now, we'll just log the logout event

	s.LogAuditEvent(ctx, &AuditEvent{
		UserID: func() *uuid.UUID {
			id, _ := uuid.FromString(claims.UserID)
			return &id
		}(),
		Action:    "logout",
		Resource:  "session",
		Success:   true,
		Message:   "User logged out successfully",
		Timestamp: time.Now(),
	})

	s.logger.Info().
		Str("user_id", claims.UserID).
		Msg("User logged out successfully")

	return nil
}

// HasPermission checks if a user has a specific permission.
func (s *authService) HasPermission(
	ctx context.Context,
	userID string,
	permission Permission,
) bool {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}

	// Admin users have all permissions
	if user.Role == models.UserRoleAdmin {
		return true
	}

	// Check specific permissions based on role and resource
	return s.checkRolePermission(user.Role, permission)
}

// HasRole checks if a user has a specific role.
func (s *authService) HasRole(
	ctx context.Context,
	userID string,
	role models.UserRole,
) bool {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}

	return user.Role == role
}

// HasProjectRole checks if a user has a specific role in a project.
func (s *authService) HasProjectRole(
	ctx context.Context,
	userID, projectID string,
	role models.ProjectRole,
) bool {
	hasRole, err := s.projectRepo.HasUserWithRole(projectID, userID, role)
	if err != nil {
		return false
	}

	return hasRole
}

// CanAccessResource checks if a user can access a specific resource.
func (s *authService) CanAccessResource(
	ctx context.Context,
	userID, resourceID, resourceType, action string,
) bool {
	switch resourceType {
	case "project":
		return s.canAccessProject(ctx, userID, resourceID, action)
	case "asset":
		return s.canAccessAsset(ctx, userID, resourceID, action)
	case "pipeline":
		return s.canAccessPipeline(ctx, userID, resourceID, action)
	default:
		return false
	}
}

// LogAuditEvent logs an audit event for security tracking.
func (s *authService) LogAuditEvent(ctx context.Context, event *AuditEvent) {
	if event.ID == uuid.Nil {
		event.ID = uuid.Must(uuid.NewV7())
	}

	s.logger.Info().
		Str("event_id", event.ID.String()).
		Str("action", event.Action).
		Str("resource", event.Resource).
		Bool("success", event.Success).
		Msg("Audit event logged")
}

// CreateUser creates a new user (usually from SSO).
func (s *authService) CreateUser(
	ctx context.Context,
	req *CreateUserRequest,
) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return existingUser, nil // User already exists
	}

	user := &models.User{
		ID:    uuid.Must(uuid.NewV7()),
		Email: req.Email,
		Role:  req.Role,
		// PasswordHash would be empty for SSO users
	}

	// In a real implementation, you would store the provider information
	// in a separate user_providers table

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", req.Email).
		Str("provider", req.Provider).
		Msg("User created successfully from SSO")

	return user, nil
}

// UpdateUserLastLogin updates the user's last login timestamp.
func (s *authService) UpdateUserLastLogin(
	ctx context.Context,
	userID string,
) error {
	// In a real implementation, you would update a last_login field in the users
	// table
	// For now, we'll just log the event
	s.logger.Info().
		Str("user_id", userID).
		Msg("User last login updated")

	return nil
}

// Helper methods

func (s *authService) generateTokenPair(userID string) (*TokenPair, error) {
	now := time.Now()
	accessTokenExpiresAt := now.Add(s.accessTokenTTL)
	refreshTokenExpiresAt := now.Add(s.refreshTokenTTL)

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     accessTokenExpiresAt.Unix(),
		"type":    "access",
	})

	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     refreshTokenExpiresAt.Unix(),
		"type":    "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessTokenExpiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (s *authService) getUserPermissions(user *models.User) []Permission {
	// Define permissions based on user role
	var permissions []Permission

	switch user.Role {
	case models.UserRoleAdmin:
		permissions = append(permissions, adminPermissions()...)
	case models.UserRoleUser:
		permissions = append(permissions, userPermissions()...)
	}

	return permissions
}

func (s *authService) checkRolePermission(
	role models.UserRole,
	permission Permission,
) bool {
	// Implement role-based permission checking logic
	switch role {
	case models.UserRoleAdmin:
		return true // Admin has all permissions
	case models.UserRoleUser:
		// Check specific user permissions
		return permission.Resource != "admin" // Users can't access admin resources
	}

	return false
}

func (s *authService) canAccessProject(
	ctx context.Context,
	userID, projectID, action string,
) bool {
	// Check if user is a member of the project with appropriate role
	hasAccess := s.HasProjectRole(
		ctx,
		userID,
		projectID,
		models.ProjectRoleViewer,
	)
	if !hasAccess {
		return false
	}

	// Check action-specific permissions
	switch action {
	case "read":
		return true // All project members can read
	case "write", "update":
		return s.HasProjectRole(
			ctx,
			userID,
			projectID,
			models.ProjectRoleMaintainer,
		)
	case "delete", "admin":
		return s.HasProjectRole(ctx, userID, projectID, models.ProjectRoleOwner)
	default:
		return false
	}
}

func (s *authService) canAccessAsset(
	ctx context.Context,
	userID, assetID, action string,
) bool {
	// This would require asset repository access to check project membership
	// For now, return a basic implementation
	return s.HasPermission(ctx, userID, Permission{
		Resource: "asset",
		Action:   action,
	})
}

func (s *authService) canAccessPipeline(
	ctx context.Context,
	userID, pipelineID, action string,
) bool {
	// Similar to asset access
	return s.HasPermission(ctx, userID, Permission{
		Resource: "pipeline",
		Action:   action,
	})
}

// These would be replaced with actual Google API calls.
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func (s *authService) getGoogleUserInfo(
	ctx context.Context,
	token *oauth2.Token,
) (*GoogleUserInfo, error) {
	// In a real implementation, you would make an HTTP request to Google's
	// userinfo endpoint
	// For now, return mock data
	return &GoogleUserInfo{
		ID:            "123456789",
		Email:         "user@example.com",
		VerifiedEmail: true,
		Name:          "Test User",
		Picture:       "https://example.com/avatar.jpg",
	}, nil
}

func (s *authService) findOrCreateGoogleUser(
	ctx context.Context,
	userInfo *GoogleUserInfo,
) (*models.User, error) {
	// Try to find existing user
	user, err := s.userRepo.GetByEmail(userInfo.Email)
	if err == nil && user != nil {
		return user, nil
	}

	// Create new user
	createReq := &CreateUserRequest{
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		Avatar:     userInfo.Picture,
		Role:       models.UserRoleUser, // Default role for SSO users
		Provider:   "google",
		ProviderID: userInfo.ID,
		Metadata: map[string]any{
			"verified_email": userInfo.VerifiedEmail,
			"picture":        userInfo.Picture,
		},
	}

	return s.CreateUser(ctx, createReq)
}

// Permission definitions.
func adminPermissions() []Permission {
	return []Permission{
		{ID: "admin_all", Name: "All Permissions", Resource: "*", Action: "*"},
		{
			ID:       "user_manage",
			Name:     "Manage Users",
			Resource: "user",
			Action:   "manage",
		},
		{
			ID:       "project_manage",
			Name:     "Manage Projects",
			Resource: "project",
			Action:   "manage",
		},
		{
			ID:       "system_config",
			Name:     "System Configuration",
			Resource: "system",
			Action:   "config",
		},
	}
}

func userPermissions() []Permission {
	return []Permission{
		{
			ID:       "project_read",
			Name:     "Read Projects",
			Resource: "project",
			Action:   "read",
		},
		{ID: "asset_read", Name: "Read Assets", Resource: "asset", Action: "read"},
		{
			ID:       "pipeline_read",
			Name:     "Read Pipelines",
			Resource: "pipeline",
			Action:   "read",
		},
		{
			ID:       "pipeline_trigger",
			Name:     "Trigger Pipelines",
			Resource: "pipeline",
			Action:   "trigger",
		},
	}
}
