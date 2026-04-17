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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestUserEndpointNoAuth(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/user", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestPoliciesEndpointRequiresClusterParam(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/policies", func(w http.ResponseWriter, r *http.Request) {
		cluster := r.URL.Query().Get("cluster")
		w.Header().Set("Content-Type", "application/json")
		if cluster == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"cluster parameter required"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"policies":[]}`))
	})

	// Without cluster param.
	req := httptest.NewRequest(http.MethodGet, "/api/policies", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// With cluster param.
	req = httptest.NewRequest(http.MethodGet, "/api/policies?cluster=test-cluster&namespace=flux-system", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatePolicyEndpointValidation(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/policies", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name            string `json:"name"`
			Namespace       string `json:"namespace"`
			ImageRepository string `json:"imageRepository"`
			SemverRange     string `json:"semverRange"`
		}

		decodeErr := json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")

		if decodeErr != nil || body.Name == "" || body.Namespace == "" || body.ImageRepository == "" || body.SemverRange == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"missing required fields"}`))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"message":"created"}`))
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
			mux.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdatePolicyEndpointValidation(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/policies/{namespace}/{name}", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SemverRange string `json:"semverRange"`
		}

		decodeErr := json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")

		if decodeErr != nil || body.SemverRange == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"semverRange required"}`))
			return
		}

		ns := r.PathValue("namespace")
		nm := r.PathValue("name")
		resp := map[string]string{"message": "updated", "namespace": ns, "name": nm}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Valid update.
	req := httptest.NewRequest(http.MethodPut, "/api/policies/default/my-policy?cluster=test", strings.NewReader(`{"semverRange":">=2.0.0"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "default", response["namespace"])
	assert.Equal(t, "my-policy", response["name"])

	// Missing semverRange.
	req = httptest.NewRequest(http.MethodPut, "/api/policies/default/my-policy?cluster=test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePolicyEndpoint(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/policies/{namespace}/{name}", func(w http.ResponseWriter, r *http.Request) {
		ns := r.PathValue("namespace")
		nm := r.PathValue("name")
		resp := map[string]string{"message": "deleted", "namespace": ns, "name": nm}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/policies/default/my-policy?cluster=test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "deleted", response["message"])
	assert.Equal(t, "default", response["namespace"])
	assert.Equal(t, "my-policy", response["name"])
}
