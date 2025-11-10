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

// Package handlers provides HTTP request handlers for project management
// operations.
package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectService is a mock implementation of services.ProjectService for
// testing.
type MockProjectService struct {
	mock.Mock
}

// Ensure MockProjectService implements the interface
var _ services.ProjectService = (*MockProjectService)(nil)

func (m *MockProjectService) Create(
	project *models.Project,
	ownerID string,
) error {
	args := m.Called(project, ownerID)
	return args.Error(0)
}

func (m *MockProjectService) GetByID(
	id, userID string,
) (*models.Project, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectService) List(
	userID string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectService) Search(
	userID, query string,
	limit, offset int,
) ([]models.Project, error) {
	args := m.Called(userID, query, limit, offset)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectService) Update(
	project *models.Project,
	userID string,
) error {
	args := m.Called(project, userID)
	return args.Error(0)
}

func (m *MockProjectService) Delete(id, userID string) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockProjectService) AddMember(
	projectID, userID, memberEmail, role string,
) (*models.ProjectUser, error) {
	args := m.Called(projectID, userID, memberEmail, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectUser), args.Error(1)
}

func (m *MockProjectService) RemoveMember(
	projectID, userID, memberUserID string,
) error {
	args := m.Called(projectID, userID, memberUserID)
	return args.Error(0)
}

func (m *MockProjectService) UpdateMemberRole(
	projectID, userID, memberUserID, role string,
) (*models.ProjectUser, error) {
	args := m.Called(projectID, userID, memberUserID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectUser), args.Error(1)
}

func (m *MockProjectService) ListMembers(
	projectID, userID string,
) ([]models.ProjectUser, error) {
	args := m.Called(projectID, userID)
	return args.Get(0).([]models.ProjectUser), args.Error(1)
}

func (m *MockProjectService) GetUserRole(
	projectID, userID string,
) (models.ProjectRole, error) {
	args := m.Called(projectID, userID)
	return args.Get(0).(models.ProjectRole), args.Error(1)
}

func (m *MockProjectService) CheckAccess(
	projectID, userID string,
	requiredRole models.ProjectRole,
) bool {
	args := m.Called(projectID, userID, requiredRole)
	return args.Bool(0)
}

// Helper function to create test project
func createTestProject() *models.Project {
	// Use fixed UUIDs for consistent testing
	projectID := uuid.Must(
		uuid.FromString("019a6cbf-3b4c-7090-9595-8236be7f5eec"),
	)
	ownerID := uuid.Must(uuid.FromString("019a6cbf-3b4c-7090-9595-8236be7f5eed"))
	userID := uuid.Must(uuid.FromString("019a6cbf-3b4c-7090-9595-8236be7f5eef"))

	return &models.Project{
		ID:          projectID,
		Name:        "Test Project",
		Description: "A test project for unit testing",
		Visibility:  "private",
		OwnerID:     ownerID,
		Owner: models.User{
			ID:    userID,
			Email: "owner@example.com",
			Role:  models.UserRoleUser,
		},
	}
}

func TestNewProjectHandler(t *testing.T) {
	t.Parallel()

	projectService := &MockProjectService{}
	userService := &MockUserService{}
	logger := createTestLoggerPtr()

	handler := NewProjectHandler(projectService, userService, logger)

	assert.NotNil(t, handler)
	assert.IsType(t, &ProjectHandler{}, handler)
}

