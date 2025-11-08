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

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
)

func setupTestEcho() (*echo.Echo, *MockUserService, *UserHandler) {
	e := echo.New()
	mockService := &MockUserService{}
	logger := zerolog.New(bytes.NewBuffer(nil))
	handler := NewUserHandler(mockService, &logger)
	return e, mockService, handler
}

func createTestUserWithParams(id, email string, role models.UserRole) *models.User {
	userID, _ := uuid.FromString(id)
	return &models.User{
		ID:    userID,
		Email: email,
		Role:  role,
	}
}

func TestUserHandler_ListUsers_AdminSuccess(t *testing.T) {
	e, mockService, handler := setupTestEcho()

	testUsers := []models.User{
		*createTestUserWithParams("123e4567-e89b-12d3-a456-426614174000", "admin@example.com", models.UserRoleAdmin),
		*createTestUserWithParams("123e4567-e89b-12d3-a456-426614174001", "user@example.com", models.UserRoleUser),
	}

	mockService.On("List", 20, 0).Return(testUsers, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user context as admin
	c.Set("user_role", "admin")

	err := handler.ListUsers(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.UserListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Users, 2)
	assert.Equal(t, "admin@example.com", response.Users[0].Email)
	assert.Equal(t, "user@example.com", response.Users[1].Email)

	mockService.AssertExpectations(t)
}

func TestUserHandler_GetUser_SelfAccessSuccess(t *testing.T) {
	e, mockService, handler := setupTestEcho()

	testUser := createTestUserWithParams("123e4567-e89b-12d3-a456-426614174000", "user@example.com", models.UserRoleUser)
	mockService.On("GetByID", "123e4567-e89b-12d3-a456-426614174000").Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

	// Set user context as the same user
	c.Set("user_id", "123e4567-e89b-12d3-a456-426614174000")
	c.Set("user_role", "user")

	err := handler.GetUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "user@example.com", response.Email)
	assert.Equal(t, "user", response.Role)

	mockService.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_AdminSuccess(t *testing.T) {
	e, mockService, handler := setupTestEcho()

	testUser := createTestUserWithParams("123e4567-e89b-12d3-a456-426614174000", "user@example.com", models.UserRoleUser)

	mockService.On("GetByID", "123e4567-e89b-12d3-a456-426614174000").Return(testUser, nil)
	mockService.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	updateReq := dto.UpdateUserRequest{
		Email: "admin@example.com",
		Role:  "admin",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest(http.MethodPut, "/users/123e4567-e89b-12d3-a456-426614174000", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

	// Set user context as admin
	c.Set("user_role", "admin")

	err := handler.UpdateUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "admin@example.com", response.Email)
	assert.Equal(t, "admin", response.Role)

	mockService.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_AdminSuccess(t *testing.T) {
	e, mockService, handler := setupTestEcho()

	testUser := createTestUserWithParams("123e4567-e89b-12d3-a456-426614174000", "user@example.com", models.UserRoleUser)
	mockService.On("GetByID", "123e4567-e89b-12d3-a456-426614174000").Return(testUser, nil)
	mockService.On("Delete", "123e4567-e89b-12d3-a456-426614174000").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

	// Set user context as admin
	c.Set("user_role", "admin")

	err := handler.DeleteUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.UserDeleteResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "User deleted successfully", response.Message)

	mockService.AssertExpectations(t)
}

func TestUserHandler_ListUsers_NonAdminAccess(t *testing.T) {
	e, mockService, handler := setupTestEcho()

	testUsers := []models.User{
		*createTestUserWithParams("123e4567-e89b-12d3-a456-426614174001", "user@example.com", models.UserRoleUser),
	}

	mockService.On("List", 20, 0).Return(testUsers, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user context as regular user (in real app, middleware would block this)
	c.Set("user_role", "user")

	err := handler.ListUsers(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.UserListResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Users, 1)
	assert.Equal(t, "user@example.com", response.Users[0].Email)

	mockService.AssertExpectations(t)
}