# Asserts and Errors

Test asserts are the part of a [test step](steps.md) that define the state to wait for Kubernetes to reach. It is possible to match specific objects by name as well as match any object that matches a defined state. Test errors define states that should not be reached.

## Format

The test assert file for a test step is found at `$index-assert.yaml`. So, if the test step index is `00`, the assert should be called `00-assert.yaml`. This file can contain any number of objects to match on. If the objects have a namespace set, it will be respected, but if a namespace is not set, then the test harness will look for the objects in the test case's namespace.

The test error file for a test step is found at `$index-errors.yaml` and works similar to the test assert file.

By default, a test step will wait for up to 30 seconds for the defined state to be reached. See the [configuration reference](reference.md#testassert) for documentation on configuring test asserts.

Note that an assertion or errors file is optional. If absent, the test step will be considered successful immediately once the object(s) in the test step have been created. It is also valid to create a test step that does not create any objects, but only has an assertion or errors file.

If a file name ends with `.gotmpl.yaml`, then it will be treated as a template for expansion.
See [templating.md](templating.md) for more information.

## Getting a Resource from the Cluster

If an object has a name set, then the harness will look specifically for that object to exist and then verify that its state matches what is defined in the assert file. For example, if the assert file has:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
status:
  phase: Successful
```

Then the test harness will wait for the `my-pod` pod in the test namespace to have `status.phase=Successful`. Note that any fields *not* specified in the assert file will be ignored, making it possible to specify only the important fields for the test step.

If this object is in the errors file, the test harness will report an error if that object exists and its state matches what is defined in the errors file.

## Listing Resources in the Cluster

If an object in the assert file has no name set, then the harness will list objects of that kind.
If the object in the assert file has `metadata.labels` field, then it will be used as a label selector for the list operation.
Then `kuttl` will expect there to be at least one object that matches. For example, an assert:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: my-app
status:
  phase: Successful
```

This example would wait for a pod with an `app` label value of `my-app` to exist in the test namespace with the `status.phase=Successful`.

If no labels were specified, *any* pod with specified status would satisfy the assertion.

If this is defined in the errors file instead, the test harness will report an error if *any* such pod exists in the test namespace with `status.phase=Successful`.

## Failures

When a failure occurs in either an `assert` or `errors` step, kuttl will print a difference (diff) in the test output showing the reason why the step was deemed to fail. While this may be helpful in most cases, it may still be insufficient to determine the exact cause of a failure. Some additional information may be required to fully explain why a step failed which provides fuller context. When the diff is not adequate to explain a failure, a [`collectors`](reference.md#collectors) object may optionally be used to gather further troubleshooting information in the form of pod logs, namespace events, or output of a command.

For example, consider a simple test case in which a pod is created as the initial step followed by an assertion that the pod is present and in a state of `ready=true`. If the pod is observed to contain the state `ready=false` the step, and test, will fail. With a `collectors` object present in the `TestAssert`, it may provide logs for the pod to help explain why this state was not reached.

`01-pod.yaml`

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: hello-world
  name: hello-world
  namespace: default
spec:
  containers:
  - image: docker.io/hello-world
    name: hello-world
```

`01-assert.yaml`

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestAssert
timeout: 5
collectors:
- type: pod
  pod: hello-world
  namespace: default
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: hello-world
  name: hello-world
  namespace: default
status:
  running: true
```

In this example, the `hello-world` container was not started with any arguments resulting in its running followed by termination as expected. Therefore, the status of `running=true` was not asserted.

In the command output, prior to the full diff kuttl displays will be shown the pod's logs.

```log
    logger.go:42: 20:06:29 | collectors/1-pod | starting test step 1-pod
    logger.go:42: 20:06:30 | collectors/1-pod | Pod:default/hello-world created
    logger.go:42: 20:06:35 | collectors/1-pod | test step failed 1-pod
    logger.go:42: 20:06:35 | collectors/1-pod | collecting log output for [type==pod,pod==hello-world,namespace: default]
    logger.go:42: 20:06:35 | collectors/1-pod | running command: [kubectl logs --prefix hello-world -n default --all-containers --tail=-1]
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world] 
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world] Hello from Docker!
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world] This message shows that your installation appears to be working correctly.
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world] 
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world] To generate this message, Docker took the following steps:
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world]  1. The Docker client contacted the Docker daemon.
    logger.go:42: 20:06:35 | collectors/1-pod | [pod/hello-world/hello-world]  2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    <snip>
    case.go:362: failed in step 1-pod
    case.go:364: --- Pod:default/hello-world
```

See the [reference page](reference.md#collectors) for more configuration options available with the `collectors` object.
