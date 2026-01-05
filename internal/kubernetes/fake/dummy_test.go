// Package fake provides fake Kubernetes client implementations for testing.
package fake

import "testing"

// TestDummy is a no-op test to satisfy coverage tools
func TestDummy(_ *testing.T) {
	// This test exists only to ensure the package is recognized by go test
	// and to avoid "go: no such tool 'covdata'" warnings
}
