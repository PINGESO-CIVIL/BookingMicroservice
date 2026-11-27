// Package api defines the public JSON transport of the BookingMS mock.
package api

import "bookingms/internal/service/demo"

type Attributes = demo.Attributes
type ExternalReference = demo.ExternalReference
type IdentityReference = demo.IdentityReference
type ResourcePolicy = demo.ResourcePolicy
type WeeklyScheduleRule = demo.WeeklyScheduleRule
type Interval = demo.Interval
type Reservation = demo.Reservation
type Resource = demo.Resource
type ResourceDetail = demo.ResourceDetail
type ResourceClosure = demo.ResourceClosure
type Availability = demo.Availability
type ErrorResponse = demo.ErrorResponse
type Error = demo.Error
type ErrorDetail = demo.ErrorDetail

type CreateReservationRequest struct {
	Interval
	ResourceID string      `json:"resourceId"`
	Attributes *Attributes `json:"attributes,omitempty"`
}

type UpdateReservationRequest struct {
	ExpectedVersion int64       `json:"expectedVersion"`
	ResourceID      *string     `json:"resourceId,omitempty"`
	StartAt         *string     `json:"startAt,omitempty"`
	EndAt           *string     `json:"endAt,omitempty"`
	Attributes      *Attributes `json:"attributes,omitempty"`
}

type CancelReservationRequest struct {
	ExpectedVersion int64 `json:"expectedVersion"`
}

type CreateResourceRequest struct {
	Name              string                `json:"name"`
	Enabled           bool                  `json:"enabled"`
	TimeZone          string                `json:"timeZone"`
	Policy            ResourcePolicy        `json:"policy"`
	ExternalReference *ExternalReference    `json:"externalReference,omitempty"`
	WeeklySchedule    *[]WeeklyScheduleRule `json:"weeklySchedule,omitempty"`
	Attributes        *Attributes           `json:"attributes,omitempty"`
}

type UpdateResourceRequest struct {
	ExpectedVersion   int64              `json:"expectedVersion"`
	Name              *string            `json:"name,omitempty"`
	Enabled           *bool              `json:"enabled,omitempty"`
	TimeZone          *string            `json:"timeZone,omitempty"`
	Policy            *ResourcePolicy    `json:"policy,omitempty"`
	ExternalReference *ExternalReference `json:"externalReference,omitempty"`
	Attributes        *Attributes        `json:"attributes,omitempty"`
}

type UpdateResourceScheduleRequest struct {
	ExpectedVersion int64                `json:"expectedVersion"`
	Policy          ResourcePolicy       `json:"policy"`
	WeeklySchedule  []WeeklyScheduleRule `json:"weeklySchedule"`
}

type CreateResourceClosureRequest struct {
	Interval
	Reason string `json:"reason"`
}

type HealthResponse struct {
	Status     string `json:"status"`
	Mode       string `json:"mode"`
	Persistent bool   `json:"persistent"`
}
