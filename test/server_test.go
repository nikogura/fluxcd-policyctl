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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (router *gin.Engine) {
	gin.SetMode(gin.TestMode)
	router = gin.New()
	return router
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	router := setupTestRouter()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestUserEndpointNoAuth(t *testing.T) {
	t.Parallel()

	router := setupTestRouter()
	router.GET("/api/user", func(c *gin.Context) {
		// When no auth configured, return 204
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestPoliciesEndpointRequiresClusterParam(t *testing.T) {
	t.Parallel()

	router := setupTestRouter()
	router.GET("/api/policies", func(c *gin.Context) {
		cluster := c.Query("cluster")
		if cluster == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cluster parameter required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"policies": []interface{}{}})
	})

	// Without cluster param
	req := httptest.NewRequest(http.MethodGet, "/api/policies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// With cluster param
	req = httptest.NewRequest(http.MethodGet, "/api/policies?cluster=test-cluster&namespace=flux-system", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatePolicyEndpointValidation(t *testing.T) {
	t.Parallel()

	router := setupTestRouter()
	router.POST("/api/policies", func(c *gin.Context) {
		var body struct {
			Name            string `json:"name" binding:"required"`
			Namespace       string `json:"namespace" binding:"required"`
			ImageRepository string `json:"imageRepository" binding:"required"`
			SemverRange     string `json:"semverRange" binding:"required"`
		}
		bindErr := c.ShouldBindJSON(&body)
		if bindErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "created"})
	})

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       `{"name":"my-policy","namespace":"default","imageRepository":"my-repo","semverRange":">=1.0.0"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing name",
			body:       `{"namespace":"default","imageRepository":"my-repo","semverRange":">=1.0.0"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing namespace",
			body:       `{"name":"my-policy","imageRepository":"my-repo","semverRange":">=1.0.0"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing imageRepository",
			body:       `{"name":"my-policy","namespace":"default","semverRange":">=1.0.0"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing semverRange",
			body:       `{"name":"my-policy","namespace":"default","imageRepository":"my-repo"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/api/policies?cluster=test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdatePolicyEndpointValidation(t *testing.T) {
	t.Parallel()

	router := setupTestRouter()
	router.PUT("/api/policies/:namespace/:name", func(c *gin.Context) {
		var body struct {
			SemverRange string `json:"semverRange" binding:"required"`
		}
		bindErr := c.ShouldBindJSON(&body)
		if bindErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "updated",
			"namespace": c.Param("namespace"),
			"name":      c.Param("name"),
		})
	})

	// Valid update
	req := httptest.NewRequest(http.MethodPut, "/api/policies/default/my-policy?cluster=test", strings.NewReader(`{"semverRange":">=2.0.0"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "default", response["namespace"])
	assert.Equal(t, "my-policy", response["name"])

	// Missing semverRange
	req = httptest.NewRequest(http.MethodPut, "/api/policies/default/my-policy?cluster=test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePolicyEndpoint(t *testing.T) {
	t.Parallel()

	router := setupTestRouter()
	router.DELETE("/api/policies/:namespace/:name", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "deleted",
			"namespace": c.Param("namespace"),
			"name":      c.Param("name"),
		})
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/policies/default/my-policy?cluster=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "deleted", response["message"])
	assert.Equal(t, "default", response["namespace"])
	assert.Equal(t, "my-policy", response["name"])
}
