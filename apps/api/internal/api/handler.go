package api

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"bookingms/internal/service"
)

type handler struct {
	service  *service.Service
	sequence atomic.Uint64
}

type operation struct {
	name              service.Operation
	admin             bool
	status            int
	request           func() any
	required, allowed string
	minFields         int
}

// New connects the HTTP routes to the application service.
func New(svc *service.Service) http.Handler {
	s := &handler{service: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		s.reply(w, 200, HealthResponse{Status: "ok", Mode: "mock", Persistent: false})
	})
	routes := []struct {
		path string
		op   operation
	}{
		{"GET /api/v1/reservas", operation{name: "listReservations"}},
		{"GET /api/v1/admin/reservas", operation{name: "listAdminReservations", admin: true}},
		{"POST /api/v1/reservas", operation{name: "createReservation", status: 201, request: func() any { return new(CreateReservationRequest) }, required: "resourceId startAt endAt", allowed: "resourceId startAt endAt attributes"}},
		{"GET /api/v1/reservas/{id}", operation{name: "getReservation"}},
		{"PATCH /api/v1/reservas/{id}", operation{name: "updateReservation", status: 200, request: func() any { return new(UpdateReservationRequest) }, minFields: 2, required: "expectedVersion", allowed: "expectedVersion resourceId startAt endAt attributes"}},
		{"POST /api/v1/reservas/{id}/cancelaciones", operation{name: "cancelReservation", status: 200, request: func() any { return new(CancelReservationRequest) }, required: "expectedVersion", allowed: "expectedVersion"}},
		{"GET /api/v1/recursos", operation{name: "listResources"}},
		{"POST /api/v1/recursos", operation{name: "createResource", admin: true, status: 201, request: func() any { return new(CreateResourceRequest) }, required: "name enabled timeZone policy", allowed: "name enabled timeZone policy externalReference weeklySchedule attributes"}},
		{"GET /api/v1/recursos/{id}", operation{name: "getResource", admin: true}},
		{"PATCH /api/v1/recursos/{id}", operation{name: "updateResource", admin: true, status: 200, request: func() any { return new(UpdateResourceRequest) }, minFields: 2, required: "expectedVersion", allowed: "expectedVersion name enabled timeZone policy externalReference attributes"}},
		{"PUT /api/v1/recursos/{id}/horarios", operation{name: "updateResourceSchedule", admin: true, status: 200, request: func() any { return new(UpdateResourceScheduleRequest) }, required: "expectedVersion policy weeklySchedule", allowed: "expectedVersion policy weeklySchedule"}},
		{"POST /api/v1/recursos/{id}/bloqueos", operation{name: "createResourceClosure", admin: true, status: 201, request: func() any { return new(CreateResourceClosureRequest) }, required: "startAt endAt reason", allowed: "startAt endAt reason"}},
		{"DELETE /api/v1/recursos/{id}/bloqueos/{closureId}", operation{name: "deleteResourceClosure", admin: true, status: 204}},
		{"GET /api/v1/recursos/{id}/disponibilidades", operation{name: "getResourceAvailability"}},
	}
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ready := svc.Ready(r.Context())
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		s.reply(w, status, struct {
			Ready bool `json:"ready"`
		}{Ready: ready})
	})
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
		case "http://localhost:5190", "http://127.0.0.1:5190", "http://localhost:5173", "http://127.0.0.1:5173", "http://127.0.0.1:5180":
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

