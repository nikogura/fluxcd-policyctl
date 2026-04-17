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
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler holds route handlers for the fluxcd-policyctl API.
type Handler struct {
	policyService *PolicyService
	kubeConfig    *KubeConfigService
	authConfig    *AuthConfig
	logger        *zap.Logger
}

// NewHandler creates a new Handler.
func NewHandler(policyService *PolicyService, kubeConfig *KubeConfigService, authConfig *AuthConfig, logger *zap.Logger) (handler *Handler) {
	handler = &Handler{
		policyService: policyService,
		kubeConfig:    kubeConfig,
		authConfig:    authConfig,
		logger:        logger,
	}
	return handler
}

// RegisterRoutes registers all API routes on the Gin engine.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Health check - no auth required
	router.GET("/health", h.handleHealth)

	// User info - no auth required (returns 204 if no auth)
	router.GET("/api/user", h.handleGetUser)

	// API group with optional auth middleware
	api := router.Group("/api")

	if h.authConfig.Enabled {
		authMiddleware, authErr := NewOIDCMiddleware(h.authConfig, h.logger)
		if authErr != nil {
			h.logger.Fatal("Failed to create OIDC middleware", zap.Error(authErr))
		}
		api.Use(authMiddleware)
	}

	api.GET("/clusters", h.handleGetClusters)
	api.GET("/namespaces", h.handleGetNamespaces)
	api.GET("/policies", h.handleListPolicies)
	api.POST("/policies", h.handleCreatePolicy)
	api.GET("/policies/:namespace/:name", h.handleGetPolicy)
	api.PUT("/policies/:namespace/:name", h.handleUpdatePolicy)
	api.DELETE("/policies/:namespace/:name", h.handleDeletePolicy)
}

// handleHealth returns a simple health check response.
func (h *Handler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// handleGetUser returns the current user's claims from the OIDC token.
func (h *Handler) handleGetUser(c *gin.Context) {
	claims, exists := c.Get("userClaims")
	if !exists {
		c.Status(http.StatusNoContent)
		return
	}

	userClaims, ok := claims.(UserClaims)
	if !ok {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, userClaims)
}

// handleGetClusters returns the list of available Kubernetes clusters.
func (h *Handler) handleGetClusters(c *gin.Context) {
	if h.kubeConfig.IsInCluster() {
		c.JSON(http.StatusOK, gin.H{
			"inCluster": true,
			"clusters":  []ClusterInfo{},
		})
		return
	}

	clusters, err := h.kubeConfig.GetClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inCluster": false,
		"clusters":  clusters,
	})
}

// handleGetNamespaces returns the list of namespaces for a cluster.
func (h *Handler) handleGetNamespaces(c *gin.Context) {
	clusterName := c.Query("cluster")

	namespaces, err := h.policyService.ListNamespaces(c.Request.Context(), clusterName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"namespaces": namespaces})
}

// handleListPolicies returns the list of ImagePolicies for a cluster and namespace.
func (h *Handler) handleListPolicies(c *gin.Context) {
	clusterName := c.Query("cluster")
	namespace := c.DefaultQuery("namespace", h.policyService.namespace)

	policies, err := h.policyService.ListPolicies(c.Request.Context(), clusterName, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"policies": policies})
}

// handleGetPolicy returns a single ImagePolicy by namespace and name.
func (h *Handler) handleGetPolicy(c *gin.Context) {
	clusterName := c.Query("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	policy, err := h.policyService.GetPolicy(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// handleCreatePolicy creates a new ImagePolicy.
func (h *Handler) handleCreatePolicy(c *gin.Context) {
	clusterName := c.Query("cluster")

	var req CreatePolicyRequest
	bindErr := c.ShouldBindJSON(&req)

	if bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	err := h.policyService.CreatePolicy(c.Request.Context(), clusterName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "ImagePolicy created successfully"})
}

// handleUpdatePolicy updates the semver range of an existing ImagePolicy.
func (h *Handler) handleUpdatePolicy(c *gin.Context) {
	clusterName := c.Query("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req UpdatePolicyRequest
	bindErr := c.ShouldBindJSON(&req)

	if bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	err := h.policyService.UpdatePolicy(c.Request.Context(), clusterName, namespace, name, req.SemverRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ImagePolicy updated successfully"})
}

// handleDeletePolicy deletes an ImagePolicy.
func (h *Handler) handleDeletePolicy(c *gin.Context) {
	clusterName := c.Query("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.policyService.DeletePolicy(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ImagePolicy deleted successfully"})
}
