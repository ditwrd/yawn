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

// Package app provides application bootstrap and dependency injection using
// uber-go/fx.
package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
	"github.com/ditwrd/yawn/api/internal/domain/services"
	"github.com/ditwrd/yawn/api/internal/infrastructure/database"
	"github.com/ditwrd/yawn/api/internal/infrastructure/logger"
	"github.com/ditwrd/yawn/api/internal/infrastructure/web"
	"github.com/ditwrd/yawn/api/internal/infrastructure/web/middleware"
	"github.com/ditwrd/yawn/api/internal/interfaces/handlers"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// NewFxApp creates a new fx application with all dependencies configured.
func NewFxApp() *fx.App {
	return fx.New(
		// Provide configuration
		fx.Provide(
			loadConfig,
			logger.NewLogger,
			database.NewDatabase,
			web.NewEcho,

			// Infrastructure providers
			newJWTService,
			newPasswordService,

			// Repository providers
			newUserRepository,
			newProjectRepository,
			newAssetRepository,
			newPipelineRepository,
			newRepositoryRepository,

			// Service providers
			newUserService,
			newAuthService,
			newProjectService,
			newAssetService,
			newPipelineService,
			newGitOpsService,

			// Middleware providers
			newAuthMiddleware,
			newAuthzMiddleware,

			// Handler providers
			newAuthHandler,
			newUserHandler,
			newProjectHandler,
			newAssetHandler,
			newPipelineHandler,
			newGitOpsHandler,
		),

		// Start HTTP server
		fx.Invoke(startServer, setupRoutes),

		// Use default fx logger for now
	)
}

// NewFxAppWithConfig creates a new fx application using provided configuration.
func NewFxAppWithConfig(cfg *config.Config) *fx.App {
	return fx.New(
		// Provide configuration
		fx.Provide(
			func() (*config.Config, error) { return cfg, nil },
			logger.NewLogger,
			database.NewDatabase,
			web.NewEcho,

			// Infrastructure providers
			newJWTService,
			newPasswordService,

			// Repository providers
			newUserRepository,
			newProjectRepository,
			newAssetRepository,
			newPipelineRepository,
			newRepositoryRepository,

			// Service providers
			newUserService,
			newAuthService,
			newProjectService,
			newAssetService,
			newPipelineService,
			newGitOpsService,

			// Middleware providers
			newAuthMiddleware,
			newAuthzMiddleware,

			// Handler providers
			newAuthHandler,
			newUserHandler,
			newProjectHandler,
			newAssetHandler,
			newPipelineHandler,
			newGitOpsHandler,
		),

		// Start HTTP server
		fx.Invoke(setupRoutes, startServer),

		// Use default fx logger for now
	)
}

// loadConfig loads the application configuration from default locations.
func loadConfig() (*config.Config, error) {
	return config.LoadConfig("")
}

