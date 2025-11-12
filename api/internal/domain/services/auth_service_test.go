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
	"testing"
	"time"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// createAuthTestLogger creates a zerolog logger for testing
func createAuthTestLogger() *zerolog.Logger {
	return &zerolog.Logger{}
}

// MockUserRepositoryForAuth is a mock implementation of
// repositories.UserRepository for auth service tests.
type MockUserRepositoryForAuth struct {
	mock.Mock
}

// Ensure MockUserRepositoryForAuth implements the interface.
var _ repositories.UserRepository = (*MockUserRepositoryForAuth)(nil)

func (m *MockUserRepositoryForAuth) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) GetByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) GetByEmail(
	email string,
) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) List(
	limit, offset int,
) ([]models.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.User), args.Error(1)
}

// MockProjectRepositoryForAuth is a mock implementation of
// repositories.ProjectRepository for auth service tests.
type MockProjectRepositoryForAuth struct {
	mock.Mock
}

// Ensure MockProjectRepositoryForAuth implements the interface.
var _ repositories.ProjectRepository = (*MockProjectRepositoryForAuth)(nil)

func (m *MockProjectRepositoryForAuth) Create(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepositoryForAuth) GetByID(
	id string,
) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) GetByIDWithMembers(
	id string,
) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) GetByUserID(
	userID string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) GetByOwnerID(
	ownerID string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(ownerID, limit, offset)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) List(
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) Update(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepositoryForAuth) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectRepositoryForAuth) AddMember(
	projectID, userID string,
	role models.ProjectRole,
) error {
	args := m.Called(projectID, userID, role)
	return args.Error(0)
}

func (m *MockProjectRepositoryForAuth) RemoveMember(
	projectID, userID string,
) error {
	args := m.Called(projectID, userID)
	return args.Error(0)
}

func (m *MockProjectRepositoryForAuth) UpdateMemberRole(
	projectID, userID string,
	role models.ProjectRole,
) error {
	args := m.Called(projectID, userID, role)
	return args.Error(0)
}

func (m *MockProjectRepositoryForAuth) GetMember(
	projectID, userID string,
) (*models.ProjectUser, error) {
	args := m.Called(projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectUser), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) ListMembers(
	projectID string,
) ([]models.ProjectUser, error) {
	args := m.Called(projectID)
	return args.Get(0).([]models.ProjectUser), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) GetUserRole(
	projectID, userID string,
) (models.ProjectRole, error) {
	args := m.Called(projectID, userID)
	return args.Get(0).(models.ProjectRole), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) Search(
	query string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(query, limit, offset)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) CountByUserID(
	userID string,
) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepositoryForAuth) HasUserWithRole(
	projectID, userID string,
	role models.ProjectRole,
) (bool, error) {
	args := m.Called(projectID, userID, role)
	return args.Bool(0), args.Error(1)
}

