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

// Package logger provides structured logging using zerolog with JSON/console
// output.
package logger

import (
	"os"

	"github.com/ditwrd/yawn/api/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLogger creates a zerolog logger with configurable level and JSON/console
// format.
func NewLogger(cfg *config.Config) *zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().
			Err(err).
			Msgf("Invalid log level '%s', using 'info'", cfg.Logger.Level)
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var output zerolog.Logger
	if cfg.Logger.Format == "console" {
		output = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}).With().Timestamp().Logger()
	} else {
		// JSON format (default)
		output = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Add service context
	output = output.With().Str("service", "yawn-api").Logger()

	return &output
}
