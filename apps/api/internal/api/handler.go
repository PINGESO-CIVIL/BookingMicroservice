package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"bookingms/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type handler struct {
	service   *service.Service
	sequence  atomic.Uint64
	authorize Authorizer
}

type operation struct {
	name              service.Operation
	admin             bool
	request           func() any
	required, allowed string
	minFields         int
}

// New connects the HTTP routes to the application service.
func New(svc *service.Service) http.Handler {
	return NewWithAuthorizer(svc, nil)
}

func NewWithAuthorizer(svc *service.Service, authorize Authorizer) http.Handler {
	s := &handler{service: svc, authorize: authorize}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		s.reply(w, 200, HealthResponse{Status: "ok", Mode: "mock", Persistent: false})
	})
	registerRoutes(newHuma(mux), s)
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
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Mock-Scenario, X-Request-Id")
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

type transportError struct {
	ErrorResponse
	status int
}

func (e *transportError) Error() string             { return e.ErrorResponse.Error.Message }
func (e *transportError) GetStatus() int            { return e.status }
func (e *transportError) ContentType(string) string { return "application/json; charset=utf-8" }

func (s *handler) fault(fault *service.Fault) *transportError {
	if fault.Message != "" {
		return &transportError{
			ErrorResponse: ErrorResponse{Error: Error{Code: "INVALID_REQUEST", Message: fault.Message, RequestID: fmt.Sprintf("request-demo-%06d", s.sequence.Add(1))}},
			status:        http.StatusBadRequest,
		}
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
	return &transportError{ErrorResponse: s.service.ErrorExample(fault.Example), status: status}
}

type validatedInputKey struct{}

// validate preserves the existing authorization order and error format before
// Huma decodes the normalized JSON into the registered transport input.
func (s *handler) validate(op operation, ctx huma.Context, next func(huma.Context)) {
	r, w := humago.Unwrap(ctx)
	scenario := r.Header.Get("X-Mock-Scenario")
	actor := service.Anonymous
	var subject string
	if s.authorize != nil {
		authenticatedSubject, fault := s.authorize(r.Header.Get("Authorization"), operationPermission(op))
		if fault != nil {
			response := s.fault(fault)
			s.reply(w, response.status, response.ErrorResponse)
			return
		}
		subject = authenticatedSubject
		actor = service.DemoUser
		if op.admin {
			actor = service.DemoAdmin
		}
	} else {
		switch r.Header.Get("Authorization") {
		case "Bearer demo-user":
			actor = service.DemoUser
		case "Bearer demo-admin":
			actor = service.DemoAdmin
		}
	}
	if fault := s.service.Authorize(op.admin, actor, scenario); fault != nil {
		response := s.fault(fault)
		s.reply(w, response.status, response.ErrorResponse)
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
	var confirmation service.Confirmation
	if op.request != nil {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			s.fail(w, 400, "INVALID_REQUEST", "Se requiere un cuerpo JSON.")
			return
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		target := op.request()
		if err := DecodeRequest(body, target, op.required, op.allowed, op.minFields); err != nil {
			s.fail(w, 400, "INVALID_REQUEST", "La solicitud contiene campos ausentes, no admitidos o con formato incorrecto.")
			return
		}
		normalized, err := json.Marshal(target)
		if err != nil {
			s.fail(w, 500, "INTERNAL_ERROR", "No se pudo preparar la solicitud JSON.")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(normalized))
		if create, ok := target.(*CreateReservationRequest); ok {
			confirmation.StartAt, confirmation.EndAt = create.StartAt, create.EndAt
		}

	}
	if op.name == "createReservation" {
		key := r.Header.Get("Idempotency-Key")
		if !ValidText(key, 128) {
			s.fail(w, 400, "INVALID_REQUEST", "Idempotency-Key debe tener entre 1 y 128 caracteres.")
			return
		}
		confirmation.Authorization = r.Header.Get("Authorization")
		confirmation.RequestID = r.Header.Get("X-Request-Id")
		if confirmation.RequestID == "" {
			confirmation.RequestID = rand.Text()
		}
		confirmation.IdempotencyKey = notificationKey(actor, subject, key)
	}
	id := r.PathValue("id")
	closureID := r.PathValue("closureId")
	if (id != "" && !ValidUUID(id)) || (closureID != "" && !ValidUUID(closureID)) {
		s.fail(w, 400, "INVALID_REQUEST", "El identificador debe ser un UUID.")
		return
	}
	id = strings.ToLower(id)
	input := service.Input{ID: id, ClosureID: strings.ToLower(closureID), Scenario: scenario, Confirmation: confirmation}
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
	next(huma.WithValue(ctx, validatedInputKey{}, input))
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

// G10 accepts UUID keys. Scope the derived key to the Agenda actor and write.
func notificationKey(actor service.Actor, subject, key string) [16]byte {
	scope := fmt.Sprintf("demo:%d", actor)
	if subject != "" {
		scope = fmt.Sprintf("jwt:%q", subject)
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("bookingms:confirmacion:v1:%s:%s", scope, key)))
	var id [16]byte
	copy(id[:], digest[:16])
	id[6] = (id[6] & 0x0f) | 0x80
	id[8] = (id[8] & 0x3f) | 0x80
	return id
}
