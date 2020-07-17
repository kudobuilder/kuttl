# KUTTL

<img src="https://kuttl.dev/images/kuttl-horizontal-logo.png" width="256">

[![CircleCI](https://circleci.com/gh/kudobuilder/kuttl.svg?style=svg)](https://circleci.com/gh/kudobuilder/kuttl)

KUbernetes Test TooL (KUTTL) provides a declarative approach to test Kubernetes Operators.

KUTTL is designed for testing operators, however it can declaratively test any kubernetes objects.

> This is a customized versions of Kuttl. Layer5 has modified Kuttl to support more features that complement the use cases of Layer5.

## Getting Started

Please refer to the [getting started guide](https://kuttl.dev/docs/) documentation.

## Additional features

### InCluster kubeConfig

Kuttl was meant to be used as a CLI tool, thus it takes kubeConfig from the ENV and loads it. But we would want to run some test in the cluster so that we have access all kubernetes resources. 
Enabling the InCluster field in the harness object will force kuttl to pull out the KubeConfig as an InCluster config from inside the Pod.
### Custom handler injection in TestSteps

Developers can pass functions to Kuttl and pass the TestStep name so that Kuttl will run this function at the end of the TestStep. It is useful for cases where we might want custom tests other than the `asserts` and `errors` by Kuttl.

### Result Collection

Each TestCase returns a result object that lists the TestSteps that succeeded and failed.

### YAML Namespace injection

In the YAML files passed to Kuttl, if one uses the placeholder `<NAMESPACE>`, then while loading the tests Kuttl will replace it with the name of the kubernetes namespace in which the tests will be run.

### Namespace Annotations

The developer can provide a list of annotations to Kuttl which will be used by Kuttl to annotate the namespace while its creation.

> For an example of how to use it, look into the SMI conformance project by Layer5.