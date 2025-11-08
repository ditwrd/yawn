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

// Package services provides JWT token management for authentication.
package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService provides JWT token operations for authentication.
type JWTService interface {
	GenerateTokenPair(
		user *models.User,
	) (accessToken, refreshToken string, err error)
	GenerateAccessToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	RefreshToken(refreshTokenString string) (string, error)
	InvalidateToken(tokenString string) error
	IsTokenBlacklisted(tokenString string) (bool, error)
}

// TokenClaims represents custom JWT claims with user information.
type TokenClaims struct {
	UserID uuid.UUID       `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	Type   TokenType       `json:"type"` // access or refresh
	jwt.RegisteredClaims
}

// TokenType represents the type of JWT token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// JWTConfig holds JWT token configuration settings.
type JWTConfig struct {
	AccessSecret  string        `json:"access_secret"`
	RefreshSecret string        `json:"refresh_secret"`
	AccessExpiry  time.Duration `json:"access_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
	Issuer        string        `json:"issuer"`
	Audience      string        `json:"audience"`
}

// jwtService implements JWTService.
type jwtService struct {
	config    *JWTConfig
	blacklist map[string]time.Time
}

// NewJWTService creates a new JWT service.
func NewJWTService(config *JWTConfig) JWTService {
	return &jwtService{
		config:    config,
		blacklist: make(map[string]time.Time),
	}
}

func (s *jwtService) GenerateTokenPair(
	user *models.User,
) (string, string, error) {
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *jwtService) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now()
	claims := &TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   user.ID.String(),
			Audience:  []string{s.config.Audience},
			ID:        uuid.Must(uuid.NewV7()).String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.config.AccessSecret))
}

func (s *jwtService) generateRefreshToken(user *models.User) (string, error) {
	now := time.Now()
	claims := &TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Type:   TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   user.ID.String(),
			Audience:  []string{s.config.Audience},
			ID:        uuid.Must(uuid.NewV7()).String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.config.RefreshSecret))
}

func (s *jwtService) ValidateToken(tokenString string) (*TokenClaims, error) {
	blacklisted, err := s.IsTokenBlacklisted(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist status: %w", err)
	}

	if blacklisted {
		return nil, errors.New("token has been invalidated")
	}

	unverifiedToken, _, err := jwt.NewParser().
		ParseUnverified(tokenString, &TokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := unverifiedToken.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims type")
	}

	var secret string

	switch claims.Type {
	case TokenTypeAccess:
		secret = s.config.AccessSecret
	case TokenTypeRefresh:
		secret = s.config.RefreshSecret
	default:
		return nil, fmt.Errorf("unknown token type: %s", claims.Type)
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					token.Header["alg"],
				)
			}

			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	validatedClaims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims type after validation")
	}

	if err := s.validateClaims(validatedClaims); err != nil {
		return nil, fmt.Errorf("claim validation failed: %w", err)
	}

	return validatedClaims, nil
}

func (s *jwtService) validateClaims(claims *TokenClaims) error {
	if claims.Issuer != s.config.Issuer {
		return fmt.Errorf(
			"invalid issuer: expected %s, got %s",
			s.config.Issuer,
			claims.Issuer,
		)
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != s.config.Audience {
		return fmt.Errorf(
			"invalid audience: expected %s, got %v",
			s.config.Audience,
			claims.Audience,
		)
	}

	if claims.ExpiresAt == nil {
		return errors.New("missing expiration claim")
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return errors.New("token has expired")
	}

	if claims.UserID == uuid.Nil {
		return errors.New("invalid user ID in token")
	}

	return nil
}

func (s *jwtService) RefreshToken(refreshTokenString string) (string, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != TokenTypeRefresh {
		return "", fmt.Errorf("expected refresh token, got %s", claims.Type)
	}

	user := &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	return s.GenerateAccessToken(user)
}

func (s *jwtService) InvalidateToken(tokenString string) error {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to validate token for invalidation: %w", err)
	}

	s.blacklist[tokenString] = claims.ExpiresAt.Time

	return nil
}

func (s *jwtService) IsTokenBlacklisted(tokenString string) (bool, error) {
	expirationTime, exists := s.blacklist[tokenString]
	if !exists {
		return false, nil
	}

	if time.Now().After(expirationTime) {
		delete(s.blacklist, tokenString)

		return false, nil
	}

	return true, nil
}

// DefaultJWTConfig returns default JWT configuration.
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		AccessSecret:  "your-access-secret-change-in-production",
		RefreshSecret: "your-refresh-secret-change-in-production",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "yawn-api",
		Audience:      "yawn-client",
	}
}
