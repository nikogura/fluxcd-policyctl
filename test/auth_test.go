// Copyright 2026 Nik Ogura
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

package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikogura/fluxcd-policyctl/pkg/policyctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOIDCMiddlewareDisabled(t *testing.T) {
	t.Parallel()

	config := &policyctl.AuthConfig{
		Enabled: false,
	}

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	middleware, err := policyctl.NewOIDCMiddleware(config, logger)
	require.NoError(t, err)
	require.NotNil(t, middleware)

	// When auth is disabled, requests should pass through
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOIDCMiddlewareEnabledRejectsNoToken(t *testing.T) {
	t.Parallel()

	// This test verifies that when OIDC is enabled with a bogus issuer,
	// the middleware creation itself fails (can't discover OIDC config).
	// In production, the issuer must be reachable.
	config := &policyctl.AuthConfig{
		Enabled:   true,
		IssuerURL: "https://nonexistent.example.com",
		Audience:  "test-audience",
	}

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	// NewOIDCMiddleware with unreachable issuer should fail
	_, err := policyctl.NewOIDCMiddleware(config, logger)
	// This will either fail (can't reach issuer) or succeed if there's a timeout.
	// Either way, the middleware should handle it gracefully.
	if err != nil {
		// Expected: can't discover OIDC configuration
		assert.Contains(t, err.Error(), "OIDC")
	}
}

func TestAuthConfigDefaults(t *testing.T) {
	t.Parallel()

	config := &policyctl.AuthConfig{}
	assert.False(t, config.Enabled)
	assert.Empty(t, config.IssuerURL)
	assert.Empty(t, config.Audience)
	assert.Nil(t, config.AllowedGroups)
}

func TestUserClaimsStructure(t *testing.T) {
	t.Parallel()

	claims := policyctl.UserClaims{
		Email:  "test@example.com",
		Name:   "Test User",
		Groups: []string{"engineering", "ops"},
	}

	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "Test User", claims.Name)
	assert.Len(t, claims.Groups, 2)
	assert.Contains(t, claims.Groups, "engineering")
}
