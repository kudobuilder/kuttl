# KUTTL Configuration Reference

## TestSuite

The `TestSuite` object specifies the settings for the entire test suite and should live in the test suite configuration file (`kuttl-test.yaml` by default, or `--config`):

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestSuite
startKIND: true
kindContainers:
- your/image:latest
testDirs:
- tests/e2e/
timeout: 120
```

Supported settings:

Field             |      Type        | Description                                                                              | Default
------------------|------------------|------------------------------------------------------------------------------------------|--------
crdDir            | string           | Path to CRDs to install before running tests. KUTTL waits for CRDs to be available prior to starting tests.                                            |
manifestDirs      | list of strings  | Paths to manifests to install before running tests.                                      |
testDirs          | list of strings  | Directories containing test cases to run.                                                |
startControlPlane | bool             | Whether or not to start a local etcd and kubernetes API server for the tests.            | false
startKIND         | bool             | Whether or not to start a local kind cluster for the tests.                              | false
kindNodeCache     | bool             | If set, each node defined in the kind configuration will have a docker volume mounted into it to persist pulled container images across test runs | false
kindConfig        | string           | Path to the KIND configuration file to use.                                              |
kindContext       | string           | KIND context to use.                                                                     | "kind"
skipDelete        | bool             | If set, do not delete the resources after running the tests (implies SkipClusterDelete). | false
skipClusterDelete | bool             | If set, do not delete the mocked control plane or kind cluster.                          | false
timeout           | int              | Override the default timeout of 30 seconds (in seconds).                                 | 30
parallel          | int              | The maximum number of tests to run at once.                                              | 8
artifactsDir      | string           | The directory to output artifacts to (current working directory if not specified).       | .
commands          | list of [Commands](#commands) | Commands to run prior to running the tests.                                   | []
kindContainers    | list of strings  | List of Docker images to load into the KIND cluster once it is started.                  | []
reportFormat      | string           | Determines the report format. If empty, no report is generated. One of: JSON, XML.       |
reportName        | string           | The name of report to create. This field is not used unless reportFormat is set.         | "kuttl-test"
namespace         | string           | The namespace to use for tests. This namespace will be created if it does not exist and removed if it was created (unless `skipDelete` is set). If no namespace is set, one will be auto-generated. |
suppress          | list of strings  | Suppresses log collection of the specified types. Currently only `events` is supported.  |

## TestStep

The `TestStep` object can be used to specify settings for a test step and can be specified in any test step YAML.

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestStep
apply:
- my-new-resource.yaml
assert:
- my-asserted-new-resource.yaml
error:
- my-errored-new-resource.yaml
unitTest: false
delete:
- apiVersion: v1
  kind: Pod
  name: my-pod
commands:
- command: helm init
kubeconfig: foo.kubeconfig
```

Supported settings:

Field    |          Type             | Description
---------|---------------------------|---------------------------------------------------------------------
apply    | list of files             | A list of files to apply as part of this step. Specified path is relative to that in which the step occurs.
assert   | list of files             | A list of files to assert as part of this step. See documentation for [asserts and errors](asserts-errors.md) for more information. Specified path is relative to that in which the step occurs.
error    | list of files             | A list of files to error as part of this step. See documentation for [asserts and errors](asserts-errors.md) for more information. Specified path is relative to that in which the step occurs.
delete   | list of object references | A list of objects to delete, if they do not already exist, at the beginning of the test step. The test harness will wait for the objects to be successfully deleted before applying the objects in the step.
index    | int                       | Override the test step's index.
commands | list of [Commands](#commands) | Commands to run prior at the beginning of the test step.
kubeconfig    | string                       | The Kubeconfig file to use to run the included steps(s).
unitTest    | bool                       | Indicates if the step is a unit test, safe to run without a real Kubernetes cluster.


Object Reference:

Field      |   Type | Description
-----------|--------|---------------------------------------------------------------------
apiVersion | string | The Kubernetes API version of the objects to delete.
kind       | string | The Kubernetes kind of the objects to delete.
name       | string | If specified, the name of the object to delete. If not specified, all objects that match the specified labels will be deleted.
namespace  | string | The namespace of the objects to delete.
labels     | map    | If specified, a label selector to use when looking up objects to delete. If both labels and name are unspecified, then all resources of the specified kind in the namespace will be deleted.

## TestAssert

The `TestAssert` object can be used to specify settings for a test step's assert and must be specified in the test step's assert YAML.

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestAssert
timeout: 30
commands:
- command: echo hello
collectors:
- type: pod
  pod: nginx
```

Supported settings:

Field   | Type | Description                                           | Default
--------|------|-------------------------------------------------------|-------------
timeout | int  | Number of seconds that the test is allowed to run for | 30
collectors | list of [collectors](#collectors) | The collectors to be invoked to gather information upon step failure | N/A
commands | list of [commands](#commands) | Commands to run prior to the beginning of the test step. | N/A

## TestFile

A `TestFile` object can be used to provide configuration concerning a single YAML test file that contains it.

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestFile
testRunSelector:
  matchLabels:
    flavor: vanilla
```

Supported settings:

| Field           | Type           | Description                                                                                                                     | Default                                                      |
|-----------------|----------------|---------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------|
| testRunSelector | label selector | If this selector does not match [labels of this test run](#test-run-labels-and-selectors), the containing file will be ignored. | Empty label selector (matches all possible test label sets). |


### Test Run Labels and Selectors

An invocation of `kuttl test` may specify a label set associated with a test run using a command line flag.
One can then use a `TestFile` object with `testRunSelector` to decide whether a given test YAML file should be included
in a test run or not.

## Collectors

The `Collectors` object is used by the `TestAssert` object as a way to collect certain information about the outcome of an `assert` or `errors` step should it fail. A collector is only invoked in cases where a failure occurs and not if the step succeeds. Collection can occur from Pod logs, Namespace events, or the output of a custom command.

Supported settings:

Field   | Type | Description                                           | Default
--------|------|-------------------------------------------------------|-------------
type | string  | Type of collector to run. Values are one of `pod`, `command`, or `events`. If the field named `command` is specified, `type` is assumed to be `command`. If the field named `pod` is specified, `type` is assumed to be `pod`. | `pod`
pod | string  | The pod name from which to access logs. | N/A
namespace | string  | Namespace in which the pod or events can be located. | N/A
container | string  | Container name inside the pod from which to fetch logs. If empty assumes all containers. | unset
selector | string  | Label query to select a pod. | N/A
tail | int  | The number of last lines to collect from a pod. | 10 (if selector); all (if pod name)
command | string  | Command to run. Requires an empty type or type `command`. Must not specify fields `pod`, `namespace`, `container`, or `selector` if present. | N/A

## Commands

The `Commands` object is used by `TestStep`, `TestAssert`, and `TestSuite` to enable running commands in tests:

Field         |   Type | Description
--------------|--------|---------------------------------------------------------------------
command       | string | The command and argument to run as a string.
script        | string | Allows a shell script to run - namespaced and command should not be used with script.  namespaced is ignored and command is an error.  env expansion is depended upon the shell but ENV is passed to the runtime env.
namespaced    | bool   | If set, the `--namespace` flag will be appended to the command with the namespace to use (the test namespace for a test step or "default" for the test suite).
ignoreFailure | bool   | If set, failures will be ignored.
background    | bool   | If this command is to be started in the background. These are only support in TestSuites.
skipLogOutput | bool   | If set, the output from the command is *not* logged. Useful for sensitive logs or to reduce noise.
timeout       | int    | Override the TestSuite timeout for this command (in seconds).

*Note*: The current working directory (CWD) for `command`/`script` is the test directory.
