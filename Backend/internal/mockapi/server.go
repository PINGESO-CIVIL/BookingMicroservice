// Package mockapi implements Hito 1 transport and immutable response fixtures.
// It deliberately does not implement production business rules, persistence, or identity.
package mockapi

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"bookingms/internal/api"
)

//go:embed fixtures.json
var fixtureJSON []byte

type object = map[string]any
type fixtureSet struct {
	Examples          map[string]json.RawMessage    `json:"examples"`
	Resources         []api.Resource                `json:"resources"`
	Reservations      []api.Reservation             `json:"reservations"`
	AdminReservations []api.Reservation             `json:"adminReservations"`
	Details           map[string]api.ResourceDetail `json:"details"`
	Availability      map[string]api.Availability   `json:"availability"`
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
	request           func() any
	required, allowed string
	minFields         int
}

// New returns an independent handler. No write changes fixtures or subsequent reads.
func New() http.Handler {
	s := &server{}
	if err := json.Unmarshal(fixtureJSON, &s.fixtures); err != nil {
		panic("invalid generated mock fixtures: " + err.Error())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		s.reply(w, 200, api.HealthResponse{Status: "ok", Mode: "mock", Persistent: false})
	})
	routes := []struct {
		path string
		op   operation
	}{
		{"GET /api/v1/reservas", operation{name: "listReservations"}},
		{"GET /api/v1/admin/reservas", operation{name: "listAdminReservations", admin: true}},
		{"POST /api/v1/reservas", operation{name: "createReservation", status: 201, example: "ConfirmedReservation", request: func() any { return new(api.CreateReservationRequest) }, required: "resourceId startAt endAt", allowed: "resourceId startAt endAt attributes"}},
		{"GET /api/v1/reservas/{id}", operation{name: "getReservation"}},
		{"PATCH /api/v1/reservas/{id}", operation{name: "updateReservation", status: 200, example: "UpdatedReservation", request: func() any { return new(api.UpdateReservationRequest) }, minFields: 2, required: "expectedVersion", allowed: "expectedVersion resourceId startAt endAt attributes"}},
		{"POST /api/v1/reservas/{id}/cancelaciones", operation{name: "cancelReservation", status: 200, example: "CancelledReservation", request: func() any { return new(api.CancelReservationRequest) }, required: "expectedVersion", allowed: "expectedVersion"}},
		{"GET /api/v1/recursos", operation{name: "listResources"}},
		{"POST /api/v1/recursos", operation{name: "createResource", admin: true, status: 201, example: "ResourceDetail", request: func() any { return new(api.CreateResourceRequest) }, required: "name enabled timeZone policy", allowed: "name enabled timeZone policy externalReference weeklySchedule attributes"}},
		{"GET /api/v1/recursos/{id}", operation{name: "getResource", admin: true}},
		{"PATCH /api/v1/recursos/{id}", operation{name: "updateResource", admin: true, status: 200, example: "UpdatedResourceDetail", request: func() any { return new(api.UpdateResourceRequest) }, minFields: 2, required: "expectedVersion", allowed: "expectedVersion name enabled timeZone policy externalReference attributes"}},
		{"PUT /api/v1/recursos/{id}/horarios", operation{name: "updateResourceSchedule", admin: true, status: 200, example: "UpdatedResourceDetail", request: func() any { return new(api.UpdateResourceScheduleRequest) }, required: "expectedVersion policy weeklySchedule", allowed: "expectedVersion policy weeklySchedule"}},
		{"POST /api/v1/recursos/{id}/bloqueos", operation{name: "createResourceClosure", admin: true, status: 201, example: "ResourceClosure", request: func() any { return new(api.CreateResourceClosureRequest) }, required: "startAt endAt reason", allowed: "startAt endAt reason"}},
		{"DELETE /api/v1/recursos/{id}/bloqueos/{closureId}", operation{name: "deleteResourceClosure", admin: true, status: 204}},
		{"GET /api/v1/recursos/{id}/disponibilidades", operation{name: "getResourceAvailability"}},
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
	s.reply(w, status, api.ErrorResponse{Error: api.Error{
		Code: code, Message: message, RequestID: fmt.Sprintf("request-demo-%06d", s.sequence.Add(1)),
	}})
}

