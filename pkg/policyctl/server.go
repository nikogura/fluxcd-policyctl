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
	"fmt"
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// AccessMode constants define supported access modes.
const (
	AccessModeLocal      = "local"
	AccessModeCluster    = "cluster"
	AccessModeNamespaces = "namespaces"
	AccessModeNamespace  = "namespace"
)

// PodNamespaceFile is the standard Kubernetes downward API path for the pod's namespace.
const PodNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// AccessConfig defines how fluxcd-policyctl accesses Kubernetes resources.
type AccessConfig struct {
	Mode              string   // "local", "cluster", "namespaces", "namespace"
	AllowedNamespaces []string // for "namespaces" mode
	PodNamespace      string   // auto-detected for "namespace" mode
}

// ValidAccessModes returns the list of valid access mode strings.
func ValidAccessModes() (modes []string) {
	modes = []string{AccessModeLocal, AccessModeCluster, AccessModeNamespaces, AccessModeNamespace}
	return modes
}

// IsValidAccessMode returns true if the given mode string is a valid access mode.
func IsValidAccessMode(mode string) (valid bool) {
	for _, m := range ValidAccessModes() {
		if m == mode {
			valid = true
			return valid
		}
	}

	valid = false
	return valid
}

// DetectPodNamespace reads the pod namespace from the Kubernetes downward API file.
// If the file does not exist, it returns the provided fallback namespace.
func DetectPodNamespace(fallback string) (ns string) {
	data, readErr := os.ReadFile(PodNamespaceFile)
	if readErr != nil {
		ns = fallback
		return ns
	}

	ns = string(data)
	return ns
}

// RunServer starts the fluxcd-policyctl web server.
func RunServer(address string, namespace string, authConfig *AuthConfig, accessConfig *AccessConfig, logger *zap.Logger) (err error) {
	// Auto-detect pod namespace for "namespace" mode.
	if accessConfig.Mode == AccessModeNamespace && accessConfig.PodNamespace == "" {
		accessConfig.PodNamespace = DetectPodNamespace(namespace)
	}

	// Initialize services.
	kubeConfigService := NewKubeConfigService(logger)
	policyService := NewPolicyService(kubeConfigService, namespace, logger)

	// Set up routes.
	mux := http.NewServeMux()
	handler := NewHandler(policyService, kubeConfigService, authConfig, accessConfig, logger)
	handler.RegisterRoutes(mux)

	// Set up UI routes for SPA serving.
	SetupUIRoutes(mux)

	logger.Info("Server starting",
		zap.String("address", address),
		zap.String("namespace", namespace),
		zap.String("accessMode", accessConfig.Mode),
		zap.Bool("authEnabled", authConfig.Enabled),
	)

	// Wrap with request logging.
	logged := loggingMiddleware(logger, mux)

	runErr := http.ListenAndServe(address, logged)
	if runErr != nil {
		err = fmt.Errorf("failed to start server: %w", runErr)
		return err
	}

	return err
}

// loggingMiddleware logs each request.
func loggingMiddleware(logger *zap.Logger, next http.Handler) (handler http.Handler) {
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})

	return handler
}

func init() { //nolint:gochecknoinits // suppress log prefix.
	log.SetFlags(0)
}