func (s *handler) reply(w http.ResponseWriter, status int, body any) {
	if status == 204 {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func (s *handler) fail(w http.ResponseWriter, status int, code, message string) {
	s.reply(w, status, ErrorResponse{Error: Error{
		Code: code, Message: message, RequestID: fmt.Sprintf("request-demo-%06d", s.sequence.Add(1)),
	}})
}

func (s *handler) writeFault(w http.ResponseWriter, fault *service.Fault) {
	if fault.Message != "" {
		s.fail(w, http.StatusBadRequest, "INVALID_REQUEST", fault.Message)
		return
	}
	status := http.StatusBadRequest
	switch fault.Example {
	case "Unauthenticated":
		status = http.StatusUnauthorized
	case "Forbidden":
		status = http.StatusForbidden
	case "ServiceUnavailable":
		status = http.StatusServiceUnavailable
	case "ReservationNotFound", "ResourceNotFound", "ClosureNotFound":
		status = http.StatusNotFound
	case "ResourceScheduleConflict", "ResourceUnavailable", "IdempotencyConflict", "VersionConflict", "ResourceVersionConflict":
		status = http.StatusConflict
	case "CancellationWindowClosed", "PolicyViolation":
		status = http.StatusUnprocessableEntity
	}
	s.reply(w, status, s.service.ErrorExample(fault.Example))
}

func (s *handler) handle(op operation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scenario := r.Header.Get("X-Mock-Scenario")
		actor := service.Anonymous
		switch r.Header.Get("Authorization") {
		case "Bearer demo-user":
			actor = service.DemoUser
		case "Bearer demo-admin":
			actor = service.DemoAdmin
		}
		if fault := s.service.Authorize(op.admin, actor, scenario); fault != nil {
			s.writeFault(w, fault)
			return
		}
		if r.Method == "GET" && r.Body != nil && r.ContentLength > 0 {
			s.fail(w, 400, "INVALID_REQUEST", "Las consultas no reciben cuerpo.")
			return
		}
		query, queryErr := url.ParseQuery(r.URL.RawQuery)
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
			if err := DecodeRequest(body, op.request(), op.required, op.allowed, op.minFields); err != nil {
				s.fail(w, 400, "INVALID_REQUEST", "La solicitud contiene campos ausentes, no admitidos o con formato incorrecto.")
				return
			}
		}
		if op.name == "createReservation" && !ValidText(r.Header.Get("Idempotency-Key"), 128) {
			s.fail(w, 400, "INVALID_REQUEST", "Idempotency-Key debe tener entre 1 y 128 caracteres.")
			return
		}
		id := r.PathValue("id")
		closureID := r.PathValue("closureId")
		if (id != "" && !ValidUUID(id)) || (closureID != "" && !ValidUUID(closureID)) {
			s.fail(w, 400, "INVALID_REQUEST", "El identificador debe ser un UUID.")
			return
		}
		id = strings.ToLower(id)
		input := service.Input{ID: id, ClosureID: strings.ToLower(closureID), Scenario: scenario}
		switch op.name {
		case "listReservations", "listAdminReservations", "listResources":
			var message string
			input.Query, message = listQuery(query, op.name)
			if message != "" {
				s.fail(w, http.StatusBadRequest, "INVALID_REQUEST", message)
				return
			}
			// Keep the existing cursor representation and scope across the refactor.
			query.Del("cursor")
			input.Query.CursorScope = r.URL.Path + "|" + r.Header.Get("Authorization") + "|" + scenario + "|" + query.Encode()
		case "getResourceAvailability":
			if queryAllowed(query, "from to") {
				from, to, err := rangeQuery(query, true)
				if err == nil {
					input.Query.From, input.Query.To = from, to
				}
			}
		}
		result, fault := s.service.Execute(op.name, input)
		if fault != nil {
			s.writeFault(w, fault)
			return
		}
		status := op.status
		if status == 0 {
			status = http.StatusOK
		}
		s.reply(w, status, result)
	}
}

func includes(list, value string) bool {
	for _, item := range strings.Fields(list) {
		if item == value {
			return true
		}
	}
	return false
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
func listQuery(q url.Values, operation service.Operation) (service.Query, string) {
	allowed := "enabled limit cursor"
	isResourceList := operation == "listResources"
	if !isResourceList {
		allowed = "resourceId status from to limit cursor"
	}
	if !queryAllowed(q, allowed) {
		return service.Query{}, "Filtro no admitido."
	}
	from, to, err := rangeQuery(q, false)
	if err != nil {
		return service.Query{}, "Completa ambos extremos de un rango válido."
	}
	enabled := true
	if value, ok := q["enabled"]; ok {
		if value[0] != "true" && value[0] != "false" {
			return service.Query{}, "enabled debe ser true o false."
		}
		enabled = value[0] == "true"
	}
	if status := q.Get("status"); q.Has("status") && status != "confirmed" && status != "cancelled" {
		return service.Query{}, "Estado inválido."
	}
	resourceID := strings.ToLower(q.Get("resourceId"))
	if q.Has("resourceId") && !ValidUUID(resourceID) {
		return service.Query{}, "resourceId debe ser un UUID."
	}
	if q.Has("cursor") && !ValidText(q.Get("cursor"), 2048) {
		return service.Query{}, "Cursor inválido."
	}
	limit := 20
	if value, ok := q["limit"]; ok {
		limit, err = strconv.Atoi(value[0])
		if err != nil || limit < 1 || limit > 100 {
			return service.Query{}, "limit debe estar entre 1 y 100."
		}
	}
	return service.Query{
		Enabled: enabled, Status: q.Get("status"), ResourceID: resourceID,
		From: from, To: to, Limit: limit, Cursor: q.Get("cursor"),
	}, ""
}
