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

			// Service providers
			newUserService,

			// Middleware providers
			newAuthMiddleware,
			newAuthzMiddleware,

			// Handler providers
			newAuthHandler,
			newUserHandler,
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

			// Service providers
			newUserService,

			// Middleware providers
			newAuthMiddleware,
			newAuthzMiddleware,

			// Handler providers
			newAuthHandler,
			newUserHandler,
		),

		// Start HTTP server
		fx.Invoke(startServer, setupRoutes),

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

// Repository providers

func newUserRepository(db *gorm.DB) repositories.UserRepository {
	return repositories.NewUserRepository(db)
}

// Service providers

func newUserService(userRepo repositories.UserRepository) services.UserService {
	return services.NewUserService(userRepo)
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

// setupRoutes configures all application routes.
func setupRoutes(
	e *echo.Echo,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) {
	web.SetupRoutes(e, &web.RouterConfig{
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		AuthMiddleware:  authMiddleware,
		AuthzMiddleware: authzMiddleware,
	})
}
