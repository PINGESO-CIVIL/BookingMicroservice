package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"bookingms/internal/api"
	"bookingms/internal/service"
	"bookingms/internal/store"
	"strings"
	"testing"
)

type object = map[string]any

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	fixtures, err := store.NewFixtures()
	if err != nil {
		t.Fatal(err)
	}
	database, err := store.OpenPostgres("")
	if err != nil {
		t.Fatal(err)
	}
	return api.New(service.New(fixtures, database))
}

func call(handler http.Handler, method, path, body, token, scenario string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)
	request.Header.Set("Idempotency-Key", "demo-key")
	request.Header.Set("X-Mock-Scenario", scenario)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
func TestAllFourteenOperations(t *testing.T) {
	fixtureJSON, err := os.ReadFile("../store/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Examples map[string]json.RawMessage `json:"examples"`
	}
	if err := json.Unmarshal(fixtureJSON, &f); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		method, path, example string
		status                int
	}{
		{"GET", "/api/v1/reservas", "", 200},
		{"GET", "/api/v1/admin/reservas", "", 200},
		{"POST", "/api/v1/reservas", "CreateRequest", 201},
		{"GET", "/api/v1/reservas/00000000-0000-4000-8000-200000000001", "", 200},
		{"PATCH", "/api/v1/reservas/00000000-0000-4000-8000-200000000001", "UpdateRequest", 200},
		{"POST", "/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones", "CancelRequest", 200},
		{"GET", "/api/v1/recursos", "", 200},
		{"POST", "/api/v1/recursos", "CreateResourceRequest", 201},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "", 200},
		{"PATCH", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "UpdateResourceRequest", 200},
		{"PUT", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/horarios", "UpdateResourceScheduleRequest", 200},
		{"POST", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos", "CreateResourceClosureRequest", 201},
		{"DELETE", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos/00000000-0000-4000-8000-300000000001", "", 204},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/disponibilidades?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z", "", 200},
	}
	handler := newHandler(t)
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			body := string(f.Examples[tc.example])
			response := call(handler, tc.method, tc.path, body, "Bearer demo-admin", "")
			if response.Code != tc.status {
				t.Fatalf("status %d, body %s", response.Code, response.Body)
			}
			if response.Header().Get("X-Mock-Response") != "true" {
				t.Fatal("missing mock marker")
			}
			if tc.status == 204 {
				if response.Body.Len() != 0 {
					t.Fatal("204 must not contain JSON")
				}
				return
			}
			if !json.Valid(response.Body.Bytes()) {
				t.Fatal("invalid JSON")
			}
		})
	}
}

