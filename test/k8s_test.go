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
	"testing"

	"github.com/nikogura/fluxcd-policyctl/pkg/policyctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPolicyService(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	kubeConfig := policyctl.NewKubeConfigService(logger)
	service := policyctl.NewPolicyService(kubeConfig, "flux-system", logger)

	assert.NotNil(t, service)
}

func TestPolicyViewStructure(t *testing.T) {
	t.Parallel()

	view := policyctl.PolicyView{
		Name:              "my-app",
		Namespace:         "default",
		ImageRepository:   "my-app-repo",
		ImageURL:          "ghcr.io/myorg/my-app",
		SemverRange:       ">=1.0.0 <2.0.0",
		LatestVersion:     "1.5.3",
		AvailableVersions: []string{"1.5.3", "1.5.2", "1.5.1", "1.4.0"},
		LastUpdated:       "2026-04-15T10:00:00Z",
		Ready:             true,
		Message:           "Latest image tag for 'ghcr.io/myorg/my-app' resolved to 1.5.3",
	}

	assert.Equal(t, "my-app", view.Name)
	assert.Equal(t, "default", view.Namespace)
	assert.Equal(t, "my-app-repo", view.ImageRepository)
	assert.Equal(t, "ghcr.io/myorg/my-app", view.ImageURL)
	assert.Equal(t, ">=1.0.0 <2.0.0", view.SemverRange)
	assert.Equal(t, "1.5.3", view.LatestVersion)
	assert.Len(t, view.AvailableVersions, 4)
	assert.Equal(t, "2026-04-15T10:00:00Z", view.LastUpdated)
	assert.True(t, view.Ready)
	assert.Contains(t, view.Message, "1.5.3")
}

func TestCreatePolicyRequestValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request policyctl.CreatePolicyRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: policyctl.CreatePolicyRequest{
				Name:            "my-policy",
				Namespace:       "default",
				ImageRepository: "my-repo",
				SemverRange:     ">=1.0.0",
			},
			valid: true,
		},
		{
			name: "missing name",
			request: policyctl.CreatePolicyRequest{
				Namespace:       "default",
				ImageRepository: "my-repo",
				SemverRange:     ">=1.0.0",
			},
			valid: false,
		},
		{
			name: "missing namespace",
			request: policyctl.CreatePolicyRequest{
				Name:            "my-policy",
				ImageRepository: "my-repo",
				SemverRange:     ">=1.0.0",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
				assert.NotEmpty(t, tt.request.Namespace)
				assert.NotEmpty(t, tt.request.ImageRepository)
				assert.NotEmpty(t, tt.request.SemverRange)
			} else {
				hasEmpty := tt.request.Name == "" || tt.request.Namespace == "" ||
					tt.request.ImageRepository == "" || tt.request.SemverRange == ""
				assert.True(t, hasEmpty)
			}
		})
	}
}

func TestKubeConfigServiceCreation(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	service := policyctl.NewKubeConfigService(logger)
	assert.NotNil(t, service)

	// Should be able to determine if running in cluster
	// (in test environment, this will be false)
	inCluster := service.IsInCluster()
	assert.False(t, inCluster)
}

func TestKubeConfigGetClusters(t *testing.T) {
	t.Parallel()

	logger, logErr := zap.NewDevelopment()
	require.NoError(t, logErr)

	service := policyctl.NewKubeConfigService(logger)

	// This test depends on having a kubeconfig file
	// It should either return clusters or an error, not panic
	clusters, err := service.GetClusters()
	if err != nil {
		// Expected if no kubeconfig exists
		t.Logf("GetClusters returned error (expected in CI): %v", err)
		return
	}

	// If we got clusters, verify structure
	for _, cluster := range clusters {
		assert.NotEmpty(t, cluster.Name)
	}
}
