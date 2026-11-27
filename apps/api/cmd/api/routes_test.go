package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bookingms/internal/api"
	"bookingms/internal/service"
	"bookingms/internal/store"
)

func testRoutes(t *testing.T, env, docsDir string, logger *slog.Logger) http.Handler {
	t.Helper()
	fixtures, err := store.NewFixtures()
	if err != nil {
		t.Fatal(err)
	}
	database, err := store.OpenPostgres("")
	if err != nil {
		t.Fatal(err)
	}
	return routes(api.New(service.New(fixtures, database)), env, docsDir, logger)
}

func TestRoutesWithoutDatabase(t *testing.T) {
	handler := testRoutes(t, "", t.TempDir(), nil)
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/healthz", http.StatusOK},
		{"/readyz", http.StatusServiceUnavailable},
		{"/api/v1/reservas", http.StatusOK},
	} {
		request := httptest.NewRequest(http.MethodGet, tc.path, nil)
		request.Header.Set("Authorization", "Bearer demo-user")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != tc.status {
			t.Errorf("%s = %d, want %d", tc.path, response.Code, tc.status)
		}
	}
}

func TestDocsEnvironment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte("openapi: 3.1.0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		env    string
		path   string
		status int
		body   string
	}{
		{"dev", "/api/docs", http.StatusTemporaryRedirect, ""},
		{"dev", "/api/docs/", http.StatusOK, "SwaggerUIBundle"},
		{"dev", "/api/docs/openapi.yaml", http.StatusOK, "openapi: 3.1.0"},
		{"test", "/api/docs", http.StatusTemporaryRedirect, ""},
		{"test", "/api/docs/", http.StatusOK, "SwaggerUIBundle"},
		{"test", "/api/docs/openapi.yaml", http.StatusOK, "openapi: 3.1.0"},
		{"", "/api/docs", http.StatusNotFound, ""},
		{"prod", "/api/docs/", http.StatusNotFound, ""},
		{"prod", "/api/docs/openapi.yaml", http.StatusNotFound, ""},
		{"prod", "/api/docs/swagger-ui-bundle.js", http.StatusNotFound, ""},
	} {
		response := httptest.NewRecorder()
		testRoutes(t, tc.env, dir, nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.body) {
			t.Errorf("API_ENV=%q %s = %d, want %d with %q", tc.env, tc.path, response.Code, tc.status, tc.body)
		}
		if tc.status == http.StatusTemporaryRedirect && response.Header().Get("Location") != "/api/docs/" {
			t.Errorf("unexpected Swagger redirect: %q", response.Header().Get("Location"))
		}
	}
}

func TestObservabilityLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	handler := testRoutes(t, "test", t.TempDir(), logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Authorization", "Bearer secreto-no-loguear")
	req.AddCookie(&http.Cookie{Name: "sesion", Value: "cookie-secreta"})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
	reqID := res.Header().Get("X-Request-Id")
	if reqID == "" {
		t.Fatal("se esperaba X-Request-Id generado en la respuesta")
	}
	raw := buf.String()
	if strings.Contains(raw, "secreto-no-loguear") || strings.Contains(raw, "cookie-secreta") {
		t.Fatalf("el log no debe exponer tokens ni cookies: %s", raw)
	}
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log no es JSON válido: %v (%s)", err, raw)
	}
	for _, key := range []string{"servicio", "version", "request_id", "metodo", "ruta", "estado", "duracion_ms"} {
		if _, ok := entry[key]; !ok {
			t.Errorf("falta clave %q en log estructurado: %s", key, raw)
		}
	}
	if entry["request_id"] != reqID {
		t.Errorf("request_id en log = %v, esperado %q", entry["request_id"], reqID)
	}
}
