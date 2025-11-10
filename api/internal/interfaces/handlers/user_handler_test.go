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

// Package handlers provides HTTP request handlers for user management
// operations.
//
// This package contains handlers for user CRUD operations with proper
// authorization. All handlers follow RESTful conventions with proper error
// handling and JSON responses.
//
// Security features:
//   - Role-based access control (admin/self access)
//   - Input validation and sanitization
//   - Proper error messages (don't leak sensitive information)
//   - Pagination support for list endpoints
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

// createTestLogger creates a zerolog logger for testing
func createTestLogger() zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter())
}

// createTestLoggerPtr creates a pointer to a zerolog logger for testing
func createTestLoggerPtr() *zerolog.Logger {
	logger := createTestLogger()
	return &logger
}

func TestNewUserHandler(t *testing.T) {
	t.Parallel()

	type args struct {
		userService services.UserService
		logger      *zerolog.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successful user handler creation",
			args: args{
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewUserHandler(tt.args.userService, tt.args.logger)

			// Verify that the handler is not nil
			assert.NotNil(t, got)

			// Verify that the handler is of the correct type
			assert.IsType(t, &UserHandler{}, got)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	t.Parallel()
	type fields struct {
		userService *MockUserService
		logger      *zerolog.Logger
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
			name: "successful user listing with default pagination",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					users := []models.User{
						{
							ID:    uuid.Must(uuid.NewV7()),
							Email: "user1@example.com",
							Role:  models.UserRoleUser,
						},
						{
							ID:    uuid.Must(uuid.NewV7()),
							Email: "user2@example.com",
							Role:  models.UserRoleAdmin,
						},
					}
					m.On("List", 20, 0).Return(users, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/users", nil)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful user listing with custom pagination",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					users := []models.User{
						{
							ID:    uuid.Must(uuid.NewV7()),
							Email: "user1@example.com",
							Role:  models.UserRoleUser,
						},
					}
					m.On("List", 10, 10).Return(users, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodGet,
						"/users?page=2&limit=10",
						nil,
					)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "database error during user listing",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("List", mock.AnythingOfType("int"), mock.AnythingOfType("int")).
						Return([]models.User{}, errors.New("database error"))
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/users", nil)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    false,
		},
		{
			name: "invalid pagination parameters (negative values)",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					users := []models.User{}
					m.On("List", 20, 0).Return(users, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodGet,
						"/users?page=-1&limit=-5",
						nil,
					)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &UserHandler{
				userService: tt.fields.userService,
				logger:      tt.fields.logger,
			}

			err := h.ListUsers(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.userService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService *MockUserService
		logger      *zerolog.Logger
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
			name: "successful user retrieval by admin",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "user@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodGet,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful user retrieval by self",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "user@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodGet,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_role", "user")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "user not found",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByID", mock.AnythingOfType("string")).
						Return(nil, errors.New("user not found"))
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodGet,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
					return c
				}(),
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name: "access denied - user trying to access another user's data",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "other@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodGet,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "different-user-id")
					c.Set("user_role", "user")
					return c
				}(),
			},
			wantStatus: http.StatusForbidden,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &UserHandler{
				userService: tt.fields.userService,
				logger:      tt.fields.logger,
			}

			err := h.GetUser(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.userService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService *MockUserService
		logger      *zerolog.Logger
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
			name: "successful user update by admin",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "old@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					m.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					updateReq := dto.UpdateUserRequest{
						Email: "new@example.com",
						Role:  "admin",
					}
					reqBody, _ := json.Marshal(updateReq)
					req := httptest.NewRequest(
						http.MethodPut,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						bytes.NewReader(reqBody),
					)
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful user update by self (email only)",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "old@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					m.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					updateReq := dto.UpdateUserRequest{
						Email: "newemail@example.com",
					}
					reqBody, _ := json.Marshal(updateReq)
					req := httptest.NewRequest(
						http.MethodPut,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						bytes.NewReader(reqBody),
					)
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_role", "user")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "user not found for update",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByID", mock.AnythingOfType("string")).
						Return(nil, errors.New("user not found"))
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					updateReq := dto.UpdateUserRequest{
						Email: "new@example.com",
					}
					reqBody, _ := json.Marshal(updateReq)
					req := httptest.NewRequest(
						http.MethodPut,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						bytes.NewReader(reqBody),
					)
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
					return c
				}(),
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name: "access denied - user trying to update another user",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "other@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					updateReq := dto.UpdateUserRequest{
						Email: "new@example.com",
					}
					reqBody, _ := json.Marshal(updateReq)
					req := httptest.NewRequest(
						http.MethodPut,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						bytes.NewReader(reqBody),
					)
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "different-user-id")
					c.Set("user_role", "user")
					return c
				}(),
			},
			wantStatus: http.StatusForbidden,
			wantErr:    false,
		},
		{
			name: "invalid request body",
			fields: fields{
				userService: &MockUserService{},
				logger:      createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodPut,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						bytes.NewReader([]byte("invalid")),
					)
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
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
			h := &UserHandler{
				userService: tt.fields.userService,
				logger:      tt.fields.logger,
			}

			err := h.UpdateUser(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.userService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Parallel()

	type fields struct {
		userService *MockUserService
		logger      *zerolog.Logger
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
			name: "successful user deletion by admin",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "user@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					m.On("Delete", mock.AnythingOfType("string")).Return(nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodDelete,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "successful user deletion by self",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "user@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					m.On("Delete", mock.AnythingOfType("string")).Return(nil)
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodDelete,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_role", "user")
					return c
				}(),
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "user not found for deletion",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					m.On("GetByID", mock.AnythingOfType("string")).
						Return(nil, errors.New("user not found"))
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodDelete,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
					return c
				}(),
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name: "access denied - user trying to delete another user",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "other@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					// Delete should not be called due to access check
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodDelete,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "different-user-id")
					c.Set("user_role", "user")
					return c
				}(),
			},
			wantStatus: http.StatusForbidden,
			wantErr:    false,
		},
		{
			name: "database error during deletion",
			fields: fields{
				userService: func() *MockUserService {
					m := &MockUserService{}
					user := &models.User{
						ID:    uuid.Must(uuid.NewV7()),
						Email: "user@example.com",
						Role:  models.UserRoleUser,
					}
					m.On("GetByID", mock.AnythingOfType("string")).Return(user, nil)
					m.On("Delete", mock.AnythingOfType("string")).
						Return(errors.New("database error"))
					return m
				}(),
				logger: createTestLoggerPtr(),
			},
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(
						http.MethodDelete,
						"/users/123e4567-e89b-12d3-a456-426614174000",
						nil,
					)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetParamNames("id")
					c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")
					c.Set("user_id", "admin-id")
					c.Set("user_role", "admin")
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
			h := &UserHandler{
				userService: tt.fields.userService,
				logger:      tt.fields.logger,
			}

			err := h.DeleteUser(tt.args.c)
			rec := tt.args.c.Response().Writer.(*httptest.ResponseRecorder)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.fields.userService.AssertExpectations(t)
		})
	}
}
