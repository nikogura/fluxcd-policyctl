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
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikogura/fluxcd-policyctl/pkg/ui"
)

// SetupUIRoutes configures the UI routes for serving the embedded frontend.
func SetupUIRoutes(router *gin.Engine) {
	// Get the subdirectory containing the built UI files from the ui package
	uiFS, err := fs.Sub(ui.Files, "dist")
	if err != nil {
		// If UI files don't exist (development), serve a placeholder
		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "fluxcd-policyctl API server running. UI not available in development mode.",
			})
		})
		return
	}

	// Serve favicon.ico directly from root
	router.GET("/favicon.ico", func(c *gin.Context) {
		faviconData, faviconErr := fs.ReadFile(uiFS, "favicon.ico")
		if faviconErr != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Header("Content-Type", "image/x-icon")
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(http.StatusOK, "image/x-icon", faviconData)
	})

	// Serve Next.js static assets
	router.StaticFS("/_next", http.FS(uiFS))

	// Serve the main application for all non-API routes (SPA fallback)
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Skip health check
		if path == "/health" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}

		// For all other routes, serve the main HTML file (SPA routing)
		indexContent, readErr := fs.ReadFile(uiFS, "index.html")
		if readErr != nil {
			c.String(http.StatusInternalServerError, "UI not available")
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.String(http.StatusOK, string(indexContent))
	})
}
