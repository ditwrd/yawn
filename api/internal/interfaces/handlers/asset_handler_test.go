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

// Package handlers provides HTTP request handlers for asset management
// operations.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAssetService is a mock implementation of services.AssetService.
type MockAssetService struct {
	mock.Mock
}

// Ensure MockAssetService implements the interface.
var _ services.AssetService = (*MockAssetService)(nil)

func (m *MockAssetService) Create(
	ctx context.Context,
	req *services.CreateAssetRequest,
) (*models.Asset, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetService) GetByID(
	ctx context.Context,
	id string,
) (*models.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetService) GetByProjectID(
	ctx context.Context,
	projectID string,
	page, limit int,
) (*services.PaginatedAssetsResponse, error) {
	args := m.Called(ctx, projectID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*services.PaginatedAssetsResponse), args.Error(1)
}

func (m *MockAssetService) GetByRepositoryID(
	ctx context.Context,
	repositoryID string,
) ([]*models.Asset, error) {
	args := m.Called(ctx, repositoryID)

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetService) List(
	ctx context.Context,
	page, limit int,
	filters services.AssetListFilters,
) (*services.PaginatedAssetsResponse, error) {
	args := m.Called(ctx, page, limit, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*services.PaginatedAssetsResponse), args.Error(1)
}

func (m *MockAssetService) Update(
	ctx context.Context,
	id string,
	req *services.UpdateAssetRequest,
) (*models.Asset, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)

	return args.Error(0)
}

func (m *MockAssetService) Search(
	ctx context.Context,
	query string,
	page, limit int,
) (*services.PaginatedAssetsResponse, error) {
	args := m.Called(ctx, query, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*services.PaginatedAssetsResponse), args.Error(1)
}

func (m *MockAssetService) GetVersionHistory(
	ctx context.Context,
	projectID, assetName string,
) ([]*models.Asset, error) {
	args := m.Called(ctx, projectID, assetName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*models.Asset), args.Error(1)
}

func (m *MockAssetService) GetLatestVersion(
	ctx context.Context,
	projectID, assetName string,
) (*models.Asset, error) {
	args := m.Called(ctx, projectID, assetName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Asset), args.Error(1)
}

func (m *MockAssetService) ValidateAccess(
	ctx context.Context,
	assetID, userID string,
	requiredRole models.ProjectRole,
) error {
	args := m.Called(ctx, assetID, userID, requiredRole)

	return args.Error(0)
}

func (m *MockAssetService) CanCreate(
	ctx context.Context,
	projectID, userID string,
) error {
	args := m.Called(ctx, projectID, userID)

	return args.Error(0)
}

// createAssetTestLogger creates a test logger instance.
func createAssetTestLogger() *zerolog.Logger {
	logger := zerolog.Nop()

	return &logger
}

// setupEchoContext creates an echo context for testing.
func setupEchoContext(
	method, path string,
	body any,
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

	return c, rec
}

// setUserContext sets user information in the echo context.
func setUserContext(c echo.Context, userID, userRole string) {
	c.Set("user_id", userID)
	c.Set("user_role", userRole)
}

func TestNewAssetHandler(t *testing.T) {
	t.Parallel()

	mockService := &MockAssetService{}
	logger := createAssetTestLogger()

	handler := NewAssetHandler(mockService, logger)

	assert.NotNil(t, handler)
	assert.IsType(t, &AssetHandler{}, handler)
}

