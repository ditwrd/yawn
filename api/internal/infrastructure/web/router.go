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

// Package web provides HTTP routing configuration for the Echo framework.
//
// Configures all API routes with appropriate middleware for authentication,
// authorization, and request handling. Supports modular route organization.
// Uses oaswrap/spec echoopenapi adapter for automatic OpenAPI specification
// generation.
package web

import (
	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/infrastructure/web/middleware"
	"github.com/ditwrd/yawn/api/internal/interfaces/dto"
	"github.com/ditwrd/yawn/api/internal/interfaces/handlers"
	"github.com/labstack/echo/v4"
	"github.com/oaswrap/spec/adapter/echoopenapi"
	"github.com/oaswrap/spec/option"
)

// RouterConfig holds all the dependencies required for setting up application
// routes.
//
// This struct provides a clean way to inject all necessary handlers and
// middleware
// into the router configuration, ensuring proper separation of concerns.
type RouterConfig struct {
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	ProjectHandler  *handlers.ProjectHandler
	AssetHandler    *handlers.AssetHandler
	PipelineHandler *handlers.PipelineHandler
	GitOpsHandler   *handlers.GitOpsHandler
	AuthMiddleware  *middleware.AuthMiddleware
	AuthzMiddleware *middleware.AuthorizationMiddleware
}

