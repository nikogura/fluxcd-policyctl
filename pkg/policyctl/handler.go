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
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// ConfigResponse is the JSON response for the /api/config endpoint.
type ConfigResponse struct {
	AccessMode         string   `json:"accessMode"`
	InCluster          bool     `json:"inCluster"`
	AllowedNamespaces  []string `json:"allowedNamespaces"`
	FixedNamespace     *string  `json:"fixedNamespace"`
	RefreshIntervalSec int      `json:"refreshIntervalSec"`
}

// Handler holds route handlers for the fluxcd-policyctl API.
type Handler struct {
	policyService *PolicyService
	kubeConfig    *KubeConfigService
	authConfig    *AuthConfig
	accessConfig  *AccessConfig
	logger        *zap.Logger
}

// NewHandler creates a new Handler.
func NewHandler(policyService *PolicyService, kubeConfig *KubeConfigService, authConfig *AuthConfig, accessConfig *AccessConfig, logger *zap.Logger) (handler *Handler) {
	handler = &Handler{
		policyService: policyService,
		kubeConfig:    kubeConfig,
		authConfig:    authConfig,
		accessConfig:  accessConfig,
		logger:        logger,
	}
	return handler
}

// RegisterRoutes registers all API routes on the mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Health check — no auth.
	mux.HandleFunc("GET /health", h.handleHealth)

	// Config endpoint — no auth (frontend needs it before auth).
	mux.HandleFunc("GET /api/config", h.handleGetConfig)

	// User info — no auth (returns 204 if disabled).
	mux.HandleFunc("GET /api/user", h.handleGetUser)

	// API routes — optionally wrapped with OIDC middleware.
	if h.authConfig.Enabled {
		authMiddleware, authErr := NewOIDCMiddleware(h.authConfig, h.logger)
		if authErr != nil {
			h.logger.Fatal("Failed to create OIDC middleware", zap.Error(authErr))
		}

		mux.Handle("GET /api/clusters", authMiddleware(http.HandlerFunc(h.handleGetClusters)))
		mux.Handle("GET /api/namespaces", authMiddleware(http.HandlerFunc(h.handleGetNamespaces)))
		mux.Handle("GET /api/policies", authMiddleware(http.HandlerFunc(h.handleListPolicies)))
		mux.Handle("POST /api/policies", authMiddleware(http.HandlerFunc(h.handleCreatePolicy)))
		mux.Handle("GET /api/policies/{namespace}/{name}", authMiddleware(http.HandlerFunc(h.handleGetPolicy)))
		mux.Handle("PUT /api/policies/{namespace}/{name}", authMiddleware(http.HandlerFunc(h.handleUpdatePolicy)))
		mux.Handle("DELETE /api/policies/{namespace}/{name}", authMiddleware(http.HandlerFunc(h.handleDeletePolicy)))
	} else {
		mux.HandleFunc("GET /api/clusters", h.handleGetClusters)
		mux.HandleFunc("GET /api/namespaces", h.handleGetNamespaces)
		mux.HandleFunc("GET /api/policies", h.handleListPolicies)
		mux.HandleFunc("POST /api/policies", h.handleCreatePolicy)
		mux.HandleFunc("GET /api/policies/{namespace}/{name}", h.handleGetPolicy)
		mux.HandleFunc("PUT /api/policies/{namespace}/{name}", h.handleUpdatePolicy)
		mux.HandleFunc("DELETE /api/policies/{namespace}/{name}", h.handleDeletePolicy)
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// handleHealth returns a simple health check response.
func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// handleGetConfig returns the current access mode and constraints for the frontend.
func (h *Handler) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	refreshSec := h.accessConfig.RefreshIntervalSec
	if refreshSec <= 0 {
		refreshSec = 30
	}

	resp := ConfigResponse{
		AccessMode:         h.accessConfig.Mode,
		InCluster:          h.accessConfig.Mode != AccessModeLocal,
		RefreshIntervalSec: refreshSec,
	}

	switch h.accessConfig.Mode {
	case AccessModeNamespaces:
		resp.AllowedNamespaces = h.accessConfig.AllowedNamespaces
	case AccessModeNamespace:
		resp.FixedNamespace = &h.accessConfig.PodNamespace
	}

	writeJSON(w, http.StatusOK, resp)
}

// isNamespaceAllowed checks whether the given namespace is permitted under the current access mode.
func (h *Handler) isNamespaceAllowed(ns string) (allowed bool) {
	switch h.accessConfig.Mode {
	case AccessModeNamespace:
		allowed = ns == h.accessConfig.PodNamespace
		return allowed
	case AccessModeNamespaces:
		for _, allowedNS := range h.accessConfig.AllowedNamespaces {
			if ns == allowedNS {
				allowed = true
				return allowed
			}
		}

		allowed = false
		return allowed
	default:
		// "local" and "cluster" modes allow all namespaces.
		allowed = true
		return allowed
	}
}

