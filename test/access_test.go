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

	"github.com/nikogura/fluxcd-policyctl/pkg/policyctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAccessConfigDefaults(t *testing.T) {
	t.Parallel()

	config := &policyctl.AccessConfig{}
	assert.Empty(t, config.Mode)
	assert.Nil(t, config.AllowedNamespaces)
	assert.Empty(t, config.PodNamespace)
}

func TestAccessConfigStructure(t *testing.T) {
	t.Parallel()

	config := &policyctl.AccessConfig{
		Mode:              policyctl.AccessModeNamespaces,
		AllowedNamespaces: []string{"dev-01", "stage-01"},
	}

	assert.Equal(t, "namespaces", config.Mode)
	assert.Len(t, config.AllowedNamespaces, 2)
	assert.Contains(t, config.AllowedNamespaces, "dev-01")
	assert.Contains(t, config.AllowedNamespaces, "stage-01")
}

func TestValidAccessModes(t *testing.T) {
	t.Parallel()

	modes := policyctl.ValidAccessModes()
	assert.Len(t, modes, 4)
	assert.Contains(t, modes, "local")
	assert.Contains(t, modes, "cluster")
	assert.Contains(t, modes, "namespaces")
	assert.Contains(t, modes, "namespace")
}

func TestIsValidAccessMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		mode  string
		valid bool
	}{
		{name: "local is valid", mode: "local", valid: true},
		{name: "cluster is valid", mode: "cluster", valid: true},
		{name: "namespaces is valid", mode: "namespaces", valid: true},
		{name: "namespace is valid", mode: "namespace", valid: true},
		{name: "empty is invalid", mode: "", valid: false},
		{name: "unknown is invalid", mode: "foobar", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := policyctl.IsValidAccessMode(tt.mode)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestDetectPodNamespaceFallback(t *testing.T) {
	t.Parallel()

	// In a test environment, the downward API file does not exist.
	// DetectPodNamespace should return the fallback.
	ns := policyctl.DetectPodNamespace("flux-system")
	assert.Equal(t, "flux-system", ns)
}

func TestAccessModeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "local", policyctl.AccessModeLocal)
	assert.Equal(t, "cluster", policyctl.AccessModeCluster)
	assert.Equal(t, "namespaces", policyctl.AccessModeNamespaces)
	assert.Equal(t, "namespace", policyctl.AccessModeNamespace)
}

func TestConfigEndpointLocalMode(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{Mode: policyctl.AccessModeLocal}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp policyctl.ConfigResponse
	decodeErr := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, decodeErr)

	assert.Equal(t, "local", resp.AccessMode)
	assert.False(t, resp.InCluster)
	assert.Nil(t, resp.AllowedNamespaces)
	assert.Nil(t, resp.FixedNamespace)
}

func TestConfigEndpointClusterMode(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{Mode: policyctl.AccessModeCluster}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp policyctl.ConfigResponse
	decodeErr := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, decodeErr)

	assert.Equal(t, "cluster", resp.AccessMode)
	assert.True(t, resp.InCluster)
	assert.Nil(t, resp.AllowedNamespaces)
	assert.Nil(t, resp.FixedNamespace)
}

func TestConfigEndpointNamespacesMode(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:              policyctl.AccessModeNamespaces,
		AllowedNamespaces: []string{"dev-01", "stage-01"},
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp policyctl.ConfigResponse
	decodeErr := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, decodeErr)

	assert.Equal(t, "namespaces", resp.AccessMode)
	assert.True(t, resp.InCluster)
	assert.Equal(t, []string{"dev-01", "stage-01"}, resp.AllowedNamespaces)
	assert.Nil(t, resp.FixedNamespace)
}

func TestConfigEndpointNamespaceMode(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:         policyctl.AccessModeNamespace,
		PodNamespace: "dev-01",
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp policyctl.ConfigResponse
	decodeErr := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, decodeErr)

	assert.Equal(t, "namespace", resp.AccessMode)
	assert.True(t, resp.InCluster)
	assert.Nil(t, resp.AllowedNamespaces)
	require.NotNil(t, resp.FixedNamespace)
	assert.Equal(t, "dev-01", *resp.FixedNamespace)
}

func TestNamespaceModeGetNamespacesReturnsFixed(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:         policyctl.AccessModeNamespace,
		PodNamespace: "my-ns",
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/namespaces", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string][]string
	decodeErr := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, decodeErr)

	assert.Equal(t, []string{"my-ns"}, resp["namespaces"])
}

func TestNamespaceModeForbidsOtherNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:         policyctl.AccessModeNamespace,
		PodNamespace: "my-ns",
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Attempt to list policies in a different namespace.
	req := httptest.NewRequest(http.MethodGet, "/api/policies?namespace=other-ns", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]string
	decodeErr := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, decodeErr)
	assert.Contains(t, resp["error"], "not allowed")
}

func TestNamespacesModeForbidsUnlistedNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:              policyctl.AccessModeNamespaces,
		AllowedNamespaces: []string{"dev-01", "stage-01"},
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Attempt to list policies in a namespace that is NOT in the allowed list.
	req := httptest.NewRequest(http.MethodGet, "/api/policies?namespace=prod-01", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNamespaceModeForbidsCreateInOtherNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:         policyctl.AccessModeNamespace,
		PodNamespace: "my-ns",
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := `{"name":"my-policy","namespace":"other-ns","imageRepository":"my-repo","semverRange":">=1.0.0"}`
	req := httptest.NewRequest(http.MethodPost, "/api/policies", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNamespaceModeForbidsUpdateInOtherNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:         policyctl.AccessModeNamespace,
		PodNamespace: "my-ns",
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := `{"semverRange":">=2.0.0"}`
	req := httptest.NewRequest(http.MethodPut, "/api/policies/other-ns/my-policy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNamespaceModeForbidsDeleteInOtherNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{
		Mode:         policyctl.AccessModeNamespace,
		PodNamespace: "my-ns",
	}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/policies/other-ns/my-policy", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLocalModeAllowsAnyNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{Mode: policyctl.AccessModeLocal}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// In local mode, the request will NOT be rejected by access control.
	// It will fail at the k8s layer (no cluster param), but NOT with 403.
	req := httptest.NewRequest(http.MethodGet, "/api/policies?namespace=any-ns", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should NOT be forbidden - will be 500 because there's no real cluster.
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestClusterModeAllowsAnyNamespace(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	policyService := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	authConfig := &policyctl.AuthConfig{Enabled: false}
	accessConfig := &policyctl.AccessConfig{Mode: policyctl.AccessModeCluster}

	handler := policyctl.NewHandler(policyService, kubeConfig, authConfig, accessConfig, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/policies?namespace=any-ns", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should NOT be forbidden.
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}
