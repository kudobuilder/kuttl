---
kep-number: 7
short-desc: Assertions for CLI commands
authors:
  - "@nfnt"
owners:
  - "@nfnt"
creation-date: 2020-10-28
last-updated: 2020-10-28
status: provisional
---

# Assertions for CLI commands

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories](#user-stories)
    - [User Story 1](#user-story-1)
    - [User Story 2](#user-story-2)
  - [Implementation Details](#implementation-details)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Graduation Criteria](#graduation-criteria)
- [Implementation History](#implementation-history)

## Summary

By asserting CLI commands, we can check operator state that isn't exposed through Kubernetes resources. This allow kuttl to support more test scenarios.

## Motivation

The state of an operator is not always fully covered by Kubernetes objects. Some state may be expressed in stdout, logs or other tooling. kuttl should allow users to assert changes to this state. The simplest way to do this is by running commands or scripts that take care of asserting this state. E.g., by parsing log output. Such a command assertion can then either succeed or fail.

### Goals

- Assert that commands examining the internal state of an operator succeed or fail
- Assert commands in the same way Kubernetes objects are asserted

### Non-Goals

- Introduce "flows" to determine when to run commands. I.e. to run commands once before or after asserting resources

## Proposal

Introduce commands as part of `TestAssert`. These commands are run at a regular interval like resource assertions until all of them either succeed or the assertion timeout is reached. The error code of the command determine success or failure. If any of the commands fail, the assert is considered to have failed. A command's return value determines if a command failed or not. I.e., command assertions behave in the same way as the existing resource assertions.

As a start, complex commands that parse output can be described using shell environments. At a later stage we can implement additional tools to simplify common tasks. E.g. to search with a regular expression in command output.

As the commands run at a regular interval, their output won't be logged on every iteration. However, if a command fail, its output will be part of the failure message. If the assert runs into its timeout, this failure message will be logged. This behavior ensures that that failing commands are logged at most once. Later, additional fields can be added to provide finer control over command output logging.

### User Stories

#### User Story 1

The JVM options of a KUDO operator can be tuned using parameters. On deployment, the JVM options are logged in the pods running the application provided by the operator. To assert that a parameter change resulted in changed JVM options we need to run `kubectl logs pod-name` and search the logs for a specific string once the operator has been deployed.

#### User Story 2

An operator can be configured to encrypt it's API. To assert that the API endpoints are indeed encrypted we can run a specific `curl` command and check that it returns successfully once the operator has been deployed.

### Implementation Details

Command assertions use most fields of the existing commands for `TestStep` and follow their behavior. The `ignoreFailure`, `background` and `timeout` fields are removed as they are not needed here:

```go
type TestAssertCommand struct {
  // The command and argument to run as a string.
  Command string `json:"command"`
  // If set, the `--namespace` flag will be appended to the command with the namespace to use.
  Namespaced bool `json:"namespaced"`
  // Ability to run a shell script from TestStep (without a script file)
  // namespaced and command should not be used with script.  namespaced is ignored and command is an error.
  // env expansion is depended upon the shell but ENV is passed to the runtime env.
  Script string `json:"script"`
  // If set, the output from the command is NOT logged.  Useful for sensitive logs or to reduce noise.
  SkipLogOutput bool `json:"skipLogOutput"`
}
```

An example that parses the logs of a pod:

```yaml
apiVersion: kuttl.dev/v1beta1
type: TestAssert
commands:
  - command: ./parse-logs.sh name-of-pod-0 string-to-search-for
    namespaced: true
timeout: 60
```

with `parse-logs.sh`:

```bash
#!/bin/bash

POD=$1
SEARCH=$2
# $3 is "--namespace"
NAMESPACE=$4

kubectl logs -n ${NAMESPACE} ${POD} | grep ${SEARCH}
```

The commands are then asserted in the same way and in the same loop Kubernetes objects are asserted. I.e. the current `TestAssert` loop is changed to:

```go
var testErrors []error
for i := 0; i < timeout; i++ {
  // start fresh
  testErrors = []error{}
  for _, expected := range objects {
    testErrors = append(testErrors, s.CheckResource(expected, namespace)...)
  }

  for _, command := range commands {
    testErrors = append(testErrors, s.CheckCommand(command)...)
  }

  if len(testErrors) == 0 {
    break
  }

  time.Sleep(time.Second)
}
```

This loop needs to be refactored, as running these commands could be expensive:

```go
testErrors := checkAll(objects, commands, time.Duration(timeout)*time.Second, s, namespace)
```

with

```go
func checkAll(objects []runtime.Object, commands []harness.TestAssertCommand, timeout time.Duration, s *Step, namespace string) []error {
  ctx, cancel := context.WithTimeout(context.TODO(), timeout)
  defer cancel()

  var testErrors []error
  for {
    testErrors = []error{}

    for _, object := range objects {
      testErrors = append(testErrors, s.CheckResource(ctx, object, namespace)...)
    }

    for _, command := range commands {
      testErrors = append(testErrors, s.CheckCommand(ctx, command)...)
    }

    if len(testErrors) == 0 {
      break
    }

    if ctx.Err() != nil {
      // context timeout
      break
    }

    time.Sleep(time.Second)
  }

  return testErrors
}
```

In the future, additional fields may be added to `TestAssert` and `TestAssertCommand` to
 * run commands before or after resources have been asserted
 * indicate that a command is supposed to fail

### Risks and Mitigations

Some commands might be expensive to run and it would be better to run them once instead of running them repeatedly. This can be worked around by running commands as part of a test step instead. Commands in a test step run once.

There could be subtle race conditions when running commands that have to run after a `TestStep` has been applied. The eventual consistency of Kubernetes might result in a command running in an environment where the changes of the `TestStep` haven't been applied yet. E.g., consider a `TestStep` that changes a `Deployment` and a command assert that examines the log of a `Pod` run by the `Deployment`. The change of the `Deployment` will restart pods, but as the restart is not immediate, the command might be run with the old pod.
As a workaround, users can create an assert that doesn't have corresponding `TestStep`. This will then get asserted after the prior assert succeeded.

## Graduation Criteria

The `commands` field described in the [Implementation Details](#implementation-details) runs and asserts command return values.

## Implementation History

- 2020-10-28 - Initial draft (@nfnt)
