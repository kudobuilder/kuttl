// Package test provides a public API for KUTTL test harness functionality.
package test

import "github.com/kudobuilder/kuttl/internal/harness"

// This type alias is here to avoid breaking the only apparent active user of kuttl Go API,
// https://github.com/kube-green/kube-green/blob/main/tests/integration/kuttl_test.go

type Harness = harness.Harness
