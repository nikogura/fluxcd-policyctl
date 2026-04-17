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
	"context"
	"fmt"
	"sort"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PolicyView represents an ImagePolicy with correlated ImageRepository data for API responses.
type PolicyView struct {
	Name              string   `json:"name"`
	Namespace         string   `json:"namespace"`
	ImageRepository   string   `json:"imageRepository"`
	ImageURL          string   `json:"imageUrl"`
	SemverRange       string   `json:"semverRange"`
	LatestVersion     string   `json:"latestVersion"`
	AvailableVersions []string `json:"availableVersions"`
	LastUpdated       string   `json:"lastUpdated"`
	Ready             bool     `json:"ready"`
	Message           string   `json:"message"`
}

// CreatePolicyRequest is the request body for creating an ImagePolicy.
type CreatePolicyRequest struct {
	Name            string `json:"name" binding:"required"`
	Namespace       string `json:"namespace" binding:"required"`
	ImageRepository string `json:"imageRepository" binding:"required"`
	SemverRange     string `json:"semverRange" binding:"required"`
}

// UpdatePolicyRequest is the request body for updating an ImagePolicy's semver range.
type UpdatePolicyRequest struct {
	SemverRange string `json:"semverRange" binding:"required"`
}

// PolicyService provides operations for managing Flux ImagePolicies.
type PolicyService struct {
	kubeConfig  *KubeConfigService
	logger      *zap.Logger
	clientCache map[string]client.Client
	namespace   string
}

// NewPolicyService creates a new PolicyService.
func NewPolicyService(kubeConfig *KubeConfigService, namespace string, logger *zap.Logger) (service *PolicyService) {
	service = &PolicyService{
		kubeConfig:  kubeConfig,
		logger:      logger,
		clientCache: make(map[string]client.Client),
		namespace:   namespace,
	}
	return service
}

// getClientForCluster returns a controller-runtime client for the specified cluster, using a cache.
func (ps *PolicyService) getClientForCluster(clusterName string) (k8sClient client.Client, err error) {
	// Check cache first
	cached, exists := ps.clientCache[clusterName]
	if exists {
		k8sClient = cached
		return k8sClient, err
	}

	// Build rest config for the cluster
	restCfg, restErr := ps.kubeConfig.CreateRestConfigForCluster(clusterName)
	if restErr != nil {
		err = fmt.Errorf("failed to create rest config for cluster %q: %w", clusterName, restErr)
		return k8sClient, err
	}

	// Build the scheme with Flux CRDs
	scheme, schemeErr := buildScheme()
	if schemeErr != nil {
		err = fmt.Errorf("failed to build scheme: %w", schemeErr)
		return k8sClient, err
	}

	// Create controller-runtime client
	k8sClient, err = client.New(restCfg, client.Options{Scheme: scheme})
	if err != nil {
		err = fmt.Errorf("failed to create client for cluster %q: %w", clusterName, err)
		return k8sClient, err
	}

	// Cache the client
	ps.clientCache[clusterName] = k8sClient

	ps.logger.Info("Created controller-runtime client for cluster", zap.String("cluster", clusterName))
	return k8sClient, err
}

// ListPolicies lists ImagePolicies and correlates them with ImageRepositories.
func (ps *PolicyService) ListPolicies(ctx context.Context, clusterName string, namespace string) (policies []PolicyView, err error) {
	k8sClient, clientErr := ps.getClientForCluster(clusterName)
	if clientErr != nil {
		err = clientErr
		return policies, err
	}

	// List ImagePolicies
	var policyList imagev1.ImagePolicyList
	listOpts := &client.ListOptions{Namespace: namespace}

	err = k8sClient.List(ctx, &policyList, listOpts)
	if err != nil {
		err = fmt.Errorf("failed to list ImagePolicies in namespace %q: %w", namespace, err)
		return policies, err
	}

	// List ImageRepositories for correlation
	var repoList imagev1.ImageRepositoryList

	err = k8sClient.List(ctx, &repoList, listOpts)
	if err != nil {
		err = fmt.Errorf("failed to list ImageRepositories in namespace %q: %w", namespace, err)
		return policies, err
	}

	// Build repo lookup map
	repoMap := make(map[string]*imagev1.ImageRepository)
	for idx := range repoList.Items {
		repo := &repoList.Items[idx]
		repoMap[repo.Name] = repo
	}

	// Build policy views
	policies = make([]PolicyView, 0, len(policyList.Items))

	for idx := range policyList.Items {
		policy := &policyList.Items[idx]
		view := buildPolicyView(policy, repoMap)
		policies = append(policies, view)
	}

	// Sort by name
	sort.Slice(policies, func(i, j int) (less bool) {
		less = policies[i].Name < policies[j].Name
		return less
	})

	return policies, err
}

// buildPolicyView creates a PolicyView from an ImagePolicy and correlated ImageRepositories.
func buildPolicyView(policy *imagev1.ImagePolicy, repoMap map[string]*imagev1.ImageRepository) (view PolicyView) {
	view = PolicyView{
		Name:      policy.Name,
		Namespace: policy.Namespace,
	}

	// Extract semver range
	if policy.Spec.Policy.SemVer != nil {
		view.SemverRange = policy.Spec.Policy.SemVer.Range
	}

	// Get the referenced ImageRepository name
	repoName := policy.Spec.ImageRepositoryRef.Name
	view.ImageRepository = repoName

	// Correlate with ImageRepository
	repo, repoExists := repoMap[repoName]
	if repoExists {
		view.ImageURL = repo.Spec.Image

		// Get available versions from last scan result
		if repo.Status.LastScanResult != nil {
			view.AvailableVersions = repo.Status.LastScanResult.LatestTags
		}
	}

	// Get latest image from policy status
	if policy.Status.LatestRef != nil {
		view.LatestVersion = policy.Status.LatestRef.Tag
	}

	// Check ready condition
	for _, condition := range policy.Status.Conditions {
		if condition.Type == "Ready" {
			view.Ready = condition.Status == metav1.ConditionTrue
			view.Message = condition.Message
			view.LastUpdated = condition.LastTransitionTime.Format("2006-01-02T15:04:05Z")
		}
	}

	return view
}

