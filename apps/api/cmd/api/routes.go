package main

import (
	_ "embed"
	"net/http"
	"path/filepath"
)

//go:embed docs.html
var docsHTML []byte

func routes(handler http.Handler, environment, docsDir string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	if environment == "dev" {
		mux.HandleFunc("GET /api/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/api/docs/", http.StatusTemporaryRedirect)
		})
		mux.HandleFunc("GET /api/docs/{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(docsHTML)
		})
		mux.HandleFunc("GET /api/docs/{asset}", func(w http.ResponseWriter, r *http.Request) {
			asset := r.PathValue("asset")
			switch asset {
			case "swagger-ui.css", "swagger-ui-bundle.js", "openapi.yaml":
				http.ServeFile(w, r, filepath.Join(docsDir, asset))
			default:
				http.NotFound(w, r)
			}
		})
	}
	return mux
}