func TestAssetHandler_ListAssets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectID := uuid.Must(uuid.NewV7())
	assetID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		queryParams map[string]string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful list",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					response := &services.PaginatedAssetsResponse{
						Assets: []*models.Asset{
							{
								ID:      assetID,
								Name:    "test-asset",
								Version: "1.0.0",
							},
						},
						Total: 1,
						Page:  1,
						Limit: 20,
					}
					m.On("List", ctx, 1, 20, mock.AnythingOfType("services.AssetListFilters")).
						Return(response, nil)

					return m
				}(),
			},
			args:           args{queryParams: map[string]string{}},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "list with filters",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					response := &services.PaginatedAssetsResponse{
						Assets: []*models.Asset{},
						Total:  0,
						Page:   1,
						Limit:  10,
					}
					filters := services.AssetListFilters{
						ProjectID:    projectID.String(),
						Name:         "test",
						Version:      "1.0.0",
						Search:       "query",
						RepositoryID: "repo-id",
					}
					m.On("List", ctx, 2, 10, filters).Return(response, nil)

					return m
				}(),
			},
			args: args{
				queryParams: map[string]string{
					"page":          "2",
					"limit":         "10",
					"project_id":    projectID.String(),
					"name":          "test",
					"version":       "1.0.0",
					"search":        "query",
					"repository_id": "repo-id",
				},
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "service error",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("List", ctx, 1, 20, mock.AnythingOfType("services.AssetListFilters")).
						Return(nil, errors.New("service error"))

					return m
				}(),
			},
			args:           args{queryParams: map[string]string{}},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			// Create request with query parameters
			path := "/assets"
			if len(tt.args.queryParams) > 0 {
				path += "?"

				var pathSb329 strings.Builder
				for k, v := range tt.args.queryParams {
					pathSb329.WriteString(k + "=" + v + "&")
				}

				path += pathSb329.String()

				path = path[:len(path)-1] // Remove trailing &
			}

			c, rec := setupEchoContext(http.MethodGet, path, nil)
			c.SetRequest(c.Request().WithContext(ctx))

			err := handler.ListAssets(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_GetAsset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		assetID string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful get",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					asset := &models.Asset{
						ID:      assetID,
						Name:    "test-asset",
						Version: "1.0.0",
					}
					m.On("GetByID", ctx, assetID.String()).Return(asset, nil)

					return m
				}(),
			},
			args: args{
				assetID: assetID.String(),
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "missing asset ID",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				assetID: "",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "invalid asset ID",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				assetID: "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "asset not found",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("GetByID", ctx, assetID.String()).
						Return(nil, errors.New("not found"))

					return m
				}(),
			},
			args: args{
				assetID: assetID.String(),
			},
			expectedStatus: http.StatusNotFound,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			path := "/assets/" + tt.args.assetID
			c, rec := setupEchoContext(http.MethodGet, path, nil)
			c.SetParamNames("id")
			c.SetParamValues(tt.args.assetID)
			c.SetRequest(c.Request().WithContext(ctx))

			err := handler.GetAsset(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_CreateAsset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectID := uuid.Must(uuid.NewV7())
	assetID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())
	repoID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		request  dto.CreateAssetRequest
		userID   string
		userRole string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful creation",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("CanCreate", ctx, projectID.String(), userID.String()).
						Return(nil)
					asset := &models.Asset{
						ID:        assetID,
						Name:      "test-asset",
						Version:   "1.0.0",
						ProjectID: projectID,
					}
					serviceReq := &services.CreateAssetRequest{
						Name:        "test-asset",
						Description: "Test description",
						Version:     "1.0.0",
						ProjectID:   projectID.String(),
					}
					m.On("Create", ctx, serviceReq).Return(asset, nil)

					return m
				}(),
			},
			args: args{
				request: dto.CreateAssetRequest{
					Name:        "test-asset",
					Description: "Test description",
					Version:     "1.0.0",
					ProjectID:   projectID.String(),
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "creation with repository",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("CanCreate", ctx, projectID.String(), userID.String()).
						Return(nil)
					asset := &models.Asset{
						ID:           assetID,
						Name:         "test-asset",
						Version:      "1.0.0",
						ProjectID:    projectID,
						RepositoryID: &repoID,
					}
					serviceReq := &services.CreateAssetRequest{
						Name:         "test-asset",
						Description:  "Test description",
						Version:      "1.0.0",
						ProjectID:    projectID.String(),
						RepositoryID: &[]string{repoID.String()}[0],
					}
					m.On("Create", ctx, serviceReq).Return(asset, nil)

					return m
				}(),
			},
			args: args{
				request: dto.CreateAssetRequest{
					Name:         "test-asset",
					Description:  "Test description",
					Version:      "1.0.0",
					ProjectID:    projectID.String(),
					RepositoryID: &[]string{repoID.String()}[0],
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "access denied",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("CanCreate", ctx, projectID.String(), userID.String()).
						Return(errors.New("access denied"))

					return m
				}(),
			},
			args: args{
				request: dto.CreateAssetRequest{
					Name:      "test-asset",
					Version:   "1.0.0",
					ProjectID: projectID.String(),
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusForbidden,
			wantErr:        false,
		},
		{
			name: "missing user authentication",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				request: dto.CreateAssetRequest{
					Name:      "test-asset",
					Version:   "1.0.0",
					ProjectID: projectID.String(),
				},
				userID:   "",
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusUnauthorized,
			wantErr:        false,
		},
		{
			name: "invalid request body",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					// Expect CanCreate call even though validation will fail
					m.On("CanCreate", ctx, "", userID.String()).Return(nil)
					// Expect Create call even though validation will fail
					m.On("Create", ctx, mock.AnythingOfType("*services.CreateAssetRequest")).
						Return(nil, errors.New("validation failed"))

					return m
				}(),
			},
			args: args{
				request: dto.CreateAssetRequest{
					Name: "", // Invalid: empty name
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			c, rec := setupEchoContext(http.MethodPost, "/assets", tt.args.request)
			c.SetRequest(c.Request().WithContext(ctx))

			if tt.args.userID != "" {
				setUserContext(c, tt.args.userID, tt.args.userRole)
			}

			err := handler.CreateAsset(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_UpdateAsset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		assetID  string
		request  dto.UpdateAssetRequest
		userID   string
		userRole string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful update",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("ValidateAccess", ctx, assetID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(nil)
					asset := &models.Asset{
						ID:      assetID,
						Name:    "updated-asset",
						Version: "2.0.0",
					}
					updateReq := &services.UpdateAssetRequest{
						Name: func() *string {
							s := "updated-asset"

							return &s
						}(),
					}
					m.On("Update", ctx, assetID.String(), updateReq).Return(asset, nil)

					return m
				}(),
			},
			args: args{
				assetID: assetID.String(),
				request: dto.UpdateAssetRequest{
					Name: func() *string {
						s := "updated-asset"

						return &s
					}(),
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "access denied",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("ValidateAccess", ctx, assetID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(errors.New("access denied"))

					return m
				}(),
			},
			args: args{
				assetID: assetID.String(),
				request: dto.UpdateAssetRequest{
					Name: func() *string {
						s := "updated-asset"

						return &s
					}(),
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusForbidden,
			wantErr:        false,
		},
		{
			name: "invalid asset ID",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				assetID: "invalid-uuid",
				request: dto.UpdateAssetRequest{
					Name: func() *string {
						s := "updated-asset"

						return &s
					}(),
				},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "empty asset ID",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				assetID:  "",
				request:  dto.UpdateAssetRequest{},
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			path := "/assets/" + tt.args.assetID
			c, rec := setupEchoContext(http.MethodPut, path, tt.args.request)
			c.SetParamNames("id")
			c.SetParamValues(tt.args.assetID)
			c.SetRequest(c.Request().WithContext(ctx))
			setUserContext(c, tt.args.userID, tt.args.userRole)

			err := handler.UpdateAsset(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_DeleteAsset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		assetID  string
		userID   string
		userRole string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful deletion",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("ValidateAccess", ctx, assetID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(nil)
					m.On("Delete", ctx, assetID.String()).Return(nil)

					return m
				}(),
			},
			args: args{
				assetID:  assetID.String(),
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "access denied",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("ValidateAccess", ctx, assetID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(errors.New("access denied"))

					return m
				}(),
			},
			args: args{
				assetID:  assetID.String(),
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusForbidden,
			wantErr:        false,
		},
		{
			name: "invalid asset ID",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				assetID:  "invalid-uuid",
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "service error",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("ValidateAccess", ctx, assetID.String(), userID.String(), models.ProjectRoleMaintainer).
						Return(nil)
					m.On("Delete", ctx, assetID.String()).
						Return(errors.New("delete failed"))

					return m
				}(),
			},
			args: args{
				assetID:  assetID.String(),
				userID:   userID.String(),
				userRole: string(models.UserRoleUser),
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			path := "/assets/" + tt.args.assetID
			c, rec := setupEchoContext(http.MethodDelete, path, nil)
			c.SetParamNames("id")
			c.SetParamValues(tt.args.assetID)
			c.SetRequest(c.Request().WithContext(ctx))
			setUserContext(c, tt.args.userID, tt.args.userRole)

			err := handler.DeleteAsset(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_SearchAssets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assetID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		query string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful search",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					response := &services.PaginatedAssetsResponse{
						Assets: []*models.Asset{
							{
								ID:      assetID,
								Name:    "test-asset",
								Version: "1.0.0",
							},
						},
						Total: 1,
						Page:  1,
						Limit: 20,
					}
					m.On("Search", ctx, "test", 1, 20).Return(response, nil)

					return m
				}(),
			},
			args: args{
				query: "test",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "missing query",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				query: "",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "empty query",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				query: "   ",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "service error",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("Search", ctx, "test", 1, 20).
						Return(nil, errors.New("search failed"))

					return m
				}(),
			},
			args: args{
				query: "test",
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			path := "/assets/search?q=" + url.QueryEscape(tt.args.query)
			c, rec := setupEchoContext(http.MethodGet, path, nil)
			c.SetRequest(c.Request().WithContext(ctx))

			err := handler.SearchAssets(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_GetAssetVersionHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	projectID := uuid.Must(uuid.NewV7())
	assetID := uuid.Must(uuid.NewV7())

	type fields struct {
		assetService *MockAssetService
	}

	type args struct {
		projectID string
		assetName string
	}

	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful version history",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					assets := []*models.Asset{
						{
							ID:      assetID,
							Name:    "test-asset",
							Version: "2.0.0",
						},
						{
							ID:      uuid.Must(uuid.NewV7()),
							Name:    "test-asset",
							Version: "1.0.0",
						},
					}
					m.On("GetVersionHistory", ctx, projectID.String(), "test-asset").
						Return(assets, nil)

					return m
				}(),
			},
			args: args{
				projectID: projectID.String(),
				assetName: "test-asset",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "missing project ID",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				projectID: "",
				assetName: "test-asset",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "missing asset name",
			fields: fields{
				assetService: &MockAssetService{},
			},
			args: args{
				projectID: projectID.String(),
				assetName: "",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name: "asset not found",
			fields: fields{
				assetService: func() *MockAssetService {
					m := &MockAssetService{}
					m.On("GetVersionHistory", ctx, projectID.String(), "non-existent").
						Return(nil, errors.New("not found"))

					return m
				}(),
			},
			args: args{
				projectID: projectID.String(),
				assetName: "non-existent",
			},
			expectedStatus: http.StatusNotFound,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &AssetHandler{
				assetService: tt.fields.assetService,
				logger:       createAssetTestLogger(),
			}

			path := "/assets/" + tt.args.projectID + "/" + tt.args.assetName + "/versions"
			c, rec := setupEchoContext(http.MethodGet, path, nil)
			c.SetParamNames("project_id", "asset_name")
			c.SetParamValues(tt.args.projectID, tt.args.assetName)
			c.SetRequest(c.Request().WithContext(ctx))

			err := handler.GetAssetVersionHistory(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				// Verify response structure for successful case
				if !tt.wantErr && tt.expectedStatus == http.StatusOK {
					var response dto.AssetVersionHistoryResponse

					err := json.Unmarshal(rec.Body.Bytes(), &response)
					require.NoError(t, err)
					assert.Equal(t, tt.args.assetName, response.AssetName)
					assert.NotEmpty(t, response.Versions)
				}
			}

			tt.fields.assetService.AssertExpectations(t)
		})
	}
}

