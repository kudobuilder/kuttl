---
kep-number: 5
short-desc: Overview of providing KUTTL for use as a Go library
title: KUTTL as a Library
authors:
  - "@mpereira"
  - "@jbarrick-mesosphere"
  - "@kensipe"
owners:
  - "@jbarrick-mesosphere"
editor: "@kensipe"
creation-date: 2020-02-13
last-updated: 2020-04-07
status: provisional
---

# KUTTL as a Library

## Table of Contents

* [Summary](#summary)
   * [Goals](#goals)
   * [Non-Goals](#non-goals)
   * [Principles](#principles)
* [Proposal](#proposal)
   * [cmd/cmd.go](#cmdcmdgo)
   * [kubectl/kubectl.go](#kubectlkubectlgo)
   * [k8s/k8s.go](#k8sk8sgo)
   * [operators](#operators)

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
* Separation of concerns between domains (Operators, Kubernetes, etc.)

## Proposal

We will improve and break-down the interface of the kuttl library to make it easier to consume as a downstream developer. Libraries that need to be broken out of kuttl:

* A generic Kubernetes client (currently this is contained inside of `pkg/test/harness.go`).
* Kubernetes object constructors (this is already pretty good in `pkg/test/utils/kubernetes.go`).
* Assertion methods (needs better documentation for library usage).
* Test setup and tear down methods (is currently coupled to steps).
* Cluster and test environment provisioning (this is currently coupled to `Harness` in `harness.go`).

Additionally, Operator-specific helper methods will be provided in the library that can be used by library users.

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

### operators

Kuttl needs a package in order to provide conveniences for interacting with operators and webhooks.

* Partial CRUD functionality for CRs / CRDs (get)
