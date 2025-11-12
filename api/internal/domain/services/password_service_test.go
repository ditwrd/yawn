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

package services

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordService_Argon2Hashing(t *testing.T) {
	// Test with default OWASP configuration
	config := DefaultPasswordConfig()
	service := NewPasswordService(config)

	t.Run("hash password successfully", func(t *testing.T) {
		password := "SecurePass123!"
		hash, err := service.HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$argon2id$v=19$"))

		// Verify hash contains expected parameters
		assert.Contains(t, hash, "m=19456") // Memory
		assert.Contains(t, hash, "t=2")     // Iterations
		assert.Contains(t, hash, "p=1")     // Parallelism
	})

	t.Run("validate correct password", func(t *testing.T) {
		password := "SecurePass123!"
		hash, err := service.HashPassword(password)
		require.NoError(t, err)

		valid, err := service.ValidatePassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("reject incorrect password", func(t *testing.T) {
		password := "SecurePass123!"
		wrongPassword := "WrongPass123!"
		hash, err := service.HashPassword(password)
		require.NoError(t, err)

		valid, err := service.ValidatePassword(wrongPassword, hash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		password1 := "SecurePass123!"
		password2 := "AnotherPass456!"

		hash1, err1 := service.HashPassword(password1)
		hash2, err2 := service.HashPassword(password2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run(
		"same password produces different hashes (different salts)",
		func(t *testing.T) {
			password := "SecurePass123!"

			hash1, err1 := service.HashPassword(password)
			hash2, err2 := service.HashPassword(password)

			require.NoError(t, err1)
			require.NoError(t, err2)
			assert.NotEqual(
				t,
				hash1,
				hash2,
			) // Different salts should produce different hashes

			// But both should validate correctly
			valid1, err1 := service.ValidatePassword(password, hash1)
			valid2, err2 := service.ValidatePassword(password, hash2)

			require.NoError(t, err1)
			require.NoError(t, err2)
			assert.True(t, valid1)
			assert.True(t, valid2)
		},
	)

	t.Run("validate invalid hash format", func(t *testing.T) {
		password := "SecurePass123!"
		invalidHash := "$invalid$hash$format"

		valid, err := service.ValidatePassword(password, invalidHash)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "invalid hash format")
	})

	t.Run("validate bcrypt hash (should fail)", func(t *testing.T) {
		password := "SecurePass123!"
		// This is a sample bcrypt hash format
		bcryptHash := "$2a$12$N9qo8uLOickgx2ZMRZoMye.Ijdjr3VGTNxBVnLd8l1QJ7kK1XeYjS"

		valid, err := service.ValidatePassword(password, bcryptHash)
		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "not an Argon2id hash")
	})
}

func TestPasswordService_Configuration(t *testing.T) {
	t.Run("default configuration values", func(t *testing.T) {
		config := DefaultPasswordConfig()

		// Argon2id parameters
		assert.Equal(t, uint32(19456), config.Memory)  // 19 MiB
		assert.Equal(t, uint32(2), config.Iterations)  // 2 iterations
		assert.Equal(t, uint8(1), config.Parallelism)  // 1 parallel thread
		assert.Equal(t, uint32(16), config.SaltLength) // 16 bytes salt
		assert.Equal(t, uint32(32), config.KeyLength)  // 32 bytes hash

		// Password strength requirements
		assert.Equal(t, 8, config.MinLength)
		assert.True(t, config.RequireUppercase)
		assert.True(t, config.RequireLowercase)
		assert.True(t, config.RequireNumbers)
		assert.True(t, config.RequireSpecialChars)
	})

	t.Run("custom configuration", func(t *testing.T) {
		config := &PasswordConfig{
			Memory:              32768, // 32 MiB
			Iterations:          3,     // 3 iterations
			Parallelism:         2,     // 2 parallel threads
			SaltLength:          32,    // 32 bytes salt
			KeyLength:           64,    // 64 bytes hash
			MinLength:           12,
			RequireUppercase:    true,
			RequireLowercase:    true,
			RequireNumbers:      true,
			RequireSpecialChars: false,
		}

		service := NewPasswordService(config)
		password := "VerySecurePassword123!"

		hash, err := service.HashPassword(password)
		require.NoError(t, err)

		// Verify hash contains custom parameters
		assert.Contains(t, hash, "m=32768")
		assert.Contains(t, hash, "t=3")
		assert.Contains(t, hash, "p=2")

		// Verify it still validates correctly
		valid, err := service.ValidatePassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestPasswordService_ParseArgon2Hash(t *testing.T) {
	config := DefaultPasswordConfig()
	service := NewPasswordService(config)

	t.Run("parse valid argon2 hash", func(t *testing.T) {
		// This is a real Argon2id hash format
		hash := "$argon2id$v=19$m=19456,t=2,p=1$c29tZXNhbHQ$RGFlbW9uSW5zZWN1cml0eUNoZWNr"

		memory, iterations, parallelism, salt, decodedHash, err := service.(*passwordService).parseArgon2Hash(
			hash,
		)

		require.NoError(t, err)
		assert.Equal(t, uint32(19456), memory)
		assert.Equal(t, uint32(2), iterations)
		assert.Equal(t, uint8(1), parallelism)
		assert.NotEmpty(t, salt)
		assert.NotEmpty(t, decodedHash)
	})

	t.Run("reject invalid format", func(t *testing.T) {
		invalidHashes := []string{
			"not-a-hash",
			"$bcrypt$something",
			"$argon2i$v=19$m=19456,t=2,p=1$salt$hash",  // argon2i instead of argon2id
			"$argon2id$v=18$m=19456,t=2,p=1$salt$hash", // wrong version
			"$argon2id$v=19$m=19456,t=2$invalid",       // missing parallelism
		}

		for _, invalidHash := range invalidHashes {
			_, _, _, _, _, err := service.(*passwordService).parseArgon2Hash(
				invalidHash,
			)
			require.Error(
				t,
				err,
				"expected error for invalid hash: %s",
				invalidHash,
			)
		}
	})
}

func TestNewPasswordService(t *testing.T) {
	t.Parallel()

	type args struct {
		config *PasswordConfig
	}

	tests := []struct {
		name string
		args args
		want PasswordService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NewPasswordService(tt.args.config); !cmp.Equal(tt.want, got) {
				t.Errorf(
					"NewPasswordService() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}

func Test_passwordService_HashPassword(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}

			got, err := s.HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"passwordService.HashPassword() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("passwordService.HashPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordService_ValidatePassword(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
		hash     string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}

			got, err := s.ValidatePassword(tt.args.password, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"passwordService.ValidatePassword() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf(
					"passwordService.ValidatePassword() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func Test_passwordService_CheckPasswordStrength(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}
			err := s.CheckPasswordStrength(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"passwordService.CheckPasswordStrength() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func Test_passwordService_hasUppercase(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}
			if got := s.hasUppercase(tt.args.password); got != tt.want {
				t.Errorf("passwordService.hasUppercase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordService_hasLowercase(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}
			if got := s.hasLowercase(tt.args.password); got != tt.want {
				t.Errorf("passwordService.hasLowercase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordService_hasNumber(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}
			if got := s.hasNumber(tt.args.password); got != tt.want {
				t.Errorf("passwordService.hasNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordService_hasSpecialChar(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}
			if got := s.hasSpecialChar(tt.args.password); got != tt.want {
				t.Errorf("passwordService.hasSpecialChar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordService_isCommonPassword(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		password string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}
			if got := s.isCommonPassword(tt.args.password); got != tt.want {
				t.Errorf(
					"passwordService.isCommonPassword() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func Test_passwordService_parseArgon2Hash(t *testing.T) {
	t.Parallel()

	type fields struct {
		config *PasswordConfig
	}

	type args struct {
		hash string
	}

	tests := []struct {
		name            string
		fields          fields
		args            args
		wantMemory      uint32
		wantIterations  uint32
		wantParallelism uint8
		wantSalt        []byte
		wantDecodedHash []byte
		wantErr         bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &passwordService{
				config: tt.fields.config,
			}

			gotMemory, gotIterations, gotParallelism, gotSalt, gotDecodedHash, err := s.parseArgon2Hash(
				tt.args.hash,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"passwordService.parseArgon2Hash() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr {
				return
			}

			if gotMemory != tt.wantMemory {
				t.Errorf(
					"passwordService.parseArgon2Hash() gotMemory = %v, want %v",
					gotMemory,
					tt.wantMemory,
				)
			}

			if gotIterations != tt.wantIterations {
				t.Errorf(
					"passwordService.parseArgon2Hash() gotIterations = %v, want %v",
					gotIterations,
					tt.wantIterations,
				)
			}

			if gotParallelism != tt.wantParallelism {
				t.Errorf(
					"passwordService.parseArgon2Hash() gotParallelism = %v, want %v",
					gotParallelism,
					tt.wantParallelism,
				)
			}

			if !cmp.Equal(tt.wantSalt, gotSalt) {
				t.Errorf(
					"passwordService.parseArgon2Hash() gotSalt = %v, want %v\ndiff=%s",
					gotSalt,
					tt.wantSalt,
					cmp.Diff(tt.wantSalt, gotSalt),
				)
			}

			if !cmp.Equal(tt.wantDecodedHash, gotDecodedHash) {
				t.Errorf(
					"passwordService.parseArgon2Hash() gotDecodedHash = %v, want %v\ndiff=%s",
					gotDecodedHash,
					tt.wantDecodedHash,
					cmp.Diff(tt.wantDecodedHash, gotDecodedHash),
				)
			}
		})
	}
}

func Test_parseArgon2Params(t *testing.T) {
	t.Parallel()

	type args struct {
		params string
	}

	tests := []struct {
		name            string
		args            args
		wantMemory      uint32
		wantIterations  uint32
		wantParallelism uint8
		wantErr         bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMemory, gotIterations, gotParallelism, err := parseArgon2Params(
				tt.args.params,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseArgon2Params() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if gotMemory != tt.wantMemory {
				t.Errorf(
					"parseArgon2Params() gotMemory = %v, want %v",
					gotMemory,
					tt.wantMemory,
				)
			}

			if gotIterations != tt.wantIterations {
				t.Errorf(
					"parseArgon2Params() gotIterations = %v, want %v",
					gotIterations,
					tt.wantIterations,
				)
			}

			if gotParallelism != tt.wantParallelism {
				t.Errorf(
					"parseArgon2Params() gotParallelism = %v, want %v",
					gotParallelism,
					tt.wantParallelism,
				)
			}
		})
	}
}

func TestDefaultPasswordConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want *PasswordConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := DefaultPasswordConfig(); !cmp.Equal(tt.want, got) {
				t.Errorf(
					"DefaultPasswordConfig() = %v, want %v\ndiff=%s",
					got,
					tt.want,
					cmp.Diff(tt.want, got),
				)
			}
		})
	}
}
