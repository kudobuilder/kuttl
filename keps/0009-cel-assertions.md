---
kep-number: 9
short-desc: This KEP discusses adding CEL-based assertions to Kuttl
title: CEL Assertions
authors:
    - "@kumar-mallikarjuna"
owners:
    - "@kumar-mallikarjuna"
    - "@porridge"
creation-date: 2024-09-23
last-updated: 2024-09-23
status: provisional
---

# CEL Assertions

## Summary

Add Common Expression Language (CEL) support for Kuttl assertions. This would extend Kuttl's ability to perform
assertions based on complex expressions that are currently not possible with the existing syntax.

## Motivation

Currently, Kuttl lacks support for complex data manipulation in assertions.
Here are a few examples where this limitation is apparent:

-   Testing whether a Deployment has created _more than `n` Pods_.
    -   While Kuttl supports equality testing, conditional matching poses a challenge.
-   Slice assertion.
    -   Kuttl's current method for handling slice comparisons is somewhat unclear. This is especially important when
        the order of elements in the slice is not fixed, which is common in Kubernetes (see
        https://github.com/kudobuilder/kuttl/issues/76#issuecomment-660944596).
    -   Additionally, partial assertions on slices are not possible (see https://github.com/kudobuilder/kuttl/issues/76).

These issues could be resolved by incorporating a CEL engine into Kuttl, allowing for expression
evaluation and more flexible assertions.

### Goals

1. Add a CEL engine using [github.com/google/cel-go/cel](https://github.com/google/cel-go/cel).
2. Update `TestAssert` CRD to specify CEL identifiers and evaluate expressions.
3. Support any/all assertion of CEL expressions.

### Non-Goals

1. Reading non-Kubernetes resources for assertions.
2. Reading resources from multiple clusters for assertions.

## Proposal

### CRD Changes

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestAssert
celAssert:
  resources:
  - apiVersion: apps/v1
      kind: Deployment
      name: coredns
      namespace: kube-system
      id: resource1
  - apiVersion: apps/v1
      kind: Deployment
      name: metrics-server
      namespace: kube-system
      id: resource2

  # Success if any expression evaluates to true
  any:
  - expression: "resource.status.readyReplicas > 1"
  - expression: ...
  - ...

  # Success only if all expressions evaluate to true
  all:
  - expression: "resource.status.readyReplicas > 0"
  - expression: ...
  - ...
```

### Implementation

At a high level, we would use the [github.com/google/cel-go/cel](https://github.com/google/cel-go/cel) library to evaluate expressions as follows:

```go
func EvaluateExpression(resources...*unstructured.Unstructured)(ref.Val, error) {
    env, err: = cel.NewEnv(
        cel.Variable("<id1>", cel.MapType(cel.StringType, cel.DynType)),
        cel.Variable("<id2>", cel.MapType(cel.StringType, cel.DynType)),
        ...
    )
    if err != nil {
        return fmt.Errorf("failed to create environment: %w", err)
    }

    ast, issues: = env.Compile(`<expression>`)
    if issues != nil && issues.Err() != nil {
        return fmt.Errorf("type-check error: %s", issues.Err())
    }

    prg, err: = env.Program(ast)
    if err != nil {
        return fmt.Errorf("program construction error: %w", err)
    }

    out, _, err: = prg.Eval(map[string] interface {} {
        "<id1>": resources[0].Object,
        "<id2>": resources[1].Object,
        ...
    })

    if err != nil {
        return fmt.Errorf("failed to evaluate program: %w", err)
    }

    return out, nil
}
```
