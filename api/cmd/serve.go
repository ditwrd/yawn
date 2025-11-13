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
)

// serveCmd represents the serve command.
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
	serveCmd.Flags().
		StringP("port", "p", "", "Port to run the Yawn platform server on")
	serveCmd.Flags().
		StringP("host", "H", "", "Host to bind the Yawn platform server to")
	serveCmd.Flags().
		Bool("dev", false, "Enable development mode with hot-reload for workflow testing")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Get config file path from flag
	configFile, _ := cmd.Flags().GetString("config")

	// Load configuration using the new config system
	config, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override with CLI flags if provided
	if port, _ := cmd.Flags().GetString("port"); port != "" {
		config.Server.Port = port
	}
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		config.Server.Host = host
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
		err := a.Start(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to start application: %w", err)
		}
	}()

	// Get dev flag for printing server info
	dev, _ := cmd.Flags().GetBool("dev")

	// Print server information
	printServerInfo(config, dev)

	// Wait for either shutdown signal or error
	select {
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %v, shutting down gracefully...\n", sig)
	case err := <-errCh:
		return err
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer shutdownCancel()

	// Stop the application
	if err := a.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop application gracefully: %w", err)
	}

	fmt.Println("Server stopped successfully")

	return nil
}


// printServerInfo prints server startup information.
func printServerInfo(config *config.Config, dev bool) {
	fmt.Printf("🚀 Starting Yawn Platform Server\n")
	fmt.Printf(
		"📡 Server: http://%s:%s\n",
		config.Server.Host,
		config.Server.Port,
	)
	fmt.Printf("💾 Database: %s\n", config.Database.Type)

	if dev {
		fmt.Printf("🔧 Development mode: enabled (hot-reload for workflows)\n")
	}

	fmt.Printf(
		"🌐 Web Interface: http://%s:%s\n",
		config.Server.Host,
		config.Server.Port,
	)
	fmt.Printf(
		"📚 API Docs: http://%s:%s/docs\n",
		config.Server.Host,
		config.Server.Port,
	)
	fmt.Println(
		"✨ Yawn (Yet Another Workflow eNgine) - Asset-centric workflows ready!",
	)
	fmt.Println("Press Ctrl+C to stop the server")
}
