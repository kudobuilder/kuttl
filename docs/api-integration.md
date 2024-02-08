# API Integration

It is possible to integrate KUTTL into your own Go test infrastructure.  KUDO provides as an example `kubectl kudo test` using the KUTTL test harness.  The following are the necessary steps.

## Add KUTTL to Go.mod

`go get github.com/kudobuilder/kuttl`

or get a specific version

`go get github.com/kudobuilder/kuttl@v0.1.0`

## Common Imports to Use

The test harness type is defined in an `apis` package similar to a Kubernetes type along with a version package.  The test harness is currently `v1beta1` and provides the main configuration for a test suite.

The `test` package contains the `test.Harness` implementation (given the configuration of the test harness configuration type previously mentioned).  The `test.Harness` provides the "run" of the test run and needs a Go `t *testing.T`.

The `testutils` package contains utilities for docker, kubernetes, loggers and testing.

```go
import (
  harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
  "github.com/kudobuilder/kuttl/pkg/test"
  testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
)
```

## Test Harness

The `harness.TestSuite` is the structure that controls how the test harness will run.

```go
options := harness.TestSuite{}
```

The Go `t *testing.T` and `harness.TestSuite` are provided to `test.Harness` which provides the implementation for testing.

```go
Run: func(cmd *cobra.Command, args []string) {
  testutils.RunTests("kudo", testToRun, options.Parallel, func(t *testing.T) {
    harness := test.Harness{
      TestSuite: options,
      T:         t,
    }

    harness.Run()
  })
},

```

A more complete example is provided in KUDOs [cmd/test.go](https://github.com/kudobuilder/kudo/blob/master/pkg/kudoctl/cmd/test.go)
