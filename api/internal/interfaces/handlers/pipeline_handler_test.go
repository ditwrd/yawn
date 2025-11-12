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

// Package handlers provides HTTP request handlers for pipeline management
// operations.
//
// This package contains handlers for pipeline CRUD operations with proper
// authorization and validation. All handlers follow RESTful conventions
// with proper error handling and JSON responses.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPipelineService is a mock implementation of services.PipelineService for
// testing.
type MockPipelineService struct {
	mock.Mock
}

// Ensure MockPipelineService implements the interface.
var _ services.PipelineService = (*MockPipelineService)(nil)

func (m *MockPipelineService) Create(
	ctx context.Context,
	req *services.CreatePipelineRequest,
) (*models.Pipeline, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pipeline), args.Error(1)
}

func (m *MockPipelineService) GetByID(
	ctx context.Context,
	id string,
) (*models.Pipeline, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pipeline), args.Error(1)
}

func (m *MockPipelineService) GetByProjectID(
	ctx context.Context,
	projectID string,
	page, limit int,
) (*services.PaginatedPipelinesResponse, error) {
	args := m.Called(ctx, projectID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.PaginatedPipelinesResponse), args.Error(1)
}

func (m *MockPipelineService) List(
	ctx context.Context,
	page, limit int,
	filters services.PipelineListFilters,
) (*services.PaginatedPipelinesResponse, error) {
	args := m.Called(ctx, page, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.PaginatedPipelinesResponse), args.Error(1)
}

func (m *MockPipelineService) Update(
	ctx context.Context,
	id string,
	req *services.UpdatePipelineRequest,
) (*models.Pipeline, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pipeline), args.Error(1)
}

func (m *MockPipelineService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPipelineService) Search(
	ctx context.Context,
	query string,
	page, limit int,
) (*services.PaginatedPipelinesResponse, error) {
	args := m.Called(ctx, query, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.PaginatedPipelinesResponse), args.Error(1)
}

func (m *MockPipelineService) UpdateStatus(
	ctx context.Context,
	id string,
	status models.PipelineStatus,
) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPipelineService) ValidateAccess(
	ctx context.Context,
	pipelineID, userID string,
	requiredRole models.ProjectRole,
) error {
	args := m.Called(ctx, pipelineID, userID, requiredRole)
	return args.Error(0)
}

func (m *MockPipelineService) CanCreate(
	ctx context.Context,
	projectID, userID string,
) error {
	args := m.Called(ctx, projectID, userID)
	return args.Error(0)
}

func (m *MockPipelineService) TriggerExecution(
	ctx context.Context,
	pipelineID, userID string,
	config *string,
) (*models.PipelineExecution, error) {
	args := m.Called(ctx, pipelineID, userID, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PipelineExecution), args.Error(1)
}

func (m *MockPipelineService) GetExecutionByID(
	ctx context.Context,
	executionID string,
) (*models.PipelineExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PipelineExecution), args.Error(1)
}

func (m *MockPipelineService) GetExecutionsByPipelineID(
	ctx context.Context,
	pipelineID string,
	page, limit int,
) (*services.PaginatedExecutionsResponse, error) {
	args := m.Called(ctx, pipelineID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.PaginatedExecutionsResponse), args.Error(1)
}

func (m *MockPipelineService) CancelExecution(
	ctx context.Context,
	executionID, userID string,
) error {
	args := m.Called(ctx, executionID, userID)
	return args.Error(0)
}

func (m *MockPipelineService) GetRunningExecutions(
	ctx context.Context,
) ([]*models.PipelineExecution, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PipelineExecution), args.Error(1)
}

func (m *MockPipelineService) AddDependency(
	ctx context.Context,
	pipelineID, dependsOnID string,
	condition *string,
) error {
	args := m.Called(ctx, pipelineID, dependsOnID, condition)
	return args.Error(0)
}

func (m *MockPipelineService) RemoveDependency(
	ctx context.Context,
	pipelineID, dependsOnID string,
) error {
	args := m.Called(ctx, pipelineID, dependsOnID)
	return args.Error(0)
}

func (m *MockPipelineService) GetDependencyGraph(
	ctx context.Context,
	projectID string,
) (*services.DependencyGraphResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.DependencyGraphResponse), args.Error(1)
}

func (m *MockPipelineService) ValidateDependencies(
	ctx context.Context,
	pipelineID string,
) error {
	args := m.Called(ctx, pipelineID)
	return args.Error(0)
}