func TestAssetHandler_assetToResponse(t *testing.T) {
	t.Parallel()

	handler := &AssetHandler{
		logger: createAssetTestLogger(),
	}

	projectID := uuid.Must(uuid.NewV7())
	repoID := uuid.Must(uuid.NewV7())
	pipelineID := uuid.Must(uuid.NewV7())

	asset := &models.Asset{
		ID:           uuid.Must(uuid.NewV7()),
		Name:         "test-asset",
		Description:  "Test description",
		Version:      "1.0.0",
		ProjectID:    projectID,
		RepositoryID: &repoID,
		Project: models.Project{
			ID:          projectID,
			Name:        "Test Project",
			Description: "Project description",
			Visibility:  "private",
		},
		Repository: &models.Repository{
			ID:         repoID,
			URL:        "https://github.com/test/repo.git",
			Branch:     "main",
			SyncStatus: models.RepositoryStatusSuccess,
			ProjectID:  projectID,
		},
		Pipelines: []models.Pipeline{
			{
				ID:        pipelineID,
				Name:      "test-pipeline",
				ProjectID: projectID,
			},
		},
	}

	response := handler.assetToResponse(asset)

	assert.Equal(t, asset.ID.String(), response.ID)
	assert.Equal(t, asset.Name, response.Name)
	assert.Equal(t, asset.Description, response.Description)
	assert.Equal(t, asset.Version, response.Version)
	assert.Equal(t, asset.ProjectID.String(), response.ProjectID)
	assert.NotNil(t, response.RepositoryID)
	assert.Equal(t, repoID.String(), *response.RepositoryID)

	// Check nested relationships
	assert.NotNil(t, response.Project)
	assert.Equal(t, projectID.String(), response.Project.ID)
	assert.Equal(t, "Test Project", response.Project.Name)

	assert.NotNil(t, response.Repository)
	assert.Equal(t, repoID.String(), response.Repository.ID)
	assert.Equal(t, "https://github.com/test/repo.git", response.Repository.URL)

	assert.Len(t, response.PipelineIDs, 1)
	assert.Equal(t, pipelineID.String(), response.PipelineIDs[0])
}