// GetPolicy retrieves a single ImagePolicy by name.
func (ps *PolicyService) GetPolicy(ctx context.Context, clusterName string, namespace string, name string) (policy PolicyView, err error) {
	k8sClient, clientErr := ps.getClientForCluster(clusterName)
	if clientErr != nil {
		err = clientErr
		return policy, err
	}

	// Get the ImagePolicy
	var imagePolicy imagev1.ImagePolicy
	key := client.ObjectKey{Namespace: namespace, Name: name}

	err = k8sClient.Get(ctx, key, &imagePolicy)
	if err != nil {
		err = fmt.Errorf("failed to get ImagePolicy %q in namespace %q: %w", name, namespace, err)
		return policy, err
	}

	// Get correlated ImageRepository
	repoMap := make(map[string]*imagev1.ImageRepository)
	repoName := imagePolicy.Spec.ImageRepositoryRef.Name

	var imageRepo imagev1.ImageRepository
	repoKey := client.ObjectKey{Namespace: namespace, Name: repoName}
	repoErr := k8sClient.Get(ctx, repoKey, &imageRepo)

	if repoErr == nil {
		repoMap[repoName] = &imageRepo
	}

	policy = buildPolicyView(&imagePolicy, repoMap)
	return policy, err
}

// CreatePolicy creates a new ImagePolicy with a semver policy.
func (ps *PolicyService) CreatePolicy(ctx context.Context, clusterName string, req CreatePolicyRequest) (err error) {
	k8sClient, clientErr := ps.getClientForCluster(clusterName)
	if clientErr != nil {
		err = clientErr
		return err
	}

	policy := &imagev1.ImagePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: imagev1.ImagePolicySpec{
			ImageRepositoryRef: meta.NamespacedObjectReference{
				Name: req.ImageRepository,
			},
			Policy: imagev1.ImagePolicyChoice{
				SemVer: &imagev1.SemVerPolicy{
					Range: req.SemverRange,
				},
			},
		},
	}

	err = k8sClient.Create(ctx, policy)
	if err != nil {
		err = fmt.Errorf("failed to create ImagePolicy %q in namespace %q: %w", req.Name, req.Namespace, err)
		return err
	}

	ps.logger.Info("Created ImagePolicy",
		zap.String("name", req.Name),
		zap.String("namespace", req.Namespace),
		zap.String("cluster", clusterName),
	)
	return err
}

// UpdatePolicy updates the semver range of an existing ImagePolicy.
func (ps *PolicyService) UpdatePolicy(ctx context.Context, clusterName string, namespace string, name string, semverRange string) (err error) {
	k8sClient, clientErr := ps.getClientForCluster(clusterName)
	if clientErr != nil {
		err = clientErr
		return err
	}

	// Get existing policy
	var imagePolicy imagev1.ImagePolicy
	key := client.ObjectKey{Namespace: namespace, Name: name}

	err = k8sClient.Get(ctx, key, &imagePolicy)
	if err != nil {
		err = fmt.Errorf("failed to get ImagePolicy %q in namespace %q: %w", name, namespace, err)
		return err
	}

	// Update semver range
	if imagePolicy.Spec.Policy.SemVer == nil {
		imagePolicy.Spec.Policy.SemVer = &imagev1.SemVerPolicy{}
	}
	imagePolicy.Spec.Policy.SemVer.Range = semverRange

	err = k8sClient.Update(ctx, &imagePolicy)
	if err != nil {
		err = fmt.Errorf("failed to update ImagePolicy %q in namespace %q: %w", name, namespace, err)
		return err
	}

	ps.logger.Info("Updated ImagePolicy",
		zap.String("name", name),
		zap.String("namespace", namespace),
		zap.String("cluster", clusterName),
		zap.String("semverRange", semverRange),
	)
	return err
}

// DeletePolicy deletes an ImagePolicy.
func (ps *PolicyService) DeletePolicy(ctx context.Context, clusterName string, namespace string, name string) (err error) {
	k8sClient, clientErr := ps.getClientForCluster(clusterName)
	if clientErr != nil {
		err = clientErr
		return err
	}

	policy := &imagev1.ImagePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err = k8sClient.Delete(ctx, policy)
	if err != nil {
		err = fmt.Errorf("failed to delete ImagePolicy %q in namespace %q: %w", name, namespace, err)
		return err
	}

	ps.logger.Info("Deleted ImagePolicy",
		zap.String("name", name),
		zap.String("namespace", namespace),
		zap.String("cluster", clusterName),
	)
	return err
}

// ListNamespaces lists all namespaces in the specified cluster.
func (ps *PolicyService) ListNamespaces(ctx context.Context, clusterName string) (namespaces []string, err error) {
	k8sClient, clientErr := ps.getClientForCluster(clusterName)
	if clientErr != nil {
		err = clientErr
		return namespaces, err
	}

	var nsList corev1.NamespaceList

	err = k8sClient.List(ctx, &nsList)
	if err != nil {
		err = fmt.Errorf("failed to list namespaces: %w", err)
		return namespaces, err
	}

	namespaces = make([]string, 0, len(nsList.Items))
	for idx := range nsList.Items {
		namespaces = append(namespaces, nsList.Items[idx].Name)
	}

	sort.Strings(namespaces)
	return namespaces, err
}