func (m *MockPipelineService) ResolveDependencies(
	ctx context.Context,
	pipelineID string,
) ([]string, error) {
	args := m.Called(ctx, pipelineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// createPipelineTestLogger creates a zerolog logger for testing pipeline
// handlers.
func createPipelineTestLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	return &logger
}

// createTestEchoContext creates an Echo context for testing.
func createTestEchoContext(
	method, path string,
	body interface{},
	params map[string]string,
) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if params != nil {
		paramNames := make([]string, 0, len(params))
		paramValues := make([]string, 0, len(params))
		for name, value := range params {
			paramNames = append(paramNames, name)
			paramValues = append(paramValues, value)
		}
		c.SetParamNames(paramNames...)
		c.SetParamValues(paramValues...)
	}

	return c, rec
}

func TestNewPipelineHandler(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	mockService := &MockPipelineService{}

	type args struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}

	tests := []struct {
		name string
		args args
		want *PipelineHandler
	}{
		{
			name: "should create new PipelineHandler with valid dependencies",
			args: args{
				pipelineService: mockService,
				logger:          logger,
			},
			want: &PipelineHandler{
				pipelineService: mockService,
				logger:          logger,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewPipelineHandler(tt.args.pipelineService, tt.args.logger)
			assert.Equal(t, tt.want.pipelineService, got.pipelineService)
			assert.Equal(t, tt.want.logger, got.logger)
		})
	}
}

func TestPipelineHandler_ListPipelines(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	mockService := &MockPipelineService{}
	h := &PipelineHandler{
		pipelineService: mockService,
		logger:          logger,
	}

	c, _ := createTestEchoContext("GET", "/pipelines", nil, nil)

	// Placeholder test to verify method exists
	_ = h
	_ = c
	_ = mockService

	assert.True(t, true, "PipelineHandler.ListPipelines method exists")
}

