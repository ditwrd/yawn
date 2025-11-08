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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file for Yawn platform settings (default is $HOME/.yawn.yaml, ./config.yaml, or ./yawn.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose logging for workflow execution and platform operations")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level for platform and workflow logging (debug, info, warn, error)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("logger.level", rootCmd.PersistentFlags().Lookup("log-level"))

	// Set default values
	viper.SetDefault("verbose", false)
	viper.SetDefault("logger.level", "info")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Set config file name and search paths
		viper.SetConfigName("config")
		viper.SetConfigName("yawn")
		viper.SetConfigType("yaml")

		// Add search paths in order of preference
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/yawn")

		// Add home directory as last resort
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigName(".yawn")
		}
	}

	// Set environment variable prefix and replacer
	viper.SetEnvPrefix("YAWN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set application defaults
	setAppDefaults()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// setAppDefaults sets default configuration values
func setAppDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./yawn.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.name", "yawn")
	viper.SetDefault("database.user", "yawn")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.ssl_mode", "disable")

	// JWT defaults
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.ttl", 3600) // 1 hour

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
}