func TestContractInputLimits(t *testing.T) {
	handler := newHandler(t)
	resource := "/api/v1/recursos/00000000-0000-4000-8000-100000000001"
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"malformed UUID", "GET", "/api/v1/reservas/no-es-uuid", "", 400},
		{"malformed filter UUID", "GET", "/api/v1/reservas?resourceId=no-es-uuid", "", 400},
		{"cursor length", "GET", "/api/v1/reservas?cursor=" + strings.Repeat("x", 2049), "", 400},
		{"resource name length", "PATCH", resource, `{"expectedVersion":1,"name":"` + strings.Repeat("á", 129) + `"}`, 400},
		{"unrecognized field", "PATCH", resource, `{"expectedVersion":1,"Enabled":false}`, 400},
		{"explicit null", "PATCH", resource, `{"expectedVersion":1,"enabled":null}`, 400},
		{"explicit false", "PATCH", resource, `{"expectedVersion":1,"enabled":false}`, 200},
		{"large version", "PATCH", resource, `{"expectedVersion":9007199254740992,"enabled":false}`, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := call(handler, tc.method, tc.path, tc.body, "Bearer demo-admin", "")
			if response.Code != tc.status {
				t.Fatalf("status %d, expected %d: %s", response.Code, tc.status, response.Body)
			}
		})
	}
}
func TestTransportErrorsAndScenarios(t *testing.T) {
	handler := newHandler(t)
	cases := []struct {
		method, path, body, token, scenario, code string
		status                                    int
	}{
		{"GET", "/api/v1/reservas", "", "", "", "UNAUTHENTICATED", 401},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "", "Bearer demo-user", "", "FORBIDDEN", 403},
		{"GET", "/api/v1/reservas/00000000-0000-4000-8000-200000000999", "", "Bearer demo-user", "", "RESERVATION_NOT_FOUND", 404},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000999", "", "Bearer demo-admin", "", "RESOURCE_NOT_FOUND", 404},
		{"POST", "/api/v1/reservas", `{"resourceId":"00000000-0000-4000-8000-100000000001","startAt":"2026-11-10T22:00:00Z","endAt":"2026-11-10T23:00:00Z","responsible":{}}`, "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"PATCH", "/api/v1/reservas/00000000-0000-4000-8000-200000000001", `{"expectedVersion":1}`, "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"POST", "/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones", `{"expectedVersion":0}`, "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"POST", "/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones", `{"expectedVersion":1}`, "Bearer demo-user", "version", "RESERVATION_VERSION_CONFLICT", 409},
		{"POST", "/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones", `{"expectedVersion":1}`, "Bearer demo-user", "deadline", "CANCELLATION_WINDOW_CLOSED", 422},
		{"PATCH", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", `{"expectedVersion":1,"enabled":false}`, "Bearer demo-admin", "conflict", "RESOURCE_SCHEDULE_CONFLICT", 409},
		{"GET", "/api/v1/reservas", "", "Bearer demo-user", "error", "SERVICE_UNAVAILABLE", 503},
		{"GET", "/api/v1/reservas?limit=101", "", "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"GET", "/api/v1/reservas?from=2026-11-10T00:00:00Z", "", "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/disponibilidades?from=2026-11-10T00:00:00Z&to=2027-01-01T00:00:00Z", "", "Bearer demo-user", "", "INVALID_REQUEST", 400},
	}
	for _, tc := range cases {
		t.Run(tc.code+tc.path+tc.scenario, func(t *testing.T) {
			response := call(handler, tc.method, tc.path, tc.body, tc.token, tc.scenario)
			if response.Code != tc.status || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"`+tc.code+`"`)) {
				t.Fatalf("%d %s", response.Code, response.Body)
			}
		})
	}
}
func TestWritesLeaveReadsUnchanged(t *testing.T) {
	handler := newHandler(t)
	before := call(handler, "GET", "/api/v1/reservas", "", "Bearer demo-user", "").Body.String()
	call(handler, "POST", "/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones", `{"expectedVersion":1}`, "Bearer demo-user", "")
	after := call(handler, "GET", "/api/v1/reservas", "", "Bearer demo-user", "").Body.String()
	if before != after {
		t.Fatal("mock write modified later reads")
	}
	before = call(handler, "GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "", "Bearer demo-admin", "").Body.String()
	call(handler, "PATCH", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", `{"expectedVersion":1,"name":"Changed"}`, "Bearer demo-admin", "")
	call(handler, "DELETE", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos/00000000-0000-4000-8000-300000000001", "", "Bearer demo-admin", "")
	after = call(handler, "GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "", "Bearer demo-admin", "").Body.String()
	if before != after {
		t.Fatal("resource mock write modified later reads")
	}
}
func TestPaginationAndCursorBinding(t *testing.T) {
	handler := newHandler(t)
	var first struct {
		Items      []object `json:"items"`
		NextCursor string   `json:"nextCursor"`
	}
	response := call(handler, "GET", "/api/v1/reservas?limit=1", "", "Bearer demo-user", "")
	if err := json.Unmarshal(response.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 1 || first.NextCursor == "" {
		t.Fatal("missing continuation")
	}
	second := call(handler, "GET", "/api/v1/reservas?limit=1&cursor="+first.NextCursor, "", "Bearer demo-user", "")
	var next struct {
		Items []object `json:"items"`
	}
	_ = json.Unmarshal(second.Body.Bytes(), &next)
	if len(next.Items) != 1 || first.Items[0]["id"] == next.Items[0]["id"] {
		t.Fatal("invalid continuation")
	}
	invalid := call(handler, "GET", "/api/v1/reservas?status=confirmed&limit=1&cursor="+first.NextCursor, "", "Bearer demo-user", "")
	if invalid.Code != 400 {
		t.Fatal("cursor accepted with different filters")
	}
}

func TestAdminReadsAllUsersWithoutChangingPersonalAccess(t *testing.T) {
	handler := newHandler(t)
	response := call(handler, "GET", "/api/v1/admin/reservas", "", "Bearer demo-admin", "")
	var page struct {
		Items []object `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || len(page.Items) != 6 {
		t.Fatal("admin list must include every fixture user", response.Body)
	}
	users := map[string]bool{}
	for _, r := range page.Items {
		users[r["responsible"].(map[string]any)["subject"].(string)] = true
	}
	if len(users) != 3 {
		t.Fatal("admin list excluded other users")
	}
	denied := call(handler, "GET", "/api/v1/admin/reservas", "", "Bearer demo-user", "")
	if denied.Code != 403 {
		t.Fatal("ordinary user accessed administrative list")
	}
	hidden := call(handler, "GET", "/api/v1/reservas/00000000-0000-4000-8000-200000000005", "", "Bearer demo-user", "")
	if hidden.Code != 404 {
		t.Fatal("personal endpoint exposed another user's reservation")
	}
	personal := call(handler, "GET", "/api/v1/reservas", "", "Bearer demo-user", "")
	if err := json.Unmarshal(personal.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 4 {
		t.Fatal("personal list changed scope")
	}
	filtered := call(handler, "GET", "/api/v1/admin/reservas?resourceId=00000000-0000-4000-8000-100000000002&status=confirmed", "", "Bearer demo-admin", "")
	if err := json.Unmarshal(filtered.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0]["id"] != "00000000-0000-4000-8000-200000000005" {
		t.Fatal("administrative filters failed", filtered.Body)
	}
}
func TestAvailabilityFixtureAndCORS(t *testing.T) {
	handler := newHandler(t)
	response := call(handler, "GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000002/disponibilidades?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z", "", "Bearer demo-user", "")
	var data struct {
		SlotDurationSeconds int      `json:"slotDurationSeconds"`
		AvailableSlots      []object `json:"availableSlots"`
	}
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data.SlotDurationSeconds != 1800 || len(data.AvailableSlots) != 28 {
		t.Fatal("wrong resource fixture", response.Body)
	}
	request := httptest.NewRequest("OPTIONS", "/api/v1/reservas", nil)
	request.Header.Set("Origin", "http://127.0.0.1:5190")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != 204 || recorder.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:5190" {
		t.Fatal("Swagger preflight rejected")
	}
	request.Header.Set("Origin", "https://untrusted.example")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected origin allowed")
	}
}
