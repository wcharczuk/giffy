package assert

import (
	"testing"
)

// Filter is a unit test filter.
type Filter string

// Filters
const (
	// Unit is a filter for unit tests.
	Unit = "unit"
	// Acceptance is a filter for acceptance tests.
	Acceptance = "acceptance"
	// Integration is a filter for integration tests.
	Integration = "integration"
)

// CheckFilter checks the filter.
func CheckFilter(t *testing.T, filter Filter) {}