func TestProjectHandler_CreateProject(t *testing.T) {
	t.Parallel()

	type fields struct {
		projectService *MockProjectService
		userService    *MockUserService
		logger         *zerolog.Logger
	}
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful project creation",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					project := createTestProject()
					m.On("Create", mock.AnythingOfType("*models.Project"), "test-user-id").
						Return(nil)
					m.On("GetByID", mock.AnythingOfType("string"), "test-user-id").
						Return(project, nil)
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := dto.CreateProjectRequest{
						Name:        "Test Project",
						Description: "A test project",
						Visibility:  "private",
					}
					reqBody, _ := json.Marshal(req)
					request := httptest.NewRequest(
						http.MethodPost,
						"/projects",
						bytes.NewReader(reqBody),
					)
					request.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "missing project name",
			fields: fields{
				projectService: &MockProjectService{},
				userService:    &MockUserService{},
				logger:         createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := dto.CreateProjectRequest{
						Description: "A test project",
						Visibility:  "private",
					}
					reqBody, _ := json.Marshal(req)
					request := httptest.NewRequest(
						http.MethodPost,
						"/projects",
						bytes.NewReader(reqBody),
					)
					request.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
		{
			name: "invalid request format",
			fields: fields{
				projectService: &MockProjectService{},
				userService:    &MockUserService{},
				logger:         createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					request := httptest.NewRequest(
						http.MethodPost,
						"/projects",
						bytes.NewReader([]byte("invalid")),
					)
					request.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
		{
			name: "project creation service error",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					m.On("Create", mock.AnythingOfType("*models.Project"), "test-user-id").
						Return(errors.New("database error"))
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := dto.CreateProjectRequest{
						Name:        "Test Project",
						Description: "A test project",
						Visibility:  "private",
					}
					reqBody, _ := json.Marshal(req)
					request := httptest.NewRequest(
						http.MethodPost,
						"/projects",
						bytes.NewReader(reqBody),
					)
					request.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &ProjectHandler{
				projectService: tt.fields.projectService,
				userService:    tt.fields.userService,
				logger:         tt.fields.logger,
			}

			err := h.CreateProject(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.projectService.AssertExpectations(t)
		})
	}
}

func TestProjectHandler_GetProject(t *testing.T) {
	t.Parallel()

	type fields struct {
		projectService *MockProjectService
		userService    *MockUserService
		logger         *zerolog.Logger
	}
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful project retrieval",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					project := createTestProject()
					projectID := project.ID.String()
					m.On("GetByID", projectID, "test-user-id").Return(project, nil)
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					project := createTestProject()
					projectID := project.ID.String()
					request := httptest.NewRequest(
						http.MethodGet,
						"/projects/"+projectID,
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.SetParamNames("id")
					c.SetParamValues(projectID)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "project not found",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					projectID := uuid.Must(uuid.FromString("019a6cc1-763a-7378-bc8f-32387b98d226")).
						String()
					m.On("GetByID", projectID, "test-user-id").
						Return(nil, errors.New("not found"))
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					projectID := uuid.Must(uuid.FromString("019a6cc1-763a-7378-bc8f-32387b98d226")).
						String()
					request := httptest.NewRequest(
						http.MethodGet,
						"/projects/"+projectID,
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.SetParamNames("id")
					c.SetParamValues(projectID)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name: "invalid UUID format",
			fields: fields{
				projectService: &MockProjectService{},
				userService:    &MockUserService{},
				logger:         createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					request := httptest.NewRequest(
						http.MethodGet,
						"/projects/invalid-uuid",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.SetParamNames("id")
					c.SetParamValues("invalid-uuid")
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &ProjectHandler{
				projectService: tt.fields.projectService,
				userService:    tt.fields.userService,
				logger:         tt.fields.logger,
			}

			err := h.GetProject(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.projectService.AssertExpectations(t)
		})
	}
}

func TestProjectHandler_ListProjects(t *testing.T) {
	t.Parallel()

	type fields struct {
		projectService *MockProjectService
		userService    *MockUserService
		logger         *zerolog.Logger
	}
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful project listing",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					projects := []models.Project{*createTestProject()}
					m.On("List", "test-user-id", 20, 0).Return(projects, nil)
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					request := httptest.NewRequest(http.MethodGet, "/projects", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful project listing with pagination",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					projects := []models.Project{*createTestProject()}
					m.On("List", "test-user-id", 10, 10).Return(projects, nil)
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					request := httptest.NewRequest(
						http.MethodGet,
						"/projects?page=2&limit=10",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful project search",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					projects := []models.Project{*createTestProject()}
					m.On("Search", "test-user-id", "test", 20, 0).Return(projects, nil)
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					request := httptest.NewRequest(
						http.MethodGet,
						"/projects?search=test",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "service error during listing",
			fields: fields{
				projectService: func() *MockProjectService {
					m := &MockProjectService{}
					m.On("List", "test-user-id", 20, 0).
						Return([]models.Project{}, errors.New("database error"))
					return m
				}(),
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					request := httptest.NewRequest(http.MethodGet, "/projects", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(request, rec)
					c.Set("user_id", "test-user-id")
					return c
				}(),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &ProjectHandler{
				projectService: tt.fields.projectService,
				userService:    tt.fields.userService,
				logger:         tt.fields.logger,
			}

			err := h.ListProjects(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.projectService.AssertExpectations(t)
		})
	}
}
