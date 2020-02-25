---
kep-number: 3
short-desc: This KEP describes how we will launch the process that is being tested as a part of KUTTL.
title: KUTTL Unit Under Test Declaration
authors:
  - "@jbarrick-mesosphere"
owners:
  - "@jbarrick-mesosphere"
editor: "@kensipe"
creation-date: 2020-02-12
last-updated: 2020-02-25
status: provisional
---

# KUTTL Unit Under Test Declaration

This KEP describes how we will launch the process that is being tested as a part of KUTTL.

## Table of Contents

* [Summary](#summary)
  * [Goals](#goals)
* [Proposal](#proposal)
* [Alternatives](#alternatives)

## Summary

In previous versions of the KUDO test harness, the KUDO manager was built-in to KUTTL and launched when the `--start-kudo` flag was set. Since KUTTL is being split out from KUDO and should be useful for operators not built on KUDO, we need to design an alternative mechanism that does not require compiling the operator into KUTTL.

### Goals

* Launch operators or controllers without requiring them to be built in to the KUTTL binary.
* Support operators not written in Go.
* Support starting an operator against a mocked control plane environment, without requiring a full Kubernetes environment.

## Proposal

The proposal is to extend the existing `commands` feature of the `TestStep` and `TestSuite` objects to support launching background processes.

Currently, it is possible to run commands as part of the bootstrapping of a test suite or as a part of a test step, for example:

```
apiVersion: kudo.dev/v1alpha1
kind: TestSuite
startControlPlane: true
commands:
- command: ./bin/kudo init
```

When the command is launched, it is provided a kubeconfig that can connect to the created control plane. This is almost suitable for launching the unit under test, however, the test suite will wait for the command to complete prior to continuing the tests. In order to support launching the unit under test, it should support launching background commands that are run. Once the background commands are started, the execution can be continued and the background commands are terminated at the end of the test suite.

The `command` objects already support two other settings: `namespaced` and `ignoreFailure`. We can extend this by adding a `background` setting that would allow running the command in the background. For example, to launch the KUDO controller, you could run:

```
apiVersion: kudo.dev/v1alpha1
kind: TestSuite
startControlPlane: true
commands:
- command: ./bin/manager
  background: true
```

The general workflow would be that the user could build the KUDO controller binary and then run KUTTL which would start up etcd and kube-apiserver and then run `./bin/manager` in the background before launching the test suite.

## Alternatives

* Build a library version of KUTTL that allows running a callback or some other user-provided implementation once the control plane has started. Unfortunately, this would require users to import KUTTL as a dependency which is quite heavy-weight with Kubernetes dependencies and would limit usefulness for non-Go controllers.
* Always use KIND for tests and drop support for control plane-only environments and then launch the unit under test via Kubernetes pods.
* Define a new setting in the TestSuite manifest rather than extending `commands`. This could potentially be more clear by making a new setting(e.g., creating a new `controllers` or `backgroundCommands` option), but this seems unnecessary.
