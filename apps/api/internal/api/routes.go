package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"bookingms/internal/service"
	"bookingms/internal/service/demo"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type MockInput struct {
	Scenario string `header:"X-Mock-Scenario"`
}

type ItemInput struct {
	MockInput
	ID string `path:"id" format:"uuid"`
}

type closureInput struct {
	ItemInput
	ClosureID string `path:"closureId" format:"uuid"`
}

type PageInput struct {
	MockInput
	Limit  int    `query:"limit" minimum:"1" maximum:"100" default:"20"`
	Cursor string `query:"cursor" minLength:"1" maxLength:"2048"`
}

type reservationListInput struct {
	PageInput
	ResourceID string `query:"resourceId" format:"uuid"`
	Status     string `query:"status" enum:"confirmed,cancelled"`
	From       string `query:"from" format:"date-time"`
	To         string `query:"to" format:"date-time"`
}

type resourceListInput struct {
	PageInput
	Enabled bool `query:"enabled" default:"true"`
}

type availabilityInput struct {
	ItemInput
	From string `query:"from" format:"date-time" required:"true"`
	To   string `query:"to" format:"date-time" required:"true"`
}

type createReservationInput struct {
	MockInput
	RequestID      string `header:"X-Request-Id"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           CreateReservationRequest
}

type createResourceInput struct {
	MockInput
	Body CreateResourceRequest
}

type itemWriteInput[T any] struct {
	ItemInput
	Body T
}

type jsonOutput[T any] struct {
	ContentType string `header:"Content-Type"`
	Body        T
}

func newHuma(mux *http.ServeMux) huma.API {
	config := huma.DefaultConfig("API genérica de reservas LPS", "0.4.0")
	config.OpenAPIPath, config.DocsPath, config.SchemasPath = "", "", ""
	config.CreateHooks = nil
	config.Formats = map[string]huma.Format{
		"application/json":                huma.DefaultJSONFormat,
		"application/json; charset=utf-8": huma.DefaultJSONFormat,
	}
	config.Components.Schemas = huma.NewMapRegistry("#/components/schemas/", func(t reflect.Type, hint string) string {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		switch t {
		case reflect.TypeFor[demo.Page[Reservation]]():
			return "ReservationPage"
		case reflect.TypeFor[demo.Page[Resource]]():
			return "ResourcePage"
		case reflect.TypeFor[huma.ErrorDetail]():
			return "ErrorDetailModel"
		case reflect.TypeFor[Interval]():
			return "AvailableWindow"
		default:
			return huma.DefaultSchemaNamer(t, hint)
		}
	})
	return humago.New(mux, config)
}

func registerRoutes(api huma.API, s *handler) {
	const reservations = "/api/v1/reservas"
	const resources = "/api/v1/recursos"

	registerJSON[reservationListInput, demo.Page[Reservation]](api, s, "GET", reservations, 200, operation{name: "listReservations"})
	registerJSON[reservationListInput, demo.Page[Reservation]](api, s, "GET", "/api/v1/admin/reservas", 200, operation{name: "listAdminReservations", admin: true})
	registerJSON[createReservationInput, Reservation](api, s, "POST", reservations, 201, operation{name: "createReservation"})
	registerJSON[ItemInput, Reservation](api, s, "GET", reservations+"/{id}", 200, operation{name: "getReservation"})
	registerJSON[itemWriteInput[UpdateReservationRequest], Reservation](api, s, "PATCH", reservations+"/{id}", 200, operation{name: "updateReservation", minFields: 2})
	registerJSON[itemWriteInput[CancelReservationRequest], Reservation](api, s, "POST", reservations+"/{id}/cancelaciones", 200, operation{name: "cancelReservation"})
	registerJSON[resourceListInput, demo.Page[Resource]](api, s, "GET", resources, 200, operation{name: "listResources"})
	registerJSON[createResourceInput, ResourceDetail](api, s, "POST", resources, 201, operation{name: "createResource", admin: true})
	registerJSON[ItemInput, ResourceDetail](api, s, "GET", resources+"/{id}", 200, operation{name: "getResource", admin: true})
	registerJSON[itemWriteInput[UpdateResourceRequest], ResourceDetail](api, s, "PATCH", resources+"/{id}", 200, operation{name: "updateResource", admin: true, minFields: 2})
	registerJSON[itemWriteInput[UpdateResourceScheduleRequest], ResourceDetail](api, s, "PUT", resources+"/{id}/horarios", 200, operation{name: "updateResourceSchedule", admin: true})
	registerJSON[itemWriteInput[CreateResourceClosureRequest], ResourceClosure](api, s, "POST", resources+"/{id}/bloqueos", 201, operation{name: "createResourceClosure", admin: true})
	registerJSON[availabilityInput, Availability](api, s, "GET", resources+"/{id}/disponibilidades", 200, operation{name: "getResourceAvailability"})
	op := operation{name: "deleteResourceClosure", admin: true}
	huma.Register(api, humaOperation(s, "DELETE", resources+"/{id}/bloqueos/{closureId}", 204, op), func(ctx context.Context, _ *closureInput) (*struct{}, error) {
		if _, fault := s.service.Execute(ctx, op.name, validatedInput(ctx)); fault != nil {
			return nil, s.fault(fault)
		}
		return nil, nil
	})
}

func configureRequest[I any](op *operation) {
	body, ok := reflect.TypeFor[I]().FieldByName("Body")
	if !ok {
		return
	}
	op.request = func() any { return reflect.New(body.Type).Interface() }
	var required, allowed []string
	for _, field := range reflect.VisibleFields(body.Type) {
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		allowed = append(allowed, name)
		if !includes(strings.ReplaceAll(options, ",", " "), "omitempty") {
			required = append(required, name)
		}
	}
	op.required = strings.Join(required, " ")
	op.allowed = strings.Join(allowed, " ")
}

func validatedInput(ctx context.Context) service.Input {
	return ctx.Value(validatedInputKey{}).(service.Input)
}

func registerJSON[I, T any](api huma.API, s *handler, method, path string, status int, op operation) {
	configureRequest[I](&op)
	huma.Register(api, humaOperation(s, method, path, status, op), func(ctx context.Context, _ *I) (*jsonOutput[T], error) {
		result, fault := s.service.Execute(ctx, op.name, validatedInput(ctx))
		if fault != nil {
			return nil, s.fault(fault)
		}
		body, ok := result.(T)
		if !ok {
			return nil, fmt.Errorf("unexpected response type %T for %s", result, op.name)
		}
		return &jsonOutput[T]{ContentType: "application/json; charset=utf-8", Body: body}, nil
	})
}

func humaOperation(s *handler, method, path string, status int, op operation) huma.Operation {
	return huma.Operation{
		OperationID: string(op.name), Method: method, Path: path, DefaultStatus: status,
		// Keep Agenda's 400 response, strict JSON decoding, and auth-first ordering.
		SkipValidateParams: true, SkipValidateBody: true,
		Middlewares: huma.Middlewares{func(ctx huma.Context, next func(huma.Context)) {
			s.validate(op, ctx, next)
		}},
	}
}
