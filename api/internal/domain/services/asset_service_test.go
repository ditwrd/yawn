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

// Package services provides business logic layer for the application.
package services

import (
	"context"
	"errors"
	"testing"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAssetRepository is a mock implementation of repositories.AssetRepository.
type MockAssetRepository struct {
	mock.Mock
}

// Ensure MockAssetRepository implements the interface.
var _ repositories.AssetRepository = (*MockAssetRepository)(nil)

func (m *MockAssetRepository) Create(
	ctx context.Context,
	asset *models.Asset,
) error {
	args := m.Called(ctx, asset)

	return args.Error(0)
}

func (m *MockAssetRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) GetByProjectID(
	ctx context.Context,
	projectID string,
	limit, offset int,
) ([]*models.Asset, error) {
	args := m.Called(ctx, projectID, limit, offset)

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) GetByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*models.Asset, error) {
	args := m.Called(ctx, repositoryID)

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) List(
	ctx context.Context,
	limit, offset int,
	filters repositories.AssetFilters,
) ([]*models.Asset, error) {
	args := m.Called(ctx, limit, offset, filters)

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) Update(
	ctx context.Context,
	asset *models.Asset,
) error {
	args := m.Called(ctx, asset)

	return args.Error(0)
}

func (m *MockAssetRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)

	return args.Error(0)
}

func (m *MockAssetRepository) Search(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]*models.Asset, error) {
	args := m.Called(ctx, query, limit, offset)

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) GetVersionHistory(
	ctx context.Context,
	projectID, assetName string,
) ([]*models.Asset, error) {
	args := m.Called(ctx, projectID, assetName)

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) GetLatestVersion(
	ctx context.Context,
	projectID, assetName string,
) (*models.Asset, error) {
	args := m.Called(ctx, projectID, assetName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetRepository) Count(
	ctx context.Context,
	filters repositories.AssetFilters,
) (int64, error) {
	args := m.Called(ctx, filters)

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAssetRepository) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	args := m.Called(ctx, id)

	return args.Bool(0), args.Error(1)
}

func (m *MockAssetRepository) ExistsByName(
	ctx context.Context,
	projectID, name string,
) (bool, error) {
	args := m.Called(ctx, projectID, name)

	return args.Bool(0), args.Error(1)
}

// MockProjectRepository for testing asset service.
type MockProjectRepository struct {
	mock.Mock
}

// Ensure MockProjectRepository implements the interface.
var _ repositories.ProjectRepository = (*MockProjectRepository)(nil)

func (m *MockProjectRepository) Create(project *models.Project) error {
	args := m.Called(project)

	return args.Error(0)
}

func (m *MockProjectRepository) GetByID(id string) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByIDWithMembers(
	id string,
) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByUserID(
	userID string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(userID, limit, offset)

	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByOwnerID(
	ownerID string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(ownerID, limit, offset)

	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepository) List(
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(limit, offset)

	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(project *models.Project) error {
	args := m.Called(project)

	return args.Error(0)
}

func (m *MockProjectRepository) Delete(id string) error {
	args := m.Called(id)

	return args.Error(0)
}

func (m *MockProjectRepository) AddMember(
	projectID, userID string,
	role models.ProjectRole,
) error {
	args := m.Called(projectID, userID, role)

	return args.Error(0)
}

func (m *MockProjectRepository) RemoveMember(projectID, userID string) error {
	args := m.Called(projectID, userID)

	return args.Error(0)
}

func (m *MockProjectRepository) UpdateMemberRole(
	projectID, userID string,
	role models.ProjectRole,
) error {
	args := m.Called(projectID, userID, role)

	return args.Error(0)
}

func (m *MockProjectRepository) GetMember(
	projectID, userID string,
) (*models.ProjectUser, error) {
	args := m.Called(projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.ProjectUser), args.Error(1)
}

func (m *MockProjectRepository) ListMembers(
	projectID string,
) ([]models.ProjectUser, error) {
	args := m.Called(projectID)

	return args.Get(0).([]models.ProjectUser), args.Error(1)
}

func (m *MockProjectRepository) GetUserRole(
	projectID, userID string,
) (models.ProjectRole, error) {
	args := m.Called(projectID, userID)

	return args.Get(0).(models.ProjectRole), args.Error(1)
}

func (m *MockProjectRepository) Search(
	query string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(query, limit, offset)

	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepository) Count() (int64, error) {
	args := m.Called()

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepository) CountByUserID(userID string) (int64, error) {
	args := m.Called(userID)

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepository) Exists(id string) (bool, error) {
	args := m.Called(id)

	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepository) HasUserWithRole(
	projectID, userID string,
	role models.ProjectRole,
) (bool, error) {
	args := m.Called(projectID, userID, role)

	return args.Bool(0), args.Error(1)
}

// MockUserRepository for testing asset service.
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements the interface.
var _ repositories.UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)

	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)

	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)

	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]models.User, error) {
	args := m.Called(limit, offset)

	return args.Get(0).([]models.User), args.Error(1)
}