// handleGetUser returns the current user's claims from the OIDC token.
func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, claims)
}

// handleGetClusters returns the list of available Kubernetes clusters.
func (h *Handler) handleGetClusters(w http.ResponseWriter, _ *http.Request) {
	if h.kubeConfig.IsInCluster() {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"inCluster": true,
			"clusters":  []ClusterInfo{},
		})
		return
	}

	clusters, err := h.kubeConfig.GetClusters()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"inCluster": false,
		"clusters":  clusters,
	})
}

// handleGetNamespaces returns the list of namespaces for a cluster.
func (h *Handler) handleGetNamespaces(w http.ResponseWriter, r *http.Request) {
	// In "namespace" mode, return only the fixed namespace without querying the cluster.
	if h.accessConfig.Mode == AccessModeNamespace {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"namespaces": []string{h.accessConfig.PodNamespace},
		})
		return
	}

	clusterName := r.URL.Query().Get("cluster")

	namespaces, err := h.policyService.ListNamespaces(r.Context(), clusterName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// In "namespaces" mode, filter to only the allowed set.
	if h.accessConfig.Mode == AccessModeNamespaces {
		namespaces = filterNamespaces(namespaces, h.accessConfig.AllowedNamespaces)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"namespaces": namespaces})
}

// filterNamespaces returns only the namespaces that appear in the allowed list.
func filterNamespaces(namespaces []string, allowed []string) (filtered []string) {
	allowedSet := make(map[string]bool, len(allowed))
	for _, ns := range allowed {
		allowedSet[ns] = true
	}

	filtered = make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		if allowedSet[ns] {
			filtered = append(filtered, ns)
		}
	}

	return filtered
}

// handleListPolicies returns the list of ImagePolicies for a cluster and namespace.
func (h *Handler) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("cluster")
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = h.policyService.namespace
	}

	if !h.isNamespaceAllowed(namespace) {
		writeError(w, http.StatusForbidden, "access to namespace "+namespace+" is not allowed")
		return
	}

	policies, err := h.policyService.ListPolicies(r.Context(), clusterName, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"policies": policies})
}

// handleGetPolicy returns a single ImagePolicy by namespace and name.
func (h *Handler) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("cluster")
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	policy, err := h.policyService.GetPolicy(r.Context(), clusterName, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, policy)
}

// handleCreatePolicy creates a new ImagePolicy.
func (h *Handler) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("cluster")

	var req CreatePolicyRequest
	decodeErr := json.NewDecoder(r.Body).Decode(&req)
	if decodeErr != nil {
		writeError(w, http.StatusBadRequest, decodeErr.Error())
		return
	}

	if req.Name == "" || req.Namespace == "" || req.ImageRepository == "" || req.SemverRange == "" {
		writeError(w, http.StatusBadRequest, "name, namespace, imageRepository, and semverRange are required")
		return
	}

	if !h.isNamespaceAllowed(req.Namespace) {
		writeError(w, http.StatusForbidden, "access to namespace "+req.Namespace+" is not allowed")
		return
	}

	err := h.policyService.CreatePolicy(r.Context(), clusterName, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "ImagePolicy created successfully"})
}

// handleUpdatePolicy updates the semver range of an existing ImagePolicy.
func (h *Handler) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("cluster")
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	if !h.isNamespaceAllowed(namespace) {
		writeError(w, http.StatusForbidden, "access to namespace "+namespace+" is not allowed")
		return
	}

	var req UpdatePolicyRequest
	decodeErr := json.NewDecoder(r.Body).Decode(&req)
	if decodeErr != nil {
		writeError(w, http.StatusBadRequest, decodeErr.Error())
		return
	}

	if req.SemverRange == "" {
		writeError(w, http.StatusBadRequest, "semverRange is required")
		return
	}

	err := h.policyService.UpdatePolicy(r.Context(), clusterName, namespace, name, req.SemverRange)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ImagePolicy updated successfully"})
}

// handleDeletePolicy deletes an ImagePolicy.
func (h *Handler) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("cluster")
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	if !h.isNamespaceAllowed(namespace) {
		writeError(w, http.StatusForbidden, "access to namespace "+namespace+" is not allowed")
		return
	}

	err := h.policyService.DeletePolicy(r.Context(), clusterName, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "ImagePolicy deleted successfully"})
}
