package server

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/sakashita/memento-memo/web"
)

func spaHandler() http.Handler {
	if web.Assets == nil {
		// Dev mode: no embedded assets, frontend served by Vite
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Frontend not available (dev mode)", http.StatusNotFound)
		})
	}

	distFS, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		// If dist directory doesn't exist in embed, serve nothing
		return http.NotFoundHandler()
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		f, err := fs.Stat(distFS, strings.TrimPrefix(path, "/"))
		if err == nil && !f.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for unknown paths
		indexFile, err := distFS.(fs.ReadFileFS).ReadFile("index.html")
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexFile)
	})
}
