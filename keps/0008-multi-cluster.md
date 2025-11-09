---
kep-number: 8
short-desc: This KEP describes how we will support more than one cluster under test
title: KUTTL Mult-Cluster Support
authors:
  - "@jbarrick-mesosphere"
owners:
  - "@jbarrick-mesosphere"
creation-date: 2021-01-08
last-updated: 2021-01-08
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

## Proposal

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

### User-specfied namespaces

The multi-kubeconfig support has subtle implications on kuttl's behaviour when a user specifies the namespace name.
This is especially true when the various kubeconfigs in fact refer to the same cluster, perhaps with different contexts.

There is no easy way to tell whether the clusters behind these kubeconfigs overlap, especially when using lazy kubeconfig loading.
If they do, then there is no clear client <-> namespace ownership relation.

When this KEP was initially implemented, the rules that kuttl obeyed had been:
- no user-specified namespace: make an ephemeral namespace name, for every client: create it (no error if already present), and clean it up afterward
- otherwise:
  - BEFORE the test, check if the supplied namespace exists, but USING THE DEFAULT CLIENT ONLY, remember this fact for all clients in the case (!)
  - if existed, do not touch it, for any client
  - if missing, then for EVERY CLIENT create it (no error if already present) and clean it up afterward

There are following issues in the above logic:
- if the ephemeral namespace exists, we do not know whether we're clashing with a third-party-created
  namespace (and should abort the test) or just seeing the same namespace resource via two different
  kubeconfigs.
- the existence of namespace for the default client should not determine (at least not unconditionally)
  the fate of the namespace for other clients.
- in particular, if a namespace was missing on default cluster, but pre-existed on auxiliary cluster,
  kuttl would clean the latter up. If it was the other way round, it would not touch any namespace,
  and likely fail due to missing namespace on the auxiliary cluster.

This logic was changed in [PR #637](https://github.com/kudobuilder/kuttl/pull/637) to try and be more thoughtful, by maintaining the presence/absence information separately for every client.
However, for backward compatibility we still take into account whether the namespace was user supplied or not.
