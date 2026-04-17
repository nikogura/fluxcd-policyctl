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

package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/nikogura/fluxcd-policyctl/pkg/policyctl"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // Cobra boilerplate
var verbose bool

//nolint:gochecknoglobals // Cobra boilerplate
var logLevel string

//nolint:gochecknoglobals // Cobra boilerplate
var address string

//nolint:gochecknoglobals // Cobra boilerplate
var namespace string

//nolint:gochecknoglobals // Cobra boilerplate
var oidcIssuer string

//nolint:gochecknoglobals // Cobra boilerplate
var oidcAudience string

//nolint:gochecknoglobals // Cobra boilerplate
var oidcGroups string

//nolint:gochecknoglobals // Cobra boilerplate
var accessMode string

//nolint:gochecknoglobals // Cobra boilerplate
var allowedNamespaces string

// rootCmd represents the base command when called without any subcommands.
//
//nolint:gochecknoglobals // Cobra boilerplate
var rootCmd = &cobra.Command{
	Use:   "fluxcd-policyctl",
	Short: "Flux CD ImagePolicy Management Tool",
	Long: `fluxcd-policyctl is a web-based management tool for Flux CD ImagePolicies.

It provides a UI and API for viewing, creating, updating, and deleting
Flux ImagePolicy resources across Kubernetes clusters.

The server provides:
- Web UI for managing Flux ImagePolicies
- Multi-cluster support via kubeconfig
- Optional OIDC authentication
- REST API for programmatic access

Kubernetes configuration:
- Uses in-cluster config when running in a pod
- Falls back to ~/.kube/config for local development

Example (local development):
  fluxcd-policyctl
  fluxcd-policyctl --bind-address=0.0.0.0:8080

Example (with OIDC authentication):
  fluxcd-policyctl --oidc-issuer=https://accounts.google.com --oidc-audience=my-app`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := zap.NewProduction()
		if err != nil {
			log.Fatalf("failed to create logger: %s", err)
		}
		defer func() {
			_ = logger.Sync()
		}()

		authConfig := &policyctl.AuthConfig{
			Enabled:   oidcIssuer != "",
			IssuerURL: oidcIssuer,
			Audience:  oidcAudience,
		}

		if oidcGroups != "" {
			authConfig.AllowedGroups = strings.Split(oidcGroups, ",")
		}

		// Resolve access mode: flag takes precedence, then env var, then default.
		resolvedMode := resolveAccessMode()
		if !policyctl.IsValidAccessMode(resolvedMode) {
			logger.Fatal("Invalid access mode",
				zap.String("mode", resolvedMode),
				zap.Strings("valid", policyctl.ValidAccessModes()),
			)
		}

		accessConfig := &policyctl.AccessConfig{
			Mode: resolvedMode,
		}

		// Resolve allowed namespaces for "namespaces" mode.
		resolvedNS := resolveAllowedNamespaces()
		if resolvedNS != "" {
			accessConfig.AllowedNamespaces = strings.Split(resolvedNS, ",")
		}

		err = policyctl.RunServer(address, namespace, authConfig, accessConfig, logger)
		if err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	},
}

// resolveAccessMode returns the access mode from the flag, falling back to the env var, then to "local".
func resolveAccessMode() (mode string) {
	if accessMode != "" {
		mode = accessMode
		return mode
	}

	envMode := os.Getenv("POLICYCTL_ACCESS_MODE")
	if envMode != "" {
		mode = envMode
		return mode
	}

	mode = "local"
	return mode
}

// resolveAllowedNamespaces returns the allowed namespaces from the flag, falling back to the env var.
func resolveAllowedNamespaces() (ns string) {
	if allowedNamespaces != "" {
		ns = allowedNamespaces
		return ns
	}

	ns = os.Getenv("POLICYCTL_NAMESPACES")
	return ns
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//nolint:gochecknoinits // Cobra boilerplate
func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log Level (debug, info, warn, error)")
	rootCmd.Flags().StringVarP(&address, "bind-address", "b", "0.0.0.0:9999", "Address (host and port) on which to listen")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "flux-system", "Default namespace for Flux resources")
	rootCmd.Flags().StringVar(&oidcIssuer, "oidc-issuer", "", "OIDC issuer URL (enables authentication when set)")
	rootCmd.Flags().StringVar(&oidcAudience, "oidc-audience", "", "OIDC audience for token validation")
	rootCmd.Flags().StringVar(&oidcGroups, "oidc-groups", "", "Comma-separated list of allowed OIDC groups")
	rootCmd.Flags().StringVarP(&accessMode, "access-mode", "m", "", "Access mode: local, cluster, namespaces, namespace (default \"local\", env: POLICYCTL_ACCESS_MODE)")
	rootCmd.Flags().StringVar(&allowedNamespaces, "allowed-namespaces", "", "Comma-separated list of allowed namespaces for 'namespaces' mode (env: POLICYCTL_NAMESPACES)")
}
