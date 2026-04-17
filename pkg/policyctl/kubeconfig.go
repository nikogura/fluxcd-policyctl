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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ClusterInfo represents a kubectl cluster.
type ClusterInfo struct {
	Name    string `json:"name"`
	Current bool   `json:"current"`
}

// KubeConfigService handles kubeconfig operations.
type KubeConfigService struct {
	logger         *zap.Logger
	inCluster      bool
	kubeconfigPath string
}

// NewKubeConfigService creates a new kubeconfig service.
func NewKubeConfigService(logger *zap.Logger) (service *KubeConfigService) {
	service = &KubeConfigService{
		logger: logger,
	}

	// Check if running in cluster
	_, err := rest.InClusterConfig()
	service.inCluster = err == nil

	if !service.inCluster {
		// Determine kubeconfig path
		service.kubeconfigPath = os.Getenv("KUBECONFIG")
		if service.kubeconfigPath == "" {
			if home := homeDir(); home != "" {
				service.kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}
	}

	return service
}

// IsInCluster returns true if running inside a Kubernetes cluster.
func (kcs *KubeConfigService) IsInCluster() (inCluster bool) {
	inCluster = kcs.inCluster
	return inCluster
}

// GetClusters returns available kubectl clusters.
func (kcs *KubeConfigService) GetClusters() (clusters []ClusterInfo, err error) {
	if kcs.inCluster {
		err = errors.New("clusters not available when running in cluster")
		return clusters, err
	}

	if kcs.kubeconfigPath == "" {
		err = errors.New("kubeconfig path not found")
		return clusters, err
	}

	// Load the kubeconfig file
	var config *clientcmdapi.Config
	config, err = clientcmd.LoadFromFile(kcs.kubeconfigPath)
	if err != nil {
		err = fmt.Errorf("failed to load kubeconfig: %w", err)
		return clusters, err
	}

	// Get current cluster from current context
	var currentCluster string
	if config.CurrentContext != "" {
		if context, exists := config.Contexts[config.CurrentContext]; exists {
			currentCluster = context.Cluster
		}
	}

	// Extract unique clusters
	clusterMap := make(map[string]bool)

	for clusterName := range config.Clusters {
		if !clusterMap[clusterName] {
			clusterMap[clusterName] = true
			clusterInfo := ClusterInfo{
				Name:    clusterName,
				Current: clusterName == currentCluster,
			}
			clusters = append(clusters, clusterInfo)
		}
	}

	// Sort clusters alphabetically, but put current cluster first
	sort.Slice(clusters, func(i, j int) (less bool) {
		if clusters[i].Current {
			less = true
			return less
		}
		if clusters[j].Current {
			less = false
			return less
		}
		less = clusters[i].Name < clusters[j].Name
		return less
	})

	return clusters, err
}

// GetCurrentContext returns the current context name.
func (kcs *KubeConfigService) GetCurrentContext() (currentContext string, err error) {
	if kcs.inCluster {
		err = errors.New("current context not available when running in cluster")
		return currentContext, err
	}

	var config *clientcmdapi.Config
	config, err = clientcmd.LoadFromFile(kcs.kubeconfigPath)
	if err != nil {
		err = fmt.Errorf("failed to load kubeconfig: %w", err)
		return currentContext, err
	}

	currentContext = config.CurrentContext
	return currentContext, err
}

// GetCurrentCluster returns the current cluster name.
func (kcs *KubeConfigService) GetCurrentCluster() (clusterName string, err error) {
	if kcs.inCluster {
		err = errors.New("current cluster not available when running in cluster")
		return clusterName, err
	}

	var config *clientcmdapi.Config
	config, err = clientcmd.LoadFromFile(kcs.kubeconfigPath)
	if err != nil {
		err = fmt.Errorf("failed to load kubeconfig: %w", err)
		return clusterName, err
	}

	if config.CurrentContext != "" {
		if context, exists := config.Contexts[config.CurrentContext]; exists {
			clusterName = context.Cluster
			return clusterName, err
		}
	}

	return clusterName, err
}

// CreateRestConfigForCluster creates a rest.Config for the specified cluster.
func (kcs *KubeConfigService) CreateRestConfigForCluster(clusterName string) (restConfig *rest.Config, err error) {
	if kcs.inCluster {
		// When in cluster, ignore cluster name and use in-cluster config
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			err = fmt.Errorf("failed to get in-cluster config: %w", err)
			return restConfig, err
		}
		return restConfig, err
	}

	// Load kubeconfig
	var config *clientcmdapi.Config
	config, err = clientcmd.LoadFromFile(kcs.kubeconfigPath)
	if err != nil {
		err = fmt.Errorf("failed to load kubeconfig: %w", err)
		return restConfig, err
	}

	// Verify cluster exists
	cluster, exists := config.Clusters[clusterName]
	if !exists {
		err = fmt.Errorf("cluster %q not found in kubeconfig", clusterName)
		return restConfig, err
	}

	// Find the most commonly used user for this cluster
	var userName string
	userName, err = kcs.findBestUserForCluster(config, clusterName)
	if err != nil {
		err = fmt.Errorf("failed to find suitable user for cluster %q: %w", clusterName, err)
		return restConfig, err
	}

	// Create a virtual context combining the cluster and user
	virtualContext := &clientcmdapi.Context{
		Cluster:  clusterName,
		AuthInfo: userName,
	}

	// Create client config with virtual context
	clientConfig := clientcmd.NewDefaultClientConfig(clientcmdapi.Config{
		Clusters:  map[string]*clientcmdapi.Cluster{clusterName: cluster},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{userName: config.AuthInfos[userName]},
		Contexts:  map[string]*clientcmdapi.Context{"virtual": virtualContext},
	}, &clientcmd.ConfigOverrides{
		CurrentContext: "virtual",
	})

	restConfig, err = clientConfig.ClientConfig()
	if err != nil {
		err = fmt.Errorf("failed to create rest config for cluster %q (user %q): %w", clusterName, userName, err)
		return restConfig, err
	}

	kcs.logger.Info("Created rest config for cluster", zap.String("cluster", clusterName), zap.String("user", userName))
	return restConfig, err
}

// findBestUserForCluster finds the most commonly used user for a given cluster.
func (kcs *KubeConfigService) findBestUserForCluster(config *clientcmdapi.Config, clusterName string) (bestUser string, err error) {
	// Count how many contexts use each user for this cluster
	userCounts := make(map[string]int)

	for _, context := range config.Contexts {
		if context.Cluster == clusterName {
			userCounts[context.AuthInfo]++
		}
	}

	if len(userCounts) == 0 {
		err = fmt.Errorf("no contexts found for cluster %q", clusterName)
		return bestUser, err
	}

	// Find the most commonly used user
	var maxCount int

	for user, count := range userCounts {
		if count > maxCount {
			maxCount = count
			bestUser = user
		}
	}

	// Verify the user exists in AuthInfos
	if _, exists := config.AuthInfos[bestUser]; !exists {
		err = fmt.Errorf("user %q not found in kubeconfig AuthInfos", bestUser)
		return bestUser, err
	}

	return bestUser, err
}

// homeDir returns the user's home directory.
func homeDir() (home string) {
	if h := os.Getenv("HOME"); h != "" {
		home = h
		return home
	}
	home = os.Getenv("USERPROFILE") // Windows
	return home
}