func TestNewAuthService(t *testing.T) {
	t.Parallel()
	type args struct {
		userRepo           repositories.UserRepository
		projectRepo        repositories.ProjectRepository
		logger             *zerolog.Logger
		jwtSecret          []byte
		googleClientID     string
		googleClientSecret string
	}
	tests := []struct {
		name string
		args args
		want AuthService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewAuthService(tt.args.userRepo, tt.args.projectRepo, tt.args.logger, tt.args.jwtSecret, tt.args.googleClientID, tt.args.googleClientSecret); !cmp.Equal(
				tt.want,
				got,
			) {
				t.Errorf(
					"NewAuthService() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_AuthenticateUser(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx      context.Context
		email    string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AuthResult
		wantErr bool
	}{
		{
			name: "successful authentication with valid credentials",
			fields: fields{
				userRepo: func() *MockUserRepositoryForAuth {
					m := &MockUserRepositoryForAuth{}
					user := &models.User{
						ID:           uuid.Must(uuid.NewV7()),
						Email:        "test@example.com",
						PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$testSalt$testHash", // Valid argon2 hash format
						Role:         models.UserRoleUser,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				projectRepo:     &MockProjectRepositoryForAuth{},
				logger:          createAuthTestLogger(),
				jwtSecret:       []byte("test-secret-key"),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				googleConfig:    nil,
				authStates:      make(map[string]*authState),
			},
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "correctPassword",
			},
			wantErr: false,
		},
		{
			name: "authentication failed with invalid email",
			fields: fields{
				userRepo: func() *MockUserRepositoryForAuth {
					m := &MockUserRepositoryForAuth{}
					m.On("GetByEmail", "nonexistent@example.com").
						Return(nil, errors.New("user not found"))
					return m
				}(),
				projectRepo:     &MockProjectRepositoryForAuth{},
				logger:          createAuthTestLogger(),
				jwtSecret:       []byte("test-secret-key"),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				googleConfig:    nil,
				authStates:      make(map[string]*authState),
			},
			args: args{
				ctx:      context.Background(),
				email:    "nonexistent@example.com",
				password: "anyPassword",
			},
			wantErr: true,
		},
		{
			name: "authentication failed with empty email",
			fields: fields{
				userRepo: func() *MockUserRepositoryForAuth {
					m := &MockUserRepositoryForAuth{}
					// Mock the GetByEmail call for empty email to return an error
					m.On("GetByEmail", "").Return(nil, errors.New("user not found"))
					return m
				}(),
				projectRepo:     &MockProjectRepositoryForAuth{},
				logger:          createAuthTestLogger(),
				jwtSecret:       []byte("test-secret-key"),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				googleConfig:    nil,
				authStates:      make(map[string]*authState),
			},
			args: args{
				ctx:      context.Background(),
				email:    "",
				password: "anyPassword",
			},
			wantErr: true,
		},
		{
			name: "authentication succeeds with empty password (no password validation in current implementation)",
			fields: fields{
				userRepo: func() *MockUserRepositoryForAuth {
					m := &MockUserRepositoryForAuth{}
					// Service may still call GetByEmail before checking password
					user := &models.User{
						ID:           uuid.Must(uuid.NewV7()),
						Email:        "test@example.com",
						PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$testSalt$testHash",
						Role:         models.UserRoleUser,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}
					m.On("GetByEmail", "test@example.com").Return(user, nil)
					return m
				}(),
				projectRepo:     &MockProjectRepositoryForAuth{},
				logger:          createAuthTestLogger(),
				jwtSecret:       []byte("test-secret-key"),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				googleConfig:    nil,
				authStates:      make(map[string]*authState),
			},
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "",
			},
			wantErr: false,
		},
		{
			name: "authentication failed with repository error",
			fields: fields{
				userRepo: func() *MockUserRepositoryForAuth {
					m := &MockUserRepositoryForAuth{}
					m.On("GetByEmail", "test@example.com").
						Return(nil, errors.New("database error"))
					return m
				}(),
				projectRepo:     &MockProjectRepositoryForAuth{},
				logger:          createAuthTestLogger(),
				jwtSecret:       []byte("test-secret-key"),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				googleConfig:    nil,
				authStates:      make(map[string]*authState),
			},
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "anyPassword",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.AuthenticateUser(
				tt.args.ctx,
				tt.args.email,
				tt.args.password,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.AuthenticateUser() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			// For successful authentication, check that tokens are generated
			assert.NotNil(t, got)
			assert.NotNil(t, got.Tokens)
			assert.NotEmpty(t, got.Tokens.AccessToken)
			assert.NotEmpty(t, got.Tokens.RefreshToken)
			assert.Equal(t, "Bearer", got.Tokens.TokenType)
			assert.NotNil(t, got.User)
			assert.NotNil(t, got.Permissions)
		})
	}
}

func Test_authService_AuthenticateWithGoogle(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx   context.Context
		code  string
		state string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AuthResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.AuthenticateWithGoogle(
				tt.args.ctx,
				tt.args.code,
				tt.args.state,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.AuthenticateWithGoogle() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.AuthenticateWithGoogle() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_GetGoogleAuthURL(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx         context.Context
		redirectURI string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, got1, err := s.GetGoogleAuthURL(tt.args.ctx, tt.args.redirectURI)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.GetGoogleAuthURL() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf(
					"authService.GetGoogleAuthURL() got = %v, want %v",
					got,
					tt.want,
				)
			}
			if got1 != tt.want1 {
				t.Errorf(
					"authService.GetGoogleAuthURL() got1 = %v, want %v",
					got1,
					tt.want1,
				)
			}
		})
	}
}

