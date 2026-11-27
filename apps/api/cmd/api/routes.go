package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

//go:embed docs.html
var docsHTML []byte

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func routes(handler http.Handler, environment, docsDir string, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	if environment == "dev" || environment == "test" {
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if requestID == "" {
			requestID = uuid.NewString()
			r.Header.Set("X-Request-Id", requestID)
		}
		w.Header().Set("X-Request-Id", requestID)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		mux.ServeHTTP(recorder, r)

		if logger != nil {
			logger.Info("http_request",
				"servicio", serviceName,
				"version", version,
				"request_id", requestID,
				"metodo", r.Method,
				"ruta", r.URL.Path,
				"estado", recorder.status,
				"duracion_ms", time.Since(start).Milliseconds(),
			)
		}
	})
}
