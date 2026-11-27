package model

import "time"

// ReservationStatus contains only the states agreed for the initial scope.
// The zero value is not a valid status. Services must reject unknown values.
type ReservationStatus string

const (
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

// Reservation occupies exactly one resource while confirmed.
// Responsible and CancellationPolicy must not be replaced by an update.
// Version supports detecting concurrent updates; it does not provide locking.
type Reservation struct {
	ID                 string
	ScopeID            string
	ResourceID         string
	Responsible        IdentityReference
	Interval           Interval
	Status             ReservationStatus
	CancellationPolicy CancellationPolicySnapshot
	Attributes         map[string]any
	Version            int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