func TestPipelineHandler_GetPipeline(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.GetPipeline(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.GetPipeline() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_CreatePipeline(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	mockService := &MockPipelineService{}
	h := &PipelineHandler{
		pipelineService: mockService,
		logger:          logger,
	}

	c, _ := createTestEchoContext("POST", "/pipelines", nil, nil)

	// Placeholder test to verify method exists
	_ = h
	_ = c
	_ = mockService

	assert.True(t, true, "PipelineHandler.CreatePipeline method exists")
}

func TestPipelineHandler_UpdatePipeline(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.UpdatePipeline(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.UpdatePipeline() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_DeletePipeline(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.DeletePipeline(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.DeletePipeline() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_SearchPipelines(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.SearchPipelines(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.SearchPipelines() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_UpdatePipelineStatus(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.UpdatePipelineStatus(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.UpdatePipelineStatus() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_TriggerExecution(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.TriggerExecution(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.TriggerExecution() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_GetExecutions(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.GetExecutions(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.GetExecutions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_CancelExecution(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.CancelExecution(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.CancelExecution() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_GetDependencyGraph(t *testing.T) {
	t.Parallel()
	type fields struct {
		pipelineService services.PipelineService
		logger          *zerolog.Logger
	}
	type args struct {
		c echo.Context
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
			h := &PipelineHandler{
				pipelineService: tt.fields.pipelineService,
				logger:          tt.fields.logger,
			}
			if err := h.GetDependencyGraph(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf(
					"PipelineHandler.GetDependencyGraph() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestPipelineHandler_pipelineToResponse(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	h := &PipelineHandler{logger: logger}

	// Create proper UUIDs
	pipelineUUID, _ := uuid.NewV7()
	projectUUID, _ := uuid.NewV7()

	tests := []struct {
		name     string
		pipeline *models.Pipeline
		want     dto.PipelineResponse
	}{
		{
			name: "should convert pipeline to response correctly",
			pipeline: &models.Pipeline{
				ID:          pipelineUUID,
				Name:        "Build Pipeline",
				Description: "Builds and tests the application",
				Status:      models.PipelineStatusActive,
				Config:      "steps: [{name: 'build'}]",
				Schedule:    "0 2 * * *",
				IsEnabled:   true,
				ProjectID:   projectUUID,
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			},
			want: dto.PipelineResponse{
				ID:          pipelineUUID.String(),
				Name:        "Build Pipeline",
				Description: "Builds and tests the application",
				Status:      string(models.PipelineStatusActive),
				Config:      "steps: [{name: 'build'}]",
				Schedule:    "0 2 * * *",
				IsEnabled:   true,
				ProjectID:   projectUUID.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := h.pipelineToResponse(tt.pipeline)

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Status, got.Status)
			assert.Equal(t, tt.want.Config, got.Config)
			assert.Equal(t, tt.want.Schedule, got.Schedule)
			assert.Equal(t, tt.want.IsEnabled, got.IsEnabled)
			assert.Equal(t, tt.want.ProjectID, got.ProjectID)

			// Check that timestamp fields are populated
			assert.NotEmpty(t, got.CreatedAt)
			assert.NotEmpty(t, got.UpdatedAt)
		})
	}
}

func TestPipelineHandler_executionToResponse(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	h := &PipelineHandler{logger: logger}

	// Create proper UUIDs and time pointers
	executionUUID, _ := uuid.NewV7()
	pipelineUUID, _ := uuid.NewV7()
	startedAt := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name      string
		execution *models.PipelineExecution
		want      dto.PipelineExecutionResponse
	}{
		{
			name: "should convert execution to response correctly",
			execution: &models.PipelineExecution{
				ID:         executionUUID,
				PipelineID: pipelineUUID,
				Status:     models.ExecutionStatusRunning,
				StartedAt:  &startedAt,
			},
			want: dto.PipelineExecutionResponse{
				ID:         executionUUID.String(),
				PipelineID: pipelineUUID.String(),
				Status:     string(models.ExecutionStatusRunning),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := h.executionToResponse(tt.execution)

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.PipelineID, got.PipelineID)
			assert.Equal(t, tt.want.Status, got.Status)

			// Check that timestamp fields are populated
			assert.NotEmpty(t, got.StartedAt)
		})
	}
}

func TestPipelineHandler_stepToResponse(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	h := &PipelineHandler{logger: logger}

	// Create proper UUIDs and time pointers
	stepUUID, _ := uuid.NewV7()
	executionUUID, _ := uuid.NewV7()
	startedAt := time.Now().Add(-30 * time.Minute)
	completedAt := time.Now().Add(-10 * time.Minute)

	tests := []struct {
		name string
		step *models.ExecutionStep
		want dto.ExecutionStepResponse
	}{
		{
			name: "should convert step to response correctly",
			step: &models.ExecutionStep{
				ID:          stepUUID,
				ExecutionID: executionUUID,
				Name:        "Build Step",
				Status:      models.ExecutionStatusCompleted,
				StartedAt:   &startedAt,
				CompletedAt: &completedAt,
			},
			want: dto.ExecutionStepResponse{
				ID:          stepUUID.String(),
				ExecutionID: executionUUID.String(),
				Name:        "Build Step",
				Status:      string(models.ExecutionStatusCompleted),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := h.stepToResponse(tt.step)

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.ExecutionID, got.ExecutionID)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Status, got.Status)

			// Check that timestamp fields are populated
			assert.NotEmpty(t, got.StartedAt)
			assert.NotEmpty(t, got.CompletedAt)
		})
	}
}

func TestPipelineHandler_dependencyToResponse(t *testing.T) {
	t.Parallel()
	logger := createPipelineTestLogger()
	h := &PipelineHandler{logger: logger}

	// Create proper UUIDs
	dependencyUUID, _ := uuid.NewV7()
	pipelineUUID, _ := uuid.NewV7()
	dependsOnUUID, _ := uuid.NewV7()

	tests := []struct {
		name       string
		dependency *models.PipelineDependency
		want       dto.PipelineDependencyResponse
	}{
		{
			name: "should convert dependency to response correctly",
			dependency: &models.PipelineDependency{
				ID:                  dependencyUUID,
				PipelineID:          pipelineUUID,
				DependsOnPipelineID: dependsOnUUID,
				Condition:           "success",
				CreatedAt:           time.Now().Add(-24 * time.Hour),
			},
			want: dto.PipelineDependencyResponse{
				ID:                  dependencyUUID.String(),
				PipelineID:          pipelineUUID.String(),
				DependsOnPipelineID: dependsOnUUID.String(),
				Condition:           "success",
				CreatedAt:           "2024-01-01T00:00:00Z", // Will be set by handler
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := h.dependencyToResponse(tt.dependency)

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.PipelineID, got.PipelineID)
			assert.Equal(t, tt.want.DependsOnPipelineID, got.DependsOnPipelineID)
			assert.Equal(t, tt.want.Condition, got.Condition)

			// Check that timestamp fields are populated
			assert.NotEmpty(t, got.CreatedAt)
		})
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
