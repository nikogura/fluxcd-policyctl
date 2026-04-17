// Copyright 2024 Nik Ogura
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policyctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// AuthConfig holds OIDC authentication configuration.
type AuthConfig struct {
	Enabled       bool
	IssuerURL     string
	Audience      string
	AllowedGroups []string
}

// UserClaims represents the claims extracted from an OIDC token.
type UserClaims struct {
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

// GetUserClaims retrieves user claims from request context.
func GetUserClaims(ctx context.Context) (claims *UserClaims) {
	val := ctx.Value(userClaimsKey)
	if val == nil {
		claims = nil
		return claims
	}

	uc, ok := val.(UserClaims)
	if !ok {
		claims = nil
		return claims
	}

	claims = &uc
	return claims
}

// NewOIDCMiddleware creates an HTTP middleware for OIDC token validation.
func NewOIDCMiddleware(config *AuthConfig, logger *zap.Logger) (middleware func(http.Handler) http.Handler, err error) {
	if !config.Enabled {
		// Passthrough when auth is disabled.
		middleware = func(next http.Handler) (wrapped http.Handler) {
			wrapped = next
			return wrapped
		}
		return middleware, err
	}

	// Create OIDC provider.
	ctx := context.Background()
	var provider *oidc.Provider

	provider, err = oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		err = fmt.Errorf("failed to create OIDC provider for issuer %q: %w", config.IssuerURL, err)
		return middleware, err
	}

	// Create token verifier.
	verifierConfig := &oidc.Config{
		ClientID: config.Audience,
	}
	verifier := provider.Verifier(verifierConfig)

	middleware = createAuthMiddleware(verifier, config, logger)
	return middleware, err
}

// extractBearerToken extracts and validates the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) (token string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		err = errors.New("missing Authorization header")
		return token, err
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		err = errors.New("invalid Authorization header format")
		return token, err
	}

	token = strings.TrimPrefix(authHeader, "Bearer ")
	return token, err
}

// extractClaims verifies the token and extracts user claims.
func extractClaims(r *http.Request, verifier *oidc.IDTokenVerifier) (claims UserClaims, err error) {
	var tokenString string
	tokenString, err = extractBearerToken(r)
	if err != nil {
		return claims, err
	}

	idToken, verifyErr := verifier.Verify(r.Context(), tokenString)
	if verifyErr != nil {
		err = fmt.Errorf("token verification failed: %w", verifyErr)
		return claims, err
	}

	var rawClaims json.RawMessage
	claimsErr := idToken.Claims(&rawClaims)
	if claimsErr != nil {
		err = fmt.Errorf("failed to extract claims: %w", claimsErr)
		return claims, err
	}

	unmarshalErr := json.Unmarshal(rawClaims, &claims)
	if unmarshalErr != nil {
		err = fmt.Errorf("failed to unmarshal claims: %w", unmarshalErr)
		return claims, err
	}

	return claims, err
}

// createAuthMiddleware builds the actual middleware handler for OIDC authentication.
func createAuthMiddleware(verifier *oidc.IDTokenVerifier, config *AuthConfig, logger *zap.Logger) (middleware func(http.Handler) http.Handler) {
	middleware = func(next http.Handler) (wrapped http.Handler) {
		wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := extractClaims(r, verifier)
			if err != nil {
				logger.Warn("Authentication failed", zap.Error(err))
				writeError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			// Check group membership if configured.
			if len(config.AllowedGroups) > 0 && !isGroupMember(claims.Groups, config.AllowedGroups) {
				logger.Warn("User not in allowed groups",
					zap.String("email", claims.Email),
					zap.Strings("userGroups", claims.Groups),
					zap.Strings("allowedGroups", config.AllowedGroups),
				)
				writeError(w, http.StatusForbidden, "insufficient group membership")
				return
			}

			// Store claims in context.
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

		return wrapped
	}

	return middleware
}

// isGroupMember checks if any of the user's groups match the allowed groups.
func isGroupMember(userGroups []string, allowedGroups []string) (member bool) {
	for _, userGroup := range userGroups {
		for _, allowedGroup := range allowedGroups {
			if userGroup == allowedGroup {
				member = true
				return member
			}
		}
	}

	member = false
	return member
}
