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
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ditwrd/yawn/api/internal/app"
	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Yawn workflow platform server",
	Long: `Start the Yawn platform server to handle workflow execution, automation,
and dashboard requests.

This command launches the central hub where engineers and non-engineers can
collaboratively build DAG pipelines, automations, and dashboards using Yawn's
asset-centric design philosophy.

The server automatically syncs with Git repositories for GitOps workflows,
provides REST APIs for workflow management, and serves the web interface
for visual workflow editing and dashboard creation.

Examples:
  # Start with default configuration
  api serve

  # Start on custom port
  api serve --port 8080

  # Start with development mode for local workflow testing
  api serve --dev --port 3000

  # Start with custom configuration
  api serve --config ./config.yaml

  # Start with specific host binding
  api serve --host 0.0.0.0 --port 8080

Once running, access:
• Web Interface: http://localhost:8080
• API Documentation: http://localhost:8080/docs
• Health Check: http://localhost:8080/health`,
	RunE: runServe,
}

func init() {
	// Add local flags for the serve command
	serveCmd.Flags().StringP("port", "p", "", "Port to run the Yawn platform server on")
	serveCmd.Flags().StringP("host", "H", "", "Host to bind the Yawn platform server to")
	serveCmd.Flags().
		Bool("dev", false, "Enable development mode with hot-reload for workflow testing")

	// Bind flags to viper
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("dev", serveCmd.Flags().Lookup("dev"))
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load configuration (this will use the viper instance from root.go)
	config, err := loadCommandConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create fx application with configuration
	a := app.NewFxAppWithConfig(config)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := a.Start(ctx); err != nil {
			errCh <- fmt.Errorf("failed to start application: %w", err)
		}
	}()

	// Print server information
	printServerInfo(config)

	// Wait for either shutdown signal or error
	select {
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %v, shutting down gracefully...\n", sig)
	case err := <-errCh:
		return err
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop the application
	if err := a.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop application gracefully: %w", err)
	}

	fmt.Println("Server stopped successfully")
	return nil
}

// loadCommandConfig loads configuration for commands, reusing the viper from root.go
func loadCommandConfig() (*config.Config, error) {
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}

// printServerInfo prints server startup information
func printServerInfo(config *config.Config) {
	fmt.Printf("🚀 Starting Yawn Platform Server\n")
	fmt.Printf("📡 Server: http://%s:%s\n", config.Server.Host, config.Server.Port)
	fmt.Printf("💾 Database: %s\n", config.Database.Type)
	if viper.GetBool("dev") {
		fmt.Printf("🔧 Development mode: enabled (hot-reload for workflows)\n")
	}
	fmt.Printf("🌐 Web Interface: http://%s:%s\n", config.Server.Host, config.Server.Port)
	fmt.Printf("📚 API Docs: http://%s:%s/docs\n", config.Server.Host, config.Server.Port)
	fmt.Println("✨ Yawn (Yet Another Workflow eNgine) - Asset-centric workflows ready!")
	fmt.Println("Press Ctrl+C to stop the server")
}
