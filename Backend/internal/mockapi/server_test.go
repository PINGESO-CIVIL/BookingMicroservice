package mockapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
func TestAllThirteenOperations(t *testing.T) {
	var f fixtureSet
	if err := json.Unmarshal(fixtureJSON, &f); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		method, path, example string
		status                int
	}{
		{"GET", "/api/v1/reservations", "", 200},
		{"POST", "/api/v1/reservations", "CreateRequest", 201},
		{"GET", "/api/v1/reservations/reservation-001", "", 200},
		{"PATCH", "/api/v1/reservations/reservation-001", "UpdateRequest", 200},
		{"POST", "/api/v1/reservations/reservation-001/cancel", "", 200},
		{"GET", "/api/v1/resources", "", 200},
		{"POST", "/api/v1/resources", "CreateResourceRequest", 201},
		{"GET", "/api/v1/resources/resource-001", "", 200},
		{"PATCH", "/api/v1/resources/resource-001", "UpdateResourceRequest", 200},
		{"PUT", "/api/v1/resources/resource-001/schedule", "UpdateResourceScheduleRequest", 200},
		{"POST", "/api/v1/resources/resource-001/closures", "CreateResourceClosureRequest", 201},
		{"DELETE", "/api/v1/resources/resource-001/closures/closure-001", "", 204},
		{"GET", "/api/v1/resources/resource-001/availability?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z", "", 200},
	}
	handler := New()
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			body := string(f.Examples[tc.example])
			if strings.HasSuffix(tc.path, "/cancel") {
				body = `{"expectedVersion":1}`
			}
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
func TestTransportErrorsAndScenarios(t *testing.T) {
	handler := New()
	cases := []struct {
		method, path, body, token, scenario, code string
		status                                    int
	}{
		{"GET", "/api/v1/reservations", "", "", "", "UNAUTHENTICATED", 401},
		{"GET", "/api/v1/resources/resource-001", "", "Bearer demo-user", "", "FORBIDDEN", 403},
		{"GET", "/api/v1/reservations/no-existe", "", "Bearer demo-user", "", "RESERVATION_NOT_FOUND", 404},
		{"GET", "/api/v1/resources/ajeno", "", "Bearer demo-admin", "", "RESOURCE_NOT_FOUND", 404},
		{"POST", "/api/v1/reservations", `{"resourceId":"resource-001","startAt":"2026-11-10T22:00:00Z","endAt":"2026-11-10T23:00:00Z","responsible":{}}`, "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"PATCH", "/api/v1/reservations/reservation-001", `{"expectedVersion":1}`, "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"POST", "/api/v1/reservations/reservation-001/cancel", `{"expectedVersion":0}`, "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"POST", "/api/v1/reservations/reservation-001/cancel", `{"expectedVersion":1}`, "Bearer demo-user", "version", "RESERVATION_VERSION_CONFLICT", 409},
		{"POST", "/api/v1/reservations/reservation-001/cancel", `{"expectedVersion":1}`, "Bearer demo-user", "deadline", "CANCELLATION_WINDOW_CLOSED", 422},
		{"PATCH", "/api/v1/resources/resource-001", `{"expectedVersion":1,"enabled":false}`, "Bearer demo-admin", "conflict", "RESOURCE_SCHEDULE_CONFLICT", 409},
		{"GET", "/api/v1/reservations", "", "Bearer demo-user", "error", "SERVICE_UNAVAILABLE", 503},
		{"GET", "/api/v1/reservations?limit=101", "", "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"GET", "/api/v1/reservations?from=2026-11-10T00:00:00Z", "", "Bearer demo-user", "", "INVALID_REQUEST", 400},
		{"GET", "/api/v1/resources/resource-001/availability?from=2026-11-10T00:00:00Z&to=2027-01-01T00:00:00Z", "", "Bearer demo-user", "", "INVALID_REQUEST", 400},
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
	handler := New()
	before := call(handler, "GET", "/api/v1/reservations", "", "Bearer demo-user", "").Body.String()
	call(handler, "POST", "/api/v1/reservations/reservation-001/cancel", `{"expectedVersion":1}`, "Bearer demo-user", "")
	after := call(handler, "GET", "/api/v1/reservations", "", "Bearer demo-user", "").Body.String()
	if before != after {
		t.Fatal("mock write modified later reads")
	}
	before = call(handler, "GET", "/api/v1/resources/resource-001", "", "Bearer demo-admin", "").Body.String()
	call(handler, "PATCH", "/api/v1/resources/resource-001", `{"expectedVersion":1,"name":"Changed"}`, "Bearer demo-admin", "")
	call(handler, "DELETE", "/api/v1/resources/resource-001/closures/closure-001", "", "Bearer demo-admin", "")
	after = call(handler, "GET", "/api/v1/resources/resource-001", "", "Bearer demo-admin", "").Body.String()
	if before != after {
		t.Fatal("resource mock write modified later reads")
	}
}
func TestPaginationAndCursorBinding(t *testing.T) {
	handler := New()
	var first struct {
		Items      []object `json:"items"`
		NextCursor string   `json:"nextCursor"`
	}
	response := call(handler, "GET", "/api/v1/reservations?limit=1", "", "Bearer demo-user", "")
	if err := json.Unmarshal(response.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 1 || first.NextCursor == "" {
		t.Fatal("missing continuation")
	}
	second := call(handler, "GET", "/api/v1/reservations?limit=1&cursor="+first.NextCursor, "", "Bearer demo-user", "")
	var next struct {
		Items []object `json:"items"`
	}
	_ = json.Unmarshal(second.Body.Bytes(), &next)
	if len(next.Items) != 1 || first.Items[0]["id"] == next.Items[0]["id"] {
		t.Fatal("invalid continuation")
	}
	invalid := call(handler, "GET", "/api/v1/reservations?status=confirmed&limit=1&cursor="+first.NextCursor, "", "Bearer demo-user", "")
	if invalid.Code != 400 {
		t.Fatal("cursor accepted with different filters")
	}
}

func TestAdminReadsAllUsersWithoutChangingPersonalAccess(t *testing.T) {
	handler := New()
	response := call(handler, "GET", "/api/v1/admin/reservations", "", "Bearer demo-admin", "")
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
	denied := call(handler, "GET", "/api/v1/admin/reservations", "", "Bearer demo-user", "")
	if denied.Code != 403 {
		t.Fatal("ordinary user accessed administrative list")
	}
	hidden := call(handler, "GET", "/api/v1/reservations/reservation-005", "", "Bearer demo-user", "")
	if hidden.Code != 404 {
		t.Fatal("personal endpoint exposed another user's reservation")
	}
	personal := call(handler, "GET", "/api/v1/reservations", "", "Bearer demo-user", "")
	if err := json.Unmarshal(personal.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 4 {
		t.Fatal("personal list changed scope")
	}
	filtered := call(handler, "GET", "/api/v1/admin/reservations?resourceId=resource-002&status=confirmed", "", "Bearer demo-admin", "")
	if err := json.Unmarshal(filtered.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0]["id"] != "reservation-005" {
		t.Fatal("administrative filters failed", filtered.Body)
	}
}
func TestAvailabilityFixtureAndCORS(t *testing.T) {
	handler := New()
	response := call(handler, "GET", "/api/v1/resources/resource-002/availability?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z", "", "Bearer demo-user", "")
	var data struct {
		SlotDurationSeconds int      `json:"slotDurationSeconds"`
		AvailableSlots      []object `json:"availableSlots"`
	}
	_ = json.Unmarshal(response.Body.Bytes(), &data)
	if data.SlotDurationSeconds != 1800 || len(data.AvailableSlots) != 28 {
		t.Fatal("wrong resource fixture", response.Body)
	}
	request := httptest.NewRequest("OPTIONS", "/api/v1/reservations", nil)
	request.Header.Set("Origin", "http://localhost:8080")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != 204 || recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8080" {
		t.Fatal("Swagger preflight rejected")
	}
	request.Header.Set("Origin", "https://untrusted.example")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected origin allowed")
	}
}