// startServer starts the HTTP server with graceful shutdown support.
func startServer(lc fx.Lifecycle, e *echo.Echo) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				err := e.Start(e.Server.Addr)
				if err != nil &&
					!errors.Is(err, http.ErrServerClosed) {
					e.Logger.Fatalf("Failed to start server: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})
}

// Infrastructure providers

func newJWTService() services.JWTService {
	return services.NewJWTService(&services.JWTConfig{
		AccessSecret:  "your-access-secret-key",  // Use config in production
		RefreshSecret: "your-refresh-secret-key", // Use config in production
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour, // 7 days
		Issuer:        "yawn-api",
		Audience:      "yawn-client",
	})
}

func newPasswordService() services.PasswordService {
	return services.NewPasswordService(&services.PasswordConfig{
		Memory:              19456, // 19 MiB in KiB
		Iterations:          2,
		Parallelism:         1,
		SaltLength:          16,
		KeyLength:           32,
		MinLength:           8,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
	})
}

// Logger adapter for repositories that expect Info(msg string, fields ...any).
type repoLoggerAdapter struct {
	logger *zerolog.Logger
}

func (r *repoLoggerAdapter) Info(msg string, fields ...any) {
	r.logger.Info().Fields(fields).Msg(msg)
}

// Repository providers

func newUserRepository(db *gorm.DB) repositories.UserRepository {
	return repositories.NewUserRepository(db)
}

func newProjectRepository(db *gorm.DB) repositories.ProjectRepository {
	return repositories.NewProjectRepository(db)
}

func newAssetRepository(
	db *gorm.DB,
	logger *zerolog.Logger,
) repositories.AssetRepository {
	return repositories.NewAssetRepository(db, &repoLoggerAdapter{logger: logger})
}

func newPipelineRepository(
	db *gorm.DB,
	logger *zerolog.Logger,
) repositories.PipelineRepository {
	return repositories.NewPipelineRepository(
		db,
		&repoLoggerAdapter{logger: logger},
	)
}

func newRepositoryRepository(
	db *gorm.DB,
	logger *zerolog.Logger,
) repositories.RepositoryRepository {
	return repositories.NewRepositoryRepository(
		db,
		&repoLoggerAdapter{logger: logger},
	)
}

// Service providers

func newUserService(userRepo repositories.UserRepository) services.UserService {
	return services.NewUserService(userRepo)
}

func newAuthService(
	userRepo repositories.UserRepository,
	projectRepo repositories.ProjectRepository,
	logger *zerolog.Logger,
	jwtService services.JWTService,
	passwordService services.PasswordService,
) services.AuthService {
	return services.NewAuthService(
		userRepo,
		projectRepo,
		logger,
		[]byte("your-jwt-secret-key"),
		"",
		"",
	)
}

func newProjectService(
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
) services.ProjectService {
	return services.NewProjectService(projectRepo, userRepo)
}

func newAssetService(
	assetRepo repositories.AssetRepository,
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
	logger *zerolog.Logger,
) services.AssetService {
	return services.NewAssetService(assetRepo, projectRepo, userRepo, logger)
}

func newPipelineService(
	pipelineRepo repositories.PipelineRepository,
	projectRepo repositories.ProjectRepository,
	assetRepo repositories.AssetRepository,
	userRepo repositories.UserRepository,
	logger *zerolog.Logger,
) services.PipelineService {
	return services.NewPipelineService(
		pipelineRepo,
		projectRepo,
		assetRepo,
		userRepo,
		logger,
	)
}

func newGitOpsService(
	repositoryRepo repositories.RepositoryRepository,
	pipelineRepo repositories.PipelineRepository,
	projectRepo repositories.ProjectRepository,
	logger *zerolog.Logger,
) services.GitOpsService {
	return services.NewGitOpsService(
		repositoryRepo,
		pipelineRepo,
		projectRepo,
		logger,
	)
}

// Middleware providers

func newAuthMiddleware(
	jwtService services.JWTService,
	logger *zerolog.Logger,
) *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(jwtService, logger)
}

func newAuthzMiddleware(
	logger *zerolog.Logger,
) *middleware.AuthorizationMiddleware {
	return middleware.NewAuthorizationMiddleware(logger)
}

// Handler providers

func newAuthHandler(
	userService services.UserService,
	jwtService services.JWTService,
	passwordService services.PasswordService,
	logger *zerolog.Logger,
) *handlers.AuthHandler {
	return handlers.NewAuthHandler(
		userService,
		jwtService,
		passwordService,
		logger,
	)
}

func newUserHandler(
	userService services.UserService,
	logger *zerolog.Logger,
) *handlers.UserHandler {
	return handlers.NewUserHandler(userService, logger)
}

func newProjectHandler(
	projectService services.ProjectService,
	userService services.UserService,
	logger *zerolog.Logger,
) *handlers.ProjectHandler {
	return handlers.NewProjectHandler(projectService, userService, logger)
}

func newAssetHandler(
	assetService services.AssetService,
	logger *zerolog.Logger,
) *handlers.AssetHandler {
	return handlers.NewAssetHandler(assetService, logger)
}

func newPipelineHandler(
	pipelineService services.PipelineService,
	logger *zerolog.Logger,
) *handlers.PipelineHandler {
	return handlers.NewPipelineHandler(pipelineService, logger)
}

func newGitOpsHandler(
	gitOpsService services.GitOpsService,
	logger *zerolog.Logger,
) *handlers.GitOpsHandler {
	return handlers.NewGitOpsHandler(gitOpsService, logger)
}

// setupRoutes configures all application routes.
func setupRoutes(
	e *echo.Echo,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	assetHandler *handlers.AssetHandler,
	pipelineHandler *handlers.PipelineHandler,
	gitOpsHandler *handlers.GitOpsHandler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) {
	web.SetupRoutes(e, &web.RouterConfig{
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		ProjectHandler:  projectHandler,
		AssetHandler:    assetHandler,
		PipelineHandler: pipelineHandler,
		GitOpsHandler:   gitOpsHandler,
		AuthMiddleware:  authMiddleware,
		AuthzMiddleware: authzMiddleware,
	})
}
