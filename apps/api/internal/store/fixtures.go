package store

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"

	"bookingms/internal/service/demo"
)

//go:embed fixtures.json
var fixtureJSON []byte

// Fixtures holds the immutable Hito 1 read data.
type Fixtures struct {
	data struct {
		Examples          map[string]json.RawMessage     `json:"examples"`
		Resources         []demo.Resource                `json:"resources"`
		Reservations      []demo.Reservation             `json:"reservations"`
		AdminReservations []demo.Reservation             `json:"adminReservations"`
		Details           map[string]demo.ResourceDetail `json:"details"`
		Availability      map[string]demo.Availability   `json:"availability"`
	}
}

func NewFixtures() (*Fixtures, error) {
	fixtures := new(Fixtures)
	if err := json.Unmarshal(fixtureJSON, &fixtures.data); err != nil {
		return nil, fmt.Errorf("fixtures de demostración inválidos: %w", err)
	}
	return fixtures, nil
}

func (f *Fixtures) Resources() []demo.Resource {
	return slices.Clone(f.data.Resources)
}

func (f *Fixtures) Reservations(admin bool) []demo.Reservation {
	if admin {
		return slices.Clone(f.data.AdminReservations)
	}
	return slices.Clone(f.data.Reservations)
}

func (f *Fixtures) Reservation(id string) (demo.Reservation, bool) {
	for _, reservation := range f.data.Reservations {
		if reservation.ID == id {
			return reservation, true
		}
	}
	return demo.Reservation{}, false
}

func (f *Fixtures) Resource(id string) (demo.ResourceDetail, bool) {
	value, ok := f.data.Details[id]
	value.Closures = slices.Clone(value.Closures)
	value.WeeklySchedule = slices.Clone(value.WeeklySchedule)
	if value.ExternalReference != nil {
		reference := *value.ExternalReference
		value.ExternalReference = &reference
	}
	return value, ok
}

func (f *Fixtures) Availability(id string) demo.Availability {
	value := f.data.Availability[id]
	value.AvailableWindows = slices.Clone(value.AvailableWindows)
	value.AvailableSlots = slices.Clone(value.AvailableSlots)
	return value
}

func (f *Fixtures) ReservationExample(name string) demo.Reservation {
	return decodeExample[demo.Reservation](f, name)
}

func (f *Fixtures) ResourceExample(name string) demo.ResourceDetail {
	return decodeExample[demo.ResourceDetail](f, name)
}

func (f *Fixtures) ClosureExample(name string) demo.ResourceClosure {
	return decodeExample[demo.ResourceClosure](f, name)
}

func (f *Fixtures) ErrorExample(name string) demo.ErrorResponse {
	return decodeExample[demo.ErrorResponse](f, name)
}

func decodeExample[T any](f *Fixtures, name string) T {
	var value T
	if err := json.Unmarshal(f.data.Examples[name], &value); err != nil {
		panic(fmt.Sprintf("ejemplo de demostración %s inválido: %v", name, err))
	}
	return value
}
