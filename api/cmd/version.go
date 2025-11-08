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

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Yawn platform version information",
	Long: `Print the version information for the Yawn (Yet Another Workflow eNgine) platform.

This displays the current version, Git commit information, build details, and
runtime version, which is useful for troubleshooting and ensuring compatibility
with workflow definitions and integrations.

Version information helps with:
• Debugging workflow execution issues
• Ensuring SDK compatibility
• Tracking GitOps deployment history
• Verifying platform capabilities`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Yawn (Yet Another Workflow eNgine)\n")
		fmt.Printf("Platform Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Go Runtime: %s\n", GoVersion)
	},
}