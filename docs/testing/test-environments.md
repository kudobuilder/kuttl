# Test Environments

The KUTTL test harness can run tests against several different test environments, allowing your test suites to be used in many different environments.

A default environment for the tests can be defined in `kuttl-test.yaml` allowing each test suite or project to easily use the correct environment.

## Live Cluster

If no configuration is provided, the tests will run against your default cluster context using whatever Kubernetes cluster is configured in your kubeconfig.

You can also provide an alternative kubeconfig file by either setting `$KUBECONFIG` or the `--kubeconfig` flag:

```bash
kubectl kuttl test --kubeconfig=mycluster.yaml
```

## Kubernetes-in-docker

KUTTL has a built in integration with [kind](https://github.com/kubernetes-sigs/kind) to start and interact with kubernetes-in-docker clusters.

To start a kind cluster in your tests either specify it on the command line:

```bash
kubectl kuttl test --start-kind=true
```

Or specify it in your `kuttl-test.yaml`:

```yaml
apiVersion: kudo.k8s.io/v1alpha1
kind: TestSuite
kindNodeCache: true
```

By default KUTTL will use the default kind cluster name of "kind". If a kind cluster is already running with that name, it will use the existing cluster.

The kind cluster name can be overridden by setting either `kindContext` in your configuration or `--kind-context` on the command line.

By setting `kindNodeCache`, the containerd directories will be mounted into a Docker volume in order to persist the images pulled during a test run across test runs.

If you want to load images into the built KIND cluster that have not been pushed, set `kindContainers`. See [Tips And Tricks](tips.md#loading-built-images-into-kind) for an example.

It is also possible to provide a custom kind configuration file. For example, to override the Kubernetes cluster version, create a kind configuration file called `kind.yaml`:

```yaml
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
nodes:
- role: control-plane
  image: kindest/node:v1.14.3
```

See the [kind documentation](https://kind.sigs.k8s.io/docs/user/quick-start/#configuring-your-kind-cluster) for all options supported by kind.

Now specify either `--kind-config` or `kindConfig` in your configuration file:

```bash
kubectl kuttl test --kind-config=kind.yaml
```

*Note*: Once the tests have been completed, the test harness will collect the kind cluster's logs and then delete it, unless `--skip-cluster-delete` has been set.

## Mocked Control Plane

The above environments are great for end to end testing, however, for integration test use-cases it may be unnecessary to create actual pods or other resources. This can make the tests a lot more flaky or slow than they need to be.

To write integration tests using the KUTTL test harness, it is possible to start a mocked control plane that starts only the Kubernetes API server and etcd. In this environment, objects can be created and operated on by custom controllers, however, there is no scheduler, nodes, or built-in controllers. This means that pods will never run and built-in types, such as, deployments cannot create pods.

Kubernetes controllers can be added to this environment by using the TestSuite configuration command in  order to start the controller:

```
commands:
  - command: ./bin/manager
    background: true
```

To start the mocked control plane, specify either `--start-control-plane` on the CLI or `startControlPlane` in the configuration file:

```bash
kubectl kuttl test --start-control-plane
```

## Environment Setup

Before running a test suite, it may be necessary to setup the Kubernetes cluster - typically, either installing required services or custom resource definitions.

Your `kuttl-test.yaml` can specify the settings needed to setup the cluster:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestSuite
startControlPlane: true
testDirs:
- tests/e2e/
manifestDirs:
- tests/manifests/
crdDir: tests/crds/
commands:
  - command: kubectl apply -f https://raw.githubusercontent.com/kudobuilder/kudo/master/docs/deployment/10-crds.yaml
```

The above configuration would start kind, install all of the CRDs in `tests/crds/`, and run all of the commands defined in `kubectl` before running the tests in `testDirs`.

See the [configuration reference](reference.md#testsuite) for documentation on configuring test suites.

### Starting a Kubernetes Controller

In some test suites, it may be useful to have a controller running. To start a controller, add a configuration as a command in the TestSuite configuration file `kuttl-test.yaml`:

For a KUDO, an example of deploying an previously released controller would look like:

```
commands:
  - command: kubectl kudo init --wait
```

The KUDO CLI has a readiness watch on the installation of the KUDO manager.  When it exits, the KUDO manager is ready.

Another commonly explain is the starting of a manager that is still in development.  The assumption of the code snippet below is that a `make manager` or Makefile target generated a manager in the `bin` folder.

```
commands:
  - command: ./bin/manager
    background: true
```

## KUTTL Mode of Testing in a Cluster

KUTTL `test` is designed to function in 2 distinct modes managed by the use of the `--namespace` flag.

1. By default, KUTTL will create a namespace, run a series of steps defined by a test, then delete the namespace.  It will create a namespace for each test running in namespace isolation.  Since, KUTTL owns the namespace, it deletes it as part of cleanup.

1. When `--namespace` specifies a namespace, it is expected that the namespace exists.  KUTTL in this mode, does **NOT** create or delete the namespace.  All tests are run and share this namespace by default.  If the namespace does NOT exist, the test fails.

### Single Namespace Testing

When running with the `--namespace`, there are potential consequences which are very important to understand.  Normally when KUTTL is in the "apply" phase, if an object doesn't exist, it is created.  If it does exist, it is merge patch updated.  When creating a series of tests which do NOT share a namespace, potentially the same object is referenced in multiple tests.  Those objects are separated by namespace and are auto-cleaned up by the deleting of the namespace. When running in the same namespace, this cleanup does NOT happen.  It is the responsibility of the test designers to delete the objects pre- or post-test.  This results in TestSuites designed to run in single namespace can be run in the default multi-namespace mode, but it is possible the reverse isn't true.  More care needs to be taken in single namespace testing for pre/post test management.

It is worth noting that extra care noted above is necessary for the "happy path". IF a test fails and does not clean up properly, some future test (during this testsuite) may be affected. For these reasons, running parallel tests for single namespace testing could also run into challenges and is not recommended.

## Permissions / RBAC Rules

KUTTL was initially designed to "own" a cluster for testing.  In its default mode, it needs to be able to create and delete namespaces, as well as create/update/view kubernetes objects in that namespace.  The RBAC needs in this mode include:

1. POST, GET, LIST, PUT, PATCH, DELETE on namespace and the objects in that namespace.
1. GET, LIST events

It is possible to turn off events with the `--suppress-log=events`.  This removes the need to GET or LIST events.

When running in single namespace testing mode, no permissions are needed for namespaces, reducing permissions to events.  In this mode, it is possible to remove KUTTLs access needs by using the `--suppress-log=events`.  In this mode, you will need access in the explicitly provided namespace to create, update and delete kubernetes objects defined in the test.

**NOTE:** This defined permissions are for KUTTL itself and do NOT take in account the test that kuttl is running.  It is possible for the test to create a namespace which is considered outside the KUTTL permission needs.
