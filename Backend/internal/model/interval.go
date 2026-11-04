package model

import "time"

// Interval represents [Start, End): the start is inclusive and the end exclusive.
// Both values represent unambiguous instants; Start must be strictly before End.
// Local recurring schedules require a separate model once their rules are agreed.
type Interval struct {
	Start time.Time
	End   time.Time
}
