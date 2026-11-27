package main

import (
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

func testRoutes(t *testing.T, env, docsDir string) http.Handler {
	t.Helper()
	fixtures, err := store.NewFixtures()
	if err != nil {
		t.Fatal(err)
	}
	database, err := store.OpenPostgres("")
	if err != nil {
		t.Fatal(err)
	}
	return routes(api.New(service.New(fixtures, database)), env, docsDir)
}

func TestRoutesWithoutDatabase(t *testing.T) {
	handler := testRoutes(t, "", t.TempDir())
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
		{"", "/api/docs", http.StatusNotFound, ""},
		{"prod", "/api/docs/", http.StatusNotFound, ""},
		{"prod", "/api/docs/openapi.yaml", http.StatusNotFound, ""},
		{"prod", "/api/docs/swagger-ui-bundle.js", http.StatusNotFound, ""},
	} {
		response := httptest.NewRecorder()
		testRoutes(t, tc.env, dir).ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.body) {
			t.Errorf("API_ENV=%q %s = %d, want %d with %q", tc.env, tc.path, response.Code, tc.status, tc.body)
		}
		if tc.status == http.StatusTemporaryRedirect && response.Header().Get("Location") != "/api/docs/" {
			t.Errorf("unexpected Swagger redirect: %q", response.Header().Get("Location"))
		}
	}
}
