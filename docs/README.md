# Getting Started

## Pre-requisites

Before you get started using KUTTL, you need to have a running Kubernetes cluster setup. If you already have a cluster there are no prerequisites.  If you want to use the mocked control plane or Kind, you will need [Kind](https://github.com/kubernetes-sigs/kind).

- Setup a Kubernetes Cluster in version `1.13` or later
- Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) in version `1.13` or later.

## Install KUTTL CLI

Install the `kubectl kuttl` plugin. To do so, please follow the [CLI plugin installation instructions](cli.md).

The KUTTL CLI leverages the kubectl plugin system, which gives you all its functionality under `kubectl kuttl`.

## Using KUTTL

Once you have a running cluster with `kubectl` installed along with the KUTTL CLI plugin, you can run tests with KUTTL like so:

```bash
$ kubectl kuttl test path/to/test-suite
```

[Learn more](what-is-kuttl.md) about KUTTL and check out how to get started with the [KUTTL test harness](kuttl-test-harness.md).
