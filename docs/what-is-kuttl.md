# What is KUTTL

## Overview

The KUbernetes Test TooL (KUTTL) provides a declarative approach to testing production-grade Kubernetes [operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It provides a way to inject an operator (subject under test) during the TestSuite setup and allows tests to be standard YAML files.  Test assertions are often partial YAML documents which assert the state defined is true.

It is also possible to have KUTTL automate the setup of a cluster.

## Motivation

Testing Kubernetes operators is not easy. As the KUDO team was building a "declarative" Kubernetes operator, it just made sense to create a declarative way to test as well.  The motivation is to leverage the existing Kubernetes eco-system for resource management (YAMLs) in a way to **setup** a test and as well as a way to **assert** state within the cluster.

## When would you use KUTTL

The testing eco-system is vast and includes at a minimum low level unit tests, integration tests and end-to-end testing.  KUTTL is built to support some kubernetes integration test scenarios and is most valuable as an end-to-end (e2e) test harness.

KUTTL is great when you want to:

* Provide tests against your Custom Resource Definitions (CRDs)
* Inject a controller and assert states in a running cluster
* Test a set of TestSuites against multiple implementations and multiple versions of Kubernetes clusters.
