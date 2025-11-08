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

// Package services provides secure password hashing and validation.
package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordService provides password hashing and validation operations.
type PasswordService interface {
	HashPassword(password string) (string, error)
	ValidatePassword(password, hash string) (bool, error)
	CheckPasswordStrength(password string) error
}

// PasswordConfig holds password hashing configuration.
//
// Argon2id parameters based on OWASP recommendations (2024/2025):
// - Memory: 19 MiB (minimum recommended)
// - Iterations: 2 (minimum recommended)
// - Parallelism: 1 (minimum recommended)
// - Salt length: 16 bytes (128 bits)
// - Hash length: 32 bytes (256 bits)
type PasswordConfig struct {
	// Argon2id parameters
	Memory      uint32 `json:"memory"`      // Memory in KiB (19 MiB = 19456 KiB)
	Iterations  uint32 `json:"iterations"`  // Number of iterations (time)
	Parallelism uint8  `json:"parallelism"` // Number of parallel threads
	SaltLength  uint32 `json:"salt_length"` // Salt length in bytes
	KeyLength   uint32 `json:"key_length"`  // Hash length in bytes

	// Password strength validation
	MinLength           int  `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
}

// passwordService implements PasswordService.
type passwordService struct {
	config *PasswordConfig
}

// NewPasswordService creates a new password service.
func NewPasswordService(config *PasswordConfig) PasswordService {
	return &passwordService{
		config: config,
	}
}

func (s *passwordService) HashPassword(password string) (string, error) {
	// Generate a cryptographically secure random salt
	salt := make([]byte, s.config.SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate Argon2id hash
	hash := argon2.IDKey([]byte(password), salt, s.config.Iterations, s.config.Memory, s.config.Parallelism, s.config.KeyLength)

	// Format: $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
	// Using PHC (Password Hashing Competition) string format
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	// Construct the PHC string format
	passwordHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		s.config.Memory, s.config.Iterations, s.config.Parallelism,
		saltBase64, hashBase64)

	return passwordHash, nil
}

func (s *passwordService) ValidatePassword(password, hash string) (bool, error) {
	// Parse the PHC string format
	// Expected format: $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
	memory, iterations, parallelism, salt, decodedHash, err := s.parseArgon2Hash(hash)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash: %w", err)
	}

	// Generate hash of the provided password using the same parameters
	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(decodedHash)))

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(decodedHash, computedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func (s *passwordService) CheckPasswordStrength(password string) error {
	if len(password) < s.config.MinLength {
		return fmt.Errorf("password must be at least %d characters long", s.config.MinLength)
	}

	if s.config.RequireUppercase && !s.hasUppercase(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if s.config.RequireLowercase && !s.hasLowercase(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if s.config.RequireNumbers && !s.hasNumber(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if s.config.RequireSpecialChars && !s.hasSpecialChar(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	if s.isCommonPassword(password) {
		return fmt.Errorf("password is too common, please choose a stronger one")
	}

	return nil
}

func (s *passwordService) hasUppercase(password string) bool {
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			return true
		}
	}
	return false
}

func (s *passwordService) hasLowercase(password string) bool {
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			return true
		}
	}
	return false
}

func (s *passwordService) hasNumber(password string) bool {
	for _, char := range password {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

func (s *passwordService) hasSpecialChar(password string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, char := range password {
		for _, special := range specialChars {
			if char == special {
				return true
			}
		}
	}
	return false
}

func (s *passwordService) isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "123456789", "12345678", "12345",
		"1234567", "1234567890", "qwerty", "abc123", "password123",
		"admin", "letmein", "welcome", "monkey", "login",
	}

	passwordLower := ""
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			passwordLower += string(char + 32)
		} else {
			passwordLower += string(char)
		}
	}

	for _, common := range commonPasswords {
		if passwordLower == common {
			return true
		}
	}

	return false
}

// parseArgon2Hash parses an Argon2id PHC string and returns its components.
// Expected format: $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
func (s *passwordService) parseArgon2Hash(hash string) (memory uint32, iterations uint32, parallelism uint8, salt []byte, decodedHash []byte, err error) {
	// Basic validation of hash format
	if len(hash) < 10 || hash[:10] != "$argon2id$" {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid hash format: not an Argon2id hash")
	}

	// Split the hash into components using strings.Split for simplicity
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid hash format: unexpected structure, got %d parts: %v", len(parts), parts)
	}

	// Parse parameters: m=<memory>,t=<iterations>,p=<parallelism>
	params := parts[3]
	memory, iterations, parallelism, err = parseArgon2Params(params)
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	decodedHash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return memory, iterations, parallelism, salt, decodedHash, nil
}


// parseArgon2Params parses the parameters string (m=<memory>,t=<iterations>,p=<parallelism>)
func parseArgon2Params(params string) (memory uint32, iterations uint32, parallelism uint8, err error) {
	// Split by comma
	paramParts := strings.Split(params, ",")
	if len(paramParts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid parameters format")
	}

	// Parse memory parameter
	memoryStr := paramParts[0]
	if len(memoryStr) < 2 || memoryStr[:2] != "m=" {
		return 0, 0, 0, fmt.Errorf("invalid memory parameter")
	}
	memory64, err := strconv.ParseUint(memoryStr[2:], 10, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid memory value: %w", err)
	}
	memory = uint32(memory64)

	// Parse iterations parameter
	iterationsStr := paramParts[1]
	if len(iterationsStr) < 2 || iterationsStr[:2] != "t=" {
		return 0, 0, 0, fmt.Errorf("invalid iterations parameter")
	}
	iterations64, err := strconv.ParseUint(iterationsStr[2:], 10, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid iterations value: %w", err)
	}
	iterations = uint32(iterations64)

	// Parse parallelism parameter
	parallelismStr := paramParts[2]
	if len(parallelismStr) < 2 || parallelismStr[:2] != "p=" {
		return 0, 0, 0, fmt.Errorf("invalid parallelism parameter")
	}
	parallelism64, err := strconv.ParseUint(parallelismStr[2:], 10, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid parallelism value: %w", err)
	}
	parallelism = uint8(parallelism64)

	return memory, iterations, parallelism, nil
}

// DefaultPasswordConfig returns default password configuration using Argon2id with OWASP 2024/2025 recommendations.
//
// Security parameters:
// - Memory: 19456 KiB (19 MiB) - OWASP minimum recommendation
// - Iterations: 2 - OWASP minimum recommendation
// - Parallelism: 1 - OWASP minimum recommendation
// - Salt length: 16 bytes (128 bits) - Strong entropy
// - Key length: 32 bytes (256 bits) - Strong hash output
//
// These parameters provide approximately 100ms hash time on modern hardware,
// balancing security and performance while preventing GPU-based attacks.
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		// Argon2id parameters (OWASP 2024/2025 recommendations)
		Memory:      19456, // 19 MiB in KiB
		Iterations:  2,     // Number of iterations
		Parallelism: 1,     // Number of parallel threads
		SaltLength:  16,    // 16 bytes (128 bits)
		KeyLength:   32,    // 32 bytes (256 bits)

		// Password strength validation
		MinLength:           8,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
	}
}
