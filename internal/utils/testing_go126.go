//go:build go1.26

package utils

// ModulePath was added to testing.testDeps in Go 1.26.
func (testDeps) ModulePath() string { return "" }
