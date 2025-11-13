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
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "Yawn (Yet Another Workflow eNgine) - Unified Workflow Platform",
	Long: `Yawn (Yet Another Workflow eNgine) is a unified platform that combines the strengths
of Airflow, Dagster, Windmill, AppSmith, and n8n into one cohesive environment.

Built with an asset-centric philosophy, Yawn enables both engineers and non-engineers to
collaboratively build DAG pipelines, automations, and dashboards in a shared workspace.

Key Features:
• Asset-centric workflow design (focus on outputs, not tasks)
• Unified platform for pipelines, automations, and dashboards
• GitOps integration with code-based workflow definitions
• Collaborative environment for technical and non-technical users
• Python SDK for custom logic and integrations

This CLI provides commands to manage the Yawn platform server and workflows.`,
}

// Execute adds all child commands to the root command. Called by main.main().
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().
		String("config", "", "config file for Yawn platform settings (default is ./config.yaml, ./yawn.yaml, or environment variables)")
}
