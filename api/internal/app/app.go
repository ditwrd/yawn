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
package app

import (
	"context"
	"net/http"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/ditwrd/yawn/api/internal/infrastructure/database"
	"github.com/ditwrd/yawn/api/internal/infrastructure/logger"
	"github.com/ditwrd/yawn/api/internal/infrastructure/web"
	"go.uber.org/fx"
	"github.com/labstack/echo/v4"
)

// NewFxApp creates a new fx application with all dependencies
func NewFxApp() *fx.App {
	return fx.New(
		// Provide configuration
		fx.Provide(
			loadConfig,
			logger.NewLogger,
			database.NewDatabase,
			web.NewEcho,
		),

		// Start HTTP server
		fx.Invoke(startServer),

		// Use default fx logger for now
	)
}

// loadConfig loads the application configuration
func loadConfig() (*config.Config, error) {
	return config.LoadConfig("")
}

// startServer starts the HTTP server
func startServer(lc fx.Lifecycle, e *echo.Echo) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(e.Server.Addr); err != nil && err != http.ErrServerClosed {
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

