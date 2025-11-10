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
	"errors"
	"testing"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepository is a mock implementation of
// repositories.ProjectRepository
type MockProjectRepository struct {
	mock.Mock
}

// Ensure MockProjectRepository implements the interface
var _ repositories.ProjectRepository = (*MockProjectRepository)(nil)

// MockUserRepository is a mock implementation of repositories.UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements the interface
var _ repositories.UserRepository = (*MockUserRepository)(nil)

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

// UserRepository mock methods
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

func TestNewProjectService(t *testing.T) {
	t.Parallel()

	projectRepo := &MockProjectRepository{}
	userRepo := &MockUserRepository{}

	service := NewProjectService(projectRepo, userRepo)

	assert.NotNil(t, service)
	assert.IsType(t, &projectService{}, service)
}

func TestProjectService_Create(t *testing.T) {
	t.Parallel()

	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}
	type args struct {
		project *models.Project
		ownerID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful project creation",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("Create", mock.AnythingOfType("*models.Project")).Return(nil)
					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				project: &models.Project{
					Name:        "Test Project",
					Description: "A test project",
					Visibility:  "private",
				},
				ownerID: userID.String(),
			},
			wantErr: false,
		},
		{
			name: "nil project",
			fields: fields{
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				project: nil,
				ownerID: userID.String(),
			},
			wantErr: true,
		},
		{
			name: "empty owner ID",
			fields: fields{
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				project: &models.Project{Name: "Test"},
				ownerID: "",
			},
			wantErr: true,
		},
		{
			name: "empty project name",
			fields: fields{
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				project: &models.Project{Name: ""},
				ownerID: userID.String(),
			},
			wantErr: true,
		},
		{
			name: "invalid visibility",
			fields: fields{
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				project: &models.Project{
					Name:       "Test Project",
					Visibility: "invalid",
				},
				ownerID: userID.String(),
			},
			wantErr: true,
		},
		{
			name: "repository error",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("Create", mock.AnythingOfType("*models.Project")).
						Return(errors.New("db error"))
					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				project: &models.Project{Name: "Test Project"},
				ownerID: userID.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &projectService{
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
			}

			err := s.Create(tt.args.project, tt.args.ownerID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify that the project was properly configured
				assert.Equal(t, userID, tt.args.project.OwnerID)
			}

			tt.fields.projectRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_GetByID(t *testing.T) {
	t.Parallel()

	projectID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}
	type args struct {
		id     string
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful project retrieval",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					project := &models.Project{
						ID:    projectID,
						Name:  "Test Project",
						Owner: models.User{ID: userID},
						Users: []models.ProjectUser{},
					}
					m.On("GetByIDWithMembers", projectID.String()).Return(project, nil)
					m.On("GetUserRole", projectID.String(), userID.String()).
						Return(models.ProjectRoleOwner, nil)
					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				id:     projectID.String(),
				userID: userID.String(),
			},
			wantErr: false,
		},
		{
			name: "empty project ID",
			fields: fields{
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				id:     "",
				userID: userID.String(),
			},
			wantErr: true,
		},
		{
			name: "project not found",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("GetByIDWithMembers", projectID.String()).
						Return(nil, errors.New("not found"))
					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				id:     projectID.String(),
				userID: userID.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &projectService{
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
			}

			result, err := s.GetByID(tt.args.id, tt.args.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Test Project", result.Name)
			}

			tt.fields.projectRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_AddMember(t *testing.T) {
	t.Parallel()

	// Use fixed UUIDs for consistent testing
	projectID := uuid.Must(
		uuid.FromString("019a6cc1-4fed-7fbe-b2c8-adf780550ad7"),
	)
	userID := uuid.Must(uuid.FromString("019a6cc1-4fed-7fbf-ab34-08582c3d3aba"))
	memberID := uuid.Must(uuid.FromString("019a6cc1-4fed-7fc0-8d8d-40bcc19e1e06"))

	type fields struct {
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}
	type args struct {
		projectID   string
		userID      string
		memberEmail string
		role        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful member addition",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					project := &models.Project{
						ID:      projectID,
						OwnerID: userID,
					}
					m.On("GetByID", projectID.String()).Return(project, nil)
					m.On("GetUserRole", projectID.String(), userID.String()).
						Return(models.ProjectRoleOwner, nil).
						Maybe()
					// GetMember before adding returns not found
					m.On("GetMember", projectID.String(), memberID.String()).
						Return(nil, errors.New("not found")).
						Once()
					m.On("AddMember", projectID.String(), memberID.String(), models.ProjectRoleViewer).
						Return(nil)
					// GetMember after adding returns the created member
					createdMember := &models.ProjectUser{
						ID:      uuid.Must(uuid.NewV7()),
						Role:    models.ProjectRoleViewer,
						User:    models.User{ID: memberID, Email: "member@example.com"},
						Project: models.Project{ID: projectID},
					}
					m.On("GetMember", projectID.String(), memberID.String()).
						Return(createdMember, nil)
					return m
				}(),
				userRepo: func() *MockUserRepository {
					m := &MockUserRepository{}
					m.On("GetByEmail", "member@example.com").Return(&models.User{
						ID:    memberID,
						Email: "member@example.com",
					}, nil)
					return m
				}(),
			},
			args: args{
				projectID:   projectID.String(),
				userID:      userID.String(),
				memberEmail: "member@example.com",
				role:        "viewer",
			},
			wantErr: false,
		},
		{
			name: "invalid role",
			fields: fields{
				projectRepo: &MockProjectRepository{},
				userRepo:    &MockUserRepository{},
			},
			args: args{
				projectID:   projectID.String(),
				userID:      userID.String(),
				memberEmail: "member@example.com",
				role:        "invalid",
			},
			wantErr: true,
		},
		{
			name: "user not found",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					// Mock GetByID for CheckAccess method
					project := &models.Project{
						ID:   projectID,
						Name: "Test Project",
					}
					m.On("GetByID", projectID.String()).Return(project, nil)
					// Add GetUserRole mock for CheckAccess method
					m.On("GetUserRole", projectID.String(), userID.String()).
						Return(models.ProjectRoleOwner, nil)
					return m
				}(),
				userRepo: func() *MockUserRepository {
					m := &MockUserRepository{}
					m.On("GetByEmail", "nonexistent@example.com").
						Return(nil, errors.New("not found"))
					return m
				}(),
			},
			args: args{
				projectID:   projectID.String(),
				userID:      userID.String(),
				memberEmail: "nonexistent@example.com",
				role:        "viewer",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &projectService{
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
			}

			result, err := s.AddMember(
				tt.args.projectID,
				tt.args.userID,
				tt.args.memberEmail,
				tt.args.role,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "member@example.com", result.User.Email)
			}

			tt.fields.projectRepo.AssertExpectations(t)
			tt.fields.userRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_CheckAccess(t *testing.T) {
	t.Parallel()

	projectID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		projectRepo *MockProjectRepository
		userRepo    *MockUserRepository
	}
	type args struct {
		projectID    string
		userID       string
		requiredRole models.ProjectRole
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult bool
	}{
		{
			name: "owner access granted",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					project := &models.Project{
						ID:      projectID,
						OwnerID: userID,
					}
					m.On("GetByID", projectID.String()).Return(project, nil)
					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				projectID:    projectID.String(),
				userID:       userID.String(),
				requiredRole: models.ProjectRoleOwner,
			},
			wantResult: true,
		},
		{
			name: "project not found",
			fields: fields{
				projectRepo: func() *MockProjectRepository {
					m := &MockProjectRepository{}
					m.On("GetByID", projectID.String()).
						Return(nil, errors.New("not found"))
					return m
				}(),
				userRepo: &MockUserRepository{},
			},
			args: args{
				projectID:    projectID.String(),
				userID:       userID.String(),
				requiredRole: models.ProjectRoleViewer,
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &projectService{
				projectRepo: tt.fields.projectRepo,
				userRepo:    tt.fields.userRepo,
			}

			result := s.CheckAccess(
				tt.args.projectID,
				tt.args.userID,
				tt.args.requiredRole,
			)

			assert.Equal(t, tt.wantResult, result)

			tt.fields.projectRepo.AssertExpectations(t)
		})
	}
}