// SetupRoutesWithOpenAPI configures all application routes with OpenAPI
// documentation generation.
//
// Sets up authentication and user management endpoints with proper
// security middleware. Organizes routes by feature with versioning support.
// Uses oaswrap/spec echoopenapi adapter for automatic OpenAPI specification
// generation.
func SetupRoutesWithOpenAPI(e *echo.Echo, cfg *RouterConfig) {
	// Create a new OpenAPI router with comprehensive configuration
	r := echoopenapi.NewRouter(
		e,
		option.WithTitle("Yawn GitOps CI/CD Platform API"),
		option.WithVersion("1.0.0"),
		option.WithDescription(
			"A comprehensive GitOps CI/CD platform API for managing projects, assets, repositories, and deployment pipelines. Features include user authentication, project-based collaboration, pipeline orchestration, and GitOps operations.",
		),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("Bearer")),
		option.WithScalar(), // Use Scalar UI for documentation
	)

	// Health check endpoint (no auth required)
	r.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	}).With(
		option.Summary("Check API health status"),
		option.Description("Returns the health status of the API service"),
		option.Tags("Health"),
	)

	// API v1 routes
	v1 := r.Group("/api/v1")
	//
	// Authentication routes (no auth required)
	authGroup := v1.Group("/auth")

	authGroup.POST("/register", cfg.AuthHandler.Register).With(
		option.Summary("Register a new user"),
		option.Description("Creates a new user account with the provided email and password"),
		option.Request(new(dto.RegisterRequest)),
		option.Response(201, new(dto.RegisterResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(409, new(dto.ErrorResponse)),
		option.Tags("Authentication"),
	)
	authGroup.POST("/login", cfg.AuthHandler.Login).With(
		option.Summary("User login"),
		option.Description("Authenticates a user with email and password, returning JWT tokens"),
		option.Request(new(dto.LoginRequest)),
		option.Response(200, new(dto.LoginResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Tags("Authentication"),
	)
	authGroup.POST("/refresh", cfg.AuthHandler.Refresh).With(
		option.Summary("Refresh access token"),
		option.Description("Refreshes an access token using a valid refresh token"),
		option.Request(new(dto.RefreshRequest)),
		option.Response(200, new(dto.RefreshResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Authentication"),
	)
	authGroup.POST("/logout", cfg.AuthHandler.Logout).With(
		option.Summary("User logout"),
		option.Description("Logs out a user by invalidating their access token"),
		option.Request(new(dto.LogoutRequest)),
		option.Response(200, new(dto.LogoutResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Authentication"),
	)

	// User management routes (auth required)
	userGroup := v1.Group("/users",
		cfg.AuthMiddleware.RequireAuth(),
		cfg.AuthzMiddleware.RequireAdmin(),
	).With(option.GroupSecurity("bearerAuth"))

	userGroup.GET("", cfg.UserHandler.ListUsers).With(
		option.Summary("List all users"),
		option.Description("Retrieves a paginated list of all users in the system (admin only)"),
		option.Response(200, new(dto.UserListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Tags("User Management"),
	)

	userGroup.DELETE("/:id", cfg.UserHandler.DeleteUser).With(
		option.Summary("Delete user"),
		option.Description("Deletes a user account by ID - admin only"),
		option.Request(new(dto.UserRequestsWithID)),
		option.Response(200, new(dto.UserDeleteResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("User Management"),
	)

	// Self or admin routes
	userSelfOrAdmin := v1.Group("/users",
		cfg.AuthMiddleware.RequireAuth(),
		cfg.AuthzMiddleware.RequireRoleOrOwnership(models.UserRoleAdmin),
	).With(option.GroupSecurity("bearerAuth"))

	userSelfOrAdmin.GET("/:id", cfg.UserHandler.GetUser).With(
		option.Summary("Get user by ID"),
		option.Description("Retrieves user information by ID (self or admin only)"),
		option.Request(new(dto.UserRequestsWithID)),
		option.Response(200, new(dto.UserResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("User Management"),
	)
	userSelfOrAdmin.PUT("/:id", cfg.UserHandler.UpdateUser).With(
		option.Summary("Update user"),
		option.Description("Updates user information (self or admin only)"),
		option.Request(new(dto.UpdateUserRequest)),
		option.Response(200, new(dto.UserResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("User Management"),
	)
	//
	// Project management routes (auth required)
	projectGroup := v1.Group("/projects",
		cfg.AuthMiddleware.RequireAuth(),
	).With(option.GroupSecurity("bearerAuth"))

	// Project CRUD routes
	projectGroup.POST("", cfg.ProjectHandler.CreateProject).With(
		option.Summary("Create project"),
		option.Description("Creates a new project with the provided details"),
		option.Request(new(dto.CreateProjectRequest)),
		option.Response(201, new(dto.ProjectResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	projectGroup.GET("", cfg.ProjectHandler.ListProjects).With(
		option.Summary("List projects"),
		option.Description("Retrieves a list of projects accessible to the authenticated user"),
		option.Response(200, new(dto.ProjectListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	projectGroup.GET("/:id", cfg.ProjectHandler.GetProject).With(
		option.Summary("Get project by ID"),
		option.Description("Retrieves project information by ID"),
		option.Request(new(dto.ProjectRequestsWithID)),
		option.Response(200, new(dto.ProjectResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	projectGroup.PUT("/:id", cfg.ProjectHandler.UpdateProject).With(
		option.Summary("Update project"),
		option.Description("Updates project information by ID"),
		option.Request(new(dto.UpdateProjectRequest)),
		option.Response(200, new(dto.ProjectResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	projectGroup.DELETE("/:id", cfg.ProjectHandler.DeleteProject).With(
		option.Summary("Delete project"),
		option.Description("Deletes a project by ID"),
		// option.Response(204),
		option.Request(new(dto.ProjectRequestsWithID)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	//
	// Project member management routes
	projectMemberGroup := v1.Group("/projects/:id/members",
		cfg.AuthMiddleware.RequireAuth(),
	).With(option.GroupSecurity("bearerAuth"))
	projectMemberGroup.POST("", cfg.ProjectHandler.AddProjectMember).With(
		option.Summary("Add project member"),
		option.Description("Adds a new member to a project"),
		option.Request(new(dto.AddProjectMemberRequest)),
		option.Response(201, new(dto.ProjectMemberResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	projectMemberGroup.GET("", cfg.ProjectHandler.ListProjectMembers).With(
		option.Summary("List project members"),
		option.Description("Retrieves a list of members for a specific project"),
		option.Request(new(dto.ProjectRequestsWithID)),
		option.Response(200, new(dto.ProjectMemberListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Project Management"),
	)
	projectMemberGroup.PUT("/:memberId", cfg.ProjectHandler.UpdateProjectMemberRole).
		With(
			option.Summary("Update project member role"),
			option.Description("Updates the role of a project member"),
			option.Request(new(dto.UpdateProjectMemberRequest)),
			option.Response(200, new(dto.ProjectMemberResponse)),
			option.Response(400, new(dto.ValidationErrorResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("Project Management"),
		)
	projectMemberGroup.DELETE("/:memberId", cfg.ProjectHandler.RemoveProjectMember).
		With(
			option.Summary("Remove project member"),
			option.Description("Removes a member from a project"),
			// option.Response(204),
			option.Request(new(dto.ProjectMembersRequests)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("Project Management"),
		)
	// Asset management routes (auth required)
	assetGroup := v1.Group("/assets",
		cfg.AuthMiddleware.RequireAuth(),
	).With(option.GroupSecurity("bearerAuth"))

	// Asset CRUD routes
	assetGroup.POST("", cfg.AssetHandler.CreateAsset).With(
		option.Summary("Create asset"),
		option.Description("Creates a new asset with the provided details"),
		option.Request(new(dto.CreateAssetRequest)),
		option.Response(201, new(dto.AssetResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	assetGroup.GET("", cfg.AssetHandler.ListAssets).With(
		option.Summary("List assets"),
		option.Description("Retrieves a list of assets accessible to the authenticated user"),
		option.Response(200, new(dto.AssetListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	assetGroup.GET("/search", cfg.AssetHandler.SearchAssets).With(
		option.Summary("Search assets"),
		option.Description("Searches for assets based on query parameters"),
		option.Request(new(dto.AssetSearchRequest)),
		option.Response(200, new(dto.AssetListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	assetGroup.GET("/:id", cfg.AssetHandler.GetAsset).With(
		option.Summary("Get asset by ID"),
		option.Description("Retrieves asset information by ID"),
		option.Request(new(dto.AssetRequestsWithID)),
		option.Response(200, new(dto.AssetResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	assetGroup.GET("/:id/versions", cfg.AssetHandler.GetAssetVersionHistory).With(
		option.Summary("Get asset version history"),
		option.Description("Retrieves the version history of an asset"),
		option.Request(new(dto.AssetVersionHistoryRequest)),
		option.Response(200, new(dto.AssetVersionHistoryResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	assetGroup.PUT("/:id", cfg.AssetHandler.UpdateAsset).With(
		option.Summary("Update asset"),
		option.Description("Updates asset information by ID"),
		option.Request(new(dto.UpdateAssetRequest)),
		option.Response(200, new(dto.AssetResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	assetGroup.DELETE("/:id", cfg.AssetHandler.DeleteAsset).With(
		option.Summary("Delete asset"),
		option.Description("Deletes an asset by ID"),
		option.Request(new(dto.AssetRequestsWithID)),
		// option.Response(204),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Asset Management"),
	)
	// Pipeline management routes (auth required)
	pipelineGroup := v1.Group("/pipelines",
		cfg.AuthMiddleware.RequireAuth(),
	).With(option.GroupSecurity("bearerAuth"))

	// Pipeline CRUD routes
	pipelineGroup.POST("", cfg.PipelineHandler.CreatePipeline).With(
		option.Summary("Create pipeline"),
		option.Description("Creates a new pipeline with the provided configuration"),
		option.Request(new(dto.CreatePipelineRequest)),
		option.Response(201, new(dto.PipelineResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Pipeline Management"),
	)
	pipelineGroup.GET("", cfg.PipelineHandler.ListPipelines).With(
		option.Summary("List pipelines"),
		option.Description("Retrieves a list of pipelines accessible to the authenticated user"),
		option.Response(200, new(dto.PipelineListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Pipeline Management"),
	)
	pipelineGroup.GET("/search", cfg.PipelineHandler.SearchPipelines).With(
		option.Summary("Search pipelines"),
		option.Description("Searches for pipelines based on query parameters"),
		option.Request(new(dto.PipelineSearchRequest)),
		option.Response(200, new(dto.PipelineSearchResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("Pipeline Management"),
	)
	pipelineGroup.GET("/:id", cfg.PipelineHandler.GetPipeline).With(
		option.Summary("Get pipeline by ID"),
		option.Description("Retrieves pipeline information by ID"),
		option.Request(new(dto.PipelineRequestsWithID)),
		option.Response(200, new(dto.PipelineResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Pipeline Management"),
	)
	pipelineGroup.PUT("/:id", cfg.PipelineHandler.UpdatePipeline).With(
		option.Summary("Update pipeline"),
		option.Description("Updates pipeline configuration by ID"),
		option.Request(new(dto.UpdatePipelineRequest)),
		option.Response(200, new(dto.PipelineResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Pipeline Management"),
	)
	pipelineGroup.DELETE("/:id", cfg.PipelineHandler.DeletePipeline).With(
		option.Summary("Delete pipeline"),
		option.Description("Deletes a pipeline by ID"),
		option.Request(new(dto.PipelineRequestsWithID)),
		option.Response(200, new(dto.PipelineDeleteResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Pipeline Management"),
	)

	// Pipeline execution routes
	pipelineGroup.POST("/:id/trigger", cfg.PipelineHandler.TriggerExecution).With(
		option.Summary("Trigger pipeline execution"),
		option.Description("Manually triggers the execution of a pipeline"),
		option.Request(new(dto.PipelineTriggerRequest)),
		option.Response(202, new(dto.TriggerExecutionResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Pipeline Execution"),
	)
	pipelineGroup.GET("/:id/executions", cfg.PipelineHandler.GetExecutions).With(
		option.Summary("Get pipeline executions"),
		option.Description("Retrieves the execution history of a pipeline"),
		option.Request(new(dto.PipelineExecutionsRequest)),
		option.Response(200, new(dto.PipelineExecutionListResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(403, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
		option.Tags("Pipeline Execution"),
	)
	pipelineGroup.PUT("/:id/status", cfg.PipelineHandler.UpdatePipelineStatus).
		With(
			option.Summary("Update pipeline status"),
			option.Description("Updates the status of a pipeline"),
			option.Request(new(dto.UpdatePipelineStatusRequest)),
			option.Response(200, new(dto.PipelineResponse)),
			option.Response(400, new(dto.ValidationErrorResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("Pipeline Management"),
		)
	pipelineGroup.DELETE("/:id/executions/:executionId", cfg.PipelineHandler.CancelExecution).
		With(
			option.Summary("Cancel pipeline execution"),
			option.Description("Cancels a running pipeline execution"),
			option.Request(new(dto.PipelineExecutionCancelRequest)),
			option.Response(200, new(dto.CancelExecutionResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Response(409, new(dto.ErrorResponse)),
			option.Tags("Pipeline Execution"),
		)

	// Pipeline visualization routes
	pipelineGroup.GET("/:id/dependencies", cfg.PipelineHandler.GetDependencyGraph).
		With(
			option.Summary("Get pipeline dependency graph"),
			option.Description("Retrieves the dependency graph of a pipeline"),
			option.Request(new(dto.PipelineDependenciesRequest)),
			option.Response(200, new(dto.DependencyGraphResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("Pipeline Visualization"),
		)
	// GitOps management routes (auth required)
	gitOpsGroup := v1.Group("/gitops",
		cfg.AuthMiddleware.RequireAuth(),
	).With(option.GroupSecurity("bearerAuth"))

	// GitOps operations
	gitOpsGroup.POST("/repositories/:repoId/sync", cfg.GitOpsHandler.SyncRepository).
		With(
			option.Summary("Sync repository"),
			option.Description("Manually triggers a GitOps sync for a repository"),
			option.Request(new(dto.GitOpsRepositorySyncRequest)),
			option.Response(202, new(dto.GitOpsSyncResponse)),
			option.Response(400, new(dto.ValidationErrorResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("GitOps"),
		)
	gitOpsGroup.GET("/repositories/:repoId/status", cfg.GitOpsHandler.GetSyncStatus).
		With(
			option.Summary("Get repository sync status"),
			option.Description("Retrieves the current sync status of a repository"),
			option.Request(new(dto.GitOpsRepositoryStatusRequest)),
			option.Response(200, new(dto.GitOpsStatusResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("GitOps"),
		)
	gitOpsGroup.GET("/repositories/:repoId/pending", cfg.GitOpsHandler.GetPendingSync).
		With(
			option.Summary("Get pending sync operations"),
			option.Description("Retrieves pending GitOps sync operations for a repository"),
			option.Request(new(dto.GitOpsPendingSyncRequest)),
			option.Response(200, new(dto.PendingSyncResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("GitOps"),
		)
	gitOpsGroup.POST("/repositories/:repoId/validate", cfg.GitOpsHandler.ValidateRepository).
		With(
			option.Summary("Validate repository"),
			option.Description("Validates a GitOps repository configuration"),
			option.Request(new(dto.GitOpsRepositoryValidateRequest)),
			option.Response(200, new(dto.ValidationResponse)),
			option.Response(400, new(dto.ValidationErrorResponse)),
			option.Response(401, new(dto.ErrorResponse)),
			option.Response(403, new(dto.ErrorResponse)),
			option.Response(404, new(dto.ErrorResponse)),
			option.Tags("GitOps"),
		)
	gitOpsGroup.POST("/webhooks", cfg.GitOpsHandler.HandleWebhook).With(
		option.Summary("Handle GitOps webhook"),
		option.Description("Processes incoming GitOps webhooks from Git providers"),
		option.Request(new(dto.WebhookRequest)),
		option.Response(200, new(dto.WebhookResponse)),
		option.Response(400, new(dto.ValidationErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Tags("GitOps"),
	)
}

// SetupRoutes maintains backward compatibility with the original function
// signature.
// This function now calls the OpenAPI-enabled version.
func SetupRoutes(e *echo.Echo, cfg *RouterConfig) {
	SetupRoutesWithOpenAPI(e, cfg)
}
