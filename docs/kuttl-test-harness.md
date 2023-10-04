# KUTTL Test Harness

KUTTL is a declarative integration testing harness for testing operators, KUDO, [Helm charts](testing/tips.md#helm-testing), and any other Kubernetes applications or controllers. Test cases are written as plain Kubernetes resources and can be run against a mocked control plane, locally in kind, or any other Kubernetes cluster.

Whether you are developing an application, controller, operator, or deploying Kubernetes clusters the KUTTL test harness helps you easily write portable end-to-end, integration, and conformance tests for Kubernetes without needing to write any code.

<h2>Table of Contents</h2>

[[toc]]

## Installation

The test harness CLI is included in the KUTTL CLI, to install we can install the CLI using [krew](https://github.com/kubernetes-sigs/krew):

```bash
krew install kuttl
```

You can now invoke the KUDO test CLI:

```bash
kubectl kuttl test --help
```

See the [KUTTL installation guide](cli.md#installation) for alternative installation methods.

## Writing Your First Test

Now that the KUTTL CLI is installed, we can write a test. The KUTTL test CLI organizes tests into suites:

* A "test step" defines a set of Kubernetes manifests to apply and a state to assert on (wait for or expect).
* A "test case" is a collection of test steps that are run serially - if any test step fails then the entire test case is considered failed.
* A "test suite" is comprised of many test cases that are run in parallel.
* The "test harness" is the tool that runs test suites (the KUTTL CLI).

Be aware that KUTTL CLI expects a kuttl-test.yaml needs to be available, see [setup the kuttl kubectl plugin](cli.md#setup-the-kuttl-kubectl-plugin) if you didn't do so yet.

### Create a Test Case

First, let's create a directory for our test suite, let's call it `tests/e2e`:

```sh
mkdir -p tests/e2e
```

Next, we'll create a directory for our test case, the test case will be called `example-test`:

```bash
mkdir tests/e2e/example-test
```

Inside of `tests/e2e/example-test/` create our first test step, `00-install.yaml`, which will create a deployment called `example-deployment`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
```

Note that in this example, the deployment does not have a `namespace` set. The test harness will create a namespace for each test case and run all of the test steps inside of it. However, if a resource already has a namespace set (or is not a namespaced resource), then the harness will respect the namespace that is set.

Each filename in the test case directory should start with an index (in this example `00`) that indicates which test step the file is a part of. Files that do not start with a step index are ignored and can be used for documentation or other test data. Test steps are run in order and each must be successful for the test case to be considered successful.

Now that we have a test step, we need to create a test assert. The assert's filename should be the test step index followed by `-assert.yaml`. Create `tests/e2e/example-test/00-assert.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
status:
  readyReplicas: 3
```

This test step will be considered completed once the pod matches the state that we have defined. If the state is not reached by the time the assert's timeout has expired (30 seconds, by default), then the test step and case will be considered failed.

### Run the Tests

Let's run this test suite:

```sh
kubectl kuttl test --start-kind=true ./tests/e2e/
```

Running this command will:

* Start a [kind (Kubernetes-in-Docker) cluster](https://github.com/kubernetes-sigs/kind), if there is not already one running.
* Create a new namespace for the test case.
* Create the resources defined in `tests/e2e/example-test/00-install.yaml`.
* Wait for the state defined in `tests/e2e/example-test/00-assert.yaml` to be reached.
* Collect the kind cluster's logs.
* Tear down the kind cluster (or you can run `kubectl kuttl test` with `--skip-cluster-delete` to keep the cluster around after the tests run).

### Write a Second Test Step

Now that we have successfully written a test case, let's add another step to it. In this step, let's increase the number of replicas on the deployment we created in the first step from 3 to 4.

Create `tests/e2e/example-test/01-scale.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
spec:
  replicas: 4
```

Now create an assert for it in `tests/e2e/example-test/01-assert.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
status:
  readyReplicas: 4
```

Run the test suite again and the test will pass:

```sh
kubectl kuttl test --start-kind=true ./tests/e2e/
```

### Test Suite Configuration

To add this test suite to your project, create a `kuttl-test.yaml` file:

```yaml
apiVersion: kuttl.dev/v1beta1
kind: TestSuite
testDirs:
- ./tests/e2e/
startKIND: true
```

Now we can run the tests just by running `kubectl kuttl test` with no arguments.

Any arguments provided on the command line will override the settings in the `kuttl-test.yaml` file, e.g. to skip using kind and run the tests against a live Kubernetes cluster, run:

```sh
kubectl kuttl test --start-kind=false
```

Now that your first test suite is configured, see [test environments](testing/test-environments.md) for documentation on customizing your test environment or the [test step documentation](testing/steps.md) to write more advanced tests.
