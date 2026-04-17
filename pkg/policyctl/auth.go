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
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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

// NewOIDCMiddleware creates a Gin middleware for OIDC token validation.
func NewOIDCMiddleware(config *AuthConfig, logger *zap.Logger) (middleware gin.HandlerFunc, err error) {
	if !config.Enabled {
		// Return passthrough middleware when auth is disabled
		middleware = func(c *gin.Context) {
			c.Next()
		}
		return middleware, err
	}

	// Create OIDC provider
	ctx := context.Background()
	var provider *oidc.Provider

	provider, err = oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		err = fmt.Errorf("failed to create OIDC provider for issuer %q: %w", config.IssuerURL, err)
		return middleware, err
	}

	// Create token verifier
	verifierConfig := &oidc.Config{
		ClientID: config.Audience,
	}
	verifier := provider.Verifier(verifierConfig)

	middleware = createAuthMiddleware(verifier, config, logger)
	return middleware, err
}

// createAuthMiddleware builds the actual Gin middleware handler for OIDC authentication.
func createAuthMiddleware(verifier *oidc.IDTokenVerifier, config *AuthConfig, logger *zap.Logger) (handler gin.HandlerFunc) {
	handler = func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		// Verify Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify the token
		idToken, verifyErr := verifier.Verify(c.Request.Context(), tokenString)
		if verifyErr != nil {
			logger.Warn("Token verification failed", zap.Error(verifyErr))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Extract claims
		var rawClaims json.RawMessage
		claimsErr := idToken.Claims(&rawClaims)

		if claimsErr != nil {
			logger.Error("Failed to extract token claims", zap.Error(claimsErr))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to process token"})
			return
		}

		var claims UserClaims
		unmarshalErr := json.Unmarshal(rawClaims, &claims)

		if unmarshalErr != nil {
			logger.Error("Failed to unmarshal token claims", zap.Error(unmarshalErr))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to process token claims"})
			return
		}

		// Check group membership if configured
		if len(config.AllowedGroups) > 0 {
			if !isGroupMember(claims.Groups, config.AllowedGroups) {
				logger.Warn("User not in allowed groups",
					zap.String("email", claims.Email),
					zap.Strings("userGroups", claims.Groups),
					zap.Strings("allowedGroups", config.AllowedGroups),
				)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient group membership"})
				return
			}
		}

		// Store claims in context
		c.Set("userClaims", claims)
		c.Next()
	}
	return handler
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