// createTestLogger creates a test logger instance.
func createTestLogger() *zerolog.Logger {
	logger := zerolog.Nop()

	return &logger
}

func TestNewAssetService(t *testing.T) {
	t.Parallel()

	assetRepo := &MockAssetRepository{}
	projectRepo := &MockProjectRepository{}
	userRepo := &MockUserRepository{}
	logger := createTestLogger()

	service := NewAssetService(assetRepo, projectRepo, userRepo, logger)

	assert.NotNil(t, service)
	assert.IsType(t, &assetService{}, service)
}

func TestAssetService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectID := uuid.Must(uuid.NewV7())
	repoID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetRepo   *MockAssetRepository
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}

	type args struct {
		req *CreateAssetRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful asset creation",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("ExistsByName", ctx, projectID.String(), "test-asset").
						Return(false, nil)
					m.On("Create", ctx, mock.AnythingOfType("*models.Asset")).Return(nil)

					return m
				}(),
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("Exists", projectID.String()).Return(true, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:        "test-asset",
					Description: "Test asset description",
					Version:     "1.0.0",
					ProjectID:   projectID.String(),
				},
			},
			wantErr: false,
		},
		{
			name: "asset with repository",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("ExistsByName", ctx, projectID.String(), "test-asset").
						Return(false, nil)
					m.On("Create", ctx, mock.AnythingOfType("*models.Asset")).Return(nil)

					return m
				}(),
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("Exists", projectID.String()).Return(true, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:         "test-asset",
					Description:  "Test asset description",
					Version:      "1.0.0",
					ProjectID:    projectID.String(),
					RepositoryID: &[]string{repoID.String()}[0],
				},
			},
			wantErr: false,
		},
		{
			name: "nil request",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				req: nil,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:      "",
					Version:   "1.0.0",
					ProjectID: projectID.String(),
				},
			},
			wantErr: true,
		},
		{
			name: "empty version",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:      "test-asset",
					Version:   "",
					ProjectID: projectID.String(),
				},
			},
			wantErr: true,
		},
		{
			name: "empty project ID",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:    "test-asset",
					Version: "1.0.0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid project ID",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:      "test-asset",
					Version:   "1.0.0",
					ProjectID: "invalid-uuid",
				},
			},
			wantErr: true,
		},
		{
			name: "asset already exists",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("ExistsByName", ctx, projectID.String(), "test-asset").
						Return(true, nil)

					return m
				}(),
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("Exists", projectID.String()).Return(true, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:      "test-asset",
					Version:   "1.0.0",
					ProjectID: projectID.String(),
				},
			},
			wantErr: true,
		},
		{
			name: "project not found",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}

					return m
				}(),
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("Exists", projectID.String()).Return(false, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:      "test-asset",
					Version:   "1.0.0",
					ProjectID: projectID.String(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid asset name format",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				req: &CreateAssetRequest{
					Name:      "test asset with spaces",
					Version:   "1.0.0",
					ProjectID: projectID.String(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &assetService{
				assetRepo:   tt.fields.assetRepo,
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
				logger:      createTestLogger(),
			}

			result, err := s.Create(ctx, tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.args.req.Name, result.Name)
				assert.Equal(t, tt.args.req.Version, result.Version)
				assert.Equal(t, projectID, result.ProjectID)
			}

			tt.fields.assetRepo.AssertExpectations(t)
			tt.fields.projectRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetRepo   *MockAssetRepository
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}

	type args struct {
		id string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful asset retrieval",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					asset := &models.Asset{
						ID:      assetID,
						Name:    "test-asset",
						Version: "1.0.0",
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
			},
			wantErr: false,
		},
		{
			name: "empty asset ID",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: "",
			},
			wantErr: true,
		},
		{
			name: "asset not found",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("GetByID", ctx, assetID.String()).
						Return(nil, errors.New("not found"))

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &assetService{
				assetRepo:   tt.fields.assetRepo,
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
				logger:      createTestLogger(),
			}

			result, err := s.GetByID(ctx, tt.args.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, assetID, result.ID)
			}

			tt.fields.assetRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())
	repoID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetRepo   *MockAssetRepository
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}

	type args struct {
		id  string
		req *UpdateAssetRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful asset update",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					asset := &models.Asset{
						ID:      assetID,
						Name:    "test-asset",
						Version: "1.0.0",
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)
					m.On("Update", ctx, mock.AnythingOfType("*models.Asset")).Return(nil)

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
				req: &UpdateAssetRequest{
					Name: func() *string {
						s := "updated-asset"
						return &s
					}(),
				},
			},
			wantErr: false,
		},
		{
			name: "update with repository",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					asset := &models.Asset{
						ID:      assetID,
						Name:    "test-asset",
						Version: "1.0.0",
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)
					m.On("Update", ctx, mock.AnythingOfType("*models.Asset")).Return(nil)

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
				req: &UpdateAssetRequest{
					RepositoryID: &[]string{repoID.String()}[0],
				},
			},
			wantErr: false,
		},
		{
			name: "remove repository",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					asset := &models.Asset{
						ID:           assetID,
						Name:         "test-asset",
						Version:      "1.0.0",
						RepositoryID: &repoID,
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)
					m.On("Update", ctx, mock.AnythingOfType("*models.Asset")).Return(nil)

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
				req: &UpdateAssetRequest{
					RepositoryID: &[]string{""}[0],
				},
			},
			wantErr: false,
		},
		{
			name: "empty asset ID",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id:  "",
				req: &UpdateAssetRequest{},
			},
			wantErr: true,
		},
		{
			name: "nil request",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id:  assetID.String(),
				req: nil,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
				req: &UpdateAssetRequest{
					Name: func() *string {
						s := ""
						return &s
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid name format",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
				req: &UpdateAssetRequest{
					Name: func() *string {
						s := "invalid name"
						return &s
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid repository ID",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
				req: &UpdateAssetRequest{
					RepositoryID: &[]string{"invalid-uuid"}[0],
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &assetService{
				assetRepo:   tt.fields.assetRepo,
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
				logger:      createTestLogger(),
			}

			result, err := s.Update(ctx, tt.args.id, tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			tt.fields.assetRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetRepo   *MockAssetRepository
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}

	type args struct {
		id string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful asset deletion",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("Exists", ctx, assetID.String()).Return(true, nil)
					m.On("Delete", ctx, assetID.String()).Return(nil)

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
			},
			wantErr: false,
		},
		{
			name: "empty asset ID",
			fields: fields{
				assetRepo:   &MockAssetRepository{},
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: "",
			},
			wantErr: true,
		},
		{
			name: "asset not found",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("Exists", ctx, assetID.String()).Return(false, nil)

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
			},
			wantErr: true,
		},
		{
			name: "delete failed",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("Exists", ctx, assetID.String()).Return(true, nil)
					m.On("Delete", ctx, assetID.String()).
						Return(errors.New("delete failed"))

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id: assetID.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &assetService{
				assetRepo:   tt.fields.assetRepo,
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
				logger:      createTestLogger(),
			}

			err := s.Delete(ctx, tt.args.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.fields.assetRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_ValidateAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetRepo   *MockAssetRepository
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}

	type args struct {
		assetID      string
		userID       string
		requiredRole models.ProjectRole
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "access granted - maintainer role",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					asset := &models.Asset{
						ID:        assetID,
						ProjectID: uuid.Must(uuid.NewV7()),
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)

					return m
				}(),
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("HasUserWithRole", mock.AnythingOfType("string"), userID.String(), models.ProjectRoleMaintainer).
						Return(true, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				assetID:      assetID.String(),
				userID:       userID.String(),
				requiredRole: models.ProjectRoleMaintainer,
			},
			wantErr: false,
		},
		{
			name: "access denied - insufficient role",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					asset := &models.Asset{
						ID:        assetID,
						ProjectID: uuid.Must(uuid.NewV7()),
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)

					return m
				}(),
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("HasUserWithRole", mock.AnythingOfType("string"), userID.String(), models.ProjectRoleMaintainer).
						Return(false, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				assetID:      assetID.String(),
				userID:       userID.String(),
				requiredRole: models.ProjectRoleMaintainer,
			},
			wantErr: true,
		},
		{
			name: "asset not found",
			fields: fields{
				assetRepo: func() *MockAssetRepository {
					m := &MockAssetRepository{}
					m.On("GetByID", ctx, assetID.String()).
						Return(nil, errors.New("not found"))

					return m
				}(),
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				assetID:      assetID.String(),
				userID:       userID.String(),
				requiredRole: models.ProjectRoleMaintainer,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &assetService{
				assetRepo:   tt.fields.assetRepo,
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
				logger:      createTestLogger(),
			}

			err := s.ValidateAccess(
				ctx,
				tt.args.assetID,
				tt.args.userID,
				tt.args.requiredRole,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.fields.assetRepo.AssertExpectations(t)
			tt.fields.projectRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_CanCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetRepo   *MockAssetRepository
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}

	type args struct {
		projectID string
		userID    string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "can create - maintainer role",
			fields: fields{
				assetRepo: &MockAssetRepository{},
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("HasUserWithRole", projectID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(true, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				projectID: projectID.String(),
				userID:    userID.String(),
			},
			wantErr: false,
		},
		{
			name: "cannot create - viewer role",
			fields: fields{
				assetRepo: &MockAssetRepository{},
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("HasUserWithRole", projectID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(false, nil)

					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				projectID: projectID.String(),
				userID:    userID.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &assetService{
				assetRepo:   tt.fields.assetRepo,
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
				logger:      createTestLogger(),
			}

			err := s.CanCreate(ctx, tt.args.projectID, tt.args.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.fields.projectRepo.AssertExpectations(t)
		})
	}
}
