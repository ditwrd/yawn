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

// Package middleware provides authentication middleware for Echo.
package middleware

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/services"
)

// AuthMiddleware provides JWT authentication middleware.
type AuthMiddleware struct {
	jwtService services.JWTService
	logger     *zerolog.Logger
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(
	jwtService services.JWTService,
	logger *zerolog.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

// JWT creates middleware that validates JWT tokens and injects user context.
func (am *AuthMiddleware) JWT() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:     []byte("dummy"),
		ParseTokenFunc: am.customParseTokenFunc,
		ErrorHandler:   am.customErrorHandler,
		ContextKey:     "user",
		TokenLookup:    "header:Authorization:Bearer ",
		Skipper: func(c echo.Context) bool {
			return false
		},
	})
}

// customParseTokenFunc provides custom token parsing with our JWT service.
func (am *AuthMiddleware) customParseTokenFunc(
	c echo.Context,
	auth string,
) (any, error) {
	claims, err := am.jwtService.ValidateToken(auth)
	if err != nil {
		am.logger.Error().
			Err(err).
			Str("token", auth[:min(len(auth), 20)]+"...").
			Msg("JWT token validation failed")

		return nil, err
	}

	c.Set("user_claims", claims)
	c.Set("user_id", claims.UserID)
	c.Set("user_email", claims.Email)
	c.Set("user_role", claims.Role)

	return &jwt.Token{
		Raw:       auth,
		Method:    jwt.SigningMethodHS256,
		Header:    map[string]any{"alg": "HS256", "typ": "JWT"},
		Claims:    claims,
		Signature: []byte{},
		Valid:     true,
	}, nil
}

// customErrorHandler provides custom error handling for JWT middleware.
func (am *AuthMiddleware) customErrorHandler(c echo.Context, err error) error {
	am.logger.Warn().
		Err(err).
		Str("path", c.Request().URL.Path).
		Str("method", c.Request().Method).
		Msg("Authentication failed")

	var extractionErr *echojwt.TokenExtractionError

	var parsingErr *echojwt.TokenParsingError

	if errors.As(err, &extractionErr) {
		return c.JSON(http.StatusUnauthorized, map[string]any{
			"error": "Authentication token required",
			"code":  "TOKEN_MISSING",
		})
	}

	if errors.As(err, &parsingErr) {
		return c.JSON(http.StatusUnauthorized, map[string]any{
			"error": "Invalid or expired authentication token",
			"code":  "TOKEN_INVALID",
		})
	}

	return c.JSON(http.StatusUnauthorized, map[string]any{
		"error": "Authentication failed",
		"code":  "AUTH_FAILED",
	})
}

// RequireAuth creates middleware that requires authentication.
func (am *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return am.JWT()
}

// OptionalAuth creates middleware that optionally extracts JWT tokens.
func (am *AuthMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:             []byte("dummy"),
		ParseTokenFunc:         am.optionalParseTokenFunc,
		ErrorHandler:           am.optionalErrorHandler,
		ContextKey:             "user",
		TokenLookup:            "header:Authorization:Bearer ",
		ContinueOnIgnoredError: true,
	})
}

func (am *AuthMiddleware) optionalParseTokenFunc(
	c echo.Context,
	auth string,
) (any, error) {
	if auth == "" {
		return nil, nil
	}

	claims, err := am.jwtService.ValidateToken(auth)
	if err != nil {
		am.logger.Debug().
			Err(err).
			Str("path", c.Request().URL.Path).
			Msg("Optional authentication failed")

		return nil, nil
	}

	c.Set("user_claims", claims)
	c.Set("user_id", claims.UserID)
	c.Set("user_email", claims.Email)
	c.Set("user_role", claims.Role)

	return &jwt.Token{
		Raw:       auth,
		Method:    jwt.SigningMethodHS256,
		Header:    map[string]any{"alg": "HS256", "typ": "JWT"},
		Claims:    claims,
		Signature: []byte{},
		Valid:     true,
	}, nil
}

func (am *AuthMiddleware) optionalErrorHandler(
	c echo.Context,
	err error,
) error {
	return nil
}

// GetUserClaims extracts user claims from the Echo context.
func GetUserClaims(c echo.Context) (*services.TokenClaims, error) {
	if claims, ok := c.Get("user_claims").(*services.TokenClaims); ok {
		return claims, nil
	}

	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, errors.New(
			"JWT library version mismatch: token not found in context or incorrect type",
		)
	}

	claims, ok := token.Claims.(*services.TokenClaims)
	if !ok {
		return nil, errors.New(
			"JWT library version mismatch: claims not found or incorrect type",
		)
	}

	return claims, nil
}

// GetUserID extracts user ID from the Echo context.
func GetUserID(c echo.Context) (string, error) {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID, nil
	}

	claims, err := GetUserClaims(c)
	if err != nil {
		return "", err
	}

	return claims.UserID.String(), nil
}

// GetUserRole extracts user role from the Echo context.
func GetUserRole(c echo.Context) (string, error) {
	if role, ok := c.Get("user_role").(string); ok {
		return role, nil
	}

	claims, err := GetUserClaims(c)
	if err != nil {
		return "", err
	}

	return string(claims.Role), nil
}