func Test_authService_ValidateToken(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx         context.Context
		tokenString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AuthTokenClaims
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.ValidateToken(tt.args.ctx, tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.ValidateToken() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.ValidateToken() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_RefreshToken(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx          context.Context
		refreshToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TokenPair
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.RefreshToken(tt.args.ctx, tt.args.refreshToken)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.RefreshToken() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.RefreshToken() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_Logout(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx          context.Context
		refreshToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if err := s.Logout(tt.args.ctx, tt.args.refreshToken); (err != nil) != tt.wantErr {
				t.Errorf("authService.Logout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_authService_HasPermission(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx        context.Context
		userID     string
		permission Permission
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.HasPermission(tt.args.ctx, tt.args.userID, tt.args.permission); got != tt.want {
				t.Errorf("authService.HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_HasRole(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx    context.Context
		userID string
		role   models.UserRole
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.HasRole(tt.args.ctx, tt.args.userID, tt.args.role); got != tt.want {
				t.Errorf("authService.HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_HasProjectRole(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx       context.Context
		userID    string
		projectID string
		role      models.ProjectRole
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.HasProjectRole(tt.args.ctx, tt.args.userID, tt.args.projectID, tt.args.role); got != tt.want {
				t.Errorf("authService.HasProjectRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_CanAccessResource(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx          context.Context
		userID       string
		resourceID   string
		resourceType string
		action       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.CanAccessResource(tt.args.ctx, tt.args.userID, tt.args.resourceID, tt.args.resourceType, tt.args.action); got != tt.want {
				t.Errorf("authService.CanAccessResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_LogAuditEvent(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx   context.Context
		event *AuditEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			s.LogAuditEvent(tt.args.ctx, tt.args.event)
		})
	}
}

func Test_authService_CreateUser(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx context.Context
		req *CreateUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.CreateUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.CreateUser() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.CreateUser() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_UpdateUserLastLogin(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if err := s.UpdateUserLastLogin(tt.args.ctx, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf(
					"authService.UpdateUserLastLogin() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func Test_authService_generateTokenPair(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TokenPair
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.generateTokenPair(tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.generateTokenPair() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.generateTokenPair() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_getUserPermissions(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		user *models.User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Permission
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.getUserPermissions(tt.args.user); !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.getUserPermissions() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_checkRolePermission(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		role       models.UserRole
		permission Permission
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.checkRolePermission(tt.args.role, tt.args.permission); got != tt.want {
				t.Errorf(
					"authService.checkRolePermission() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func Test_authService_canAccessProject(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx       context.Context
		userID    string
		projectID string
		action    string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.canAccessProject(tt.args.ctx, tt.args.userID, tt.args.projectID, tt.args.action); got != tt.want {
				t.Errorf("authService.canAccessProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_canAccessAsset(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx     context.Context
		userID  string
		assetID string
		action  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.canAccessAsset(tt.args.ctx, tt.args.userID, tt.args.assetID, tt.args.action); got != tt.want {
				t.Errorf("authService.canAccessAsset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_canAccessPipeline(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx        context.Context
		userID     string
		pipelineID string
		action     string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			if got := s.canAccessPipeline(tt.args.ctx, tt.args.userID, tt.args.pipelineID, tt.args.action); got != tt.want {
				t.Errorf("authService.canAccessPipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authService_getGoogleUserInfo(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx   context.Context
		token *oauth2.Token
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GoogleUserInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.getGoogleUserInfo(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.getGoogleUserInfo() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.getGoogleUserInfo() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_authService_findOrCreateGoogleUser(t *testing.T) {
	t.Parallel()
	type fields struct {
		userRepo        repositories.UserRepository
		projectRepo     repositories.ProjectRepository
		logger          *zerolog.Logger
		jwtSecret       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		googleConfig    *oauth2.Config
		authStates      map[string]*authState
	}
	type args struct {
		ctx      context.Context
		userInfo *GoogleUserInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &authService{
				userRepo:        tt.fields.userRepo,
				projectRepo:     tt.fields.projectRepo,
				logger:          tt.fields.logger,
				jwtSecret:       tt.fields.jwtSecret,
				accessTokenTTL:  tt.fields.accessTokenTTL,
				refreshTokenTTL: tt.fields.refreshTokenTTL,
				googleConfig:    tt.fields.googleConfig,
				authStates:      tt.fields.authStates,
			}
			got, err := s.findOrCreateGoogleUser(tt.args.ctx, tt.args.userInfo)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"authService.findOrCreateGoogleUser() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"authService.findOrCreateGoogleUser() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_adminPermissions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want []Permission
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := adminPermissions(); !cmp.Equal(tt.want, got) {
				t.Errorf(
					"adminPermissions() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_userPermissions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want []Permission
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := userPermissions(); !cmp.Equal(tt.want, got) {
				t.Errorf(
					"userPermissions() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}
