---
kep-number: 0
short-desc: Overview of providing KUTTL for use as a Go library
title: KUTTL as a Library
authors:
  - "@mpereira"
  - "@jbarrick-mesosphere"
owners:
  - "@mpereira"
  - "@jbarrick-mesosphere"
editor: TBD
creation-date: 2020-02-13
last-updated: 2020-02-13
status: provisional
see-also:
  - KEP-2
---

# KUTTL as a Library

## Table of Contents

* [Summary](#summary)
    * [Goals](#goals)
    * [Non-Goals](#non-goals)
* [Proposal](#proposal)


## Summary

Currently, the test harness supports only writing tests in YAML which works well for the majority of cases but there are some use-cases that are currently hard to support with the fully declarative tests (for example, asserts that require context from multiple fields or objects or fuzzy asserts). Additionally, there are some developers who would prefer to write tests in Go due to their personal familiarity with the language. There are also other developers that would prefer to write fully declarative tests and not need to deal with a full on programming language.

Rather than building an entirely new library, we will unify the library and test harness efforts to ensure that improvements to one benefit both and that we do not duplicate work or fragment the community.

### Goals

* Support writing tests in Go for developers that prefer a traditional programming approach.
* Increase code reuse, prevent the need for each operator to have their own special testing library.
* Share expertise and code between fully declarative and Go-based tests.
* Allow for "breaking the glass" when declarative tests are not expressive enough.

### Non-Goals

* Creating a new spin-off testing project.
* Fragmentation of our operator developer community.
* Dropping support for YAML-based declarative tests.

### Principles

* Useful logging by default
* Depend on existing libraries whenever possible
* Retrying by default where it makes sense
* Ease of adding retrying to non-retried functions
* It is better to be consistent than to be "perfect"
* Err on the side of flexibility
* Err on the side of returning errors instead of panicking/Fatalf/etc.
* Cohesive logging output (parameters, formatting, etc.)
* Separation of concerns between domains (KUDO, Kubernetes, etc.)

## Proposal

We will improve and break-down the interface of the kuttl library to make it easier to consume as a downstream developer. Libraries that need to be broken out of kuttl:

* A generic Kubernetes client (currently this is contained inside of `pkg/test/harness.go`).
* Kubernetes object constructors (this is already pretty good in `pkg/test/utils/kubernetes.go`).
* Assertion methods (needs better documentation for library usage).
* Test setup and tear down methods (is currently coupled to steps).
* Cluster and test environment provisioning (this is currently coupled to `Harness` in `harness.go`).

Additionally, KUDO-specific helper methods will be provided in the library that can be used by library users.

We will also endeavor to stay close to the API used by the kudo-cassandra-operator testing library to ensure that migration of existing tests is easy.

### cmd/cmd.go

The cmd package should provide conveniences for running arbitrary commands (i.e. shelling out) from Go code in integration tests.

Current functionality:
* Providing an Exec function that returns exit status, stdout and stderr

```
package cmd

func Exec(
  command string,
  arguments []string,
  environmentVariables map[string]string,
) (
  exitStatus int,
  stdout *bytes.Buffer,
  stderr *bytes.Buffer,
  error error,
)
```

### kubectl/kubectl.go

The kubectl package should provide conveniences for interacting with a Kubernetes cluster as one would do via kubectl from the shell.

```
package kubectl

func Exec(
  namespaceName string,
  containerName string,
  command string,
  argvuments []string,
  environmentVariables map[string]string,
) error
```

### k8s/k8s.go

The k8s package should provide conveniences for interacting with Kubernetes objects.

* Partial CRUD functionality for namespaces
* CRUD functionality for all useful Kubernetes objects
* Current known use-cases:
* Creating and deleting namespaces during test runs

```
package k8s

func Init(kubectlOptions *kubectl.KubectlOptions) error

func CreateNamespace(namespaceName string) error

func DeleteNamespace(namespaceName string) error
```

### kudo/kudo.go

The kudo package should provide conveniences for interacting with the KUDO "system" and KUDO operators.

* Partial CRUD functionality for instances (get)
* Convenience function for getting the "instance aggregated status" from an instance
* Install operator from directory
* Uninstall operator

```
package kudo

import (
  "github.com/kudobuilder/kudo/pkg/apis/kudo/v1alpha1"
  kubectl "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
)

func Init(kubectlOptions *kubectl.KubectlOptions) error

func GetInstance(
  namespaceName string,
  instanceName string,
) (*v1alpha1.Instance, error)

func GetInstanceAggregatedStatus(
  namespaceName string,
  instanceName string,
) (*v1alpha1.ExecutionStatus, error)

func WaitForOperatorDeployComplete(
  namespaceName string,
  instanceName string,
) error

func InstallOperatorFromDirectory(
  directory string,
  namespaceName string,
  instanceName string,
  parameters []string,
) error

func UninstallOperator(
  operatorName string,
  namespaceName string,
  instanceName string,
) error

func UpdateInstanceParameters(
  namespaceName string,
  instanceName string,
  parameters map[string]string,
) error

func Exec(
  namespaceName string,
  instanceName string,
  podName string,
  containerName string,
  command string,
  arguments string,
  environment string[],
) error
```