func exampleAs[T any](fixtures fixtureSet, name string) T {
	var result T
	if err := json.Unmarshal(fixtures.Examples[name], &result); err != nil {
		panic(fmt.Sprintf("invalid generated mock example %s: %v", name, err))
	}
	return result
}

func (s *server) example(w http.ResponseWriter, status int, name string) {
	switch name {
	case "ConfirmedReservation", "UpdatedReservation", "CancelledReservation":
		s.reply(w, status, exampleAs[api.Reservation](s.fixtures, name))
	case "ResourceDetail", "UpdatedResourceDetail":
		s.reply(w, status, exampleAs[api.ResourceDetail](s.fixtures, name))
	case "ResourceClosure":
		s.reply(w, status, exampleAs[api.ResourceClosure](s.fixtures, name))
	default:
		s.reply(w, status, exampleAs[api.ErrorResponse](s.fixtures, name))
	}
}

func intervalTimes(interval api.Interval) (time.Time, time.Time) {
	start, _ := time.Parse(time.RFC3339Nano, interval.StartAt)
	end, _ := time.Parse(time.RFC3339Nano, interval.EndAt)
	return start, end
}

func overlap(interval api.Interval, from, to time.Time) bool {
	start, end := intervalTimes(interval)
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
		if scenario != "" && !includes("normal success empty paginated cancelled expired started notFound disabled conflict version deadline policy error unauthorized forbidden", scenario) {
			s.fail(w, 400, "INVALID_REQUEST", "Escenario de demostración no admitido.")
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
		_, queryErr := url.ParseQuery(r.URL.RawQuery)
		allowsQuery := op.name == "listReservations" || op.name == "listAdminReservations" ||
			op.name == "listResources" || op.name == "getResourceAvailability"
		if queryErr != nil || (!allowsQuery && r.URL.RawQuery != "") {
			s.fail(w, 400, "INVALID_REQUEST", "Parámetro no admitido.")
			return
		}
		if op.request != nil {
			mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || mediaType != "application/json" {
				s.fail(w, 400, "INVALID_REQUEST", "Se requiere un cuerpo JSON.")
				return
			}
			body := http.MaxBytesReader(w, r.Body, 1<<20)
			if err := api.DecodeRequest(body, op.request(), op.required, op.allowed, op.minFields); err != nil {
				s.fail(w, 400, "INVALID_REQUEST", "La solicitud contiene campos ausentes, no admitidos o con formato incorrecto.")
				return
			}
		}
		if op.name == "createReservation" && !api.ValidText(r.Header.Get("Idempotency-Key"), 128) {
			s.fail(w, 400, "INVALID_REQUEST", "Idempotency-Key debe tener entre 1 y 128 caracteres.")
			return
		}
		id := r.PathValue("id")
		closureID := r.PathValue("closureId")
		if (id != "" && !api.ValidUUID(id)) || (closureID != "" && !api.ValidUUID(closureID)) {
			s.fail(w, 400, "INVALID_REQUEST", "El identificador debe ser un UUID.")
			return
		}
		id = strings.ToLower(id)
		var reservation *api.Reservation
		if strings.Contains(op.name, "Reservation") && id != "" {
			for _, row := range s.fixtures.Reservations {
				if row.ID == id {
					reservation = &row
					break
				}
			}
			if reservation == nil || scenario == "notFound" {
				s.example(w, 404, "ReservationNotFound")
				return
			}
		}
		if _, exists := s.fixtures.Details[id]; id != "" && strings.Contains(r.URL.Path, "/recursos/") && (!exists || scenario == "notFound") {
			s.example(w, 404, "ResourceNotFound")
			return
		}
		if op.name == "deleteResourceClosure" && (strings.ToLower(closureID) != "00000000-0000-4000-8000-300000000001" || id != "00000000-0000-4000-8000-100000000001") {
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
				row := exampleAs[api.Reservation](s.fixtures, "CancelledReservation")
				row.ID = id
				reservation = &row
			}
			if scenario == "expired" || scenario == "started" {
				source := "00000000-0000-4000-8000-200000000003"
				if scenario == "started" {
					source = "00000000-0000-4000-8000-200000000004"
				}
				for _, row := range s.fixtures.Reservations {
					if row.ID == source {
						row.ID = id
						reservation = &row
					}
				}
			}
			s.reply(w, 200, reservation)
		case "getResource":
			s.reply(w, 200, s.fixtures.Details[id])
		case "getResourceAvailability":
			s.availability(w, r, id, scenario)
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
	if !required && !q.Has("from") && !q.Has("to") {
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
	isResourceList := op.name == "listResources"
	if !isResourceList {
		allowed = "resourceId status from to limit cursor"
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
	if status := q.Get("status"); q.Has("status") && status != "confirmed" && status != "cancelled" {
		s.fail(w, 400, "INVALID_REQUEST", "Estado inválido.")
		return
	}
	resourceID := strings.ToLower(q.Get("resourceId"))
	if q.Has("resourceId") && !api.ValidUUID(resourceID) {
		s.fail(w, 400, "INVALID_REQUEST", "resourceId debe ser un UUID.")
		return
	}
	if q.Has("cursor") && !api.ValidText(q.Get("cursor"), 2048) {
		s.fail(w, 400, "INVALID_REQUEST", "Cursor inválido.")
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
	if isResourceList {
		filtered := []api.Resource{}
		if scenario != "empty" {
			for _, row := range s.fixtures.Resources {
				if row.Enabled == enabled {
					filtered = append(filtered, row)
				}
			}
		}
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
		replyPage(s, w, r, q, scenario, limit, filtered)
		return
	}
	rows := s.fixtures.Reservations
	if op.name == "listAdminReservations" {
		rows = s.fixtures.AdminReservations
	}
	filtered := []api.Reservation{}
	status := q.Get("status")
	if scenario != "empty" {
		for _, row := range rows {
			if (resourceID == "" || row.ResourceID == resourceID) &&
				(status == "" || row.Status == status) &&
				(from.IsZero() || overlap(row.Interval, from, to)) {
				filtered = append(filtered, row)
			}
		}
	}
	replyPage(s, w, r, q, scenario, limit, filtered)
}

func replyPage[T api.Resource | api.Reservation](s *server, w http.ResponseWriter, r *http.Request, q url.Values, scenario string, limit int, filtered []T) {
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
	var next *string
	if end < len(filtered) {
		raw, _ := json.Marshal(struct {
			Offset int
			Query  string
		}{end, hash})
		cursor := base64.RawURLEncoding.EncodeToString(raw)
		next = &cursor
	}
	s.reply(w, 200, api.Page[T]{Items: filtered[offset:end], NextCursor: next})
}

func (s *server) availability(w http.ResponseWriter, r *http.Request, id, scenario string) {
	q := r.URL.Query()
	from, to, err := rangeQuery(q, true)
	if !queryAllowed(q, "from to") || err != nil || to.Sub(from) > 31*24*time.Hour {
		s.fail(w, 400, "INVALID_REQUEST", "Consulta un rango válido de hasta 31 días.")
		return
	}
	result := s.fixtures.Availability[id]
	result.From = from.UTC().Format(time.RFC3339Nano)
	result.To = to.UTC().Format(time.RFC3339Nano)
	result.AvailableWindows, result.AvailableSlots = []api.Interval{}, []api.Interval{}
	if scenario != "empty" && scenario != "disabled" {
		fixture := s.fixtures.Availability[id]
		for _, window := range fixture.AvailableWindows {
			start, end := intervalTimes(window)
			if !start.Before(to) || !end.After(from) {
				continue
			}
			if start.Before(from) {
				window.StartAt = result.From
			}
			if end.After(to) {
				window.EndAt = result.To
			}
			result.AvailableWindows = append(result.AvailableWindows, window)
		}
		for _, slot := range fixture.AvailableSlots {
			start, end := intervalTimes(slot)
			if !start.Before(from) && !end.After(to) {
				result.AvailableSlots = append(result.AvailableSlots, slot)
			}
		}
	}
	if scenario == "disabled" {
		result.Enabled = false
	}
	s.reply(w, 200, result)
}
