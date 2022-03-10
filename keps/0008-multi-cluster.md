---
kep-number: 8
short-desc: This KEP describes how we will support more than one cluster under test
title: KUTTL Mult-Cluster Support
authors:
  - "@jbarrick-mesosphere"
  - "miles-Garnsey"
owners:
  - "@jbarrick-mesosphere"
  - "miles-Garnsey"
creation-date: 2021-01-08
last-updated: 2022-03-09
status: provisional
---

# KUTTL Multi-Cluster Support

This KEP describes how we will support more than one cluster in KUTTL.

## Table of Contents

* [Summary](#summary)
  * [Goals](#goals)
* [Proposal](#proposal)
* [Alternatives](#alternatives)

## Summary

When working with more than one cluster (for example, federated clusters), then it may be desirable to apply or assert on resources in multiple clusters. In my use-case, I want to ensure that federated resources are properly propagated to the target cluster.

Currently, KUTTL only supports a single test cluster. This KEP describes how we can allow custom Kubernetes contexts per test step.

### Goals

* Support applying resources across more than one test cluster.
* Support asserting on resources across more than one test cluster.
* Support configuring more than one cluster for testing (KinD initially).

## Proposal - cluster configuration and TestSuite changes

We propose to add additional fields to the `TestSuite` API to configure the clusters that the tests will run on. 

In the struct `TestSuite`:

```
// Config for multiple clusters.
	MultiClusterConfig MultiClusterConfig `json:"multiClusterConfig"`
```

The definition for `MultiClusterConfig` and associated supporting API elements:

```
//KindConfig contains settings for a single kind cluster.
type KindConfig struct {
	// Whether or not to start a local kind cluster for the tests.
	StartKIND bool `json:"startKIND"`
	// Path to the KIND configuration file to use.
	KINDConfig string `json:"kindConfig"`
	// KIND context to use.
	KINDContext string `json:"kindContext"`
	// If set, each node defined in the kind configuration will have a docker named volume mounted into it to persist
	// pulled container images across test runs.
	KINDNodeCache bool `json:"kindNodeCache"`
	// Containers to load to each KIND node prior to running the tests.
	KINDContainers []string `json:"kindContainers"`
	// If set, do not delete the resources after running the tests (implies SkipClusterDelete).
	SkipDelete bool `json:"skipDelete"`
	// If set, do not delete the mocked control plane or kind cluster.
	SkipClusterDelete bool `json:"skipClusterDelete"`
}
type MapKindConfig map[string]KindConfig
type MultiClusterConfig struct {
	// Type of cluster, KinD, external, etc.
	ClusterType      string     `json:"globalKindConfig,omitempty"`
	// Number of clusters to create from a global cluster spec.
	NumClusters      *int       `json:"numClusters,omitempty"`
	// Global config for kind clusters
	GlobalKindConfig KindConfig `json:"globalKindConfig,omitempty"`
	// Map of configurations for individual kind clusters.
	MapKindConfig    `json:",inline,omitempty"`
}
```

The objectives of the above structs are as follows:

1. Allow for different kinds of clusters to be used as specified by `ClusterType` (only KinD implemented initially).
2. Allow these clusters to be configured either at a global level (via `GlobalKindConfig` and a `NumClusters` parameter), or individually via MapKindConfig.
3. Leave latitude for the addition of other cluster types at a later time.
4. Provide the interface already present in the API for configuring the kind clusters, but encapsulate it in `KindConfig`.

## Proposal - teststep

The proposal is to add a new setting to the `TestStep` object: `kubeconfig`. This setting would allow the user to specify an alternative kubeconfig path to use for a given test step.

```
apiVersion: kudo.dev/v1alpha1
kind: TestStep
kubeconfig: ./secondary-cluster.yaml
```

If the kubeconfig setting is not set, then the global Kubernetes client is used.

If the kubeconfig setting is set, then it will be used for all Kubernetes operations within the step: commands (the KUBECONFIG environment variable will be set to the kubeconfig setting, relative to the KUTTL CLI's working directory), applied resources, asserts, and errors. This means that a single step can only be configured to use a single kubeconfig, but multiple steps can be used if a test case needs to interact with more than one cluster. This allows, for example, a federated resource to be created in one step and then another step can be used to verify that it actually exists on the destination cluster.

Note that the `kubeconfig` setting on the `TestStep` would be unaffected by the global Kubernetes configuration, so the `--kubeconfig` flag, `$KUBECONFIG` environment variable, etc, will be ignored for these steps.

A namespace is generated for each `TestCase` and this needs to be created in each cluster referenced by `TestSteps` within the `TestCase`. At the beginning of the `TestCase`, the generated namespace will be created in every cluster used in the `TestCase`. The namespaces will also be deleted at the end if `--skip-delete` is not set.
