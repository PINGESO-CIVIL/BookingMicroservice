package model

import "time"

// ExternalReference links a reservable resource to its owning catalog, if any.
// System identifies the catalog and ID identifies the resource within it.
type ExternalReference struct {
	System string
	ID     string
}

// Resource is an exclusively reservable unit within a scope.
// Identifiers are opaque; their generation and storage format are not fixed here.
type Resource struct {
	ID                string
	ScopeID           string
	Name              string
	Enabled           bool
	TimeZone          string // IANA location name, not the server's local zone.
	ExternalReference *ExternalReference
	Attributes        map[string]any // Arbitrary JSON-data dependes on the resource type and the context. 
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
