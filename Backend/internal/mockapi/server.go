// Package mockapi implements Hito 1 transport and immutable response fixtures.
// It deliberately does not implement production business rules, persistence, or identity.
package mockapi

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

//go:embed fixtures.json
var fixtureJSON []byte

type object = map[string]any
type fixtureSet struct {
	Examples          map[string]json.RawMessage `json:"examples"`
	Resources         []object                   `json:"resources"`
	Reservations      []object                   `json:"reservations"`
	AdminReservations []object                   `json:"adminReservations"`
	Details           map[string]json.RawMessage `json:"details"`
	Availability      map[string]json.RawMessage `json:"availability"`
}
type server struct {
	fixtures fixtureSet
	sequence atomic.Uint64
}
type operation struct {
	name              string
	admin             bool
	status            int
	example           string
	required, allowed string
}

// New returns an independent handler. No write changes fixtures or subsequent reads.
func New() http.Handler {
	s := &server{}
	if err := json.Unmarshal(fixtureJSON, &s.fixtures); err != nil {
		panic("invalid generated mock fixtures: " + err.Error())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		s.reply(w, 200, object{"status": "ok", "mode": "mock", "persistent": false})
	})
	routes := []struct {
		path string
		op   operation
	}{
		{"GET /api/v1/reservations", operation{name: "listReservations"}},
		{"GET /api/v1/admin/reservations", operation{name: "listAdminReservations", admin: true}},
		{"POST /api/v1/reservations", operation{name: "createReservation", status: 201, example: "ConfirmedReservation", required: "resourceId startAt endAt", allowed: "resourceId startAt endAt attributes"}},
		{"GET /api/v1/reservations/{id}", operation{name: "getReservation"}},
		{"PATCH /api/v1/reservations/{id}", operation{name: "updateReservation", status: 200, example: "UpdatedReservation", required: "expectedVersion", allowed: "expectedVersion resourceId startAt endAt attributes"}},
		{"POST /api/v1/reservations/{id}/cancel", operation{name: "cancelReservation", status: 200, example: "CancelledReservation", required: "expectedVersion", allowed: "expectedVersion"}},
		{"GET /api/v1/resources", operation{name: "listResources"}},
		{"POST /api/v1/resources", operation{name: "createResource", admin: true, status: 201, example: "ResourceDetail", required: "name enabled timeZone policy", allowed: "name enabled timeZone policy externalReference weeklySchedule attributes"}},
		{"GET /api/v1/resources/{id}", operation{name: "getResource", admin: true}},
		{"PATCH /api/v1/resources/{id}", operation{name: "updateResource", admin: true, status: 200, example: "UpdatedResourceDetail", required: "expectedVersion", allowed: "expectedVersion name enabled timeZone policy externalReference attributes"}},
		{"PUT /api/v1/resources/{id}/schedule", operation{name: "updateResourceSchedule", admin: true, status: 200, example: "UpdatedResourceDetail", required: "expectedVersion policy weeklySchedule", allowed: "expectedVersion policy weeklySchedule"}},
		{"POST /api/v1/resources/{id}/closures", operation{name: "createResourceClosure", admin: true, status: 201, example: "ResourceClosure", required: "startAt endAt reason", allowed: "startAt endAt reason"}},
		{"DELETE /api/v1/resources/{id}/closures/{closureId}", operation{name: "deleteResourceClosure", admin: true, status: 204}},
		{"GET /api/v1/resources/{id}/availability", operation{name: "getResourceAvailability"}},
	}
	for _, route := range routes {
		mux.HandleFunc(route.path, s.handle(route.op))
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.fail(w, 404, "RESOURCE_NOT_FOUND", "Ruta de demostración no encontrada.")
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock-Response", "true")
		w.Header().Set("Cache-Control", "no-store")
		origin := r.Header.Get("Origin")
		switch origin {
		case "http://localhost:8080", "http://127.0.0.1:8080", "http://localhost:5173", "http://127.0.0.1:5173", "http://127.0.0.1:5180":
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Mock-Scenario")
			w.Header().Set("Access-Control-Expose-Headers", "X-Mock-Response")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *server) reply(w http.ResponseWriter, status int, body any) {
	if status == 204 {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func (s *server) fail(w http.ResponseWriter, status int, code, message string) {
	s.reply(w, status, object{"error": object{"code": code, "message": message, "requestId": fmt.Sprintf("request-demo-%06d", s.sequence.Add(1))}})
}
func (s *server) example(w http.ResponseWriter, status int, name string) {
	s.reply(w, status, s.fixtures.Examples[name])
}
func decode(raw json.RawMessage) object {
	var result object
	_ = json.Unmarshal(raw, &result)
	return result
}
func instant(value any) (time.Time, error) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("not a date")
	}
	return time.Parse(time.RFC3339Nano, text)
}
func overlap(row object, from, to time.Time) bool {
	start, _ := instant(row["startAt"])
	end, _ := instant(row["endAt"])
	return start.Before(to) && end.After(from)
}
func includes(list, value string) bool {
	for _, item := range strings.Fields(list) {
		if item == value {
			return true
		}
	}
	return false
}

func (s *server) handle(op operation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scenario := r.Header.Get("X-Mock-Scenario")
		token := r.Header.Get("Authorization")
		if scenario == "unauthorized" || (token != "Bearer demo-user" && token != "Bearer demo-admin") {
			s.example(w, 401, "Unauthenticated")
			return
		}
		if scenario == "forbidden" || (op.admin && token != "Bearer demo-admin") {
			s.example(w, 403, "Forbidden")
			return
		}
		if scenario == "error" {
			s.example(w, 503, "ServiceUnavailable")
			return
		}
		if r.Method == "GET" && r.Body != nil && r.ContentLength > 0 {
			s.fail(w, 400, "INVALID_REQUEST", "Las consultas no reciben cuerpo.")
			return
		}
		if r.Method != "GET" && len(r.URL.Query()) != 0 {
			s.fail(w, 400, "INVALID_REQUEST", "Parámetro no admitido.")
			return
		}
		if op.allowed != "" {
			if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
				s.fail(w, 400, "INVALID_REQUEST", "Se requiere un cuerpo JSON.")
				return
			}
			var body object
			reader := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
			err := reader.Decode(&body)
			if err != nil || body == nil || reader.Decode(new(any)) != io.EOF || !validBody(body, op.required, op.allowed) {
				s.fail(w, 400, "INVALID_REQUEST", "La solicitud contiene campos ausentes, no admitidos o con formato incorrecto.")
				return
			}
			if (op.name == "updateReservation" || op.name == "updateResource") && len(body) < 2 {
				s.fail(w, 400, "INVALID_REQUEST", "Incluye al menos un campo modificable.")
				return
			}
		}
		if op.name == "createReservation" && strings.TrimSpace(r.Header.Get("Idempotency-Key")) == "" {
			s.fail(w, 400, "INVALID_REQUEST", "Falta Idempotency-Key.")
			return
		}
		id := r.PathValue("id")
		var reservation object
		if strings.Contains(op.name, "Reservation") && id != "" {
			for _, row := range s.fixtures.Reservations {
				if row["id"] == id {
					reservation = row
					break
				}
			}
			if reservation == nil || scenario == "notFound" {
				s.example(w, 404, "ReservationNotFound")
				return
			}
		}
		if op.name != "listResources" && strings.Contains(r.URL.Path, "/resources/") && (s.fixtures.Details[id] == nil || scenario == "notFound") {
			s.example(w, 404, "ResourceNotFound")
			return
		}
		if op.name == "deleteResourceClosure" && (r.PathValue("closureId") != "closure-001" || id != "resource-001") {
			s.example(w, 404, "ClosureNotFound")
			return
		}
		if r.Method != "GET" && s.writeScenario(w, op, scenario) {
			return
		}
		switch op.name {
		case "listReservations", "listAdminReservations", "listResources":
			s.list(w, r, op, scenario)
		case "getReservation":
			if scenario == "cancelled" {
				reservation = decode(s.fixtures.Examples["CancelledReservation"])
				reservation["id"] = id
			}
			if scenario == "expired" || scenario == "started" {
				source := "reservation-003"
				if scenario == "started" {
					source = "reservation-004"
				}
				for _, row := range s.fixtures.Reservations {
					if row["id"] == source {
						raw, _ := json.Marshal(row)
						reservation = decode(raw)
						reservation["id"] = id
					}
				}
			}
			s.reply(w, 200, reservation)
		case "getResource":
			s.reply(w, 200, s.fixtures.Details[id])
		case "getResourceAvailability":
			s.availability(w, r, scenario)
		default:
			if op.name == "createReservation" {
				log.Print("[MOCK] Confirmación de reserva: notificación simulada; no se envió correo ni se llamó a G10.")
			}
			if op.status == 204 {
				s.reply(w, 204, nil)
			} else {
				s.example(w, op.status, op.example)
			}
		}
	}
}

