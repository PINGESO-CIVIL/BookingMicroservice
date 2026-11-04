package model

import "time"

// CancellationPolicySnapshot records the cancellation terms accepted when a
// reservation was confirmed. Notice must be non-negative and remains unchanged
// when configuration, the reserved resource, or the reservation interval changes.
// PolicyID and PolicyVersion identify the original policy, not the current one.
type CancellationPolicySnapshot struct {
	PolicyID      string
	PolicyVersion int64
	Notice        time.Duration
}
