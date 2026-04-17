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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RunServer starts the fluxcd-policyctl web server.
func RunServer(address string, namespace string, authConfig *AuthConfig, logger *zap.Logger) (err error) {
	gin.SetMode(gin.ReleaseMode)

	// Initialize services
	kubeConfigService := NewKubeConfigService(logger)
	policyService := NewPolicyService(kubeConfigService, namespace, logger)

	// Set up router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Set up routes
	handler := NewHandler(policyService, kubeConfigService, authConfig, logger)
	handler.RegisterRoutes(router)

	// Set up UI routes for SPA serving
	SetupUIRoutes(router)

	logger.Info("Server starting",
		zap.String("address", address),
		zap.String("namespace", namespace),
		zap.Bool("authEnabled", authConfig.Enabled),
	)

	runErr := router.Run(address)
	if runErr != nil {
		err = fmt.Errorf("failed to start server: %w", runErr)
		return err
	}

	return err
}
