---
kep-number: 0
short-desc: Use this template for new KEPs
title: KUTTL Rego Support
authors:
  - "@jbarrick-mesosphere"
owners:
  - "@jbarrick-mesosphere"
editor: TBD
creation-date: 2020-02-14
last-updated: 2020-02-14
status: provisional
---

# KUTTL Rego Support

## Table of Contents

* [Summary](#summary)
* [Motivation](#motivation)
    * [Goals](#goals)
    * [Non-Goals](#non-goals)
* [Proposal](#proposal)
    * [Rego inputs](#rego-inputs)
        * [Custom Resource Definitions](#custom-resource-definitions)
    * [Running the asserts](#running-the-asserts)
* [Alternatives](#alternatives)
* [Other Rego Users](#other-rego-users)

## Summary

KUTTL currently only supports YAML for writing asserts, however, this can sometimes be limited. This KEP outlines how we can use the Rego query language for writing much more advanced asserts.

## Motivation

There are cases where pure YAML asserts are not expressive enough. A simple example might be where the user needs to verify a range of values (for example, RAM requirement is greater than `512MiB`). A more complex example might involve context of multiple fields or resources (for example, the number of pods is equal to the number of replicas on a Deployment).

In order to support more advanced use-cases like these, we can add support for Rego-based asserts.

Rego is a query language built into the Open Policy Agent that allows writing asserts over data. Open Policy Agent and Rego are used widely in the Kubernetes community, making it suitable for use in tooling targeted for Kubernetes developers.

### Goals

* Support a more expressive assertion language to make writing complex assertions possible.

### Non-Goals

* Increasing DRYness of YAML definitions (a good goal, but out of scope).

## Proposal

KUTTL currently only supports YAML or JSON files for specifying manifests. These files all have the `.yaml` or `.json` file extensions. KUTTL will search for assert files that end in `.rego` and import those as rego code.

### Rego inputs

The biggest problem to solve is how the assert's inputs should be provided. YAML asserts specify the resource kind and name directly in their manifest, but Rego queries do not have the same sort of metadata. It seems that there isn't a ton of consistency across tools using Rego, so we are relatively free to use a scheme that works for us. Unfortunately, we'll lose out on interoperability of Rego files, but this doesn't seem to be possible today anyway.

* [Preflight](https://github.com/jetstack/preflight) provides an input object: `import input["k8s/pods"] as pods`.
* [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/) provides a large map of resources by type and namespace: `import data.kubernetes.ingresses`.
* [conftest](https://github.com/instrumenta/conftest) reads local files into an input map: `input["deployment.yaml"]["spec"]["selector"]["matchLabels"]["app"]`.

As Open Policy Agent is the primary reference implementation, we'll go with OPA's structure. We will also set `input.request.namespace` to the namespace the tests are supposed to target.

```
package assert

import data.kubernetes.pods
import data.kubernetes.pipelineruns

three_pods[msg] {
    # fail if number of pods does not equal 3
    count(pods[input.request.namespace]) == 3
    msg := sprintf("three pods expected, got %d", [count(pods[input.request.namespace])])
}

all_tasks_successful[msg] {
    pr := pipelineruns[input.request.namespace][_]
    
    # pipeline run was successful
    pr.status.conditions[0].status == "True"
    
    # all taskruns are also successful
    taskRunStatus := pr.status.taskRuns[_]
    taskRunStatus.status.conditions[0].status == "True"

    msg := "task run failed"
}
```

#### Custom Resource Definitions

OPA typically supports importing CRDs by allowing importing them by name, e.g., `data.kubernetes.pipelineruns`. However, this doesn't allow specifying an API group or version. We can also support specifying the full group `data.kubernetes["pipelineruns.tekton.dev"]` to be less ambiguous.

##### TODO: how do I specify API version?

### Running the asserts

The asserts will be run the same as the YAML asserts: they will be tried periodically for the duration of the timeout. If the assert has never been successful at the end of the timeout, then the assert will be considered failed.

## Alternatives

* [Starlark](https://github.com/bazelbuild/starlark) is a Python-based language built by Google for use in describing builds (namely for use in Bazel). Starlark could be a good option both for writing asserts and for generating YAML. We opted to go with Rego as this is the path most testing tools have taken. For example, [conftest](https://github.com/instrumenta/conftest) is a Rego-based configuration tester that superceded the Starlark-based [kubetest](https://github.com/garethr/kubetest).
* [Cue](https://cuelang.org/) is another language for validating and placing constraints on data. It can also be used for generating JSON objects. It tends to be fairly complex and isn't as widely used as Rego.
* We will likely also support KUTTL-as-a-Go-library in a separate KEP.

## Other Rego Users

* [Preflight](https://github.com/jetstack/preflight)
* [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/)
* [conftest](https://github.com/instrumenta/conftest)
* [Gatekeeper](https://kubernetes.io/blog/2019/08/06/opa-gatekeeper-policy-and-governance-for-kubernetes/)
