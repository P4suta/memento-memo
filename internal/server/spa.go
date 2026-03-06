package server

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
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
		p := r.URL.Path
		if p == "/" {
			p = "/index.html"
		}

		// Check if file exists
		cleanPath := strings.TrimPrefix(p, "/")
		f, err := fs.Stat(distFS, cleanPath)
		if err == nil && !f.IsDir() {
			// For non-HTML static assets (JS, CSS, etc.), serve directly
			ext := path.Ext(cleanPath)
			if ext != ".html" {
				fileServer.ServeHTTP(w, r)
				return
			}
			// For HTML files, read and inject nonce
			html, err := distFS.(fs.ReadFileFS).ReadFile(cleanPath)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			serveHTMLWithNonce(w, r, html)
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
		serveHTMLWithNonce(w, r, indexFile)
	})
}

func injectNonce(html []byte, nonce string) []byte {
	return bytes.ReplaceAll(html, []byte("<script>"), []byte(fmt.Sprintf(`<script nonce="%s">`, nonce)))
}

func serveHTMLWithNonce(w http.ResponseWriter, r *http.Request, html []byte) {
	nonce := GetCSPNonce(r.Context())
	if nonce != "" {
		html = injectNonce(html, nonce)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}