func (s *server) writeScenario(w http.ResponseWriter, op operation, scenario string) bool {
	if scenario == "" || scenario == "normal" || scenario == "success" {
		return false
	}
	status, name := 400, "InvalidRequest"
	switch scenario {
	case "conflict":
		if op.admin && op.name != "createResource" && op.name != "deleteResourceClosure" {
			status, name = 409, "ResourceScheduleConflict"
		} else if !op.admin && op.name != "cancelReservation" {
			status, name = 409, "ResourceUnavailable"
		}
	case "version":
		if op.name == "createReservation" {
			status, name = 409, "IdempotencyConflict"
		} else if op.name == "updateReservation" || op.name == "cancelReservation" {
			status, name = 409, "VersionConflict"
		} else if op.name == "updateResource" || op.name == "updateResourceSchedule" {
			status, name = 409, "ResourceVersionConflict"
		}
	case "deadline":
		if !op.admin && op.name != "createReservation" {
			status, name = 422, "CancellationWindowClosed"
		}
	case "policy":
		if op.name != "deleteResourceClosure" && op.name != "cancelReservation" {
			status, name = 422, "PolicyViolation"
		}
	}
	s.example(w, status, name)
	return true
}

func queryAllowed(q url.Values, allowed string) bool {
	for key, values := range q {
		if !includes(allowed, key) || len(values) != 1 {
			return false
		}
	}
	return true
}
func rangeQuery(q url.Values, required bool) (time.Time, time.Time, error) {
	if !required && q.Get("from") == "" && q.Get("to") == "" {
		return time.Time{}, time.Time{}, nil
	}
	from, err := time.Parse(time.RFC3339Nano, q.Get("from"))
	if err != nil {
		return from, from, err
	}
	to, err := time.Parse(time.RFC3339Nano, q.Get("to"))
	if err != nil || !from.Before(to) {
		return from, to, fmt.Errorf("invalid range")
	}
	return from, to, nil
}
func (s *server) list(w http.ResponseWriter, r *http.Request, op operation, scenario string) {
	q := r.URL.Query()
	allowed := "enabled limit cursor"
	rows := s.fixtures.Resources
	if op.name == "listReservations" || op.name == "listAdminReservations" {
		allowed = "resourceId status from to limit cursor"
		rows = s.fixtures.Reservations
		if op.name == "listAdminReservations" {
			rows = s.fixtures.AdminReservations
		}
	}
	if !queryAllowed(q, allowed) {
		s.fail(w, 400, "INVALID_REQUEST", "Filtro no admitido.")
		return
	}
	from, to, err := rangeQuery(q, false)
	if err != nil {
		s.fail(w, 400, "INVALID_REQUEST", "Completa ambos extremos de un rango válido.")
		return
	}
	enabled := true
	if value, ok := q["enabled"]; ok {
		if value[0] != "true" && value[0] != "false" {
			s.fail(w, 400, "INVALID_REQUEST", "enabled debe ser true o false.")
			return
		}
		enabled = value[0] == "true"
	}
	if status := q.Get("status"); status != "" && status != "confirmed" && status != "cancelled" {
		s.fail(w, 400, "INVALID_REQUEST", "Estado inválido.")
		return
	}
	limit := 20
	if value, ok := q["limit"]; ok {
		limit, err = strconv.Atoi(value[0])
		if err != nil || limit < 1 || limit > 100 {
			s.fail(w, 400, "INVALID_REQUEST", "limit debe estar entre 1 y 100.")
			return
		}
	}
	if scenario == "paginated" {
		limit = 2
	}
	filtered := []object{}
	for _, row := range rows {
		if op.name == "listResources" {
			if row["enabled"] != enabled {
				continue
			}
		} else {
			if q.Get("resourceId") != "" && row["resourceId"] != q.Get("resourceId") {
				continue
			}
			if q.Get("status") != "" && row["status"] != q.Get("status") {
				continue
			}
			if !from.IsZero() && !overlap(row, from, to) {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	if op.name == "listResources" {
		sort.Slice(filtered, func(i, j int) bool { return filtered[i]["id"].(string) < filtered[j]["id"].(string) })
	}
	if scenario == "empty" {
		filtered = []object{}
	}
	binding := r.URL.Path + "|" + r.Header.Get("Authorization") + "|" + scenario + "|"
	q.Del("cursor")
	binding += q.Encode()
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(binding)))
	offset := 0
	if value := r.URL.Query().Get("cursor"); value != "" {
		raw, e := base64.RawURLEncoding.DecodeString(value)
		var c struct {
			Offset int
			Query  string
		}
		if e != nil || json.Unmarshal(raw, &c) != nil || c.Query != hash || c.Offset < 0 || c.Offset > len(filtered) {
			s.fail(w, 400, "INVALID_REQUEST", "Cursor inválido para esta consulta.")
			return
		}
		offset = c.Offset
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	var next any
	if end < len(filtered) {
		raw, _ := json.Marshal(struct {
			Offset int
			Query  string
		}{end, hash})
		next = base64.RawURLEncoding.EncodeToString(raw)
	}
	s.reply(w, 200, object{"items": filtered[offset:end], "nextCursor": next})
}

func (s *server) availability(w http.ResponseWriter, r *http.Request, scenario string) {
	q := r.URL.Query()
	from, to, err := rangeQuery(q, true)
	if !queryAllowed(q, "from to") || err != nil || to.Sub(from) > 31*24*time.Hour {
		s.fail(w, 400, "INVALID_REQUEST", "Consulta un rango válido de hasta 31 días.")
		return
	}
	result := decode(s.fixtures.Availability[r.PathValue("id")])
	result["from"] = from.UTC().Format(time.RFC3339Nano)
	result["to"] = to.UTC().Format(time.RFC3339Nano)
	for _, key := range []string{"availableWindows", "availableSlots"} {
		selected := []object{}
		for _, value := range result[key].([]any) {
			row := value.(map[string]any)
			start, _ := instant(row["startAt"])
			end, _ := instant(row["endAt"])
			if key == "availableSlots" {
				if start.Before(from) || end.After(to) {
					continue
				}
			} else {
				if !overlap(row, from, to) {
					continue
				}
				if start.Before(from) {
					row["startAt"] = from.UTC().Format(time.RFC3339Nano)
				}
				if end.After(to) {
					row["endAt"] = to.UTC().Format(time.RFC3339Nano)
				}
			}
			if scenario != "empty" && scenario != "disabled" {
				selected = append(selected, row)
			}
		}
		result[key] = selected
	}
	if scenario == "disabled" {
		result["enabled"] = false
	}
	s.reply(w, 200, result)
}

// validBody checks only the public JSON shape, never availability or business transitions.
func validBody(body object, required, allowed string) bool {
	for _, key := range strings.Fields(required) {
		if _, ok := body[key]; !ok {
			return false
		}
	}
	for key, value := range body {
		if value == nil || !includes(allowed, key) {
			return false
		}
		switch key {
		case "expectedVersion":
			n, ok := value.(float64)
			if !ok || n < 1 || n != float64(int64(n)) {
				return false
			}
		case "enabled":
			if _, ok := value.(bool); !ok {
				return false
			}
		case "startAt", "endAt":
			if _, err := instant(value); err != nil {
				return false
			}
		case "attributes":
			obj, ok := value.(map[string]any)
			if !ok || len(obj) != 0 {
				return false
			}
		case "policy":
			p, ok := value.(map[string]any)
			if !ok || len(p) != 4 {
				return false
			}
			for _, k := range []string{"cancellationNoticeSeconds", "minDurationSeconds", "maxDurationSeconds", "slotDurationSeconds"} {
				n, ok := p[k].(float64)
				if !ok || n != float64(int64(n)) || n < 0 || (k != "cancellationNoticeSeconds" && n < 60) || (k == "slotDurationSeconds" && int64(n)%60 != 0) {
					return false
				}
			}
		case "externalReference":
			p, ok := value.(map[string]any)
			if !ok || !validBody(p, "system id", "system id") {
				return false
			}
		case "weeklySchedule":
			rules, ok := value.([]any)
			if !ok {
				return false
			}
			for _, rule := range rules {
				p, ok := rule.(map[string]any)
				if !ok || len(p) != 3 {
					return false
				}
				n, ok := p["dayOfWeek"].(float64)
				if !ok || n < 1 || n > 7 || n != float64(int(n)) {
					return false
				}
				for _, k := range []string{"startLocalTime", "endLocalTime"} {
					v, ok := p[k].(string)
					if !ok {
						return false
					}
					parsed, err := time.Parse("15:04", v)
					if err != nil || parsed.Format("15:04") != v {
						return false
					}
				}
			}
		default:
			text, ok := value.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return false
			}
		}
	}
	return true
}
