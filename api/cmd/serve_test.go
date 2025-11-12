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
	"testing"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestServeCommand(t *testing.T) {
	// Test that serveCmd exists and has correct properties
	if serveCmd == nil {
		t.Fatal("serveCmd is nil")
	}

	if serveCmd.Use != "serve" {
		t.Errorf("Expected serveCmd.Use to be 'serve', got '%s'", serveCmd.Use)
	}

	if serveCmd.Short != "Start the Yawn workflow platform server" {
		t.Errorf(
			"Expected serveCmd.Short to be 'Start the Yawn workflow platform server', got '%s'",
			serveCmd.Short,
		)
	}

	// Test flags
	portFlag := serveCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("Expected 'port' flag to exist")
	}

	hostFlag := serveCmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Error("Expected 'host' flag to exist")
	}

	devFlag := serveCmd.Flags().Lookup("dev")
	if devFlag == nil {
		t.Error("Expected 'dev' flag to exist")
	}
}

func TestServeCommandFlagBinding(t *testing.T) {
	// Reset viper for this test
	viper.Reset()

	// Test flag binding to viper using existing serveCmd
	if serveCmd == nil {
		t.Fatal("serveCmd is nil")
	}

	// Test setting flag
	err := serveCmd.Flags().Set("port", "9000")
	if err != nil {
		t.Errorf("Failed to set port flag: %v", err)
	}

	// Test that viper gets the value (after binding)
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))

	if viper.GetString("server.port") != "9000" {
		t.Errorf(
			"Expected viper to get port value '9000', got '%s'",
			viper.GetString("server.port"),
		)
	}
}

func TestRootCommandStructure(t *testing.T) {
	// Test root command properties
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "api" {
		t.Errorf("Expected rootCmd.Use to be 'api', got '%s'", rootCmd.Use)
	}

	// Test that serve command is added to root
	found := false

	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "serve" {
			found = true

			break
		}
	}

	if !found {
		t.Error("Expected 'serve' command to be added to root command")
	}
}

func Test_runServe(t *testing.T) {
	t.Parallel()

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := runServe(tt.args.cmd, tt.args.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadCommandConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    *config.Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := loadCommandConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadCommandConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(tt.want, got) {
				t.Errorf(
					"loadCommandConfig() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_printServerInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		config *config.Config
	}

	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			printServerInfo(tt.args.config)
		})
	}
}
