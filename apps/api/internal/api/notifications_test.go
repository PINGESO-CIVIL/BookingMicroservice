package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"bookingms/internal/api"
	"bookingms/internal/service"
	"bookingms/internal/store"
	"bookingms/internal/store/notificaciones"

	"github.com/google/uuid"
)

func TestConfirmationThroughNotificationsContract(t *testing.T) {
	var keys []string
	var mu sync.Mutex
	g10 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/notificaciones" ||
			r.Header.Get("Authorization") != "Bearer demo-user" ||
			r.Header.Get("X-Request-Id") != "agenda-test" {
			t.Error("solicitud inesperada a G10")
		}
		key := r.Header.Get("Idempotency-Key")
		if _, err := uuid.Parse(key); err != nil {
			t.Errorf("se esperaba una clave UUID: %v", err)
		}
		mu.Lock()
		keys = append(keys, key)
		count := len(keys)
		mu.Unlock()
		var body notificaciones.NotificacionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
			body.Destinatario != "agenda@example.test" ||
			!strings.Contains(body.Cuerpo, "2026-11-10T22:00:00Z") {
			t.Errorf("payload de confirmación inesperado: %v", err)
		}
		if count == 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"estado":"enviado_mock"}`))
	}))
	defer g10.Close()

	sender, err := notificaciones.NewSender(g10.URL, "agenda@example.test")
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := store.NewFixtures()
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(fixtures, nil)
	svc.SetNotifier(sender)
	handler := api.New(svc)

	for _, expected := range []int{201, 201, 503} {
		body := `{"resourceId":"00000000-0000-4000-8000-100000000001","startAt":"2026-11-10T22:00:00Z","endAt":"2026-11-10T23:00:00Z"}`
		req := httptest.NewRequest("POST", "/api/v1/reservas", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer demo-user")
		req.Header.Set("X-Request-Id", "agenda-test")
		req.Header.Set("Idempotency-Key", "intencion-agenda")
		result := httptest.NewRecorder()
		handler.ServeHTTP(result, req)
		if result.Code != expected {
			t.Fatalf("HTTP %d, esperado %d: %s", result.Code, expected, result.Body)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(keys) != 3 || keys[0] != keys[1] || keys[1] != keys[2] {
		t.Fatal("la misma intención debe usar la misma clave de G10")
	}
}
