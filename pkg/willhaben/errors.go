package willhaben

import (
	"errors"
	"fmt"
)

// ErrUnauthorized is returned when willhaben's search API rejects a
// request as unauthorized (unexpected for the public search API, but
// handled defensively).
var ErrUnauthorized = errors.New("willhaben: unauthorized")

// APIError is returned for non-200, non-401 responses from willhaben's
// search API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("willhaben: unexpected status %d: %s", e.StatusCode, e.Body)
}
