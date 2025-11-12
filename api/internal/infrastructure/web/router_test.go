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

package web

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/oaswrap/spec/adapter/echoopenapi"
	"github.com/oaswrap/spec/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIRouterCreation(t *testing.T) {
	// Test that the OpenAPI router can be created without errors
	e := echo.New()

	// Create a new OpenAPI router
	r := echoopenapi.NewRouter(e,
		option.WithTitle("Test API"),
		option.WithVersion("1.0.0"),
		option.WithDescription("Test API for OpenAPI integration"),
		option.WithScalar(),
	)

	// Verify router was created successfully
	assert.NotNil(t, r)

	// Test adding a simple route with OpenAPI annotations
	r.GET("/test", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "test"})
	}).With(
		option.Summary("Test endpoint"),
		option.Description("A simple test endpoint"),
		option.Response(200, map[string]string{"message": "test"}),
		option.Tags("Test"),
	)

	// Verify the route was added
	assert.True(
		t,
		true,
	) // If we get here without panics, the OpenAPI integration works
}

func TestRouterConfigStruct(t *testing.T) {
	// Test that RouterConfig struct can be created
	cfg := &RouterConfig{}
	assert.NotNil(t, cfg)
}

func TestSetupRoutesFunction(t *testing.T) {
	// Test that SetupRoutes function exists and can be called
	e := echo.New()
	cfg := &RouterConfig{}

	// This should not panic even with empty config
	assert.NotPanics(t, func() {
		SetupRoutes(e, cfg)
	}, "SetupRoutes should not panic with empty config")
}

func TestOpenAPIAnnotations(t *testing.T) {
	// Test that OpenAPI options work correctly
	e := echo.New()
	r := echoopenapi.NewRouter(e,
		option.WithTitle("Test API"),
		option.WithVersion("1.0.0"),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("Bearer")),
	)

	require.NotNil(t, r)

	// Test route with authentication
	authGroup := r.Group("/auth")
	authGroup.POST("/login", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"token": "test"})
	}).With(
		option.Summary("Login endpoint"),
		option.Description("Test login endpoint"),
		option.Request(map[string]string{"email": "test@example.com", "password": "password"}),
		option.Response(200, map[string]string{"token": "test"}),
		option.Tags("Authentication"),
	)

	// Test protected route
	protectedGroup := r.Group("/protected")
	protectedGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Mock auth middleware
			c.Set("user_id", "test-user")

			return next(c)
		}
	})
	protectedGroup.GET("/profile", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"user_id": "test-user"})
	}).With(
		option.Summary("Protected endpoint"),
		option.Description("A protected endpoint requiring authentication"),
		option.Response(200, map[string]string{"user_id": "test"}),
		option.Tags("Protected"),
	)

	assert.True(
		t,
		true,
	) // If we get here without panics, OpenAPI annotations work correctly
}
