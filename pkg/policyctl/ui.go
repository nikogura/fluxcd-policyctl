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
	"strings"

	"github.com/nikogura/fluxcd-policyctl/pkg/ui"
)

// SetupUIRoutes configures the UI routes for serving the embedded frontend.
func SetupUIRoutes(mux *http.ServeMux) {
	// Get the subdirectory containing the built UI files.
	uiFS, err := fs.Sub(ui.Files, "dist")
	if err != nil {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{
				"message": "fluxcd-policyctl API server. UI not available — run make build-ui first.",
			})
		})

		return
	}

	// Serve _next/ static assets (JS chunks, CSS, etc.) directly.
	nextFS, subErr := fs.Sub(uiFS, "_next")
	if subErr == nil {
		mux.Handle("/_next/", http.StripPrefix("/_next/", http.FileServer(http.FS(nextFS))))
	}

	// Serve favicon.
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		faviconData, faviconErr := fs.ReadFile(uiFS, "favicon.ico")
		if faviconErr != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(faviconData)
	})

	// SPA fallback: serve index.html for all non-API, non-static routes.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// API routes should not fall through here.
		if strings.HasPrefix(path, "/api") {
			writeError(w, http.StatusNotFound, "API endpoint not found")
			return
		}

		if path == "/health" {
			writeError(w, http.StatusNotFound, "Not found")
			return
		}

		// Serve index.html for SPA routing.
		indexContent, readErr := fs.ReadFile(uiFS, "index.html")
		if readErr != nil {
			http.Error(w, "UI not available", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexContent)
	})
}
