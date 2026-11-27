// Package service coordinates the Hito 1 use cases against read-only demo data.
package service

import (
	"context"
	"log"
	"slices"
	"time"

	"bookingms/internal/model/demo"
)

type FixtureStore interface {
	Resources() []demo.Resource
	Reservations(admin bool) []demo.Reservation
	Reservation(id string) (demo.Reservation, bool)
	Resource(id string) (demo.ResourceDetail, bool)
	Availability(id string) demo.Availability
	ReservationExample(name string) demo.Reservation
	ResourceExample(name string) demo.ResourceDetail
	ClosureExample(name string) demo.ResourceClosure
	ErrorExample(name string) demo.ErrorResponse
}

type AvailabilityStore interface {
	Ready(context.Context) bool
}

type Service struct {
	fixtures FixtureStore
	database AvailabilityStore
}

func New(fixtures FixtureStore, database AvailabilityStore) *Service {
	return &Service{fixtures: fixtures, database: database}
}

type Operation string

type Actor uint8

const (
	Anonymous Actor = iota
	DemoUser
	DemoAdmin
)

type Input struct {
	ID        string
	ClosureID string
	Scenario  string
	Query     Query
}

type Query struct {
	Enabled     bool
	Status      string
	ResourceID  string
	From, To    time.Time
	Limit       int
	Cursor      string
	CursorScope string
}

// Fault identifies a simulated failure or an invalid request.
type Fault struct {
	Example string
	Message string
}

func (s *Service) Ready(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.database.Ready(ctx)
}

func (s *Service) ErrorExample(name string) demo.ErrorResponse {
	return s.fixtures.ErrorExample(name)
}

func (s *Service) Authorize(adminOnly bool, actor Actor, scenario string) *Fault {
	if scenario == "unauthorized" || actor == Anonymous {
		return &Fault{Example: "Unauthenticated"}
	}
	if scenario == "forbidden" || (adminOnly && actor != DemoAdmin) {
		return &Fault{Example: "Forbidden"}
	}
	switch scenario {
	case "error":
		return &Fault{Example: "ServiceUnavailable"}
	case "", "normal", "success", "empty", "paginated", "cancelled", "expired", "started", "notFound", "disabled", "conflict", "version", "deadline", "policy":
		return nil
	default:
		return &Fault{Message: "Escenario de demostración no admitido."}
	}
}

func (s *Service) Execute(operation Operation, input Input) (any, *Fault) {
	switch operation {
	case "listReservations", "listAdminReservations":
		return s.listReservations(operation == "listAdminReservations", input)
	case "listResources":
		return s.listResources(input)
	case "getReservation", "updateReservation", "cancelReservation":
		reservation, found := s.fixtures.Reservation(input.ID)
		if !found || input.Scenario == "notFound" {
			return nil, &Fault{Example: "ReservationNotFound"}
		}
		if operation == "getReservation" {
			return s.reservationScenario(reservation, input.Scenario), nil
		}
	case "getResource", "updateResource", "updateResourceSchedule", "createResourceClosure", "deleteResourceClosure", "getResourceAvailability":
		resource, found := s.fixtures.Resource(input.ID)
		if !found || input.Scenario == "notFound" {
			return nil, &Fault{Example: "ResourceNotFound"}
		}
		switch operation {
		case "getResource":
			return resource, nil
		case "getResourceAvailability":
			return s.availability(input)
		case "deleteResourceClosure":
			if !slices.ContainsFunc(resource.Closures, func(closure demo.ResourceClosure) bool { return closure.ID == input.ClosureID }) {
				return nil, &Fault{Example: "ClosureNotFound"}
			}
		}
	}
	return s.simulateWrite(operation, input.Scenario)
}

func (s *Service) reservationScenario(reservation demo.Reservation, scenario string) demo.Reservation {
	id := reservation.ID
	switch scenario {
	case "cancelled":
		reservation = s.fixtures.ReservationExample("CancelledReservation")
	case "expired", "started":
		source := "00000000-0000-4000-8000-200000000003"
		if scenario == "started" {
			source = "00000000-0000-4000-8000-200000000004"
		}
		if fixture, found := s.fixtures.Reservation(source); found {
			reservation = fixture
		}
	}
	reservation.ID = id
	return reservation
}

func (s *Service) simulateWrite(operation Operation, scenario string) (any, *Fault) {
	if scenario != "" && scenario != "normal" && scenario != "success" {
		return nil, &Fault{Example: writeFailure(operation, scenario)}
	}
	switch operation {
	case "createReservation":
		log.Print("[MOCK] Confirmación de reserva: notificación simulada.")
		return s.fixtures.ReservationExample("ConfirmedReservation"), nil
	case "updateReservation":
		return s.fixtures.ReservationExample("UpdatedReservation"), nil
	case "cancelReservation":
		return s.fixtures.ReservationExample("CancelledReservation"), nil
	case "createResource":
		return s.fixtures.ResourceExample("ResourceDetail"), nil
	case "updateResource", "updateResourceSchedule":
		return s.fixtures.ResourceExample("UpdatedResourceDetail"), nil
	case "createResourceClosure":
		return s.fixtures.ClosureExample("ResourceClosure"), nil
	case "deleteResourceClosure":
		return nil, nil
	default:
		return nil, &Fault{Example: "InvalidRequest"}
	}
}

func writeFailure(operation Operation, scenario string) string {
	switch scenario {
	case "conflict":
		switch operation {
		case "updateResource", "updateResourceSchedule", "createResourceClosure":
			return "ResourceScheduleConflict"
		case "createReservation", "updateReservation":
			return "ResourceUnavailable"
		}
	case "version":
		switch operation {
		case "createReservation":
			return "IdempotencyConflict"
		case "updateReservation", "cancelReservation":
			return "VersionConflict"
		case "updateResource", "updateResourceSchedule":
			return "ResourceVersionConflict"
		}
	case "deadline":
		if operation == "updateReservation" || operation == "cancelReservation" {
			return "CancellationWindowClosed"
		}
	case "policy":
		if operation != "deleteResourceClosure" && operation != "cancelReservation" {
			return "PolicyViolation"
		}
	}
	return "InvalidRequest"
}
