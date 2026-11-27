// Package demo contains the read models shared by the fixture store and the mock service.
package demo

type Attributes struct{}

type ExternalReference struct {
	System string `json:"system"`
	ID     string `json:"id"`
}

type IdentityReference struct {
	Issuer  string `json:"issuer"`
	Subject string `json:"subject"`
}

type ResourcePolicy struct {
	CancellationNoticeSeconds int64 `json:"cancellationNoticeSeconds"`
	MinDurationSeconds        int64 `json:"minDurationSeconds"`
	MaxDurationSeconds        int64 `json:"maxDurationSeconds"`
	SlotDurationSeconds       int64 `json:"slotDurationSeconds"`
}

type WeeklyScheduleRule struct {
	DayOfWeek      int    `json:"dayOfWeek"`
	StartLocalTime string `json:"startLocalTime"`
	EndLocalTime   string `json:"endLocalTime"`
}

type Interval struct {
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
}

type Reservation struct {
	Interval
	ID                        string            `json:"id"`
	ResourceID                string            `json:"resourceId"`
	Responsible               IdentityReference `json:"responsible"`
	Status                    string            `json:"status"`
	Attributes                Attributes        `json:"attributes"`
	CancellationNoticeSeconds int64             `json:"cancellationNoticeSeconds"`
	CancellationDeadline      string            `json:"cancellationDeadline"`
	Version                   int64             `json:"version"`
	CreatedAt                 string            `json:"createdAt"`
	UpdatedAt                 string            `json:"updatedAt"`
}

type Resource struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Enabled    bool       `json:"enabled"`
	TimeZone   string     `json:"timeZone"`
	Attributes Attributes `json:"attributes"`
}

type ResourceDetail struct {
	Resource
	Policy            ResourcePolicy       `json:"policy"`
	WeeklySchedule    []WeeklyScheduleRule `json:"weeklySchedule"`
	Closures          []ResourceClosure    `json:"closures"`
	ExternalReference *ExternalReference   `json:"externalReference,omitempty"`
	Version           int64                `json:"version"`
	CreatedAt         string               `json:"createdAt"`
	UpdatedAt         string               `json:"updatedAt"`
}

type ResourceClosure struct {
	Interval
	ID         string `json:"id"`
	ResourceID string `json:"resourceId"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"createdAt"`
}

type Availability struct {
	ResourceID          string     `json:"resourceId"`
	Enabled             bool       `json:"enabled"`
	TimeZone            string     `json:"timeZone"`
	From                string     `json:"from"`
	To                  string     `json:"to"`
	AvailableWindows    []Interval `json:"availableWindows"`
	AvailableSlots      []Interval `json:"availableSlots"`
	SlotDurationSeconds int64      `json:"slotDurationSeconds"`
}

type Page[T Resource | Reservation] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"nextCursor"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	RequestID string        `json:"requestId"`
	Details   []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}
