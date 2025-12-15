package testing

import (
	"testing"
	"testing/quick"
)

// PropertyTestConfig holds configuration for property-based tests
type PropertyTestConfig struct {
	MaxCount int // Number of iterations (minimum 100 as per requirements)
}

// DefaultPropertyConfig returns the default configuration for property tests
func DefaultPropertyConfig() *quick.Config {
	return &quick.Config{
		MaxCount: 100, // Minimum 100 iterations as per requirements
	}
}

// RunPropertyTest runs a property-based test with the default configuration
func RunPropertyTest(t *testing.T, f interface{}, testName string) {
	t.Helper()
	config := DefaultPropertyConfig()
	
	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property test failed for %s: %v", testName, err)
	}
}