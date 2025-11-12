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

package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessAllowedOrigins(t *testing.T) {
	tests := []struct {
		name               string
		origins            []string
		enableWildcardPort bool
		expected           []string
	}{
		{
			name:               "No wildcard with wildcard disabled",
			origins:            []string{"http://localhost:3000"},
			enableWildcardPort: false,
			expected:           []string{"http://localhost:3000"},
		},
		{
			name:               "No wildcard with wildcard enabled",
			origins:            []string{"http://localhost:3000"},
			enableWildcardPort: true,
			expected:           []string{"http://localhost:3000"},
		},
		{
			name:               "Wildcard pattern with wildcard enabled",
			origins:            []string{"http://localhost:*"},
			enableWildcardPort: true,
			expected: []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:5173",
				"http://localhost:8000",
				"http://localhost:8080",
				"http://localhost:9000",
			},
		},
		{
			name:               "Wildcard pattern with wildcard disabled",
			origins:            []string{"http://localhost:*"},
			enableWildcardPort: false,
			expected:           []string{"http://localhost:*"},
		},
		{
			name:               "Multiple origins with mixed patterns",
			origins:            []string{"https://example.com", "http://localhost:*", "http://127.0.0.1:3000"},
			enableWildcardPort: true,
			expected: []string{
				"https://example.com",
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:5173",
				"http://localhost:8000",
				"http://localhost:8080",
				"http://localhost:9000",
				"http://127.0.0.1:3000",
			},
		},
		{
			name:               "Invalid wildcard pattern - host wildcard",
			origins:            []string{"http://*:3000"},
			enableWildcardPort: true,
			expected: []string{
				"http://*:3000",
				"http://*:3001",
				"http://*:5173",
				"http://*:8000",
				"http://*:8080",
				"http://*:9000",
			},
		},
		{
			name:               "Empty origins list",
			origins:            []string{},
			enableWildcardPort: true,
			expected:           nil,
		},
		{
			name:               "HTTPS wildcard with different host",
			origins:            []string{"https://dev.example.com:*"},
			enableWildcardPort: true,
			expected: []string{
				"https://dev.example.com:3000",
				"https://dev.example.com:3001",
				"https://dev.example.com:5173",
				"https://dev.example.com:8000",
				"https://dev.example.com:8080",
				"https://dev.example.com:9000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processAllowedOrigins(tt.origins, tt.enableWildcardPort)
			assert.Equal(t, tt.expected, result)
		})
	}
}