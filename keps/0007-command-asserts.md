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

By asserting CLI commands, kuttl can support many additional test scenarios that examing the internal state of an operator.

## Motivation

Operators and custom resources can be used to deploy complex applications. For example, KUDO provides operators for Apache Cassandra and Apache Kafka. Deploying custom resources of these operators may not always result in changed Kubernetes resources but also in changes that aren't surfaced by Kubernetes resources but are internal to the operator. Asserting these changes only works through accessing service APIs or running commands and parsing their output. kuttl should provide tools to assert such output.

### Goals

- Assert that commands examining the internal state of an operator succeed or fail
- Keep eventual consistency when asserting commands

### Non-Goals

- Introduce "flows" to determine when to run commands. I.e. to run commands once before or after asserting resources

## Proposal

Introduce commands as part of `TestAssert`. These commands are run at a regular interval like resource assertions until they either succeed or a timeout of the assert is reached. If any of the commands fail, the assert is considered to have failed. A command's return value determines if a command failed or not.

As a start, complex commands that parse output can be described using shell environments. At a later stage we can implement additional tools to simplify common tasks. E.g. to search with a regular expression in command output.

### User Stories

#### User Story 1

The JVM options of a KUDO operator can be tuned using parameters. On deployment, the JVM options are logged in the pods running the application provided by the operator. To assert that a parameter change resulted in changed JVM options we need to run `kubectl logs pod-name` and search the logs for a specific string once the operator has been deployed.

#### User Story 2

An operator can be configured to encrypt it's API. To assert that the API endpoints are indeed encrypted we can run a specific `curl` command and check that it returns successfully once the operator has been deployed.

### Implementation Details

Command assertions use most fields of the existing commands for `TestStep` and follow their behavior. The `ignoreFailure`, `background` and `timeout` fields are removed as they are not needed here.

An example that parses the logs of a pod:

```yaml
apiVersion: kuttl.dev/v1beta1
type: TestAssert
commands:
  - command: ./parse-logs.sh name-of-pod-0 string-to-search-for
    namespaced: true
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

### Risks and Mitigations

Some commands might be expensive to run and it would be better to run them once instead of running them repeatedly. This can be worked around by running commands as part of a test step instead. Commands in a test step run once.

There could be subtle race conditions when running commands that have to run after a `TestStep` has been applied. The eventual consistency of Kubernetes might result in a command running in an environment where the changes of the `TestStep` haven't been applied yet. E.g., consider a `TestStep` that changes a `Deployment` and a command assert that examines the log of a `Pod` run by the `Deployment`. The change of the `Deployment` will restart pods, but as the restart is not immediate, the command might be run with the old pod.
As a workaround, users can create an assert that doesn't have corresponding `TestStep`. This will then get asserted after the prior assert succeeded.

## Graduation Criteria

The `commands` field described in the [Implementation Details](#implementation-details) runs and asserts command return values.

## Implementation History

- 2020-10-28 - Initial draft (@nfnt)
