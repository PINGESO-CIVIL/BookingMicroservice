package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMockReservationFlow(t *testing.T) {
	base := strings.TrimRight(os.Getenv("BOOKING_E2E_API_URL"), "/")
	if base == "" {
		t.Skip("BOOKING_E2E_API_URL debe apuntar a la API de Compose")
	}
	client := &http.Client{Timeout: 5 * time.Second}
	call := func(method, path, token, body string) (int, []byte) {
		t.Helper()
		request, err := http.NewRequestWithContext(t.Context(), method, base+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		if body != "" {
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "mock-e2e")
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, data
	}

	live, _ := call(http.MethodGet, "/healthz", "", "")
	ready, _ := call(http.MethodGet, "/readyz", "", "")
	forbidden, _ := call(http.MethodGet, "/api/v1/admin/reservas", "demo-user", "")
	before, original := call(http.MethodGet, "/api/v1/reservas", "demo-user", "")
	body := `{"resourceId":"00000000-0000-4000-8000-100000000001","startAt":"2026-11-10T21:00:00Z","endAt":"2026-11-10T22:00:00Z"}`
	created, _ := call(http.MethodPost, "/api/v1/reservas", "demo-user", body)
	after, current := call(http.MethodGet, "/api/v1/reservas", "demo-user", "")
	decode := func(data []byte) any {
		t.Helper()
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}

	report := fmt.Sprintf("# Reserva simulada sobre Compose\n\n"+
		"- Salud HTTP: %d\n- PostgreSQL disponible: %d\n"+
		"- Listado administrativo con rol de usuario: %d\n"+
		"- Listado anterior a la creación: %d\n- Creación simulada: %d\n"+
		"- Listado posterior a la creación: %d\n- Los listados coinciden: %t\n",
		live, ready, forbidden, before, created, after, reflect.DeepEqual(decode(original), decode(current)))
	expected, err := os.ReadFile("testdata/reserva-simulada.golden.md")
	if err != nil {
		t.Fatal(err)
	}
	if report != strings.ReplaceAll(string(expected), "\r\n", "\n") {
		t.Fatalf("reporte inesperado:\n%s\nEsperado:\n%s", report, expected)
	}
}
